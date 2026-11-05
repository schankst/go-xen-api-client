package xenapi_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"time"

	xenapi "github.com/schankst/go-xen-api-client"
	"github.com/schankst/go-xen-api-client/xmlrpc"
)

func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	xapi, err := xenapi.NewClient("https://10.0.0.10")
	if err != nil {
		log.Fatal(err)
	}
	session, err := xapi.Session.LoginWithPassword(ctx, "root", "password", "1.0", "example")
	if err != nil {
		log.Fatal(err)
	}
	defer xapi.Session.Logout(ctx, session)

	// One call for all VMs; avoid one call per object.
	vms, err := xapi.VM.GetAllRecords(ctx, session)
	if err != nil {
		log.Fatal(err)
	}
	for _, vm := range vms {
		if !vm.IsATemplate && !vm.IsControlDomain {
			log.Println(vm.NameLabel, vm.PowerState)
		}
	}
}

// A host with a self-signed certificate: trust that certificate.
func ExampleWithTLSConfig() {
	pem, err := os.ReadFile("host-cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(pem)

	xapi, err := xenapi.NewClient("https://10.0.0.10", xenapi.WithTLSConfig(&tls.Config{RootCAs: roots}))
	if err != nil {
		log.Fatal(err)
	}
	_ = xapi
}

func ExampleError() {
	var (
		ctx     context.Context
		xapi    *xenapi.Client
		session xenapi.SessionRef
		vm      xenapi.VMRef
	)

	err := xapi.VM.Start(ctx, session, vm, false, false)
	switch {
	case errors.Is(err, xenapi.ErrVMBadPowerState):
		var apiErr *xenapi.Error
		errors.As(err, &apiErr)
		log.Printf("VM is not halted: %v", apiErr.Params)
	case err != nil:
		log.Fatal(err)
	}
}

// Watch VMs for changes: Event.From returns snapshots in wire form, which
// DecodeVMRecord turns into VMRecords.
func ExampleDecodeVMRecord() {
	var (
		ctx     context.Context
		xapi    *xenapi.Client
		session xenapi.SessionRef
	)

	token := ""
	for {
		batch, err := xapi.Event.From(ctx, session, []string{"vm"}, token, 30)
		if err != nil {
			log.Fatal(err)
		}
		token, _ = batch["token"].(string)
		events, _ := batch["events"].([]any)
		for _, e := range events {
			event, _ := e.(xmlrpc.Struct)
			snapshot, ok := event["snapshot"].(xmlrpc.Struct)
			if !ok {
				continue // e.g. a deletion
			}
			vm, err := xenapi.DecodeVMRecord(snapshot)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(event["operation"], vm.NameLabel, vm.PowerState)
		}
	}
}
