package xenapi

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"sync"
	"testing"
	"time"
)

var reMethodName = regexp.MustCompile(`<methodName>([^<]+)</methodName>`)

// fakeXAPI answers each call with respond(method) and records the methods
// called.
type fakeXAPI struct {
	mu      sync.Mutex
	methods []string
}

func newFakeXAPI(t *testing.T, respond func(method string) string) (*Client, *fakeXAPI) {
	t.Helper()
	f := &fakeXAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		method := reMethodName.FindStringSubmatch(string(body))[1]
		f.mu.Lock()
		f.methods = append(f.methods, method)
		f.mu.Unlock()
		fmt.Fprint(w, respond(method))
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return client, f
}

func (f *fakeXAPI) called(method string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Contains(f.methods, method)
}

func taskRecord(status, result string, errorInfo ...string) string {
	return successResponse(`<struct>` +
		`<member><name>status</name><value>` + status + `</value></member>` +
		`<member><name>result</name><value>` + html.EscapeString(result) + `</value></member>` +
		`<member><name>error_info</name><value>` + stringArray(errorInfo...) + `</value></member>` +
		`</struct>`)
}

func TestAsyncCall(t *testing.T) {
	client, f := newFakeXAPI(t, func(string) string { return successResponse(`OpaqueRef:task`) })
	task, err := client.AsyncCall(t.Context(), "VM.clone", SessionRef("s"), VMRef("vm"), "copy")
	if err != nil || task != "OpaqueRef:task" {
		t.Fatalf("AsyncCall = %q, %v", task, err)
	}
	if !f.called("Async.VM.clone") {
		t.Errorf("methods called = %v, want Async.VM.clone", f.methods)
	}
}

func TestWaitTask(t *testing.T) {
	polls := 0
	pendingThen := func(final string) func(string) string {
		return func(method string) string {
			if method == "task.get_record" {
				polls++
				if polls < 3 {
					return taskRecord("pending", "")
				}
				return final
			}
			return successResponse(`<string></string>`)
		}
	}

	t.Run("success", func(t *testing.T) {
		polls = 0
		client, f := newFakeXAPI(t, pendingThen(taskRecord("success", "<value>OpaqueRef:new</value>")))
		result, err := client.WaitTask(t.Context(), "s", "OpaqueRef:task")
		if err != nil || result != "OpaqueRef:new" {
			t.Fatalf("WaitTask = %v, %v; want OpaqueRef:new", result, err)
		}
		if !f.called("task.destroy") {
			t.Error("finished task was not destroyed")
		}
	})

	t.Run("failure", func(t *testing.T) {
		polls = 0
		client, _ := newFakeXAPI(t, pendingThen(taskRecord("failure", "", "VM_BAD_POWER_STATE", "OpaqueRef:vm", "halted", "running")))
		_, err := client.WaitTask(t.Context(), "s", "OpaqueRef:task")
		var apiErr *Error
		if !errors.Is(err, ErrVMBadPowerState) || !errors.As(err, &apiErr) || len(apiErr.Params) != 3 {
			t.Fatalf("WaitTask error = %v, want ErrVMBadPowerState with 3 params", err)
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		polls = 0
		client, _ := newFakeXAPI(t, pendingThen(taskRecord("cancelled", "")))
		if _, err := client.WaitTask(t.Context(), "s", "OpaqueRef:task"); !errors.Is(err, ErrTaskCancelled) {
			t.Fatalf("WaitTask error = %v, want ErrTaskCancelled", err)
		}
	})
}

func TestWaitTaskCancelsOnContextEnd(t *testing.T) {
	client, f := newFakeXAPI(t, func(method string) string {
		if method == "task.get_record" {
			return taskRecord("pending", "")
		}
		return successResponse(`<string></string>`)
	})
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()

	_, err := client.WaitTask(ctx, "s", "OpaqueRef:task")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitTask error = %v, want context.DeadlineExceeded", err)
	}
	if !f.called("task.cancel") {
		t.Errorf("methods called = %v, want task.cancel", f.methods)
	}
}
