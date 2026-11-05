package xenapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

// vmRecords returns n VM records in their decoded wire form, as
// VM.get_all_records delivers them.
func vmRecords(b *testing.B, n int) xmlrpc.Struct {
	b.Helper()
	records := xmlrpc.Struct{}
	for i := range n {
		rec := VMRecord{
			UUID:        fmt.Sprintf("uuid-%d", i),
			NameLabel:   fmt.Sprintf("vm-%d", i),
			PowerState:  VMPowerStateRunning,
			VCPUsMax:    4,
			OtherConfig: map[string]string{"a": "b", "c": "d"},
			VBDs:        []VBDRef{"OpaqueRef:1", "OpaqueRef:2"},
		}
		records[fmt.Sprintf("OpaqueRef:vm-%d", i)] = convertVMRecordToXen(rec)
	}
	return records
}

// BenchmarkConvertVMRecords measures decoding a VM.get_all_records result
// of 1000 VMs into Go values, without the XML layer.
func BenchmarkConvertVMRecords(b *testing.B) {
	input := vmRecords(b, 1000)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := convertVMRefToVMRecordMapToGo(input); err != nil {
			b.Fatal(err)
		}
	}
}

type staticTransport string

func (s staticTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader(string(s)))}, nil
}

// BenchmarkGetAllRecords measures VM.GetAllRecords end to end, XML
// decoding included, for 100 VMs.
func BenchmarkGetAllRecords(b *testing.B) {
	var body strings.Builder
	body.WriteString(`<?xml version="1.0"?><methodResponse><params><param><value><struct>` +
		`<member><name>Status</name><value><string>Success</string></value></member>` +
		`<member><name>Value</name><value><struct>`)
	for i := range 100 {
		fmt.Fprintf(&body, `<member><name>OpaqueRef:vm-%d</name><value><struct>`+
			`<member><name>uuid</name><value>uuid-%d</value></member>`+
			`<member><name>name_label</name><value>vm-%d</value></member>`+
			`<member><name>power_state</name><value>Running</value></member>`+
			`<member><name>VCPUs_max</name><value>4</value></member>`+
			`<member><name>other_config</name><value><struct><member><name>a</name><value>b</value></member></struct></value></member>`+
			`<member><name>VBDs</name><value><array><data><value>OpaqueRef:1</value><value>OpaqueRef:2</value></data></array></value></member>`+
			`</struct></value></member>`, i, i, i)
	}
	body.WriteString(`</struct></value></member></struct></value></param></params></methodResponse>`)

	client, err := NewClient("http://xapi.invalid", WithHTTPClient(&http.Client{Transport: staticTransport(body.String())}))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := client.VM.GetAllRecords(b.Context(), ""); err != nil {
			b.Fatal(err)
		}
	}
}
