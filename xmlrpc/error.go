package xmlrpc

import "fmt"

// Error represents a genuine XML-RPC protocol-level <fault> response.
//
// XenAPI itself doesn't use these - it always returns a normal
// (non-fault) value carrying its own Status/ErrorDescription fields,
// which the xenapi package turns into its own xenapi.Error - so this type
// only matters for other XML-RPC servers, or transport-level failures
// that occur before XenAPI's own dispatch logic runs.
type Error struct {
	code    string
	message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("xmlrpc fault %s: %s", e.code, e.message)
}

// Code returns the fault's faultCode, as a string regardless of whether
// the server sent it as an XML-RPC int or string.
func (e *Error) Code() string {
	return e.code
}

// Message returns the fault's faultString.
func (e *Error) Message() string {
	return e.message
}

// maxErrorBody caps how much of a non-2xx response body HTTPError keeps.
const maxErrorBody = 512

// HTTPError is returned when the server answers with a non-2xx HTTP status,
// e.g. from an authenticating proxy or a failing load balancer. The body
// is not decoded as XML-RPC.
type HTTPError struct {
	StatusCode int
	Status     string // e.g. "502 Bad Gateway"
	Body       string // first maxErrorBody bytes of the response body
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return "xmlrpc: HTTP " + e.Status
	}
	return fmt.Sprintf("xmlrpc: HTTP %s: %q", e.Status, e.Body)
}
