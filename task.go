package xenapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)

// AsyncCall starts the XenAPI method asynchronously ("Async." + method,
// e.g. "VM.clone") and returns the task that tracks it. Parameters are
// passed as for APICall. Only methods XAPI implements asynchronously accept
// this; others fail with ErrMessageMethodUnknown. Use WaitTask for the
// outcome.
func (client *Client) AsyncCall(ctx context.Context, method string, params ...any) (TaskRef, error) {
	return callAndConvert(ctx, client, "Async."+method, toString[TaskRef], params...)
}

// WaitTask polls task until it finishes and returns its result, decoded
// from XML-RPC like an APICall result (nil for methods without one). A
// failed task is returned as *Error, a cancelled one as ErrTaskCancelled.
// Once finished, the task is destroyed.
//
// If ctx ends first, WaitTask asks XAPI to cancel the task and returns
// ctx's error.
func (client *Client) WaitTask(ctx context.Context, session SessionRef, task TaskRef) (any, error) {
	delay := 100 * time.Millisecond
	for {
		record, err := client.Task.GetRecord(ctx, session, task)
		if err != nil {
			if ctx.Err() != nil {
				return nil, client.cancelTask(ctx, session, task)
			}
			return nil, err
		}

		switch record.Status {
		case TaskStatusTypeSuccess:
			defer client.destroyTask(ctx, session, task)
			result, err := xmlrpc.DecodeValue(record.Result)
			if err != nil {
				return nil, fmt.Errorf("xenapi: task %s: decoding result: %w", task, err)
			}
			return result, nil
		case TaskStatusTypeFailure:
			defer client.destroyTask(ctx, session, task)
			if len(record.ErrorInfo) == 0 {
				return nil, fmt.Errorf("xenapi: task %s failed without error info", task)
			}
			return nil, &Error{Code: ErrorCode(record.ErrorInfo[0]), Params: record.ErrorInfo[1:]}
		case TaskStatusTypeCancelled:
			defer client.destroyTask(ctx, session, task)
			return nil, ErrTaskCancelled
		}

		select {
		case <-ctx.Done():
			return nil, client.cancelTask(ctx, session, task)
		case <-time.After(delay):
		}
		delay = min(2*delay, 2*time.Second)
	}
}

// cancelTask asks XAPI to cancel task after ctx ended and returns ctx's
// error, joined with the cancellation's if that failed too.
func (client *Client) cancelTask(ctx context.Context, session SessionRef, task TaskRef) error {
	cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := client.Task.Cancel(cctx, session, task); err != nil {
		return errors.Join(ctx.Err(), fmt.Errorf("xenapi: cancelling task %s: %w", task, err))
	}
	return ctx.Err()
}

// destroyTask removes a finished task; failure only leaves it for XAPI's
// own cleanup, so it is ignored.
func (client *Client) destroyTask(ctx context.Context, session SessionRef, task TaskRef) {
	cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_ = client.Task.Destroy(cctx, session, task)
}
