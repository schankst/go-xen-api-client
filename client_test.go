package xenapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// These exercise the hand-written core of this package - NewClient, APICall,
// and Error's accessors - against recorded XenAPI response bodies. The
// generated *_gen.go files are not covered here: they are mechanical
// wrappers around APICall, regenerated wholesale every week, and testing
// each of the ~70 per-class methods would pin the shape of generated code
// rather than the behaviour of this package.
//
// APICall is the only place where the XML-RPC envelope's Status and
// ErrorDescription fields are interpreted, so it is also the only place
// where a host answering in an unexpected shape can take the caller down.

// newTestClient returns a Client whose RPC calls are answered by body.
func newTestClient(t *testing.T, body string) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, nil)
	return client
}

// successResponse renders a XenAPI success envelope around an XML-RPC value.
func successResponse(value string) string {
	return `<?xml version="1.0"?><methodResponse><params><param><value><struct>` +
		`<member><name>Status</name><value><string>Success</string></value></member>` +
		`<member><name>Value</name><value>` + value + `</value></member>` +
		`</struct></value></param></params></methodResponse>`
}

// failureResponse renders a XenAPI failure envelope with the given
// ErrorDescription array, or with the whole field omitted when
// errorDescription is the empty string.
func failureResponse(errorDescription string) string {
	members := `<member><name>Status</name><value><string>Failure</string></value></member>`
	if errorDescription != "" {
		members += `<member><name>ErrorDescription</name><value>` + errorDescription + `</value></member>`
	}
	return `<?xml version="1.0"?><methodResponse><params><param><value><struct>` +
		members +
		`</struct></value></param></params></methodResponse>`
}

// stringArray renders a []interface{} of XML-RPC strings as an <array>.
func stringArray(values ...string) string {
	body := `<array><data>`
	for _, v := range values {
		body += `<value><string>` + v + `</string></value>`
	}
	return body + `</data></array>`
}

func TestLoginWithPasswordSuccess(t *testing.T) {
	client := newTestClient(t, successResponse(`<string>OpaqueRef:session-1</string>`))

	session, err := client.Session.LoginWithPassword("root", "secret", "1.0", "test")
	if err != nil {
		t.Fatalf("LoginWithPassword: %v", err)
	}
	if session != SessionRef("OpaqueRef:session-1") {
		t.Fatalf("session = %q, want %q", session, "OpaqueRef:session-1")
	}
}

func TestAPICallSuccessExposesValue(t *testing.T) {
	client := newTestClient(t, successResponse(`<int>7</int>`))

	result, err := client.APICall("some.method")
	if err != nil {
		t.Fatalf("APICall: %v", err)
	}
	if result.Value != int64(7) {
		t.Fatalf("Value = %v (%T), want int64(7)", result.Value, result.Value)
	}
}

func TestAPICallErrorResponse(t *testing.T) {
	client := newTestClient(t, failureResponse(
		stringArray("HANDLE_INVALID", "VM", "OpaqueRef:dead-vm")))

	_, err := client.APICall("VM.get_record", "OpaqueRef:dead-vm")
	if err == nil {
		t.Fatal("expected an error for a Failure response, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("got error of type %T, want *Error", err)
	}
	if got := apiErr.Code(); got != ERR_HANDLE_INVALID {
		t.Errorf("Code() = %q, want %q", got, ERR_HANDLE_INVALID)
	}
	if got := apiErr.Type(); got != "VM" {
		t.Errorf("Type() = %q, want %q", got, "VM")
	}
	if got := apiErr.UUID(); got != "OpaqueRef:dead-vm" {
		t.Errorf("UUID() = %q, want %q", got, "OpaqueRef:dead-vm")
	}
	if got, want := apiErr.Error(), "API Error: HANDLE_INVALID VM OpaqueRef:dead-vm"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// Error codes that carry no type/uuid (most of them) must not render empty
// fields - see Error()'s implementation.
func TestAPICallErrorResponseWithoutTypeOrUUID(t *testing.T) {
	client := newTestClient(t, failureResponse(stringArray("PERMISSION_DENIED")))

	_, err := client.APICall("session.get_this_host")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("got error of type %T, want *Error", err)
	}
	if apiErr.Type() != "" || apiErr.UUID() != "" {
		t.Errorf("Type()/UUID() = %q/%q, want empty", apiErr.Type(), apiErr.UUID())
	}
	if got, want := apiErr.Error(), "API Error: PERMISSION_DENIED"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// A host that answers in a shape this client did not expect must produce an
// error, never a panic: a lost connection, a proxy error page, or a newer
// XAPI varianting the envelope are all reasons a caller would rather see.
func TestAPICallMalformedResponseReturnsError(t *testing.T) {
	cases := []struct {
		name     string
		response string
	}{
		{
			name:     "no Status field",
			response: `<?xml version="1.0"?><methodResponse><params><param><value><struct></struct></value></param></params></methodResponse>`,
		},
		{
			name: "Status is not a string",
			response: `<?xml version="1.0"?><methodResponse><params><param><value><struct>` +
				`<member><name>Status</name><value><int>1</int></value></member>` +
				`</struct></value></param></params></methodResponse>`,
		},
		{
			name:     "Failure without ErrorDescription",
			response: failureResponse(""),
		},
		{
			name:     "ErrorDescription is not an array",
			response: failureResponse(`<string>HANDLE_INVALID</string>`),
		},
		{
			name:     "ErrorDescription array is empty",
			response: failureResponse(`<array><data></data></array>`),
		},
		{
			name:     "ErrorDescription code is not a string",
			response: failureResponse(`<array><data><value><int>1</int></value></data></array>`),
		},
		{
			name:     "ErrorDescription type is not a string",
			response: failureResponse(stringArray("HANDLE_INVALID")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("APICall panicked instead of returning an error: %v", r)
				}
			}()

			client := newTestClient(t, tc.response)
			_, err := client.APICall("some.method")
			if err == nil {
				t.Fatal("expected an error for a malformed response, got nil")
			}
		})
	}
}

// The type/uuid slots are explicitly nullable in the XenAPI envelope, so a
// nil there is normal rather than malformed.
func TestAPICallErrorResponseWithNilTypeAndUUID(t *testing.T) {
	response := `<?xml version="1.0"?><methodResponse><params><param><value><struct>` +
		`<member><name>Status</name><value><string>Failure</string></value></member>` +
		`<member><name>ErrorDescription</name><value><array><data>` +
		`<value><string>HANDLE_INVALID</string></value>` +
		`<value></value><value></value>` +
		`</data></array></value></member>` +
		`</struct></value></param></params></methodResponse>`

	client := newTestClient(t, response)
	_, err := client.APICall("VM.get_record", "OpaqueRef:x")
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("got error of type %T (%v), want *Error", err, err)
	}
	if apiErr.Code() != ERR_HANDLE_INVALID {
		t.Errorf("Code() = %q, want %q", apiErr.Code(), ERR_HANDLE_INVALID)
	}
	if apiErr.Type() != "" || apiErr.UUID() != "" {
		t.Errorf("Type()/UUID() = %q/%q, want empty", apiErr.Type(), apiErr.UUID())
	}
	if strings.Contains(apiErr.Error(), "  ") {
		t.Errorf("Error() has an empty field: %q", apiErr.Error())
	}
}
