package xmlrpc

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// This replaces the old client_test.go, which required a live server on
// localhost:40080 and could never pass in CI. httptest.Server gives a
// real, working, CI-safe integration test instead of a network dependency
// or a skipped test.

func TestClientCallSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("server: reading request body: %v", err)
		}
		if got := string(body); !strings.Contains(got, "<methodName>session.login_with_password</methodName>") {
			t.Errorf("server received unexpected request body: %s", got)
		}

		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param>` +
			`<value><string>OpaqueRef:abc-123</string></value>` +
			`</param></params></methodResponse>`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)

	var reply string
	err := client.Call(t.Context(), "session.login_with_password", []any{"root", "pw"}, &reply)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if reply != "OpaqueRef:abc-123" {
		t.Fatalf("reply = %q, want %q", reply, "OpaqueRef:abc-123")
	}
}

func TestClientCallFault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0"?><methodResponse><fault><value><struct>` +
			`<member><name>faultCode</name><value><int>1</int></value></member>` +
			`<member><name>faultString</name><value><string>nope</string></value></member>` +
			`</struct></value></fault></methodResponse>`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)

	var reply any
	err := client.Call(t.Context(), "some.method", nil, &reply)
	if err == nil {
		t.Fatal("expected an error for a fault response, got nil")
	}
	if _, ok := err.(*Error); !ok {
		t.Fatalf("got error of type %T, want *Error", err)
	}
}

// callAgainstResponse runs one Call against a server that always returns
// the given methodResponse body, and returns whatever Call returned.
func callAgainstResponse(t *testing.T, responseBody string, reply any) error {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	return client.Call(t.Context(), "some.method", nil, reply)
}

// A decoded value whose Go type doesn't match the reply's element type
// must be reported as an error, not stored. Before this was handled,
// reflect.Value.Set panicked ("reflect.Set: value of type string is not
// assignable to type int64") and took the caller's goroutine down with
// it - the caller could only ever see the panic, never a returned error.
func TestClientCallReplyTypeMismatchReturnsError(t *testing.T) {
	cases := []struct {
		name     string
		response string
		reply    any
	}{
		{
			// Legal on the wire: an empty <value/> decodes as "" (string).
			name:     "empty value into *int64",
			response: `<methodResponse><params><param><value></value></param></params></methodResponse>`,
			reply:    new(int64),
		},
		{
			// XML-RPC <int> always decodes as int64, never as int.
			name:     "int into *int",
			response: `<methodResponse><params><param><value><int>42</int></value></param></params></methodResponse>`,
			reply:    new(int),
		},
		{
			name:     "dateTime into *int64",
			response: `<methodResponse><params><param><value><dateTime.iso8601>20261021T10:00:00</dateTime.iso8601></value></param></params></methodResponse>`,
			reply:    new(int64),
		},
		{
			name:     "array into *Struct",
			response: `<methodResponse><params><param><value><array><data></data></array></value></param></params></methodResponse>`,
			reply:    new(Struct),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Call panicked instead of returning an error: %v", r)
				}
			}()

			err := callAgainstResponse(t, tc.response, tc.reply)
			if err == nil {
				t.Fatalf("Call returned nil error for a reply of type %T", tc.reply)
			}
			if !strings.Contains(err.Error(), "some.method") {
				t.Errorf("error should name the method, got: %v", err)
			}
		})
	}
}

// The canonical reply types each XML-RPC type maps onto must keep working
// unchanged - the mismatch handling above must not reject these.
func TestClientCallReplyTypeMatchStillWorks(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var reply string
		err := callAgainstResponse(t,
			`<methodResponse><params><param><value><string>hello</string></value></param></params></methodResponse>`,
			&reply)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if reply != "hello" {
			t.Fatalf("reply = %q, want %q", reply, "hello")
		}
	})

	t.Run("int64", func(t *testing.T) {
		var reply int64
		err := callAgainstResponse(t,
			`<methodResponse><params><param><value><int>42</int></value></param></params></methodResponse>`,
			&reply)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if reply != 42 {
			t.Fatalf("reply = %d, want 42", reply)
		}
	})

	t.Run("time.Time", func(t *testing.T) {
		want := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
		var reply time.Time
		err := callAgainstResponse(t,
			`<methodResponse><params><param><value><dateTime.iso8601>20261021T10:00:00</dateTime.iso8601></value></param></params></methodResponse>`,
			&reply)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if !reply.Equal(want) {
			t.Fatalf("reply = %v, want %v", reply, want)
		}
	})

	t.Run("Struct", func(t *testing.T) {
		var reply Struct
		err := callAgainstResponse(t,
			`<methodResponse><params><param><value><struct>`+
				`<member><name>Status</name><value><string>Success</string></value></member>`+
				`</struct></value></param></params></methodResponse>`,
			&reply)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if reply["Status"] != "Success" {
			t.Fatalf("reply = %v, want Status=Success", reply)
		}
	})

	t.Run("any", func(t *testing.T) {
		var reply any
		err := callAgainstResponse(t,
			`<methodResponse><params><param><value><int>7</int></value></param></params></methodResponse>`,
			&reply)
		if err != nil {
			t.Fatalf("Call: %v", err)
		}
		if reply != int64(7) {
			t.Fatalf("reply = %v (%T), want int64(7)", reply, reply)
		}
	})
}

