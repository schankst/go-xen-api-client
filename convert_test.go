package xenapi

import (
	"testing"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

func TestConvertSimpleTypesToGo(t *testing.T) {
	if v, err := convertStringToGo("test", "hello"); err != nil || v != "hello" {
		t.Fatalf("convertStringToGo: got (%q, %v), want (\"hello\", nil)", v, err)
	}

	if v, err := convertBoolToGo("test", true); err != nil || v != true {
		t.Fatalf("convertBoolToGo: got (%v, %v), want (true, nil)", v, err)
	}

	// XML-RPC ints arrive as strings on the wire and get parsed here.
	if v, err := convertIntToGo("test", "42"); err != nil || v != 42 {
		t.Fatalf("convertIntToGo: got (%v, %v), want (42, nil)", v, err)
	}

	if _, err := convertIntToGo("test", "not a number"); err == nil {
		t.Fatal("convertIntToGo: expected an error for a non-numeric string, got nil")
	}
}

func TestConvertEnumVMOperationsToGoKnownValue(t *testing.T) {
	value, err := convertEnumVMOperationsToGo("test", "start")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != VMOperationsStart {
		t.Fatalf("got %q, want %q", value, VMOperationsStart)
	}
}

// This is the fork's central behavior: a value the schema didn't know
// about (this happened for real with "sysprep", added by XCP-ng after
// the upstream v0.0.2 snapshot) must be passed through as-is instead of
// failing the whole record parse. Baked into the generator template
// itself (convertEnumTypeToGoFuncTemplate in xenapi.go) since v0.2.1, so
// every regeneration gets it automatically with no separate patch step.
func TestConvertEnumVMOperationsToGoUnknownValuePassesThrough(t *testing.T) {
	value, err := convertEnumVMOperationsToGo("test", "some_future_operation_xcp_ng_added")
	if err != nil {
		t.Fatalf("unexpected error for an unknown enum value: %v", err)
	}
	if string(value) != "some_future_operation_xcp_ng_added" {
		t.Fatalf("got %q, want the raw value passed through unchanged", value)
	}
}

// convertTimeToGo normally just unwraps a time.Time the xmlrpc decoder
// already parsed from a proper <dateTime.iso8601> element - that's the
// common case for almost every "datetime" field in the schema.
func TestConvertTimeToGoNormalValue(t *testing.T) {
	want := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	value, err := convertTimeToGo("test", want)
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
	value, err := convertTimeToGo("test", "1784931535.")
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
	value, err := convertTimeToGo("test", "1784931535")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1784931535, 0)
	if !value.Equal(want) {
		t.Fatalf("got %v, want %v", value, want)
	}
}

func TestConvertTimeToGoInvalidValue(t *testing.T) {
	if _, err := convertTimeToGo("test", "not a time"); err == nil {
		t.Fatal("convertTimeToGo: expected an error for a non-numeric, non-time.Time value, got nil")
	}
}

func TestConvertVMRecordToGo(t *testing.T) {
	record, err := convertVMRecordToGo("test", xmlrpc.Struct{
		"uuid":        "abc-123",
		"name_label":  "my-vm",
		"power_state": "Running",
		"tags":        []interface{}{"prod", "web"},
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
