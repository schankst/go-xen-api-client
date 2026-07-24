package xmlrpc

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

// Client is a minimal XML-RPC client over HTTP. Each Call is one
// synchronous HTTP round trip - there's no request multiplexing to
// justify anything more elaborate, since XML-RPC (unlike the transport
// net/rpc.Client was designed around) has no notion of a persistent
// bidirectional connection.
//
// Client is safe for concurrent use by multiple goroutines.
type Client struct {
	url        string
	httpClient *http.Client

	mu      sync.Mutex
	cookies []*http.Cookie
}

// NewClient creates a Client for the XML-RPC endpoint at url. If
// transport is nil, http.DefaultTransport's zero-value equivalent
// (&http.Transport{}) is used.
func NewClient(url string, transport *http.Transport) (*Client, error) {
	if transport == nil {
		transport = &http.Transport{}
	}
	return &Client{
		url:        url,
		httpClient: &http.Client{Transport: transport},
	}, nil
}

// Call issues an XML-RPC call for method with params and, on success,
// stores the decoded result into *reply (reply must be a non-nil
// pointer; pass nil to discard the result). On a genuine XML-RPC <fault>
// response the returned error is a *Error.
//
// The first response's cookies are captured and replayed on every
// subsequent call on this Client, since XCP-ng/XenServer pools use this
// to pin a client to a specific host after a redirect.
func (c *Client) Call(method string, params []interface{}, reply interface{}) error {
	body, err := buildRequest(method, params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")

	c.mu.Lock()
	cookies := c.cookies
	c.mu.Unlock()
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.mu.Lock()
	if c.cookies == nil {
		c.cookies = resp.Cookies()
	}
	c.mu.Unlock()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	value, err := decodeResponse(respBody)
	if err != nil {
		return err
	}

	if reply == nil {
		return nil
	}
	rv := reflect.ValueOf(reply)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("xmlrpc: reply must be a non-nil pointer, got %T", reply)
	}
	rv.Elem().Set(reflect.ValueOf(value))
	return nil
}
