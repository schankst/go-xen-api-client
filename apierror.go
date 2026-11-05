package xenapi

import "fmt"

// ErrorCode is a XenAPI error code such as ErrHandleInvalid. It implements
// error so it can be matched with errors.Is:
//
//	if errors.Is(err, xenapi.ErrHandleInvalid) { ... }
//
// Codes this package has no constant for, e.g. from a newer host, are
// still reported; compare them as strings.
type ErrorCode string

func (c ErrorCode) Error() string { return "xenapi: " + string(c) }

// Error is a failure reported by XenAPI: an error code and the parameters
// XenAPI sent with it. What the parameters mean depends on the code;
// ErrHandleInvalid, for example, carries the class name and the invalid
// reference, and SR_BACKEND_FAILURE_* carries the storage backend's code,
// stdout and stderr.
type Error struct {
	Code   ErrorCode
	Params []string
}

func (e *Error) Error() string {
	if len(e.Params) == 0 {
		return "xenapi: " + string(e.Code)
	}
	return fmt.Sprintf("xenapi: %s %q", string(e.Code), e.Params)
}

// Is reports whether target is the ErrorCode of e.
func (e *Error) Is(target error) bool {
	code, ok := target.(ErrorCode)
	return ok && code == e.Code
}
