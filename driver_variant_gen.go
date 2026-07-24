//
// This file is generated. To change the content of this file, please do not
// apply the change to this file because it will get overwritten. Instead,
// change xenapi.go and execute 'go generate'.
//

package xenapi

import (
	"fmt"
	"github.com/amfranz/go-xmlrpc-client"
	"reflect"
	"strconv"
	"time"
)

var _ = fmt.Errorf
var _ = xmlrpc.NewClient
var _ = reflect.TypeOf
var _ = strconv.Atoi
var _ = time.UTC

type DriverVariantRecord struct {
	// Unique identifier/object reference
	UUID string
	// Name identifying the driver variant within the driver
	Name string
	// Driver this variant is a part of
	Driver HostDriverRef
	// Unique version of this driver variant
	Version string
	// True if the hardware for this variant is present on the host
	HardwarePresent bool
	// Priority; this needs an explanation how this is ordered
	Priority float64
	// Development and release status of this variant, like 'alpha'
	Status string
}

type DriverVariantRef string

// UNSUPPORTED. Variant of a host driver
type DriverVariantClass struct {
	client *Client
}

// GetAllRecords Return a map of Driver_variant references to Driver_variant records for all Driver_variants known to the system.
func (_class DriverVariantClass) GetAllRecords(sessionID SessionRef) (_retval map[DriverVariantRef]DriverVariantRecord, _err error) {
	_method := "Driver_variant.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRefToDriverVariantRecordMapToGo(_method + " -> ", _result.Value)
	return
}

// GetAll Return a list of all the Driver_variants known to the system.
func (_class DriverVariantClass) GetAll(sessionID SessionRef) (_retval []DriverVariantRef, _err error) {
	_method := "Driver_variant.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRefSetToGo(_method + " -> ", _result.Value)
	return
}

// Select UNSUPPORTED Select this variant of a driver to become active after reboot or immediately if currently no version is active
func (_class DriverVariantClass) Select(sessionID SessionRef, self DriverVariantRef) (_err error) {
	_method := "Driver_variant.select"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err =  _class.client.APICall(_method, _sessionIDArg, _selfArg)
	return
}

// GetStatus Get the status field of the given Driver_variant.
func (_class DriverVariantClass) GetStatus(sessionID SessionRef, self DriverVariantRef) (_retval string, _err error) {
	_method := "Driver_variant.get_status"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method + " -> ", _result.Value)
	return
}

// GetPriority Get the priority field of the given Driver_variant.
func (_class DriverVariantClass) GetPriority(sessionID SessionRef, self DriverVariantRef) (_retval float64, _err error) {
	_method := "Driver_variant.get_priority"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertFloatToGo(_method + " -> ", _result.Value)
	return
}

// GetHardwarePresent Get the hardware_present field of the given Driver_variant.
func (_class DriverVariantClass) GetHardwarePresent(sessionID SessionRef, self DriverVariantRef) (_retval bool, _err error) {
	_method := "Driver_variant.get_hardware_present"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBoolToGo(_method + " -> ", _result.Value)
	return
}

// GetVersion Get the version field of the given Driver_variant.
func (_class DriverVariantClass) GetVersion(sessionID SessionRef, self DriverVariantRef) (_retval string, _err error) {
	_method := "Driver_variant.get_version"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method + " -> ", _result.Value)
	return
}

// GetDriver Get the driver field of the given Driver_variant.
func (_class DriverVariantClass) GetDriver(sessionID SessionRef, self DriverVariantRef) (_retval HostDriverRef, _err error) {
	_method := "Driver_variant.get_driver"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostDriverRefToGo(_method + " -> ", _result.Value)
	return
}

// GetName Get the name field of the given Driver_variant.
func (_class DriverVariantClass) GetName(sessionID SessionRef, self DriverVariantRef) (_retval string, _err error) {
	_method := "Driver_variant.get_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method + " -> ", _result.Value)
	return
}

// GetUUID Get the uuid field of the given Driver_variant.
func (_class DriverVariantClass) GetUUID(sessionID SessionRef, self DriverVariantRef) (_retval string, _err error) {
	_method := "Driver_variant.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToGo(_method + " -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the Driver_variant instance with the specified UUID.
func (_class DriverVariantClass) GetByUUID(sessionID SessionRef, uuid string) (_retval DriverVariantRef, _err error) {
	_method := "Driver_variant.get_by_uuid"
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
	_retval, _err = convertDriverVariantRefToGo(_method + " -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given Driver_variant.
func (_class DriverVariantClass) GetRecord(sessionID SessionRef, self DriverVariantRef) (_retval DriverVariantRecord, _err error) {
	_method := "Driver_variant.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRecordToGo(_method + " -> ", _result.Value)
	return
}
