package xenapi

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

// Converters between decoded XML-RPC values and Go types. The generated
// code in convert_gen.go composes these for the schema's records, sets,
// maps and options.

// A conversionError reports a response value whose type doesn't match the
// schema. path locates the value within the response; it is built up as
// the error travels out, so successful conversions never pay for it.
type conversionError struct {
	path   string
	want   string
	got    any
	detail error // optional cause, e.g. a strconv range error
}

func (e *conversionError) Error() string {
	msg := fmt.Sprintf("got %T %v, want %s", e.got, e.got, e.want)
	if e.detail != nil {
		msg += " (" + e.detail.Error() + ")"
	}
	if e.path == "" {
		return msg
	}
	return e.path + ": " + msg
}

// inside prefixes err's path with segment, e.g. "[3]" or ".name_label".
func inside(err error, segment string) error {
	if ce, ok := err.(*conversionError); ok {
		ce.path = segment + ce.path
	}
	return err
}

// callAndConvert issues method and converts its result.
func callAndConvert[T any](ctx context.Context, c *Client, method string, convert func(any) (T, error), params ...any) (T, error) {
	result, err := c.APICall(ctx, method, params...)
	if err != nil {
		var zero T
		return zero, err
	}
	value, err := convert(result.Value)
	if err != nil {
		return value, fmt.Errorf("xenapi: %s: unexpected response: %w", method, err)
	}
	return value, nil
}

// toValue asserts input to T; nil yields T's zero value.
func toValue[T any](input any) (T, error) {
	value, ok := input.(T)
	if !ok && input != nil {
		return value, &conversionError{want: fmt.Sprintf("%T", value), got: input}
	}
	return value, nil
}

// toString converts a string to a string-based type such as a Ref or an
// enum. Enum values this package doesn't know, e.g. from a newer host,
// pass through unchanged. nil yields "".
func toString[T ~string](input any) (T, error) {
	s, err := toValue[string](input)
	return T(s), err
}

// toInt parses XenAPI's 64-bit int, which travels as a decimal string.
func toInt(input any) (int64, error) {
	s, ok := input.(string)
	if !ok {
		return 0, &conversionError{want: "int64 as a decimal string", got: input}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			err = numErr.Err
		}
		return 0, &conversionError{want: "int64 as a decimal string", got: input, detail: err}
	}
	return n, nil
}

// toTime accepts an XML-RPC dateTime and, as a fallback, the Unix
// timestamp string some deprecated fields (e.g. event.timestamp) carry,
// which may have a trailing dot ("1784931839."). Timestamps outside years
// 1970-9999 are rejected. nil yields the zero time.
func toTime(input any) (time.Time, error) {
	switch v := input.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		if seconds, err := strconv.ParseFloat(v, 64); err == nil && seconds >= 0 && seconds <= maxUnixSeconds {
			whole := int64(seconds)
			return time.Unix(whole, int64((seconds-float64(whole))*1e9)), nil
		}
	}
	return time.Time{}, &conversionError{want: "time.Time", got: input}
}

// maxUnixSeconds is 9999-12-31T23:59:59Z. The comparison also rejects NaN.
const maxUnixSeconds = 253402300799

// toOption converts a value that may be absent; nil yields T's zero value.
func toOption[T any](input any, convert func(any) (T, error)) (T, error) {
	if input == nil {
		var zero T
		return zero, nil
	}
	return convert(input)
}

func toSlice[T any](input any, convert func(any) (T, error)) ([]T, error) {
	items, ok := input.([]any)
	if !ok {
		return nil, &conversionError{want: "array", got: input}
	}
	out := make([]T, len(items))
	for i, item := range items {
		v, err := convert(item)
		if err != nil {
			return nil, inside(err, "["+strconv.Itoa(i)+"]")
		}
		out[i] = v
	}
	return out, nil
}

func toMap[K comparable, V any](input any, key func(any) (K, error), value func(any) (V, error)) (map[K]V, error) {
	s, ok := input.(xmlrpc.Struct)
	if !ok {
		return nil, &conversionError{want: "struct", got: input}
	}
	out := make(map[K]V, len(s))
	for k, v := range s {
		gk, err := key(k)
		if err != nil {
			return nil, inside(err, "["+k+"]")
		}
		gv, err := value(v)
		if err != nil {
			return nil, inside(err, "["+k+"]")
		}
		out[gk] = gv
	}
	return out, nil
}

// field converts the member name of s into *dst. A missing or nil member
// leaves *dst at its zero value.
func field[T any](s xmlrpc.Struct, name string, dst *T, convert func(any) (T, error)) error {
	v, ok := s[name]
	if !ok || v == nil {
		return nil
	}
	var err error
	*dst, err = convert(v)
	if err != nil {
		return inside(err, "."+name)
	}
	return nil
}

func fromString[T ~string](v T) string { return string(v) }

func fromInt(v int64) string { return strconv.FormatInt(v, 10) }

func fromValue[T any](v T) T { return v }

func fromSlice[T, X any](s []T, convert func(T) X) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = convert(v)
	}
	return out
}

func fromMap[K comparable, V, X any](m map[K]V, key func(K) string, value func(V) X) xmlrpc.Struct {
	out := make(xmlrpc.Struct, len(m))
	for k, v := range m {
		out[key(k)] = value(v)
	}
	return out
}
