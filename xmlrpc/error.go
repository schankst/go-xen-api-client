package xmlrpc

import "fmt"

// Error is an XML-RPC <fault> response. XenAPI reports its own failures in
// a regular response instead (see xenapi.Error).
type Error struct {
	code    string
	message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("xmlrpc fault %s: %s", e.code, e.message)
}

// Code returns the fault's faultCode as a string.
func (e *Error) Code() string {
	return e.code
}

// Message returns the fault's faultString.
func (e *Error) Message() string {
	return e.message
}

// maxErrorBody caps how much of a non-2xx response body HTTPError keeps.
const maxErrorBody = 512

// HTTPError reports a non-2xx HTTP response, e.g. from a proxy.
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
