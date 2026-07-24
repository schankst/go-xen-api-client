package xmlrpc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
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
