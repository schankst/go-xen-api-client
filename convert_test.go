package xenapi

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

func TestConvertSimpleTypesToGo(t *testing.T) {
	if v, err := toValue[string]("hello"); err != nil || v != "hello" {
		t.Fatalf("toValue[string]: got (%q, %v), want (\"hello\", nil)", v, err)
	}

	if v, err := toValue[bool](true); err != nil || v != true {
		t.Fatalf("toValue[bool]: got (%v, %v), want (true, nil)", v, err)
	}

	// XML-RPC ints arrive as strings on the wire and get parsed here.
	if v, err := toInt("42"); err != nil || v != 42 {
		t.Fatalf("toInt: got (%v, %v), want (42, nil)", v, err)
	}

	if _, err := toInt("not a number"); err == nil {
		t.Fatal("toInt: expected an error for a non-numeric string, got nil")
	}
}

func TestConvertEnumVMOperationsToGoKnownValue(t *testing.T) {
	value, err := toString[VMOperations]("start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != VMOperationsStart {
		t.Fatalf("got %q, want %q", value, VMOperationsStart)
	}
}

// Enum values newer than the schema (e.g. "sysprep" once) pass through.
func TestConvertEnumVMOperationsToGoUnknownValuePassesThrough(t *testing.T) {
	value, err := toString[VMOperations]("some_future_operation_xcp_ng_added")
	if err != nil {
		t.Fatalf("unexpected error for an unknown enum value: %v", err)
	}
	if string(value) != "some_future_operation_xcp_ng_added" {
		t.Fatalf("got %q, want the raw value passed through unchanged", value)
	}
}

// toTime normally just unwraps a time.Time the xmlrpc decoder
// already parsed from a proper <dateTime.iso8601> element - that's the
// common case for almost every "datetime" field in the schema.
func TestConvertTimeToGoNormalValue(t *testing.T) {
	want := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	value, err := toTime(want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !value.Equal(want) {
		t.Fatalf("got %v, want %v", value, want)
	}
}

// This happened for real against a live XCP-ng host: Event.from/Event.next's
// deprecated "timestamp" field (Deprecated_s in the schema) arrived as an
// OCaml-style float-as-string Unix timestamp - note the trailing dot, from
// OCaml's string_of_float on a whole number - instead of a wire-tagged
// dateTime.iso8601 value, which the xmlrpc decoder faithfully passes
// through as a Go string rather than erroring on the wire format itself.
// Tolerating that here - instead of hard-failing - is what let the "power"
// use case in the xen CLI actually consume task-completion events instead
// of falling back to polling alone.
func TestConvertTimeToGoOCamlFloatStringFallback(t *testing.T) {
	value, err := toTime("1784931535.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1784931535, 0)
	if !value.Equal(want) {
		t.Fatalf("got %v, want %v", value, want)
	}
}

// A plain integer string (no trailing dot) works too - ParseFloat accepts
// both forms, so there's no need for a separate integer-only code path.
func TestConvertTimeToGoPlainIntegerStringFallback(t *testing.T) {
	value, err := toTime("1784931535")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1784931535, 0)
	if !value.Equal(want) {
		t.Fatalf("got %v, want %v", value, want)
	}
}

func TestConvertTimeToGoInvalidValue(t *testing.T) {
	if _, err := toTime("not a time"); err == nil {
		t.Fatal("toTime: expected an error for a non-numeric, non-time.Time value, got nil")
	}
}

func TestConvertVMRecordToGo(t *testing.T) {
	record, err := convertVMRecordToGo(xmlrpc.Struct{
		"uuid":        "abc-123",
		"name_label":  "my-vm",
		"power_state": "Running",
		"tags":        []any{"prod", "web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.UUID != "abc-123" {
		t.Errorf("UUID = %q, want %q", record.UUID, "abc-123")
	}
	if record.NameLabel != "my-vm" {
		t.Errorf("NameLabel = %q, want %q", record.NameLabel, "my-vm")
	}
	if record.PowerState != VMPowerStateRunning {
		t.Errorf("PowerState = %q, want %q", record.PowerState, VMPowerStateRunning)
	}
	if len(record.Tags) != 2 || record.Tags[0] != "prod" || record.Tags[1] != "web" {
		t.Errorf("Tags = %v, want [prod web]", record.Tags)
	}
}

func TestConvertRecordRoundTrip(t *testing.T) {
	want := VMRecord{
		UUID:              "abc-123",
		NameLabel:         "my-vm",
		PowerState:        VMPowerStateRunning,
		VCPUsMax:          4,
		MemoryStaticMax:   1 << 30,
		AllowedOperations: []VMOperations{VMOperationsStart, "future_op"},
		OtherConfig:       map[string]string{"k": "v"},
		VBDs:              []VBDRef{"OpaqueRef:1", "OpaqueRef:2"},
		IsATemplate:       true,
		SnapshotTime:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	got, err := convertVMRecordToGo(convertVMRecordToXen(want))
	if err != nil {
		t.Fatal(err)
	}
	// The wire form has no nil arrays or structs, so compare on the wire.
	if !reflect.DeepEqual(convertVMRecordToXen(got), convertVMRecordToXen(want)) {
		t.Errorf("round trip changed the record:\ngot  %+v\nwant %+v", got, want)
	}
	if got.NameLabel != want.NameLabel || !got.SnapshotTime.Equal(want.SnapshotTime) || got.AllowedOperations[1] != "future_op" {
		t.Errorf("round trip lost values: %+v", got)
	}
}

// A mismatch deep inside a result is reported with its location.
func TestConvertErrorPath(t *testing.T) {
	input := xmlrpc.Struct{
		"OpaqueRef:vm": xmlrpc.Struct{
			"name_label": "ok",
			"VBDs":       []any{"OpaqueRef:1", int64(7)},
		},
	}
	_, err := convertVMRefToVMRecordMapToGo(input)
	if err == nil {
		t.Fatal("expected an error for an int64 inside VBDs, got nil")
	}
	if want := "[OpaqueRef:vm].VBDs[1]: got int64 7, want string"; err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestConvertMissingAndNilFieldsKeepZeroValue(t *testing.T) {
	record, err := convertVMRecordToGo(xmlrpc.Struct{"name_label": nil, "VBDs": nil})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(record, VMRecord{}) {
		t.Errorf("record = %+v, want the zero value", record)
	}
}

func TestCallAndConvertWrapsMismatch(t *testing.T) {
	client := newTestClient(t, successResponse(`<int>5</int>`))
	_, err := client.VM.GetNameLabel(t.Context(), "", "")
	if err == nil || !strings.Contains(err.Error(), "VM.get_name_label: unexpected response: got int64 5, want string") {
		t.Fatalf("error = %v, want one naming the method and the mismatch", err)
	}
}
