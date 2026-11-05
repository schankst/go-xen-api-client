//go:generate go run ./internal/cmd/genapi

package xenapi

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

// SchemaXAPIRelease is the newest XAPI release in the schema the bindings
// were generated from. Enum values from newer releases pass through
// unchanged.
const SchemaXAPIRelease = "26.21.1"

// APIResult is the untyped result of APICall: a string, int64, float64,
// bool, time.Time, xmlrpc.Struct or []any, as decoded from XML-RPC.
type APIResult struct {
	Value any
}

// APICall calls the XenAPI method with its wire name, e.g.
// "VM.get_all_records", and returns the untyped result. The generated
// methods wrap it; use it for calls they don't cover. A failure reported
// by XenAPI is returned as *Error.
//
// Canceling ctx abandons the request; an operation XAPI has already
// started keeps running on the host.
func (client *Client) APICall(ctx context.Context, method string, params ...any) (result APIResult, err error) {
	rpcResult := xmlrpc.Struct{}

	err = client.rpc.Call(ctx, method, params, &rpcResult)
	if err != nil {
		return
	}

	status, ok := rpcResult["Status"].(string)
	if !ok {
		err = fmt.Errorf("Expected a field named %q with a string value in the response", "Status")
		return
	}

	if status != "Success" {
		// ErrorDescription is [code, param...]; it is host-supplied, so
		// its shape is checked rather than asserted.
		details, ok := rpcResult["ErrorDescription"].([]any)
		if !ok {
			err = fmt.Errorf("xenapi: %s failed with status %q but the response has no ErrorDescription array", method, status)
			return
		}
		if len(details) == 0 {
			err = fmt.Errorf("xenapi: %s failed with status %q but the ErrorDescription array is empty", method, status)
			return
		}
		code, ok := details[0].(string)
		if !ok {
			err = fmt.Errorf("xenapi: %s failed with status %q but its error code is %T, not a string", method, status, details[0])
			return
		}

		apiErr := &Error{Code: ErrorCode(code)}
		for _, d := range details[1:] {
			apiErr.Params = append(apiErr.Params, errorParam(d))
		}
		err = apiErr
		return
	}

	result.Value = rpcResult["Value"]
	return
}

// errorParam renders one ErrorDescription parameter. XenAPI sends strings;
// anything else is kept in its printed form rather than dropped.
func errorParam(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

// Option configures a Client created by NewClient.
type Option func(*clientOptions)

type clientOptions struct {
	httpClient *http.Client
	tlsConfig  *tls.Config
}

// WithHTTPClient makes the Client send its requests through c, e.g. to use
// a custom RoundTripper for logging or tests. TLS settings belong to c's
// transport then, so it cannot be combined with WithTLSConfig.
func WithHTTPClient(c *http.Client) Option {
	return func(o *clientOptions) { o.httpClient = c }
}

// WithTLSConfig sets the TLS configuration of the default HTTP client, e.g.
// RootCAs to trust a host's self-signed certificate. Verification is only
// disabled if cfg sets InsecureSkipVerify.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(o *clientOptions) { o.tlsConfig = cfg }
}

// NewClient creates a Client for the XenAPI server at rawURL, an absolute
// http or https URL such as "https://10.0.0.10". By default it uses
// xmlrpc.DefaultHTTPClient and verifies the host's TLS certificate against
// the system roots.
func NewClient(rawURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("xenapi: invalid URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("xenapi: URL %q must be absolute with an http or https scheme", rawURL)
	}

	var o clientOptions
	for _, opt := range opts {
		opt(&o)
	}
	if o.httpClient != nil && o.tlsConfig != nil {
		return nil, errors.New("xenapi: WithHTTPClient and WithTLSConfig cannot be combined; configure TLS on the HTTP client's transport")
	}
	httpClient := o.httpClient
	if httpClient == nil {
		httpClient = xmlrpc.DefaultHTTPClient(o.tlsConfig)
	}

	return prepClient(xmlrpc.NewClient(rawURL, httpClient)), nil
}
