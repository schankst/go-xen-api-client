package xmlrpc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var reply string
	err = client.Call("session.login_with_password", []interface{}{"root", "pw"}, &reply)
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

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var reply interface{}
	err = client.Call("some.method", nil, &reply)
	if err == nil {
		t.Fatal("expected an error for a fault response, got nil")
	}
	if _, ok := err.(*Error); !ok {
		t.Fatalf("got error of type %T, want *Error", err)
	}
}

// callAgainstResponse runs one Call against a server that always returns
// the given methodResponse body, and returns whatever Call returned.
func callAgainstResponse(t *testing.T, responseBody string, reply interface{}) error {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client.Call("some.method", nil, reply)
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
		reply    interface{}
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

	t.Run("interface{}", func(t *testing.T) {
		var reply interface{}
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

// XCP-ng/XenServer pools use cookies to pin a client to whichever host
// handled the first request. Call must capture the first response's
// cookies and replay them on every subsequent call.
func TestClientCallReplaysCookies(t *testing.T) {
	var receivedCookies []string
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if cookie, err := r.Cookie("pool_session"); err == nil {
			receivedCookies = append(receivedCookies, cookie.Value)
		} else {
			receivedCookies = append(receivedCookies, "")
		}

		if requestCount == 1 {
			http.SetCookie(w, &http.Cookie{Name: "pool_session", Value: "pinned-to-host-1"})
		}
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param>` +
			`<value><string>ok</string></value></param></params></methodResponse>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	for i := 0; i < 3; i++ {
		var reply string
		if err := client.Call("noop", nil, &reply); err != nil {
			t.Fatalf("Call #%d: %v", i+1, err)
		}
	}

	if len(receivedCookies) != 3 {
		t.Fatalf("got %d requests, want 3", len(receivedCookies))
	}
	if receivedCookies[0] != "" {
		t.Errorf("first request should have no cookie yet, got %q", receivedCookies[0])
	}
	if receivedCookies[1] != "pinned-to-host-1" || receivedCookies[2] != "pinned-to-host-1" {
		t.Errorf("subsequent requests should replay the captured cookie, got %v", receivedCookies[1:])
	}
}

func TestClientCallConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param>` +
			`<value><string>ok</string></value></param></params></methodResponse>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var reply string
			errs <- client.Call("noop", nil, &reply)
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("concurrent Call failed: %v", err)
		}
	}
}
