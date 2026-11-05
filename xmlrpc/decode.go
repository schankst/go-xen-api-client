package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// decodeResponse parses a methodResponse document; a <fault> is returned
// as *Error. Every token read checks its error, so truncated input can't
// make it loop.
func decodeResponse(body []byte) (any, error) {
	dec := xml.NewDecoder(bytes.NewReader(body))

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
		faultValue, err := decodeValue(dec, 0)
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
		return decodeValue(dec, 0)
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

// maxDepth bounds value nesting so a hostile response can't exhaust the
// stack. XenAPI values nest a handful of levels.
const maxDepth = 64

// decodeValue decodes the content of a <value> whose start tag was already
// read, through its end tag.
func decodeValue(dec *xml.Decoder, depth int) (any, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("xmlrpc: values nested deeper than %d levels", maxDepth)
	}
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("xmlrpc: reading value: %w", err)
	}

	switch t := tok.(type) {
	case xml.EndElement:
		// Empty <value></value>.
		return "", nil

	case xml.CharData:
		// Untyped text is a string; whitespace before a typed element is
		// just formatting.
		text := string(t)
		next, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("xmlrpc: reading value: %w", err)
		}
		switch n := next.(type) {
		case xml.EndElement:
			if n.Name.Local != "value" {
				return nil, fmt.Errorf("xmlrpc: expected </value>, got </%s>", n.Name.Local)
			}
			return text, nil
		case xml.StartElement:
			if strings.TrimSpace(text) != "" {
				return nil, fmt.Errorf("xmlrpc: unexpected text %q before <%s>", text, n.Name.Local)
			}
			return decodeTypedValue(dec, n.Name.Local, depth)
		default:
			return nil, fmt.Errorf("xmlrpc: unexpected token %T in value", next)
		}

	case xml.StartElement:
		return decodeTypedValue(dec, t.Name.Local, depth)

	default:
		return nil, fmt.Errorf("xmlrpc: unexpected token %T in value", tok)
	}
}

// decodeTypedValue decodes a typed element and the </value> after it.
func decodeTypedValue(dec *xml.Decoder, name string, depth int) (any, error) {
	value, err := decodeTyped(dec, name, depth)
	if err != nil {
		return nil, err
	}
	if err := expectEnd(dec, "value"); err != nil {
		return nil, err
	}
	return value, nil
}

// decodeTyped decodes a typed element whose start tag was already read,
// through its end tag.
func decodeTyped(dec *xml.Decoder, name string, depth int) (any, error) {
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
		return decodeStruct(dec, depth)

	case "array":
		return decodeArray(dec, depth)

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

// expectEnd reads up to the end element name, allowing only whitespace in
// between.
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

func decodeStruct(dec *xml.Decoder, depth int) (Struct, error) {
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
			name, value, err := decodeMember(dec, depth)
			if err != nil {
				return nil, err
			}
			result[name] = value
		}
	}
}

func decodeMember(dec *xml.Decoder, depth int) (name string, value any, err error) {
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
				value, err = decodeValue(dec, depth+1)
				if err != nil {
					return "", nil, err
				}
			default:
				return "", nil, fmt.Errorf("xmlrpc: unexpected <%s> in member", t.Name.Local)
			}
		}
	}
}

func decodeArray(dec *xml.Decoder, depth int) ([]any, error) {
	if _, err := findStart(dec, "data"); err != nil {
		return nil, fmt.Errorf("xmlrpc: reading array: %w", err)
	}
	result := []any{}
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
			value, err := decodeValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
	}
}

// DecodeValue decodes a single XML-RPC <value> element, such as the result
// a XenAPI task stores. An empty string decodes to nil.
func DecodeValue(s string) (any, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	dec := xml.NewDecoder(strings.NewReader(s))
	if _, err := findStart(dec, "value"); err != nil {
		return nil, fmt.Errorf("xmlrpc: %w", err)
	}
	return decodeValue(dec, 0)
}
