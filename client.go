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

// SchemaXAPIRelease is the newest XAPI release name/version present in the
// xenapi.json schema the *_gen.go files in this fork were generated from
// (see https://xapi-project.github.io/xen-api/releases.html for what a
// release name/version maps to). Bindings are verified/typed up through
// this release; anything newer falls back to the enum tolerance built
// into convert_gen.go's converters (see convertEnumTypeToGoFuncTemplate
// in internal/cmd/genapi).
const SchemaXAPIRelease = "26.21.1"

// APIResult is the raw, untyped result of an APICall - Value holds whatever
// the XML-RPC response decoded to (a string, xmlrpc.Struct, []any,
// etc., depending on the call). Generated methods decode this into a typed
// result themselves; most callers should use those instead of APICall.
type APIResult struct {
	Value any
}

// APICall issues a raw XenAPI XML-RPC call (method is the wire-level name,
// e.g. "VM.get_all_records", not the Go method name) and returns its
// untyped result. Every generated method (Client.VM.GetAllRecords, etc.)
// is a thin, typed wrapper around this; call it directly only for
// functionality this fork doesn't expose a typed wrapper for. Returns an
// *Error if XenAPI reports a protocol-level failure.
//
// ctx bounds the call; once it is canceled or past its deadline, the call
// returns an error matching ctx.Err() with errors.Is. Note that canceling
// only abandons the request: an operation XAPI has already started (e.g.
// VM.start) keeps running on the host.
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
		// The envelope's ErrorDescription is [code, objtype?, uuid?] with the
		// last two explicitly nullable, but everything about it is host-
		// supplied: a proxy, a different XAPI variant, or a truncated response
		// can leave the field missing, the wrong kind, or empty. A caller
		// handling an API failure must not be crashed by the shape of the
		// failure, so every step here is checked rather than asserted.
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

		_objtype, _ := detailString(details, 1)
		_uuid, _ := detailString(details, 2)
		err = &Error{
			code:    code,
			objtype: _objtype, // might be nil
			uuid:    _uuid,    // optional
		}
		return
	}

	result.Value = rpcResult["Value"]
	return
}

// detailString returns details[index] when it is present and holds a string.
// The type and uuid slots of ErrorDescription are nullable by design
// (XenAPI sends null for codes that don't carry them), so a missing, nil,
// or non-string value means "absent" here rather than a malformed response.
func detailString(details []any, index int) (string, bool) {
	if index >= len(details) {
		return "", false
	}
	s, ok := details[index].(string)
	return s, ok
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

// WithTLSConfig sets the TLS configuration of the default HTTP client.
// Use it to trust a host's self-signed certificate:
//
//	pool := x509.NewCertPool()
//	pool.AppendCertsFromPEM(hostCertPEM)
//	client, err := xenapi.NewClient(url, xenapi.WithTLSConfig(&tls.Config{RootCAs: pool}))
//
// Certificate verification is only disabled if cfg sets InsecureSkipVerify.
func WithTLSConfig(cfg *tls.Config) Option {
	return func(o *clientOptions) { o.tlsConfig = cfg }
}

// NewClient creates a Client for the XenAPI server at rawURL, which must be
// an absolute http or https URL such as "https://10.0.0.10". Call
// Client.Session.LoginWithPassword before using it for anything else.
//
// By default, requests go through xmlrpc.DefaultHTTPClient and the host's
// TLS certificate is verified against the system roots; see WithTLSConfig
// for hosts with a self-signed certificate.
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
