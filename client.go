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
// this release; anything newer falls back to the enum-tolerance patch in
// convert_gen.go.
const SchemaXAPIRelease = "26.16.1-next"

// APIResult is the raw, untyped result of an APICall - Value holds whatever
// the XML-RPC response decoded to (a string, xmlrpc.Struct, []interface{},
// etc., depending on the call). Generated methods decode this into a typed
// result themselves; most callers should use those instead of APICall.
type APIResult struct {
	Value interface{}
}

// APICall issues a raw XenAPI XML-RPC call (method is the wire-level name,
// e.g. "VM.get_all_records", not the Go method name) and returns its
// untyped result. Every generated method (Client.VM.GetAllRecords, etc.)
// is a thin, typed wrapper around this; call it directly only for
// functionality this fork doesn't expose a typed wrapper for. Returns an
// *Error if XenAPI reports a protocol-level failure.
func (client *Client) APICall(method string, params ...interface{}) (result APIResult, err error) {
	rpcParams := xmlrpc.Params{
		Params: params,
	}

	rpcResult := xmlrpc.Struct{}

	err = client.rpc.Call(method, rpcParams, &rpcResult)
	if err != nil {
		return
	}

	status, ok := rpcResult["Status"].(string)
	if !ok {
		err = fmt.Errorf("Expected a field named %q with a string value in the response", "Status")
		return
	}

	if status != "Success" {
		details := rpcResult["ErrorDescription"].([]interface{})
		_objtype := ""
		if len(details) > 1 && details[1] != nil {
			_objtype = details[1].(string)
		}
		var _uuid string
		if len(details) > 2 && details[2] != nil {
			_uuid = details[2].(string)
		}
		err = &Error{
			code:    details[0].(string),
			objtype: _objtype, // might be nil
			uuid:    _uuid,    // optional
		}
		return
	}

	result.Value = rpcResult["Value"]
	return
}

// NewClient creates a Client for the XenAPI server at url (e.g.
// "https://10.0.0.10/"). Call Client.Session.LoginWithPassword before
// using it for anything else.
//
// If transport is nil, a default *http.Transport is used with
// InsecureSkipVerify: true - i.e. TLS certificate verification is off by
// default, matching XCP-ng/XenServer's common self-signed-certificate
// setup. Pass a transport of your own to verify certificates.
func NewClient(url string, transport *http.Transport) (*Client, error) {
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	rpc, err := xmlrpc.NewClient(url, transport)
	if err != nil {
		return nil, err
	}

	return prepClient(rpc), nil
}
