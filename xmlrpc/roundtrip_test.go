package xmlrpc

import (
	"reflect"
	"testing"
)

// These exercise buildValueElement (request.go) and getValue (result.go)
// together: build the XML-RPC wire representation of a value, then parse
// it back, and check the round trip preserves the value. Both sides of
// this were touched by the hang-risk/swallowed-error fix in v0.1.3, so
// this is the most direct check that fix didn't change the happy path.
func TestValueRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
	}{
		{"string", "hello"},
		{"string with special chars", `<tag> & "quoted" 'text'`},
		{"empty string", ""},
		{"int", 42},
		{"negative int", -7},
		{"bool true", true},
		{"bool false", false},
		{"struct", Struct{"name": "vm1", "count": int64(3)}}, // ints always come back as int64, see the "int" case above
		{"array", []interface{}{"a", "b", "c"}},
		{"nested struct in array", []interface{}{Struct{"k": "v"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire := buildValueElement(tc.value)

			result, err := parseValue([]byte(wire))
			if err != nil {
				t.Fatalf("parseValue(%q) failed: %v", wire, err)
			}

			switch want := tc.value.(type) {
			case int:
				// Integers always round-trip through XML-RPC's <int> as
				// int64 on the way back.
				got, ok := result.(int64)
				if !ok || got != int64(want) {
					t.Fatalf("got %#v (%T), want %d", result, result, want)
				}
			default:
				if !reflect.DeepEqual(result, tc.value) {
					t.Fatalf("got %#v, want %#v", result, tc.value)
				}
			}
		})
	}
}

func TestBuildValueElementUnsupportedTypePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic for an unsupported value type, got none")
		}
	}()

	type unsupported struct{}
	buildValueElement(unsupported{})
}

// Malformed/truncated XML must return an error, not hang - this is exactly
// the class of bug fixed in v0.1.3 (parser.Token() errors were previously
// discarded in several loops, which could spin forever on truncated input).
func TestParseValueMalformedDoesNotHang(t *testing.T) {
	cases := []string{
		`<value><struct><member><name>a</name><value><string>b</string>`, // truncated, no closing tags
		`<value><array><data><value><string>a</string>`,                  // truncated array
		`<value><struct>`, // truncated right after struct open
	}

	for _, xml := range cases {
		_, err := parseValue([]byte(xml))
		if err == nil {
			t.Fatalf("parseValue(%q): expected an error for truncated XML, got nil", xml)
		}
	}
}
