//
// This file is generated. To change the content of this file, please do not
// apply the change to this file because it will get overwritten. Instead,
// change xenapi.go and execute 'go generate'.
//

package xenapi

import (
	"fmt"
	"github.com/schankst/go-xen-api-client/xmlrpc"
	"reflect"
	"strconv"
	"time"
)

var _ = fmt.Errorf
var _ = xmlrpc.NewClient
var _ = reflect.TypeOf
var _ = strconv.Atoi
var _ = time.UTC

type RateLimitRecord struct {
	// Unique identifier/object reference
	UUID string
	// a human-readable name
	NameLabel string
	// a notes field containing human-readable description
	NameDescription string
	// The set of callers attached to this rate limiter
	Callers []CallerRef
	// Maximum tokens that the bucket can hold
	BurstSize float64
	// Tokens added to the bucket per second
	FillRate float64
}

type RateLimitRef string

// A rate limiter associated with one or more callers
type RateLimitClass struct {
	client *Client
}

// GetAllRecords Return a map of Rate_limit references to Rate_limit records for all Rate_limits known to the system.
func (_class RateLimitClass) GetAllRecords(sessionID SessionRef) (_retval map[RateLimitRef]RateLimitRecord, _err error) {
	_method := "Rate_limit.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefToRateLimitRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the Rate_limits known to the system.
func (_class RateLimitClass) GetAll(sessionID SessionRef) (_retval []RateLimitRef, _err error) {
	_method := "Rate_limit.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefSetToGo(_method+" -> ", _result.Value)
	return
}

// SetFillRate Set the fill rate of the rate limiter
func (_class RateLimitClass) SetFillRate(sessionID SessionRef, self RateLimitRef, value float64) (_err error) {
	_method := "Rate_limit.set_fill_rate"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertFloatToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// SetBurstSize Set the burst size of the rate limiter
func (_class RateLimitClass) SetBurstSize(sessionID SessionRef, self RateLimitRef, value float64) (_err error) {
	_method := "Rate_limit.set_burst_size"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertFloatToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// RemoveCaller Detach the given caller from this rate limiter
func (_class RateLimitClass) RemoveCaller(sessionID SessionRef, self RateLimitRef, caller CallerRef) (_err error) {
	_method := "Rate_limit.remove_caller"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_callerArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "caller"), caller)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _callerArg)
	return
}

// AddCaller Attach the given caller to this rate limiter. Replaces any rate limiter previously attached to the caller.
func (_class RateLimitClass) AddCaller(sessionID SessionRef, self RateLimitRef, caller CallerRef) (_err error) {
	_method := "Rate_limit.add_caller"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_callerArg, _err := convertCallerRefToXen(fmt.Sprintf("%s(%s)", _method, "caller"), caller)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _callerArg)
	return
}

// SetNameDescription Set the name/description field of the given Rate_limit.
func (_class RateLimitClass) SetNameDescription(sessionID SessionRef, self RateLimitRef, value string) (_err error) {
	_method := "Rate_limit.set_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// SetNameLabel Set the name/label field of the given Rate_limit.
func (_class RateLimitClass) SetNameLabel(sessionID SessionRef, self RateLimitRef, value string) (_err error) {
	_method := "Rate_limit.set_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// GetFillRate Get the fill_rate field of the given Rate_limit.
func (_class RateLimitClass) GetFillRate(sessionID SessionRef, self RateLimitRef) (_retval float64, _err error) {
	_method := "Rate_limit.get_fill_rate"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertFloatToGo(_method+" -> ", _result.Value)
	return
}

// GetBurstSize Get the burst_size field of the given Rate_limit.
func (_class RateLimitClass) GetBurstSize(sessionID SessionRef, self RateLimitRef) (_retval float64, _err error) {
	_method := "Rate_limit.get_burst_size"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertFloatToGo(_method+" -> ", _result.Value)
	return
}

// GetCallers Get the callers field of the given Rate_limit.
func (_class RateLimitClass) GetCallers(sessionID SessionRef, self RateLimitRef) (_retval []CallerRef, _err error) {
	_method := "Rate_limit.get_callers"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertCallerRefSetToGo(_method+" -> ", _result.Value)
	return
}

// GetNameDescription Get the name/description field of the given Rate_limit.
func (_class RateLimitClass) GetNameDescription(sessionID SessionRef, self RateLimitRef) (_retval string, _err error) {
	_method := "Rate_limit.get_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetNameLabel Get the name/label field of the given Rate_limit.
func (_class RateLimitClass) GetNameLabel(sessionID SessionRef, self RateLimitRef) (_retval string, _err error) {
	_method := "Rate_limit.get_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetUUID Get the uuid field of the given Rate_limit.
func (_class RateLimitClass) GetUUID(sessionID SessionRef, self RateLimitRef) (_retval string, _err error) {
	_method := "Rate_limit.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method+" -> ", _result.Value)
	return
}

// GetByNameLabel Get all the Rate_limit instances with the given label.
func (_class RateLimitClass) GetByNameLabel(sessionID SessionRef, label string) (_retval []RateLimitRef, _err error) {
	_method := "Rate_limit.get_by_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_labelArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "label"), label)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _labelArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefSetToGo(_method+" -> ", _result.Value)
	return
}

// Destroy Destroy the specified Rate_limit instance.
func (_class RateLimitClass) Destroy(sessionID SessionRef, self RateLimitRef) (_err error) {
	_method := "Rate_limit.destroy"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg)
	return
}

// Create Create a new Rate_limit instance, and return its handle. The constructor args are: name_label, name_description, burst_size, fill_rate (* = non-optional).
func (_class RateLimitClass) Create(sessionID SessionRef, args RateLimitRecord) (_retval RateLimitRef, _err error) {
	_method := "Rate_limit.create"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_argsArg, _err := convertRateLimitRecordToXen(fmt.Sprintf("%s(%s)", _method, "args"), args)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _argsArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefToGo(_method+" -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the Rate_limit instance with the specified UUID.
func (_class RateLimitClass) GetByUUID(sessionID SessionRef, uuid string) (_retval RateLimitRef, _err error) {
	_method := "Rate_limit.get_by_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_uuidArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "uuid"), uuid)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _uuidArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given Rate_limit.
func (_class RateLimitClass) GetRecord(sessionID SessionRef, self RateLimitRef) (_retval RateLimitRecord, _err error) {
	_method := "Rate_limit.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertRateLimitRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertRateLimitRecordToGo(_method+" -> ", _result.Value)
	return
}
