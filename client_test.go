package xenapi

import (
	"flag"
	"log"
	"os"
	"testing"
)

// TestAuthentication requires a real (or mock) XenAPI server on
// localhost:40080; there isn't one in CI, so this only runs when opted
// into explicitly via -run TestAuthentication against a real endpoint.
func TestAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("requires a live XenAPI server on localhost:40080; skipped with -short")
	}

	client, err := NewClient("http://localhost:40080", nil)

	sessionRef, err := client.Session.LoginWithPassword("terraform", "testing", "1.0", "terraform")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	err = client.Session.Logout(sessionRef)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	log.SetOutput(os.Stdout)
	os.Exit(m.Run())
}
