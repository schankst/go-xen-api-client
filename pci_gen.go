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

type PciDom0Access string

const (
	// dom0 can access this device as normal
	PciDom0AccessEnabled PciDom0Access = "enabled"
	// On host reboot dom0 will be blocked from accessing this device
	PciDom0AccessDisableOnReboot PciDom0Access = "disable_on_reboot"
	// dom0 cannot access this device
	PciDom0AccessDisabled PciDom0Access = "disabled"
	// On host reboot dom0 will be allowed to access this device
	PciDom0AccessEnableOnReboot PciDom0Access = "enable_on_reboot"
)

type PCIRecord struct {
	// Unique identifier/object reference
	UUID string
	// PCI class ID
	ClassID string
	// PCI class name
	ClassName string
	// Vendor ID
	VendorID string
	// Vendor name
	VendorName string
	// Device ID
	DeviceID string
	// Device name
	DeviceName string
	// Physical machine that owns the PCI device
	Host HostRef
	// PCI ID of the physical device
	PciID string
	// List of dependent PCI devices
	Dependencies []PCIRef
	// Additional configuration
	OtherConfig map[string]string
	// Subsystem vendor ID
	SubsystemVendorID string
	// Subsystem vendor name
	SubsystemVendorName string
	// Subsystem device ID
	SubsystemDeviceID string
	// Subsystem device name
	SubsystemDeviceName string
	// Driver name
	DriverName string
}

type PCIRef string

// A PCI device
type PCIClass struct {
	client *Client
}

// GetAllRecords Return a map of PCI references to PCI records for all PCIs known to the system.
func (_class PCIClass) GetAllRecords(sessionID SessionRef) (_retval map[PCIRef]PCIRecord, _err error) {
	_method := "PCI.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPCIRefToPCIRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the PCIs known to the system.
func (_class PCIClass) GetAll(sessionID SessionRef) (_retval []PCIRef, _err error) {
	_method := "PCI.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPCIRefSetToGo(_method+" -> ", _result.Value)
	return
}

// GetDom0AccessStatus Return a PCI device dom0 access status.
func (_class PCIClass) GetDom0AccessStatus(sessionID SessionRef, self PCIRef) (_retval PciDom0Access, _err error) {
	_method := "PCI.get_dom0_access_status"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertEnumPciDom0AccessToGo(_method+" -> ", _result.Value)
	return
}

// EnableDom0Access Unhide a PCI device from the dom0 kernel. (Takes affect after next boot.)
func (_class PCIClass) EnableDom0Access(sessionID SessionRef, self PCIRef) (_retval PciDom0Access, _err error) {
	_method := "PCI.enable_dom0_access"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertEnumPciDom0AccessToGo(_method+" -> ", _result.Value)
	return
}

// DisableDom0Access Hide a PCI device from the dom0 kernel. (Takes affect after next boot.)
func (_class PCIClass) DisableDom0Access(sessionID SessionRef, self PCIRef) (_retval PciDom0Access, _err error) {
	_method := "PCI.disable_dom0_access"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertEnumPciDom0AccessToGo(_method+" -> ", _result.Value)
	return
}

// RemoveFromOtherConfig Remove the given key and its corresponding value from the other_config field of the given PCI.  If the key is not in that Map, then do nothing.
func (_class PCIClass) RemoveFromOtherConfig(sessionID SessionRef, self PCIRef, key string) (_err error) {
	_method := "PCI.remove_from_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_keyArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "key"), key)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _keyArg)
	return
}

// AddToOtherConfig Add the given key-value pair to the other_config field of the given PCI.
func (_class PCIClass) AddToOtherConfig(sessionID SessionRef, self PCIRef, key string, value string) (_err error) {
	_method := "PCI.add_to_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _keyArg, _valueArg)
	return
}

// SetOtherConfig Set the other_config field of the given PCI.
func (_class PCIClass) SetOtherConfig(sessionID SessionRef, self PCIRef, value map[string]string) (_err error) {
	_method := "PCI.set_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToStringMapToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// GetDriverName Get the driver_name field of the given PCI.
func (_class PCIClass) GetDriverName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_driver_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSubsystemDeviceName Get the subsystem_device_name field of the given PCI.
func (_class PCIClass) GetSubsystemDeviceName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_subsystem_device_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSubsystemDeviceID Get the subsystem_device_id field of the given PCI.
func (_class PCIClass) GetSubsystemDeviceID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_subsystem_device_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSubsystemVendorName Get the subsystem_vendor_name field of the given PCI.
func (_class PCIClass) GetSubsystemVendorName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_subsystem_vendor_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSubsystemVendorID Get the subsystem_vendor_id field of the given PCI.
func (_class PCIClass) GetSubsystemVendorID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_subsystem_vendor_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetOtherConfig Get the other_config field of the given PCI.
func (_class PCIClass) GetOtherConfig(sessionID SessionRef, self PCIRef) (_retval map[string]string, _err error) {
	_method := "PCI.get_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertStringToStringMapToGo(_method+" -> ", _result.Value)
	return
}

// GetDependencies Get the dependencies field of the given PCI.
func (_class PCIClass) GetDependencies(sessionID SessionRef, self PCIRef) (_retval []PCIRef, _err error) {
	_method := "PCI.get_dependencies"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPCIRefSetToGo(_method+" -> ", _result.Value)
	return
}

// GetPciID Get the pci_id field of the given PCI.
func (_class PCIClass) GetPciID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_pci_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetHost Get the host field of the given PCI.
func (_class PCIClass) GetHost(sessionID SessionRef, self PCIRef) (_retval HostRef, _err error) {
	_method := "PCI.get_host"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertHostRefToGo(_method+" -> ", _result.Value)
	return
}

// GetDeviceName Get the device_name field of the given PCI.
func (_class PCIClass) GetDeviceName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_device_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetDeviceID Get the device_id field of the given PCI.
func (_class PCIClass) GetDeviceID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_device_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetVendorName Get the vendor_name field of the given PCI.
func (_class PCIClass) GetVendorName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_vendor_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetVendorID Get the vendor_id field of the given PCI.
func (_class PCIClass) GetVendorID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_vendor_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetClassName Get the class_name field of the given PCI.
func (_class PCIClass) GetClassName(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_class_name"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetClassID Get the class_id field of the given PCI.
func (_class PCIClass) GetClassID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_class_id"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetUUID Get the uuid field of the given PCI.
func (_class PCIClass) GetUUID(sessionID SessionRef, self PCIRef) (_retval string, _err error) {
	_method := "PCI.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetByUUID Get a reference to the PCI instance with the specified UUID.
func (_class PCIClass) GetByUUID(sessionID SessionRef, uuid string) (_retval PCIRef, _err error) {
	_method := "PCI.get_by_uuid"
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
	_retval, _err = convertPCIRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given PCI.
func (_class PCIClass) GetRecord(sessionID SessionRef, self PCIRef) (_retval PCIRecord, _err error) {
	_method := "PCI.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertPCIRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertPCIRecordToGo(_method+" -> ", _result.Value)
	return
}
