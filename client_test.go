package xenapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

// newTestClient returns a Client whose RPC calls are answered by body.
func newTestClient(t *testing.T, body string) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
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

// stringArray renders a []any of XML-RPC strings as an <array>.
func stringArray(values ...string) string {
	var b strings.Builder
	b.WriteString(`<array><data>`)
	for _, v := range values {
		b.WriteString(`<value><string>` + v + `</string></value>`)
	}
	b.WriteString(`</data></array>`)
	return b.String()
}

func TestLoginWithPasswordSuccess(t *testing.T) {
	client := newTestClient(t, successResponse(`<string>OpaqueRef:session-1</string>`))

	session, err := client.Session.LoginWithPassword(t.Context(), "root", "secret", "1.0", "test")
	if err != nil {
		t.Fatalf("LoginWithPassword: %v", err)
	}
	if session != SessionRef("OpaqueRef:session-1") {
		t.Fatalf("session = %q, want %q", session, "OpaqueRef:session-1")
	}
}

func TestAPICallSuccessExposesValue(t *testing.T) {
	client := newTestClient(t, successResponse(`<int>7</int>`))

	result, err := client.APICall(t.Context(), "some.method")
	if err != nil {
		t.Fatalf("APICall: %v", err)
	}
	if result.Value != int64(7) {
		t.Fatalf("Value = %v (%T), want int64(7)", result.Value, result.Value)
	}
}

func TestAPICallErrorResponse(t *testing.T) {
	cases := []struct {
		name        string
		description string
		want        Error
		wantString  string
	}{
		{
			name:        "no params",
			description: stringArray("PERMISSION_DENIED"),
			want:        Error{Code: ErrPermissionDenied},
			wantString:  "xenapi: PERMISSION_DENIED",
		},
		{
			name:        "two params",
			description: stringArray("HANDLE_INVALID", "VM", "OpaqueRef:dead-vm"),
			want:        Error{Code: ErrHandleInvalid, Params: []string{"VM", "OpaqueRef:dead-vm"}},
			wantString:  `xenapi: HANDLE_INVALID ["VM" "OpaqueRef:dead-vm"]`,
		},
		{
			name:        "three params",
			description: stringArray("VM_BAD_POWER_STATE", "OpaqueRef:vm", "halted", "running"),
			want:        Error{Code: ErrVMBadPowerState, Params: []string{"OpaqueRef:vm", "halted", "running"}},
			wantString:  `xenapi: VM_BAD_POWER_STATE ["OpaqueRef:vm" "halted" "running"]`,
		},
		{
			name:        "backend output with spaces and newlines",
			description: stringArray("SR_BACKEND_FAILURE_47", "", "The SR is not available", "stderr line 1\nline 2"),
			want:        Error{Code: ErrorCode("SR_BACKEND_FAILURE_47"), Params: []string{"", "The SR is not available", "stderr line 1\nline 2"}},
			wantString:  `xenapi: SR_BACKEND_FAILURE_47 ["" "The SR is not available" "stderr line 1\nline 2"]`,
		},
		{
			name: "nil and non-string params are kept",
			description: `<array><data><value><string>HANDLE_INVALID</string></value>` +
				`<value></value><value><int>3</int></value></data></array>`,
			want:       Error{Code: ErrHandleInvalid, Params: []string{"", "3"}},
			wantString: `xenapi: HANDLE_INVALID ["" "3"]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(t, failureResponse(tc.description))
			_, err := client.APICall(t.Context(), "some.method")

			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("APICall error = %v (%T), want *Error", err, err)
			}
			if apiErr.Code != tc.want.Code || !slices.Equal(apiErr.Params, tc.want.Params) {
				t.Errorf("error = %+v, want %+v", *apiErr, tc.want)
			}
			if got := apiErr.Error(); got != tc.wantString {
				t.Errorf("Error() = %s, want %s", got, tc.wantString)
			}
		})
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("APICall panicked instead of returning an error: %v", r)
				}
			}()

			client := newTestClient(t, tc.response)
			_, err := client.APICall(t.Context(), "some.method")
			if err == nil {
				t.Fatal("expected an error for a malformed response, got nil")
			}
		})
	}
}

func TestGeneratedMethodHonorsContext(t *testing.T) {
	client := newTestClient(t, successResponse(`<string>OpaqueRef:session-1</string>`))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := client.Session.LoginWithPassword(ctx, "root", "secret", "1.0", "test")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("LoginWithPassword with a canceled context: error = %v, want context.Canceled", err)
	}
}

func TestErrorIs(t *testing.T) {
	client := newTestClient(t, failureResponse(stringArray("HANDLE_INVALID", "VM", "OpaqueRef:x")))
	_, err := client.APICall(t.Context(), "VM.get_record", "OpaqueRef:x")
	wrapped := fmt.Errorf("looking up VM: %w", err)

	if !errors.Is(wrapped, ErrHandleInvalid) {
		t.Errorf("errors.Is(%v, ErrHandleInvalid) = false, want true", wrapped)
	}
	if errors.Is(wrapped, ErrPermissionDenied) {
		t.Errorf("errors.Is(%v, ErrPermissionDenied) = true, want false", wrapped)
	}
	if !errors.Is(wrapped, ErrorCode("HANDLE_INVALID")) {
		t.Error("errors.Is with an ErrorCode built from the wire string = false, want true")
	}
}