// Cookies (e.g. a load balancer's affinity cookie) follow normal cookie
// semantics: set on any response, updated, and deleted.
func TestClientCallCookies(t *testing.T) {
	steps := []struct {
		set  *http.Cookie // cookie the server sets in its response
		want string       // value the request must carry, "" for none
	}{
		{set: nil, want: ""},
		{set: &http.Cookie{Name: "lb", Value: "host-1"}, want: ""},
		{set: &http.Cookie{Name: "lb", Value: "host-2"}, want: "host-1"},
		{set: &http.Cookie{Name: "lb", MaxAge: -1}, want: "host-2"},
		{set: nil, want: ""},
	}

	var step int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := ""
		if c, err := r.Cookie("lb"); err == nil {
			got = c.Value
		}
		if got != steps[step].want {
			t.Errorf("request %d carried lb=%q, want %q", step+1, got, steps[step].want)
		}
		if steps[step].set != nil {
			http.SetCookie(w, steps[step].set)
		}
		step++
		_, _ = io.WriteString(w, `<?xml version="1.0"?><methodResponse><params><param>`+
			`<value><string>ok</string></value></param></params></methodResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	for i := range steps {
		if err := client.Call(t.Context(), "noop", nil, nil); err != nil {
			t.Fatalf("Call %d: %v", i+1, err)
		}
	}
}

func TestClientCallConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param>` +
			`<value><string>ok</string></value></param></params></methodResponse>`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() {
			var reply string
			errs <- client.Call(t.Context(), "noop", nil, &reply)
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("concurrent Call failed: %v", err)
		}
	}
}

func TestClientCallHTTPErrorStatus(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"unauthorized", http.StatusUnauthorized, "denied"},
		{"server error with XML body", http.StatusInternalServerError, `<?xml version="1.0"?><methodResponse><params><param><value>ok</value></param></params></methodResponse>`},
		{"proxy HTML page", http.StatusBadGateway, "<html><body>Bad Gateway</body></html>"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()

			var reply any
			err := NewClient(server.URL, nil).Call(t.Context(), "some.method", nil, &reply)

			var httpErr *HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("Call error = %v (%T), want *HTTPError", err, err)
			}
			if httpErr.StatusCode != tc.status {
				t.Errorf("StatusCode = %d, want %d", httpErr.StatusCode, tc.status)
			}
			if !strings.Contains(err.Error(), strconv.Itoa(tc.status)) {
				t.Errorf("Error() = %q, want it to name the status code", err)
			}
			if reply != nil {
				t.Errorf("reply = %v, want it left untouched", reply)
			}
		})
	}
}

func TestHTTPErrorTruncatesBody(t *testing.T) {
	long := strings.Repeat("x", 10*maxErrorBody)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, long)
	}))
	defer server.Close()

	err := NewClient(server.URL, nil).Call(t.Context(), "some.method", nil, nil)

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("Call error = %v, want *HTTPError", err)
	}
	if len(httpErr.Body) != maxErrorBody {
		t.Errorf("len(Body) = %d, want %d", len(httpErr.Body), maxErrorBody)
	}
}

// blockingServer answers only after the request's context is done or the
// test ends, so a Call against it returns only through cancellation.
func blockingServer(t *testing.T) *httptest.Server {
	t.Helper()
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(func() {
		close(release)
		server.Close()
	})
	return server
}

func TestClientCallCanceled(t *testing.T) {
	server := blockingServer(t)
	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(20*time.Millisecond, cancel)

	err := NewClient(server.URL, nil).Call(ctx, "slow.method", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Call error = %v, want context.Canceled", err)
	}
}

func TestClientCallDeadline(t *testing.T) {
	server := blockingServer(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	err := NewClient(server.URL, nil).Call(ctx, "slow.method", nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Call error = %v, want context.DeadlineExceeded", err)
	}
}

func TestClientCallAlreadyCanceledSendsNothing(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := NewClient(server.URL, nil).Call(ctx, "some.method", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Call error = %v, want context.Canceled", err)
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("server received %d requests, want 0", n)
	}
}
