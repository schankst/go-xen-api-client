package xmlrpc

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// decodeResponse parses a complete XML-RPC methodResponse document. On a
// <fault> response, the returned error is a *Error and value is nil.
//
// Every token read here is checked for an error before use - a v0.1.x bug
// in the previous (vendored) implementation of this package skipped that
// check in several places, which could spin forever re-fetching the same
// error on malformed or truncated input instead of returning it.
func decodeResponse(body []byte) (interface{}, error) {
	dec := xml.NewDecoder(strings.NewReader(string(body)))

	if _, err := findStart(dec, "methodResponse"); err != nil {
		return nil, err
	}

	root, err := nextStart(dec)
	if err != nil {
		return nil, fmt.Errorf("xmlrpc: reading methodResponse: %w", err)
	}

	switch root.Name.Local {
	case "fault":
		if _, err := findStart(dec, "value"); err != nil {
			return nil, fmt.Errorf("xmlrpc: reading fault: %w", err)
		}
		faultValue, err := decodeValue(dec)
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: decoding fault: %w", err)
		}
		fault, ok := faultValue.(Struct)
		if !ok {
			return nil, fmt.Errorf("xmlrpc: fault value is %T, want a struct", faultValue)
		}
		return nil, &Error{
			code:    fmt.Sprintf("%v", fault["faultCode"]),
			message: fmt.Sprintf("%v", fault["faultString"]),
		}
	case "params":
		if _, err := findStart(dec, "value"); err != nil {
			return nil, fmt.Errorf("xmlrpc: reading params: %w", err)
		}
		return decodeValue(dec)
	default:
		return nil, fmt.Errorf("xmlrpc: unexpected <%s> inside methodResponse", root.Name.Local)
	}
}

// findStart scans forward for the next StartElement named name, returning
// an error - never hanging - if the document ends first.
func findStart(dec *xml.Decoder, name string) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("looking for <%s>: %w", name, err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == name {
			return se, nil
		}
	}
}

// nextStart returns the next StartElement token, whatever its name.
func nextStart(dec *xml.Decoder) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, err
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se, nil
		}
	}
}

// decodeValue decodes a <value> element's content. Call sites must have
// already consumed the opening <value> StartElement (e.g. via
// findStart/nextStart); decodeValue consumes through the matching
// </value>.
func decodeValue(dec *xml.Decoder) (interface{}, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("xmlrpc: reading value: %w", err)
	}

	switch t := tok.(type) {
	case xml.EndElement:
		// Empty <value></value>.
		return "", nil

	case xml.CharData:
		// Bare text with no type wrapper defaults to <string> per the
		// XML-RPC spec.
		text := string(t)
		if err := expectEnd(dec, "value"); err != nil {
			return nil, err
		}
		return text, nil

	case xml.StartElement:
		value, err := decodeTyped(dec, t.Name.Local)
		if err != nil {
			return nil, err
		}
		if err := expectEnd(dec, "value"); err != nil {
			return nil, err
		}
		return value, nil

	default:
		return nil, fmt.Errorf("xmlrpc: unexpected token %T in value", tok)
	}
}

// decodeTyped decodes the content of a typed value element whose opening
// tag (name) was already consumed, and consumes through its own matching
// end element.
func decodeTyped(dec *xml.Decoder, name string) (interface{}, error) {
	switch name {
	case "string":
		return readText(dec, name)

	case "base64":
		text, err := readText(dec, name)
		return Base64(text), err

	case "int", "i4", "i8":
		text, err := readText(dec, name)
		if err != nil {
			return nil, err
		}
		n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: invalid <%s> value %q: %w", name, text, err)
		}
		return n, nil

	case "double":
		text, err := readText(dec, name)
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: invalid <double> value %q: %w", text, err)
		}
		return f, nil

	case "boolean":
		text, err := readText(dec, name)
		if err != nil {
			return nil, err
		}
		switch strings.TrimSpace(text) {
		case "1":
			return true, nil
		case "0":
			return false, nil
		default:
			return nil, fmt.Errorf("xmlrpc: invalid <boolean> value %q", text)
		}

	case "dateTime.iso8601":
		text, err := readText(dec, name)
		if err != nil {
			return nil, err
		}
		text = strings.TrimSpace(text)
		if t, terr := time.Parse("20060102T15:04:05", text); terr == nil {
			return t, nil
		}
		t, err := time.Parse("20060102T15:04:05Z07:00", text)
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: invalid <dateTime.iso8601> value %q", text)
		}
		return t, nil

	case "struct":
		return decodeStruct(dec)

	case "array":
		return decodeArray(dec)

	default:
		return nil, fmt.Errorf("xmlrpc: unsupported value type <%s>", name)
	}
}

