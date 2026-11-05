package xenapi

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTLSServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, successResponse(`<string>ok</string>`))
	}))
	// Rejected handshakes are the point of some tests; don't log them.
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	t.Cleanup(server.Close)
	return server
}

func TestNewClientVerifiesTLSByDefault(t *testing.T) {
	server := newTLSServer(t)

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.APICall(t.Context(), "some.method")

	var certErr *tls.CertificateVerificationError
	if !errors.As(err, &certErr) {
		t.Fatalf("APICall against a self-signed host: error = %v, want a certificate verification error", err)
	}
}

func TestNewClientTLSOptions(t *testing.T) {
	server := newTLSServer(t)
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())

	cases := []struct {
		name string
		opt  Option
	}{
		{"trusted root", WithTLSConfig(&tls.Config{RootCAs: roots})},
		{"explicit insecure", WithTLSConfig(&tls.Config{InsecureSkipVerify: true})},
		{"own HTTP client", WithHTTPClient(server.Client())},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(server.URL, tc.opt)
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.APICall(t.Context(), "some.method")
			if err != nil {
				t.Fatalf("APICall: %v", err)
			}
			if result.Value != "ok" {
				t.Errorf("Value = %v, want ok", result.Value)
			}
		})
	}
}

func TestNewClientRejectsConflictingOptions(t *testing.T) {
	_, err := NewClient("https://10.0.0.10", WithHTTPClient(http.DefaultClient), WithTLSConfig(&tls.Config{}))
	if err == nil {
		t.Fatal("NewClient with WithHTTPClient and WithTLSConfig succeeded, want an error")
	}
}

func TestNewClientValidatesURL(t *testing.T) {
	for _, u := range []string{"", "10.0.0.10", "ftp://10.0.0.10", "https://", "://bad", "/relative"} {
		if _, err := NewClient(u); err == nil {
			t.Errorf("NewClient(%q) succeeded, want an error", u)
		}
	}
	for _, u := range []string{"https://10.0.0.10", "https://10.0.0.10/", "http://host.example:8080"} {
		if _, err := NewClient(u); err != nil {
			t.Errorf("NewClient(%q) = %v, want no error", u, err)
		}
	}
}
