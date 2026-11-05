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

type BlobRecord struct {
	// Unique identifier/object reference
	UUID string
	// a human-readable name
	NameLabel string
	// a notes field containing human-readable description
	NameDescription string
	// Size of the binary data, in bytes
	Size int
	// True if the blob is publicly accessible
	Public bool
	// Time at which the data in the blob was last updated
	LastUpdated time.Time
	// The mime type associated with this object. Defaults to 'application/octet-stream' if the empty string is supplied
	MimeType string
}

type BlobRef string

// A placeholder for a binary blob
type BlobClass struct {
	client *Client
}

// GetAllRecords Return a map of blob references to blob records for all blobs known to the system.
func (_class BlobClass) GetAllRecords(ctx context.Context, sessionID SessionRef) (_retval map[BlobRef]BlobRecord, _err error) {
	_method := "blob.get_all_records"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBlobRefToBlobRecordMapToGo(_method+" -> ", _result.Value)
	return
}

// GetAll Return a list of all the blobs known to the system.
func (_class BlobClass) GetAll(ctx context.Context, sessionID SessionRef) (_retval []BlobRef, _err error) {
	_method := "blob.get_all"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBlobRefSetToGo(_method+" -> ", _result.Value)
	return
}

// Destroy
func (_class BlobClass) Destroy(ctx context.Context, sessionID SessionRef, self BlobRef) (_err error) {
	_method := "blob.destroy"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	return
}

// Create Create a placeholder for a binary blob
func (_class BlobClass) Create(ctx context.Context, sessionID SessionRef, mimeType string, public bool) (_retval BlobRef, _err error) {
	_method := "blob.create"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_mimeTypeArg, _err := convertStringToXen(fmt.Sprintf("%s(%s)", _method, "mime_type"), mimeType)
	if _err != nil {
		return
	}
	_publicArg, _err := convertBoolToXen(fmt.Sprintf("%s(%s)", _method, "public"), public)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _mimeTypeArg, _publicArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBlobRefToGo(_method+" -> ", _result.Value)
	return
}

// SetPublic Set the public field of the given blob.
func (_class BlobClass) SetPublic(ctx context.Context, sessionID SessionRef, self BlobRef, value bool) (_err error) {
	_method := "blob.set_public"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_valueArg, _err := convertBoolToXen(fmt.Sprintf("%s(%s)", _method, "value"), value)
	if _err != nil {
		return
	}
	_, _err = _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg, _valueArg)
	return
}

// SetNameDescription Set the name/description field of the given blob.
func (_class BlobClass) SetNameDescription(ctx context.Context, sessionID SessionRef, self BlobRef, value string) (_err error) {
	_method := "blob.set_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// SetNameLabel Set the name/label field of the given blob.
func (_class BlobClass) SetNameLabel(ctx context.Context, sessionID SessionRef, self BlobRef, value string) (_err error) {
	_method := "blob.set_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetMimeType Get the mime_type field of the given blob.
func (_class BlobClass) GetMimeType(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval string, _err error) {
	_method := "blob.get_mime_type"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetLastUpdated Get the last_updated field of the given blob.
func (_class BlobClass) GetLastUpdated(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval time.Time, _err error) {
	_method := "blob.get_last_updated"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetPublic Get the public field of the given blob.
func (_class BlobClass) GetPublic(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval bool, _err error) {
	_method := "blob.get_public"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetSize Get the size field of the given blob.
func (_class BlobClass) GetSize(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval int, _err error) {
	_method := "blob.get_size"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetNameDescription Get the name/description field of the given blob.
func (_class BlobClass) GetNameDescription(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval string, _err error) {
	_method := "blob.get_name_description"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetNameLabel Get the name/label field of the given blob.
func (_class BlobClass) GetNameLabel(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval string, _err error) {
	_method := "blob.get_name_label"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetUUID Get the uuid field of the given blob.
func (_class BlobClass) GetUUID(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval string, _err error) {
	_method := "blob.get_uuid"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
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

// GetByNameLabel Get all the blob instances with the given label.
func (_class BlobClass) GetByNameLabel(ctx context.Context, sessionID SessionRef, label string) (_retval []BlobRef, _err error) {
	_method := "blob.get_by_name_label"
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
	_retval, _err = convertBlobRefSetToGo(_method+" -> ", _result.Value)
	return
}

// GetByUUID Get a reference to the blob instance with the specified UUID.
func (_class BlobClass) GetByUUID(ctx context.Context, sessionID SessionRef, uuid string) (_retval BlobRef, _err error) {
	_method := "blob.get_by_uuid"
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
	_retval, _err = convertBlobRefToGo(_method+" -> ", _result.Value)
	return
}

// GetRecord Get a record containing the current state of the given blob.
func (_class BlobClass) GetRecord(ctx context.Context, sessionID SessionRef, self BlobRef) (_retval BlobRecord, _err error) {
	_method := "blob.get_record"
	_sessionIDArg, _err := convertSessionRefToXen(fmt.Sprintf("%s(%s)", _method, "session_id"), sessionID)
	if _err != nil {
		return
	}
	_selfArg, _err := convertBlobRefToXen(fmt.Sprintf("%s(%s)", _method, "self"), self)
	if _err != nil {
		return
	}
	_result, _err := _class.client.APICall(ctx, _method, _sessionIDArg, _selfArg)
	if _err != nil {
		return
	}
	_retval, _err = convertBlobRecordToGo(_method+" -> ", _result.Value)
	return
}