// readText reads the text content of a simple leaf element and consumes
// through its matching end element.
func readText(dec *xml.Decoder, name string) (string, error) {
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", fmt.Errorf("xmlrpc: reading <%s>: %w", name, err)
		}
		switch t := tok.(type) {
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			if t.Name.Local != name {
				return "", fmt.Errorf("xmlrpc: expected </%s>, got </%s>", name, t.Name.Local)
			}
			return text.String(), nil
		}
	}
}

// expectEnd consumes tokens - tolerating only insignificant whitespace -
// until the end element named name, erroring on anything else (including
// a read error, so this can never spin).
func expectEnd(dec *xml.Decoder, name string) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("xmlrpc: expecting </%s>: %w", name, err)
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local != name {
				return fmt.Errorf("xmlrpc: expected </%s>, got </%s>", name, t.Name.Local)
			}
			return nil
		case xml.CharData:
			if len(strings.TrimSpace(string(t))) != 0 {
				return fmt.Errorf("xmlrpc: unexpected text %q before </%s>", string(t), name)
			}
		default:
			return fmt.Errorf("xmlrpc: unexpected token %T before </%s>", tok, name)
		}
	}
}

func decodeStruct(dec *xml.Decoder) (Struct, error) {
	result := Struct{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: reading struct: %w", err)
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local != "struct" {
				return nil, fmt.Errorf("xmlrpc: expected </struct>, got </%s>", t.Name.Local)
			}
			return result, nil
		case xml.StartElement:
			if t.Name.Local != "member" {
				return nil, fmt.Errorf("xmlrpc: unexpected <%s> in struct", t.Name.Local)
			}
			name, value, err := decodeMember(dec)
			if err != nil {
				return nil, err
			}
			result[name] = value
		}
	}
}

func decodeMember(dec *xml.Decoder) (name string, value interface{}, err error) {
	for {
		tok, terr := dec.Token()
		if terr != nil {
			return "", nil, fmt.Errorf("xmlrpc: reading member: %w", terr)
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local != "member" {
				return "", nil, fmt.Errorf("xmlrpc: expected </member>, got </%s>", t.Name.Local)
			}
			return name, value, nil
		case xml.StartElement:
			switch t.Name.Local {
			case "name":
				name, err = readText(dec, "name")
				if err != nil {
					return "", nil, err
				}
			case "value":
				value, err = decodeValue(dec)
				if err != nil {
					return "", nil, err
				}
			default:
				return "", nil, fmt.Errorf("xmlrpc: unexpected <%s> in member", t.Name.Local)
			}
		}
	}
}

func decodeArray(dec *xml.Decoder) ([]interface{}, error) {
	if _, err := findStart(dec, "data"); err != nil {
		return nil, fmt.Errorf("xmlrpc: reading array: %w", err)
	}
	result := []interface{}{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: reading array data: %w", err)
		}
		switch t := tok.(type) {
		case xml.EndElement:
			switch t.Name.Local {
			case "data":
				continue // next token should be </array>
			case "array":
				return result, nil
			default:
				return nil, fmt.Errorf("xmlrpc: expected </data> or </array>, got </%s>", t.Name.Local)
			}
		case xml.StartElement:
			if t.Name.Local != "value" {
				return nil, fmt.Errorf("xmlrpc: unexpected <%s> in array data", t.Name.Local)
			}
			value, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
	}
}
