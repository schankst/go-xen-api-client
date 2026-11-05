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

type HostDriverRecord struct {
	// Unique identifier/object reference
	UUID string
	// Host where this driver is installed
	Host HostRef
	// Name identifying the driver uniquely
	Name string
	// Descriptive name, not used for identification
	FriendlyName string
	// Variants of this driver available for selection
	Variants []DriverVariantRef
	// Currently active variant of this driver, if any, or Null.
	ActiveVariant DriverVariantRef
	// Variant (if any) selected to become active after reboot. Or Null
	SelectedVariant DriverVariantRef
	// Device type this driver supports, like network or storage
	Type string
	// Description of the driver
	Description string
	// Information about the driver
	Info string
}

type HostDriverRef string

// UNSUPPORTED. A multi-version driver on a host
type HostDriverClass struct {
	client *Client
}

// GetAllRecords Return a map of Host_driver references to Host_driver records for all Host_drivers known to the system.
func (_class HostDriverClass) GetAllRecords(ctx context.Context, sessionID SessionRef) (_retval map[HostDriverRef]HostDriverRecord, _err error) {
	_method := "Host_driver.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostDriverRefToHostDriverRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the Host_drivers known to the system.
func (_class HostDriverClass) GetAll(ctx context.Context, sessionID SessionRef) (_retval []HostDriverRef, _err error) {
	_method := "Host_driver.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostDriverRefSetToGo(_method+" -> ", _result.Value)
	return
}

// Rescan UNSUPPORTED. Re-scan a host's drivers and update information about them. This is mostly  for trouble shooting.
func (_class HostDriverClass) Rescan(ctx context.Context, sessionID SessionRef, host HostRef) (_err error) {
	_method := "Host_driver.rescan"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_hostArg, _err := convertHostRefToXen(fmt.Sprintf("%s(%s)", _method, "host"), host)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _hostArg)
	return
}

// Deselect UNSUPPORTED. Deselect the currently active variant of this driver after reboot. No action will be taken if no variant is currently active.
func (_class HostDriverClass) Deselect(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_err error) {
	_method := "Host_driver.deselect"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	return
}

// Select UNSUPPORTED. Select a variant of the driver to become active after reboot or immediately if currently no version is active
func (_class HostDriverClass) Select(ctx context.Context, sessionID SessionRef, self HostDriverRef, variant DriverVariantRef) (_err error) {
	_method := "Host_driver.select"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_variantArg, _err := convertDriverVariantRefToXen(fmt.Sprintf("%s(%s)", _method, "variant"), variant)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _variantArg)
	return
}

// GetInfo Get the info field of the given Host_driver.
func (_class HostDriverClass) GetInfo(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_info"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetDescription Get the description field of the given Host_driver.
func (_class HostDriverClass) GetDescription(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetType Get the type field of the given Host_driver.
func (_class HostDriverClass) GetType(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_type"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSelectedVariant Get the selected_variant field of the given Host_driver.
func (_class HostDriverClass) GetSelectedVariant(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval DriverVariantRef, _err error) {
	_method := "Host_driver.get_selected_variant"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRefToGo(_method+" -> ", _result.Value)
	return
}

// GetActiveVariant Get the active_variant field of the given Host_driver.
func (_class HostDriverClass) GetActiveVariant(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval DriverVariantRef, _err error) {
	_method := "Host_driver.get_active_variant"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRefToGo(_method+" -> ", _result.Value)
	return
}

// GetVariants Get the variants field of the given Host_driver.
func (_class HostDriverClass) GetVariants(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval []DriverVariantRef, _err error) {
	_method := "Host_driver.get_variants"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertDriverVariantRefSetToGo(_method+" -> ", _result.Value)
	return
}

// GetFriendlyName Get the friendly_name field of the given Host_driver.
func (_class HostDriverClass) GetFriendlyName(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_friendly_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetName Get the name field of the given Host_driver.
func (_class HostDriverClass) GetName(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetHost Get the host field of the given Host_driver.
func (_class HostDriverClass) GetHost(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval HostRef, _err error) {
	_method := "Host_driver.get_host"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostRefToGo(_method+" -> ", _result.Value)
	return
}

// GetUUID Get the uuid field of the given Host_driver.
func (_class HostDriverClass) GetUUID(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval string, _err error) {
	_method := "Host_driver.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetByUUID Get a reference to the Host_driver instance with the specified UUID.
func (_class HostDriverClass) GetByUUID(ctx context.Context, sessionID SessionRef, uuid string) (_retval HostDriverRef, _err error) {
	_method := "Host_driver.get_by_uuid"
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
	_retval, _err = convertHostDriverRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given Host_driver.
func (_class HostDriverClass) GetRecord(ctx context.Context, sessionID SessionRef, self HostDriverRef) (_retval HostDriverRecord, _err error) {
	_method := "Host_driver.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertHostDriverRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostDriverRecordToGo(_method+" -> ", _result.Value)
	return
}
