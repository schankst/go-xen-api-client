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
//
// There is no error return: http.NewRequest is what would normally fail at
// construction time (a malformed URL), and that happens inside Call, where
// it is reported per call. Validating the URL here as well would mean
// duplicating net/url's parsing rules against http.NewRequest's.
func NewClient(url string, transport *http.Transport) *Client {
	if transport == nil {
		transport = &http.Transport{}
	}
	return &Client{
		url:        url,
		httpClient: &http.Client{Transport: transport},
	}
}

// Call issues an XML-RPC call for method with params and, on success,
// stores the decoded result into *reply (reply must be a non-nil
// pointer; pass nil to discard the result). On a genuine XML-RPC <fault>
// response the returned error is a *Error.
//
// The decoded value is stored as-is when its Go type is assignable to
// reply's element type; it is never converted. An XML-RPC <int> always
// decodes to int64, <dateTime.iso8601> to time.Time, <array> to
// []interface{}, <struct> to Struct, and both <string> and an empty
// <value/> to string - so declare reply as a pointer to one of those
// (or to interface{}), not to int, time.Time aliases, or other types
// that merely look equivalent. A mismatch is returned as an error rather
// than panicking.
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

	// reflect.Value.Set panics when the source type isn't assignable to the
	// destination type, so that has to be checked here: a server answering
	// with a different wire type than the caller's reply expects (e.g. an
	// empty <value/> where an int was expected) must surface as an error,
	// never as a panic in the caller's goroutine. No implicit conversion is
	// attempted - int64 from <int> does not become int, and a time.Time
	// does not become a Unix timestamp - because a silent conversion would
	// hide exactly the kind of wire-format surprise this check exists for.
	dst := rv.Elem()
	if dst.Kind() == reflect.Interface && dst.NumMethod() == 0 {
		// interface{} (or any empty interface) accepts every decoded value.
		dst.Set(reflect.ValueOf(value))
		return nil
	}
	decoded := reflect.ValueOf(value)
	if !decoded.Type().AssignableTo(dst.Type()) {
		return fmt.Errorf("xmlrpc: cannot store %s result of %q into %s: types are not assignable (see Client.Call's documentation for the Go type each XML-RPC type decodes to)",
			decoded.Type(), method, dst.Type())
	}
	dst.Set(decoded)
	return nil
}
