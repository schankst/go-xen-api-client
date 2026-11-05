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

type PIFMetricsRecord struct {
	// Unique identifier/object reference
	UUID string
	// Read bandwidth (KiB/s)
	IoReadKbs float64
	// Write bandwidth (KiB/s)
	IoWriteKbs float64
	// Report if the PIF got a carrier or not
	Carrier bool
	// Report vendor ID
	VendorID string
	// Report vendor name
	VendorName string
	// Report device ID
	DeviceID string
	// Report device name
	DeviceName string
	// Speed of the link in Mbit/s (if available)
	Speed int
	// Full duplex capability of the link (if available)
	Duplex bool
	// PCI bus path of the pif (if available)
	PciBusPath string
	// Time at which this information was last updated
	LastUpdated time.Time
	// additional configuration
	OtherConfig map[string]string
	// The LLDP information for the physical NIC of the corresponding PIF: the effective LLDP state (key 'state', one of enabled, disabled or blocked) and, when a neighbour is seen, the received TLVs (keys 'system_name', 'port_id' and 'port_description'). Empty for PIFs that are not managed physical NICs.
	LldpNeighbor map[string]string
}

type PIFMetricsRef string

// The metrics associated with a physical network interface
type PIFMetricsClass struct {
	client *Client
}

// GetAllRecords Return a map of PIF_metrics references to PIF_metrics records for all PIF_metrics instances known to the system.
func (_class PIFMetricsClass) GetAllRecords(ctx context.Context, sessionID SessionRef) (_retval map[PIFMetricsRef]PIFMetricsRecord, _err error) {
	_method := "PIF_metrics.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPIFMetricsRefToPIFMetricsRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the PIF_metrics instances known to the system.
func (_class PIFMetricsClass) GetAll(ctx context.Context, sessionID SessionRef) (_retval []PIFMetricsRef, _err error) {
	_method := "PIF_metrics.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPIFMetricsRefSetToGo(_method+" -> ", _result.Value)
	return
}

// RemoveFromOtherConfig Remove the given key and its corresponding value from the other_config field of the given PIF_metrics.  If the key is not in that Map, then do nothing.
func (_class PIFMetricsClass) RemoveFromOtherConfig(ctx context.Context, sessionID SessionRef, self PIFMetricsRef, key string) (_err error) {
	_method := "PIF_metrics.remove_from_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_keyArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "key"), key)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _keyArg)
	return
}

// AddToOtherConfig Add the given key-value pair to the other_config field of the given PIF_metrics.
func (_class PIFMetricsClass) AddToOtherConfig(ctx context.Context, sessionID SessionRef, self PIFMetricsRef, key string, value string) (_err error) {
	_method := "PIF_metrics.add_to_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_keyArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "key"), key)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _keyArg, _valueArg)
	return
}

// SetOtherConfig Set the other_config field of the given PIF_metrics.
func (_class PIFMetricsClass) SetOtherConfig(ctx context.Context, sessionID SessionRef, self PIFMetricsRef, value map[string]string) (_err error) {
	_method := "PIF_metrics.set_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToStringMapToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// GetLldpNeighbor Get the lldp_neighbor field of the given PIF_metrics.
func (_class PIFMetricsClass) GetLldpNeighbor(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval map[string]string, _err error) {
	_method := "PIF_metrics.get_lldp_neighbor"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToStringMapToGo(_method+" -> ", _result.Value)
	return
}

// GetOtherConfig Get the other_config field of the given PIF_metrics.
func (_class PIFMetricsClass) GetOtherConfig(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval map[string]string, _err error) {
	_method := "PIF_metrics.get_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToStringMapToGo(_method+" -> ", _result.Value)
	return
}

// GetLastUpdated Get the last_updated field of the given PIF_metrics.
func (_class PIFMetricsClass) GetLastUpdated(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval time.Time, _err error) {
	_method := "PIF_metrics.get_last_updated"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetPciBusPath Get the pci_bus_path field of the given PIF_metrics.
func (_class PIFMetricsClass) GetPciBusPath(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_pci_bus_path"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetDuplex Get the duplex field of the given PIF_metrics.
func (_class PIFMetricsClass) GetDuplex(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval bool, _err error) {
	_method := "PIF_metrics.get_duplex"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSpeed Get the speed field of the given PIF_metrics.
func (_class PIFMetricsClass) GetSpeed(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval int, _err error) {
	_method := "PIF_metrics.get_speed"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetDeviceName Get the device_name field of the given PIF_metrics.
func (_class PIFMetricsClass) GetDeviceName(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_device_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetDeviceID Get the device_id field of the given PIF_metrics.
func (_class PIFMetricsClass) GetDeviceID(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_device_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetVendorName Get the vendor_name field of the given PIF_metrics.
func (_class PIFMetricsClass) GetVendorName(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_vendor_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetVendorID Get the vendor_id field of the given PIF_metrics.
func (_class PIFMetricsClass) GetVendorID(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_vendor_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetCarrier Get the carrier field of the given PIF_metrics.
func (_class PIFMetricsClass) GetCarrier(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval bool, _err error) {
	_method := "PIF_metrics.get_carrier"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetIoWriteKbs Get the io/write_kbs field of the given PIF_metrics.
func (_class PIFMetricsClass) GetIoWriteKbs(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval float64, _err error) {
	_method := "PIF_metrics.get_io_write_kbs"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetIoReadKbs Get the io/read_kbs field of the given PIF_metrics.
func (_class PIFMetricsClass) GetIoReadKbs(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval float64, _err error) {
	_method := "PIF_metrics.get_io_read_kbs"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetUUID Get the uuid field of the given PIF_metrics.
func (_class PIFMetricsClass) GetUUID(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval string, _err error) {
	_method := "PIF_metrics.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetByUUID Get a reference to the PIF_metrics instance with the specified UUID.
func (_class PIFMetricsClass) GetByUUID(ctx context.Context, sessionID SessionRef, uuid string) (_retval PIFMetricsRef, _err error) {
	_method := "PIF_metrics.get_by_uuid"
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
	_retval, _err = convertPIFMetricsRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given PIF_metrics.
func (_class PIFMetricsClass) GetRecord(ctx context.Context, sessionID SessionRef, self PIFMetricsRef) (_retval PIFMetricsRecord, _err error) {
	_method := "PIF_metrics.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPIFMetricsRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPIFMetricsRecordToGo(_method+" -> ", _result.Value)
	return
}
