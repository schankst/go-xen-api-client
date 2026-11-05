//go:generate go run xenapi.go

package xenapi

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

// SchemaXAPIRelease is the newest XAPI release name/version present in the
// xenapi.json schema the *_gen.go files in this fork were generated from
// (see https://xapi-project.github.io/xen-api/releases.html for what a
// release name/version maps to). Bindings are verified/typed up through
// this release; anything newer falls back to the enum tolerance built
// into convert_gen.go's converters (see convertEnumTypeToGoFuncTemplate
// in xenapi.go).
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
func (client *Client) APICall(method string, params ...any) (result APIResult, err error) {
	rpcResult := xmlrpc.Struct{}

	err = client.rpc.Call(method, params, &rpcResult)
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

// NewClient creates a Client for the XenAPI server at url (e.g.
// "https://10.0.0.10/"). Call Client.Session.LoginWithPassword before
// using it for anything else.
//
// If transport is nil, a default *http.Transport is used with
// InsecureSkipVerify: true - i.e. TLS certificate verification is off by
// default, matching XCP-ng/XenServer's common self-signed-certificate
// setup. Pass a transport of your own to verify certificates.
//
// There is no error return: nothing is connected to or parsed until the
// first call, which reports a malformed URL or an unreachable host from
// Session.LoginWithPassword.
func NewClient(url string, transport *http.Transport) *Client {
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return prepClient(xmlrpc.NewClient(url, transport))
}
