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

type SecretRecord struct {
	// Unique identifier/object reference
	UUID string
	// the secret
	Value string
	// other_config
	OtherConfig map[string]string
}

type SecretRef string

// A secret
type SecretClass struct {
	client *Client
}

// GetAllRecords Return a map of secret references to secret records for all secrets known to the system.
func (_class SecretClass) GetAllRecords(ctx context.Context, sessionID SessionRef) (_retval map[SecretRef]SecretRecord, _err error) {
	_method := "secret.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertSecretRefToSecretRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the secrets known to the system.
func (_class SecretClass) GetAll(ctx context.Context, sessionID SessionRef) (_retval []SecretRef, _err error) {
	_method := "secret.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertSecretRefSetToGo(_method+" -> ", _result.Value)
	return
}

// RemoveFromOtherConfig Remove the given key and its corresponding value from the other_config field of the given secret.  If the key is not in that Map, then do nothing.
func (_class SecretClass) RemoveFromOtherConfig(ctx context.Context, sessionID SessionRef, self SecretRef, key string) (_err error) {
	_method := "secret.remove_from_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// AddToOtherConfig Add the given key-value pair to the other_config field of the given secret.
func (_class SecretClass) AddToOtherConfig(ctx context.Context, sessionID SessionRef, self SecretRef, key string, value string) (_err error) {
	_method := "secret.add_to_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// SetOtherConfig Set the other_config field of the given secret.
func (_class SecretClass) SetOtherConfig(ctx context.Context, sessionID SessionRef, self SecretRef, value map[string]string) (_err error) {
	_method := "secret.set_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// SetValue Set the value field of the given secret.
func (_class SecretClass) SetValue(ctx context.Context, sessionID SessionRef, self SecretRef, value string) (_err error) {
	_method := "secret.set_value"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetOtherConfig Get the other_config field of the given secret.
func (_class SecretClass) GetOtherConfig(ctx context.Context, sessionID SessionRef, self SecretRef) (_retval map[string]string, _err error) {
	_method := "secret.get_other_config"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetValue Get the value field of the given secret.
func (_class SecretClass) GetValue(ctx context.Context, sessionID SessionRef, self SecretRef) (_retval string, _err error) {
	_method := "secret.get_value"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetUUID Get the uuid field of the given secret.
func (_class SecretClass) GetUUID(ctx context.Context, sessionID SessionRef, self SecretRef) (_retval string, _err error) {
	_method := "secret.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// Destroy Destroy the specified secret instance.
func (_class SecretClass) Destroy(ctx context.Context, sessionID SessionRef, self SecretRef) (_err error) {
	_method := "secret.destroy"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	return
}

// Create Create a new secret instance, and return its handle. The constructor args are: value*, other_config (* = non-optional).
func (_class SecretClass) Create(ctx context.Context, sessionID SessionRef, args SecretRecord) (_retval SecretRef, _err error) {
	_method := "secret.create"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_argsArg, _err := convertSecretRecordToXen(fmt.Sprintf("%s(%s)", _method, "args"), args)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _argsArg)
	if _err != nil {
		return
	}
	_retval, _err = convertSecretRefToGo(_method+" -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the secret instance with the specified UUID.
func (_class SecretClass) GetByUUID(ctx context.Context, sessionID SessionRef, uuid string) (_retval SecretRef, _err error) {
	_method := "secret.get_by_uuid"
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
	_retval, _err = convertSecretRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given secret.
func (_class SecretClass) GetRecord(ctx context.Context, sessionID SessionRef, self SecretRef) (_retval SecretRecord, _err error) {
	_method := "secret.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertSecretRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertSecretRecordToGo(_method+" -> ", _result.Value)
	return
}
