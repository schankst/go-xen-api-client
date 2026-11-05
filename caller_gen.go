//
// This file is generated. To change the content of this file, please do not
// apply the change to this file because it will get overwritten. Instead,
// change internal/cmd/genapi and run 'go generate'.
//

package xenapi

import (
	"context"
	"fmt"
	"github.com/schankst/go-xen-api-client/xmlrpc"
	"reflect"
	"strconv"
	"time"
)

var _ = context.Background
var _ = fmt.Errorf
var _ = xmlrpc.NewClient
var _ = reflect.TypeOf
var _ = strconv.Atoi
var _ = time.UTC

type CallerRecord struct {
	// Unique identifier/object reference
	UUID string
	// a human-readable name
	NameLabel string
	// a notes field containing human-readable description
	NameDescription string
	// User agent matching pattern. Empty string is a full wildcard; a trailing '*' makes the field a prefix pattern; otherwise the field is matched exactly.
	UserAgent string
	// Client IP matching pattern. Same wildcard semantics as user_agent.
	ClientIP string
	// Last time a call was received from this caller
	LastAccess time.Time
	// Groups to which this caller has been assigned
	Groups []string
	// Rate limiter attached to this caller, if any. Populated via Rate_limit.add_caller rather than set directly.
	RateLimit RateLimitRef
	// True if this caller was created automatically by the rate limiter rather than by an administrator. Auto-registered callers are subject to the max-auto-registered-callers cap.
	AutoRegistered bool
}

type CallerRef string

// XAPI caller description and rate limiting
type CallerClass struct {
	client *Client
}

// GetAllRecords Return a map of Caller references to Caller records for all Callers known to the system.
func (_class CallerClass) GetAllRecords(ctx context.Context, sessionID SessionRef) (_retval map[CallerRef]CallerRecord, _err error) {
	_method := "Caller.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefToCallerRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the Callers known to the system.
func (_class CallerClass) GetAll(ctx context.Context, sessionID SessionRef) (_retval []CallerRef, _err error) {
	_method := "Caller.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefSetToGo(_method+" -> ", _result.Value)
	return
}

// QueryAllUsage Return per-caller usage for every known caller, sorted by token use descending.
func (_class CallerClass) QueryAllUsage(ctx context.Context, sessionID SessionRef) (_retval [][]string, _err error) {
	_method := "Caller.query_all_usage"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringSetSetToGo(_method+" -> ", _result.Value)
	return
}

// QueryGroupCallCount Return number of calls made since Xapi startup by the callers in the named group.
func (_class CallerClass) QueryGroupCallCount(ctx context.Context, sessionID SessionRef, group string) (_retval int, _err error) {
	_method := "Caller.query_group_call_count"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_groupArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "group"), group)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _groupArg)
	if _err != nil {
		return
	}
	_retval, _err = convertIntToGo(_method+" -> ", _result.Value)
	return
}

// QueryGroupTokenUsage Return tokens used since Xapi startup by the callers in the named group.
func (_class CallerClass) QueryGroupTokenUsage(ctx context.Context, sessionID SessionRef, group string) (_retval float64, _err error) {
	_method := "Caller.query_group_token_usage"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_groupArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "group"), group)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _groupArg)
	if _err != nil {
		return
	}
	_retval, _err = convertFloatToGo(_method+" -> ", _result.Value)
	return
}

// QueryCallCount Return number of calls made by this caller since Xapi startup
func (_class CallerClass) QueryCallCount(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval int, _err error) {
	_method := "Caller.query_call_count"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertIntToGo(_method+" -> ", _result.Value)
	return
}

// QueryTokenUsage Return tokens used by this caller since Xapi startup
func (_class CallerClass) QueryTokenUsage(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval float64, _err error) {
	_method := "Caller.query_token_usage"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertFloatToGo(_method+" -> ", _result.Value)
	return
}

// RemoveGroup Remove a caller from a group
func (_class CallerClass) RemoveGroup(ctx context.Context, sessionID SessionRef, self CallerRef, group string) (_err error) {
	_method := "Caller.remove_group"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_groupArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "group"), group)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _groupArg)
	return
}

