package xmlrpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"reflect"
	"strings"
)

// Client is an XML-RPC client over HTTP; each Call is one HTTP request.
// It is safe for concurrent use.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient creates a Client for the XML-RPC endpoint at url that sends
// its requests through httpClient. If httpClient is nil, it uses
// DefaultHTTPClient(nil).
//
// The URL is not validated here; a malformed URL is reported by Call.
func NewClient(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = DefaultHTTPClient(nil)
	}
	return &Client{url: url, httpClient: httpClient}
}

// DefaultHTTPClient returns an http.Client with a cookie jar and a clone of
// http.DefaultTransport, using a clone of tlsConfig if it is non-nil. It
// has no overall timeout, since calls such as VM migration can run for a
// long time; bound each call with its context instead.
func DefaultHTTPClient(tlsConfig *tls.Config) *http.Client {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		transport = transport.Clone()
	} else {
		transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig.Clone()
	}
	jar, _ := cookiejar.New(nil) // never fails
	return &http.Client{Transport: transport, Jar: jar}
}

// Call calls method with params and stores the result in *reply, which
// must be a non-nil pointer; pass a nil reply to discard the result. A
// <fault> response is returned as *Error, a non-2xx HTTP status as
// *HTTPError.
//
// The result is stored without conversion: <int> decodes to int64,
// <double> to float64, <boolean> to bool, <dateTime.iso8601> to time.Time,
// <array> to []any, <struct> to Struct, <base64> to Base64, and <string>
// or an untyped value to string. If *reply can't hold that type, Call
// returns an error.
func (c *Client) Call(ctx context.Context, method string, params []any, reply any) error {
	var dst reflect.Value
	if reply != nil {
		rv := reflect.ValueOf(reply)
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return fmt.Errorf("xmlrpc: reply must be a non-nil pointer, got %T", reply)
		}
		dst = rv.Elem()
	}

	body, err := buildRequest(method, params)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		excerpt, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: string(excerpt)}
	}

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
	// Check assignability, since reflect.Value.Set would panic on a server
	// answering with an unexpected type.
	if dst.Kind() == reflect.Interface && dst.NumMethod() == 0 {
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
