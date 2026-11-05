package xmlrpc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// buildRequest renders method and params as a complete XML-RPC methodCall
// document.
func buildRequest(method string, params []any) (string, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><methodCall><methodName>`)
	if err := writeEscaped(&b, method); err != nil {
		return "", fmt.Errorf("xmlrpc: encoding method name: %w", err)
	}
	b.WriteString(`</methodName><params>`)
	for i, p := range params {
		b.WriteString(`<param>`)
		if err := writeValue(&b, p); err != nil {
			return "", fmt.Errorf("xmlrpc: encoding param %d: %w", i, err)
		}
		b.WriteString(`</param>`)
	}
	b.WriteString(`</params></methodCall>`)
	return b.String(), nil
}

// writeValue writes value as a complete <value>...</value> element to b.
func writeValue(b *strings.Builder, value any) error {
	b.WriteString(`<value>`)
	if err := writeValueBody(b, value); err != nil {
		return err
	}
	b.WriteString(`</value>`)
	return nil
}

func writeValueBody(b *strings.Builder, value any) error {
	switch v := value.(type) {
	case Struct:
		return writeStruct(b, v)
	case Base64:
		b.WriteString(`<base64>`)
		if err := writeEscaped(b, string(v)); err != nil {
			return err
		}
		b.WriteString(`</base64>`)
	case string:
		b.WriteString(`<string>`)
		if err := writeEscaped(b, v); err != nil {
			return err
		}
		b.WriteString(`</string>`)
	case bool:
		if v {
			b.WriteString(`<boolean>1</boolean>`)
		} else {
			b.WriteString(`<boolean>0</boolean>`)
		}
	case int, int8, int16, int32, int64:
		fmt.Fprintf(b, "<int>%d</int>", v)
	case float32:
		return writeDouble(b, float64(v), 32)
	case float64:
		return writeDouble(b, v, 64)
	case time.Time:
		b.WriteString(`<dateTime.iso8601>`)
		b.WriteString(v.Format("20060102T15:04:05"))
		b.WriteString(`</dateTime.iso8601>`)
	case nil:
		// XML-RPC has no nil; XenAPI never takes one.
		return errors.New("xmlrpc: nil is not a supported XML-RPC value")
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return fmt.Errorf("xmlrpc: unsupported value type %T", value)
		}
		return writeArray(b, rv)
	}
	return nil
}

// writeDouble writes the shortest decimal that parses back to v. XML-RPC
// doubles have no exponent and no NaN or infinity.
func writeDouble(b *strings.Builder, v float64, bitSize int) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("xmlrpc: %v is not a valid XML-RPC double", v)
	}
	b.WriteString(`<double>`)
	b.WriteString(strconv.FormatFloat(v, 'f', -1, bitSize))
	b.WriteString(`</double>`)
	return nil
}

func writeStruct(b *strings.Builder, s Struct) error {
	b.WriteString(`<struct>`)
	for name, value := range s {
		b.WriteString(`<member><name>`)
		if err := writeEscaped(b, name); err != nil {
			return err
		}
		b.WriteString(`</name>`)
		if err := writeValue(b, value); err != nil {
			return fmt.Errorf("member %q: %w", name, err)
		}
		b.WriteString(`</member>`)
	}
	b.WriteString(`</struct>`)
	return nil
}

func writeArray(b *strings.Builder, rv reflect.Value) error {
	b.WriteString(`<array><data>`)
	for i := range rv.Len() {
		if err := writeValue(b, rv.Index(i).Interface()); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	b.WriteString(`</data></array>`)
	return nil
}

func writeEscaped(b *strings.Builder, s string) error {
	return xml.EscapeText(b, []byte(s))
}