// AddGroup Add a caller to a group
func (_class CallerClass) AddGroup(ctx context.Context, sessionID SessionRef, self CallerRef, group string) (_err error) {
	_method := "Caller.add_group"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_groupArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "group"), group)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _groupArg)
	return
}

// SetNameDescription Set the name/description field of the given Caller.
func (_class CallerClass) SetNameDescription(ctx context.Context, sessionID SessionRef, self CallerRef, value string) (_err error) {
	_method := "Caller.set_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// SetNameLabel Set the name/label field of the given Caller.
func (_class CallerClass) SetNameLabel(ctx context.Context, sessionID SessionRef, self CallerRef, value string) (_err error) {
	_method := "Caller.set_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// GetAutoRegistered Get the auto_registered field of the given Caller.
func (_class CallerClass) GetAutoRegistered(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval bool, _err error) {
	_method := "Caller.get_auto_registered"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBoolToGo(_method+" -> ", _result.Value)
	return
}

// GetRateLimit Get the rate_limit field of the given Caller.
func (_class CallerClass) GetRateLimit(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval RateLimitRef, _err error) {
	_method := "Caller.get_rate_limit"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefToGo(_method+" -> ", _result.Value)
	return
}

// GetGroups Get the groups field of the given Caller.
func (_class CallerClass) GetGroups(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval []string, _err error) {
	_method := "Caller.get_groups"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringSetToGo(_method+" -> ", _result.Value)
	return
}

// GetLastAccess Get the last_access field of the given Caller.
func (_class CallerClass) GetLastAccess(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval time.Time, _err error) {
	_method := "Caller.get_last_access"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertTimeToGo(_method+" -> ", _result.Value)
	return
}

// GetClientIP Get the client_ip field of the given Caller.
func (_class CallerClass) GetClientIP(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval string, _err error) {
	_method := "Caller.get_client_ip"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetUserAgent Get the user_agent field of the given Caller.
func (_class CallerClass) GetUserAgent(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval string, _err error) {
	_method := "Caller.get_user_agent"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetNameDescription Get the name/description field of the given Caller.
func (_class CallerClass) GetNameDescription(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval string, _err error) {
	_method := "Caller.get_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetNameLabel Get the name/label field of the given Caller.
func (_class CallerClass) GetNameLabel(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval string, _err error) {
	_method := "Caller.get_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetUUID Get the uuid field of the given Caller.
func (_class CallerClass) GetUUID(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval string, _err error) {
	_method := "Caller.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetByNameLabel Get all the Caller instances with the given label.
func (_class CallerClass) GetByNameLabel(ctx context.Context, sessionID SessionRef, label string) (_retval []CallerRef, _err error) {
	_method := "Caller.get_by_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_labelArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "label"), label)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _labelArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefSetToGo(_method+" -> ", _result.Value)
	return
}

// Destroy Destroy the specified Caller instance.
func (_class CallerClass) Destroy(ctx context.Context, sessionID SessionRef, self CallerRef) (_err error) {
	_method := "Caller.destroy"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	return
}

// Create Create a new Caller instance, and return its handle. The constructor args are: name_label, name_description, user_agent, client_ip (* = non-optional).
func (_class CallerClass) Create(ctx context.Context, sessionID SessionRef, args CallerRecord) (_retval CallerRef, _err error) {
	_method := "Caller.create"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_argsArg, _err := convertCallerRecordToXen(fmt.Sprintf("%s(%s)", _method, "args"), args)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _argsArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefToGo(_method+" -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the Caller instance with the specified UUID.
func (_class CallerClass) GetByUUID(ctx context.Context, sessionID SessionRef, uuid string) (_retval CallerRef, _err error) {
	_method := "Caller.get_by_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_uuidArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "uuid"), uuid)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _uuidArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given Caller.
func (_class CallerClass) GetRecord(ctx context.Context, sessionID SessionRef, self CallerRef) (_retval CallerRecord, _err error) {
	_method := "Caller.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRecordToGo(_method+" -> ", _result.Value)
	return
}
