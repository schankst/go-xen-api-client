package xenapi

import (
	"testing"

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

// This is the fork's central patch: a value the schema didn't know about
// (this happened for real with "sysprep", added by XCP-ng after the
// upstream v0.0.2 snapshot) must be passed through as-is instead of
// failing the whole record parse. See convert_gen.go's generated default
// case and patch_enums.go, which re-applies this after every regeneration.
func TestConvertEnumVMOperationsToGoUnknownValuePassesThrough(t *testing.T) {
	value, err := convertEnumVMOperationsToGo("test", "some_future_operation_xcp_ng_added")
	if err != nil {
		t.Fatalf("unexpected error for an unknown enum value: %v", err)
	}
	if string(value) != "some_future_operation_xcp_ng_added" {
		t.Fatalf("got %q, want the raw value passed through unchanged", value)
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
