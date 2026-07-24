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

type PlacementPolicy string

const (
	// Anti-affinity placement policy
	PlacementPolicyAntiAffinity PlacementPolicy = "anti_affinity"
	// Default placement policy
	PlacementPolicyNormal PlacementPolicy = "normal"
)

type VMGroupRecord struct {
	// Unique identifier/object reference
	UUID string
	// a human-readable name
	NameLabel string
	// a notes field containing human-readable description
	NameDescription string
	// The placement policy of the VM group
	Placement PlacementPolicy
	// The list of VMs associated with the group
	VMs []VMRef
}

type VMGroupRef string

// A VM group
type VMGroupClass struct {
	client *Client
}

// GetAllRecords Return a map of VM_group references to VM_group records for all VM_groups known to the system.
func (_class VMGroupClass) GetAllRecords(sessionID SessionRef) (_retval map[VMGroupRef]VMGroupRecord, _err error) {
	_method := "VM_group.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertVMGroupRefToVMGroupRecordMapToGo(_method + " -> ", _result.Value)
	return
}

// GetAll Return a list of all the VM_groups known to the system.
func (_class VMGroupClass) GetAll(sessionID SessionRef) (_retval []VMGroupRef, _err error) {
	_method := "VM_group.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertVMGroupRefSetToGo(_method + " -> ", _result.Value)
	return
}

// SetNameDescription Set the name/description field of the given VM_group.
func (_class VMGroupClass) SetNameDescription(sessionID SessionRef, self VMGroupRef, value string) (_err error) {
	_method := "VM_group.set_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err =  _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// SetNameLabel Set the name/label field of the given VM_group.
func (_class VMGroupClass) SetNameLabel(sessionID SessionRef, self VMGroupRef, value string) (_err error) {
	_method := "VM_group.set_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err =  _class.client.APICall(_method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// GetVMs Get the VMs field of the given VM_group.
func (_class VMGroupClass) GetVMs(sessionID SessionRef, self VMGroupRef) (_retval []VMRef, _err error) {
	_method := "VM_group.get_VMs"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertVMRefSetToGo(_method + " -> ", _result.Value)
	return
}

// GetPlacement Get the placement field of the given VM_group.
func (_class VMGroupClass) GetPlacement(sessionID SessionRef, self VMGroupRef) (_retval PlacementPolicy, _err error) {
	_method := "VM_group.get_placement"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertEnumPlacementPolicyToGo(_method + " -> ", _result.Value)
	return
}

// GetNameDescription Get the name/description field of the given VM_group.
func (_class VMGroupClass) GetNameDescription(sessionID SessionRef, self VMGroupRef) (_retval string, _err error) {
	_method := "VM_group.get_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetNameLabel Get the name/label field of the given VM_group.
func (_class VMGroupClass) GetNameLabel(sessionID SessionRef, self VMGroupRef) (_retval string, _err error) {
	_method := "VM_group.get_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetUUID Get the uuid field of the given VM_group.
func (_class VMGroupClass) GetUUID(sessionID SessionRef, self VMGroupRef) (_retval string, _err error) {
	_method := "VM_group.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetByNameLabel Get all the VM_group instances with the given label.
func (_class VMGroupClass) GetByNameLabel(sessionID SessionRef, label string) (_retval []VMGroupRef, _err error) {
	_method := "VM_group.get_by_name_label"
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
	_retval, _err = convertVMGroupRefSetToGo(_method + " -> ", _result.Value)
	return
}

// Destroy Destroy the specified VM_group instance.
func (_class VMGroupClass) Destroy(sessionID SessionRef, self VMGroupRef) (_err error) {
	_method := "VM_group.destroy"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err =  _class.client.APICall(_method, _sessionIDArg, _selfArg)
	return
}

// Create Create a new VM_group instance, and return its handle. The constructor args are: name_label, name_description, placement (* = non-optional).
func (_class VMGroupClass) Create(sessionID SessionRef, args VMGroupRecord) (_retval VMGroupRef, _err error) {
	_method := "VM_group.create"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_argsArg, _err := convertVMGroupRecordToXen(fmt.Sprintf("%s(%s)", _method, "args"), args)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _argsArg)
	if _err != nil {
		return
	}
	_retval, _err = convertVMGroupRefToGo(_method + " -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the VM_group instance with the specified UUID.
func (_class VMGroupClass) GetByUUID(sessionID SessionRef, uuid string) (_retval VMGroupRef, _err error) {
	_method := "VM_group.get_by_uuid"
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
	_retval, _err = convertVMGroupRefToGo(_method + " -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given VM_group.
func (_class VMGroupClass) GetRecord(sessionID SessionRef, self VMGroupRef) (_retval VMGroupRecord, _err error) {
	_method := "VM_group.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertVMGroupRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(_method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertVMGroupRecordToGo(_method + " -> ", _result.Value)
	return
}
