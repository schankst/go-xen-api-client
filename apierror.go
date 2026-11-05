package xenapi

import "fmt"

// Error is a failure reported by XenAPI: an error code, e.g.
// ERR_HANDLE_INVALID, and the parameters XenAPI sent with it. What the
// parameters mean depends on the code; HANDLE_INVALID, for example,
// carries the class name and the invalid reference, and
// SR_BACKEND_FAILURE_* carries the storage backend's code, stdout and
// stderr.
type Error struct {
	Code   string
	Params []string
}

func (e *Error) Error() string {
	if len(e.Params) == 0 {
		return "xenapi: " + e.Code
	}
	return fmt.Sprintf("xenapi: %s %q", e.Code, e.Params)
}
