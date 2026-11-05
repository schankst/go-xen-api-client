package xmlrpc

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

// wrapAsResponse embeds a <value> fragment (as produced by writeValue) into
// a full <methodResponse><params><param>...</param></params></methodResponse>
// envelope, so decodeResponse can be exercised the same way it would be
// against a real server.
func wrapAsResponse(value string) string {
	return fmt.Sprintf(
		`<?xml version="1.0"?><methodResponse><params><param>%s</param></params></methodResponse>`,
		value,
	)
}

// These exercise writeValue (encode.go) and decodeResponse/decodeValue
// (decode.go) together: build the XML-RPC wire representation of a value,
// wrap it as a response, then decode it back, and check the round trip
// preserves the value.
func TestValueRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{"string", "hello"},
		{"string with special chars", `<tag> & "quoted" 'text'`},
		{"empty string", ""},
		{"int", 42},
		{"negative int", -7},
		{"bool true", true},
		{"bool false", false},
		{"struct", Struct{"name": "vm1", "count": int64(3)}}, // ints always come back as int64
		{"array", []any{"a", "b", "c"}},
		{"nested struct in array", []any{Struct{"k": "v"}}},
		{"empty struct", Struct{}},
		{"empty array", []any{}},
		{"double", 0.1},
		{"tiny double", 1e-9},
		{"huge double", 1e300},
		{"negative double", -2.5},
		{"zero double", 0.0},
		{"max double", math.MaxFloat64},
		{"smallest denormal", math.SmallestNonzeroFloat64},
		{"float32", float32(0.1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder
			if err := writeValue(&sb, tc.value); err != nil {
				t.Fatalf("writeValue: %v", err)
			}
			b := []byte(wrapAsResponse(sb.String()))

			result, err := decodeResponse(b)
			if err != nil {
				t.Fatalf("decodeResponse(%q) failed: %v", b, err)
			}

			switch want := tc.value.(type) {
			case int:
				got, ok := result.(int64)
				if !ok || got != int64(want) {
					t.Fatalf("got %#v (%T), want %d", result, result, want)
				}
			case float32:
				got, ok := result.(float64)
				if !ok || float32(got) != want {
					t.Fatalf("got %#v (%T), want %v", result, result, want)
				}
			default:
				if !reflect.DeepEqual(result, tc.value) {
					t.Fatalf("got %#v, want %#v", result, tc.value)
				}
			}
		})
	}
}

func TestWriteValueUnsupportedTypeErrors(t *testing.T) {
	type unsupported struct{}
	var sb strings.Builder
	if err := writeValue(&sb, unsupported{}); err == nil {
		t.Fatal("expected an error for an unsupported value type, got none")
	}
}

func TestWriteValueDoubleWireFormat(t *testing.T) {
	cases := map[float64]string{
		1e-9:  "<value><double>0.000000001</double></value>",
		1e21:  "<value><double>1000000000000000000000</double></value>",
		-0.25: "<value><double>-0.25</double></value>",
		3:     "<value><double>3</double></value>",
	}
	for in, want := range cases {
		var sb strings.Builder
		if err := writeValue(&sb, in); err != nil {
			t.Fatalf("writeValue(%v): %v", in, err)
		}
		if got := sb.String(); got != want {
			t.Errorf("writeValue(%v) = %s, want %s", in, got, want)
		}
	}
}

func TestWriteValueNonFiniteDoubleErrors(t *testing.T) {
	for _, v := range []any{math.NaN(), math.Inf(1), math.Inf(-1), float32(math.Inf(1))} {
		var sb strings.Builder
		if err := writeValue(&sb, v); err == nil {
			t.Errorf("writeValue(%v) succeeded with %s, want an error", v, sb.String())
		}
	}
}

func TestWriteValueNilErrors(t *testing.T) {
	var sb strings.Builder
	if err := writeValue(&sb, nil); err == nil {
		t.Fatal("expected an error for a nil value, got none")
	}
}

// Malformed or truncated responses must return an error, not hang.
func TestDecodeResponseMalformedDoesNotHang(t *testing.T) {
	cases := []string{
		`<methodResponse><params><param><value><struct><member><name>a</name><value><string>b</string>`,
		`<methodResponse><params><param><value><array><data><value><string>a</string>`,
		`<methodResponse><params><param><value><struct>`,
		`<methodResponse>`,
		``,
		`not xml at all`,
	}

	for _, xml := range cases {
		_, err := decodeResponse([]byte(xml))
		if err == nil {
			t.Fatalf("decodeResponse(%q): expected an error for malformed input, got nil", xml)
		}
	}
}

func TestDecodeResponseFault(t *testing.T) {
	body := `<?xml version="1.0"?><methodResponse><fault><value><struct>` +
		`<member><name>faultCode</name><value><int>7</int></value></member>` +
		`<member><name>faultString</name><value><string>boom</string></value></member>` +
		`</struct></value></fault></methodResponse>`

	_, err := decodeResponse([]byte(body))
	if err == nil {
		t.Fatal("expected an error for a <fault> response, got nil")
	}
	xerr, ok := err.(*Error)
	if !ok {
		t.Fatalf("got error of type %T, want *Error", err)
	}
	if xerr.Code() != "7" {
		t.Errorf("Code() = %q, want %q", xerr.Code(), "7")
	}
	if xerr.Message() != "boom" {
		t.Errorf("Message() = %q, want %q", xerr.Message(), "boom")
	}
}
