//go:build ignore
// +build ignore

// xenapi.go is the code generator behind every *_gen.go file in this
// package (run via `go generate` or `go run xenapi.go`). It reads
// xenapi.json - a machine-readable description of the XenAPI, covering
// every class, field, enum, and message (RPC method), plus each one's
// doc string - and turns it into idiomatic Go: one Go type per XenAPI
// class/record/enum, one Go method per XenAPI message, and a matching
// pair of Go<->XenAPI value converters for every distinct type shape the
// schema uses.
//
// Pipeline (see run, at the bottom): load xenapi.json into the xapi*
// structs below -> prepare the text/template set used to emit Go source
// -> for each class, write its <class>_gen.go (enums, record struct, the
// class's method wrapper type, and one function per message) -> write
// convert_gen.go (every Go<->XenAPI value converter function referenced
// along the way, deduplicated by type) -> write client_gen.go (the
// top-level Client struct tying every class together).
//
// The tricky part in practice is goTypeForXenType/funcPartialForXenType/
// buildConverterFunc: XenAPI's type strings ("VM ref", "int -> string
// map", "enum vm_operations", ...) are a small compositional grammar, not
// a fixed enum, and occasionally the schema introduces a shape this
// generator has never seen (as happened with "an event batch", "<class>
// record", and the "X option" pattern - see the case arms below). When
// that happens, generation panics with "Unsupported XenAPI type: ..." or
// similar, naming the offending type string; teaching the generator that
// new shape means adding a case to all three of goTypeForXenType,
// funcPartialForXenType, and buildConverterFunc (and usually a matching
// template constant).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/serenize/snaker"
)

// These match XenAPI's compound type grammar, e.g. "VM ref", "VM record",
// "enum vm_operations", "(string -> string) map", "VM operations set",
// "int option". A type string can nest arbitrarily (e.g. "VM ref set"),
// which is why goTypeForXenType/funcPartialForXenType/buildConverterFunc
// recurse on the captured inner type rather than handling each shape as
// a one-off.
var (
	reXenRefType    = regexp.MustCompile("^(.+?) ref$")
	reXenSetType    = regexp.MustCompile("^(.+?) set$")
	reXenRecordType = regexp.MustCompile("^(.+?) record$")
	reXenEnumType   = regexp.MustCompile("^enum (.+)$")
	reXenMapType    = regexp.MustCompile("^\\((.+?) -> (.+?)\\) map$")
	reXenOptionType = regexp.MustCompile("^(.+?) option$")
)

// goTypeForXenType maps an XenAPI type string (as it appears in
// xenapi.json - "string", "VM ref", "(string -> string) map", ...) to the
// Go type used to represent it (string, VMRef, map[string]string, ...).
// Compound shapes recurse on their captured inner type(s).
func goTypeForXenType(xenType string) (goType string, err error) {
	var match []string
	if xenType == "bool" {
		goType = "bool"
	} else if xenType == "int" {
		goType = "int"
	} else if xenType == "float" {
		goType = "float64"
	} else if xenType == "string" {
		goType = "string"
	} else if xenType == "datetime" {
		goType = "time.Time"
	} else if xenType == "an event batch" {
		// Opaque struct (token/events/valid_ref_counts) introduced by event
		// batching; not modeled as a proper record in the schema, and unused
		// by this client, so it is passed through untyped.
		goType = "xmlrpc.Struct"
	} else if xenType == "<class> record" {
		// Event.snapshot: the actual record type depends on the event's
		// class at runtime and can't be fixed statically, so it is passed
		// through untyped.
		goType = "xmlrpc.Struct"
	} else if match = reXenOptionType.FindStringSubmatch(xenType); match != nil {
		// "option" just means the value may be absent; represented with the
		// same Go type as the non-optional form (nil/zero value stands in
		// for "none").
		goType, err = goTypeForXenType(match[1])
	} else if match = reXenSetType.FindStringSubmatch(xenType); match != nil {
		var goItemType string
		goItemType, err = goTypeForXenType(match[1])
		if err != nil {
			return
		}
		goType = "[]" + goItemType
	} else if match = reXenRefType.FindStringSubmatch(xenType); match != nil {
		goType = snaker.SnakeToCamel(match[1]) + "Ref"
	} else if match = reXenRecordType.FindStringSubmatch(xenType); match != nil {
		goType = snaker.SnakeToCamel(match[1]) + "Record"
	} else if match = reXenEnumType.FindStringSubmatch(xenType); match != nil {
		goType = snaker.SnakeToCamel(match[1])
	} else if match = reXenMapType.FindStringSubmatch(xenType); match != nil {
		var goKeyType string
		goKeyType, err = goTypeForXenType(match[1])
		if err != nil {
			return
		}
		var goValueType string
		goValueType, err = goTypeForXenType(match[2])
		if err != nil {
			return
		}
		goType = "map[" + goKeyType + "]" + goValueType
	} else {
		err = fmt.Errorf("Unsupported XenAPI type: %s", xenType)
	}
	return
}

// funcPartialForXenType maps an XenAPI type string to the identifier
// fragment used to name its converter functions (see
// convertXenTypeFuncName) - e.g. "VM ref" -> "VMRef", so its converters
// are named convertVMRefToGo/convertVMRefToXen. Mirrors
// goTypeForXenType's cases; kept as a separate function (rather than
// deriving the name from the Go type string) so the two can diverge
// where the Go type alone would be ambiguous or Go-syntax-unfriendly as
// an identifier (e.g. "[]string" isn't a valid function name fragment,
// so sets get the "Set" suffix instead).
func funcPartialForXenType(xenType string) (partial string, err error) {
	var match []string
	if xenType == "bool" {
		partial = "Bool"
	} else if xenType == "int" {
		partial = "Int"
	} else if xenType == "float" {
		partial = "Float"
	} else if xenType == "string" {
		partial = "String"
	} else if xenType == "datetime" {
		partial = "Time"
	} else if xenType == "an event batch" {
		partial = "EventBatch"
	} else if xenType == "<class> record" {
		partial = "PolymorphicRecord"
	} else if match = reXenOptionType.FindStringSubmatch(xenType); match != nil {
		var innerPartial string
		innerPartial, err = funcPartialForXenType(match[1])
		if err != nil {
			return
		}
		partial = innerPartial + "Option"
	} else if match = reXenSetType.FindStringSubmatch(xenType); match != nil {
		var itemPartial string
		itemPartial, err = funcPartialForXenType(match[1])
		if err != nil {
			return
		}
		partial = itemPartial + "Set"
	} else if match = reXenRefType.FindStringSubmatch(xenType); match != nil {
		partial = snaker.SnakeToCamel(match[1]) + "Ref"
	} else if match = reXenRecordType.FindStringSubmatch(xenType); match != nil {
		partial = snaker.SnakeToCamel(match[1]) + "Record"
	} else if match = reXenEnumType.FindStringSubmatch(xenType); match != nil {
		partial = "Enum" + snaker.SnakeToCamel(match[1])
	} else if match = reXenMapType.FindStringSubmatch(xenType); match != nil {
		var keyPartial string
		keyPartial, err = funcPartialForXenType(match[1])
		if err != nil {
			return
		}
		var valuePartial string
		valuePartial, err = funcPartialForXenType(match[2])
		if err != nil {
			return
		}
		partial = keyPartial + "To" + valuePartial + "Map"
	} else {
		err = fmt.Errorf("Unsupported XenAPI type: %s", xenType)
	}
	return
}

// convertXenTypeFuncName returns the name of the Go function that
// converts xenType in the given direction ("ToGo" or "ToXen") - e.g.
// convertXenTypeFuncName("VM ref", "ToGo") -> "convertVMRefToGo". Each
// such function has the shared signature
// func(context string, input interface{}) (T, error) (ToGo) or
// func(context string, value T) (interface{}, error)-ish (ToXen); see
// buildConverterFunc for how the body is actually generated.
func convertXenTypeFuncName(xenType string, direction string) (funcName string, err error) {
	funcPartial, err := funcPartialForXenType(xenType)
	if err != nil {
		return
	}

	funcName = "convert" + funcPartial + direction
	return
}

var reBeginningOfLine = regexp.MustCompile("(?m)^")

// formatGoDoc turns an XenAPI description string into a Go doc comment by
// prefixing every line with "// ". Used by the "godoc" template function
// (see prepTemplates) to carry field/class/message descriptions from
// xenapi.json straight into the generated Go source as real doc comments.
func formatGoDoc(input string) string {
	return reBeginningOfLine.ReplaceAllString(input, "// ")
}

// formatSingleLine collapses a (possibly multi-line) description into one
// line, for use where a doc comment must stay on the function's own
// comment line (see messageFuncTemplate).
func formatSingleLine(input string) string {
	return strings.Replace(input, "\n", " ", -1)
}

// exportedGoIdentifier converts an XenAPI snake_case (or hyphenated)
// identifier into an exported Go identifier, e.g. "name_label" ->
// "NameLabel". Used for anything that becomes a public Go symbol: type
// names, struct field names, method names.
func exportedGoIdentifier(input string) string {
	input = strings.Replace(input, "-", "_", -1)
	return snaker.SnakeToCamel(input)
}

// internalGoIdentifier converts an XenAPI snake_case identifier into an
// unexported, camelCase Go identifier, e.g. "session_id" -> "sessionID"
// courtesy of snaker's initialism handling. Used for local variable and
// method-parameter names. A couple of XenAPI names collide with Go
// keywords ("type", "interface") and get renamed outright.
func internalGoIdentifier(input string) (ident string) {
	input = strings.Replace(input, "-", "_", -1)

	// The first component of the name should be all lowercase.
	_index := strings.IndexRune(input, '_')
	if _index == -1 {
		ident = strings.ToLower(input)
	} else {
		ident = strings.ToLower(input[:_index]) + snaker.SnakeToCamel(input[_index+1:])
	}

	// Rename XenAPI identifiers that conflict with Go identifiers.
	switch ident {
	case "type":
		ident = "atype"
	case "interface":
		ident = "iface"
	}

	return
}

// executeTemplateToString renders the named template against data and
// returns the result as a string, rather than writing it directly to a
// file - used wherever the generator needs to compose a snippet (e.g. one
// converter function's body) before deciding where it ultimately goes.
func executeTemplateToString(templates *template.Template, name string, data interface{}) (text string, err error) {
	var buf bytes.Buffer

	err = templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return
	}

	text = buf.String()
	return
}

// The xapi* types below are the input model: they mirror xenapi.json's own
// structure field-for-field (see loadXenAPI, which json.Unmarshals the
// whole schema straight into a []*xapiClass), not anything Go-shaped yet.
// xapiClass is the root: one per XenAPI class (VM, Host, SR, ...), each
// holding its own enums, fields (-> one Go record type), and messages
// (-> one Go method per message). Everything under "Templates for
// generated Go source" further down consumes these as template data.

// xapiLifecycle is one entry in a field/message/class's publication
// history - e.g. {Release: "rio", Transition: "published"}. Parsed but
// not currently used for anything beyond documentation.
type xapiLifecycle struct {
	Description string `json:"description"`
	Release     string `json:"release"`
	Transition  string `json:"transition"`
}

// xapiLifecycleInfo wraps a field/message/class's full lifecycle history.
// Changed from a bare array to this {state, transitions} shape at some
// point upstream; see goTypeForXenType's git history/README for context.
type xapiLifecycleInfo struct {
	State       string           `json:"state"`
	Transitions []*xapiLifecycle `json:"transitions"`
}

// xapiEnumValue is one named value of an XenAPI enum, e.g. {Name: "Halted"}
// for the vm_power_state enum.
type xapiEnumValue struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
}

// xapiEnum is one enum type declared under a class, e.g. vm_power_state
// under VM. Becomes a Go string type plus one constant per value (see
// enumTypeTemplate).
type xapiEnum struct {
	Values []*xapiEnumValue `json:"values"`
	Name   string           `json:"name"`
}

// xapiField is one record field, e.g. VM.name_label. Becomes one field
// in the class's generated Go record struct (see recordTypeTemplate).
type xapiField struct {
	Default     string             `json:"default,omitempty"`
	Lifecycle   *xapiLifecycleInfo `json:"lifecycle"`
	Tag         string             `json:"tag"`
	Qualifier   string             `json:"qualifier"`
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Name        string             `json:"name"`
}

// GoType returns the Go type this field's XenAPI type maps to (e.g.
// "string", "VMRef", "map[string]string").
func (field *xapiField) GoType() (string, error) {
	return goTypeForXenType(field.Type)
}

// xapiParam is one parameter of an XenAPI message, e.g. VM.start's "force"
// parameter.
type xapiParam struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// GoType returns the Go type this parameter's XenAPI type maps to.
func (param *xapiParam) GoType() (string, error) {
	return goTypeForXenType(param.Type)
}

// xapiResult is a message's [type, doc] pair from xenapi.json (a 2-element
// array, hence the []string rather than a struct); only the type (index
// 0) is used here.
type xapiResult []string

// Type returns the XenAPI type string of the result, e.g. "VM ref" or
// "void" for a message with no return value.
func (result *xapiResult) Type() string {
	return (*result)[0]
}

// GoType returns the Go type this result's XenAPI type maps to.
func (result *xapiResult) GoType() (string, error) {
	return goTypeForXenType(result.Type())
}

// IsVoid reports whether the message returns nothing, in which case the
// generated Go method omits the _retval return value entirely (see
// messageFuncTemplate).
func (result *xapiResult) IsVoid() bool {
	return result.Type() == "void"
}

// xapiError is one of the named errors a message can raise (documented in
// XenAPI, but not represented as a distinct Go type - see error.go/
// gen_errors.go instead, which cover the ERR_* constants generically).
type xapiError struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
}

// xapiMessage is one RPC method on a class, e.g. VM.start. Becomes one Go
// method on the class's Class struct (see messageFuncTemplate).
type xapiMessage struct {
	Implicit    bool               `json:"implicit"`
	Lifecycle   *xapiLifecycleInfo `json:"lifecycle"`
	Tag         string             `json:"tag"`
	Roles       []string           `json:"roles"`
	Errors      []*xapiError       `json:"errors"`
	Params      []*xapiParam       `json:"params"`
	Result      *xapiResult        `json:"result"`
	Description string             `json:"description"`
	Name        string             `json:"name"`
}

// xapiClass is one top-level XenAPI class, e.g. VM, Host, SR - the root
// unit of code generation: each becomes its own <name>_gen.go file
// containing that class's enums, record type, Class wrapper type, and one
// method per message (see generateClassAPI).
type xapiClass struct {
	Tag         string             `json:"tag"`
	Lifecycle   *xapiLifecycleInfo `json:"lifecycle"`
	Enums       []*xapiEnum        `json:"enums"`
	Messages    []*xapiMessage     `json:"messages"`
	Fields      []*xapiField       `json:"fields"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
}

// Templates for generated Go source, below. Each is parsed into
// generator.templates (see prepTemplates) under the name given in
// templateLedger and rendered via executeTemplateToString or directly to
// a file handle. The pipe filters used throughout (godoc, singleLine,
// exported, internal, convertToGo, convertToXen) are registered in
// prepTemplates too.
//
// The convert*Template constants come in ToGo/ToXen pairs, one pair per
// distinct *shape* of XenAPI type (simple scalar, ref, set, option,
// record, map, enum) rather than one pair per concrete type - e.g. there
// is one convertSetTypeToGoFuncTemplate that generates a correctly typed
// converter for "VM ref set", "string set", or any other "X set", by
// parameterizing on the item type's own converter (see
// buildSetConverterFunc). ToGo converters all share the signature
// func(context string, input interface{}) (T, error); ToXen converters
// take a T and return something assignable into an xmlrpc.Struct/[]any
// slot (see client.go's APICall and the xmlrpc package).

// fileHeaderTemplate is prepended to every generated file: the
// do-not-edit notice, package clause, and shared imports. The var _ =
// lines exist because not every generated file ends up using every
// import (e.g. a class with no time.Time fields won't reference "time"),
// and Go doesn't allow unused imports.
const fileHeaderTemplate string = `//
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
`

// enumTypeTemplate renders one XenAPI enum as a Go string type plus one
// constant per value (see xapiEnum).
const enumTypeTemplate string = `
type {{ .Name|exported }} string

const ({{ range .Values }}
	{{ .Doc|godoc }}
	{{ (printf "%s_%s" $.Name .Name)|exported }} {{ $.Name|exported }} = {{ printf "%q" .Name }}{{ end }}
)
`

// recordTypeTemplate renders a class's fields as a Go struct, e.g.
// VMRecord (see xapiField). Only emitted for classes that have fields
// (see generateClassAPI).
const recordTypeTemplate string = `
type {{ .Name|exported }}Record struct {{ "{" }}{{ range .Fields }}
	{{ .Description|godoc }}
	{{ .Name|exported }} {{ .GoType }}{{ end }}
}
`

// classTypeTemplate renders a class's method-wrapper type, e.g. VMClass -
// the receiver every one of that class's generated methods hangs off of
// (see messageFuncTemplate). Holds a back-reference to the owning Client
// so its methods can call client.APICall.
const classTypeTemplate string = `
{{ .Description|godoc }}
type {{ .Name|exported }}Class struct {
	client *Client
}
`

// refTypeTemplate renders a class's reference type, e.g. VMRef - the
// opaque handle XenAPI uses to identify one instance of that class,
// underneath just a string (see goTypeForXenType's " ref" case).
const refTypeTemplate string = `
type {{ .Name|exported }}Ref string
`

// messageFuncTemplate renders one XenAPI message as a Go method on its
// class's Class type, e.g. VMClass.Start. Every parameter is converted
// Go->XenAPI, the call is dispatched via Client.APICall, and (unless the
// message is void) the result is converted XenAPI->Go before returning -
// see convertToXen/convertToGo, the template functions that resolve to
// the right converter for each parameter/result's type (getOrCreateConverterFunc).
const messageFuncTemplate string = `
{{ .Message.Name|exported|godoc }} {{ .Message.Description|singleLine }}{{ if .Message.Errors }}
//
// Errors:{{ range .Message.Errors }}
//  {{ .Name }} - {{ .Doc }}{{ end }}{{ end }}
func (_class {{ .Class.Name|exported }}Class) {{ .Message.Name|exported }}({{ range $index, $param := .Message.Params }}{{ if gt $index 0 }}, {{ end }}{{ .Name|internal }} {{ .GoType }}{{ end }}) ({{ if not .Message.Result.IsVoid }}_retval {{ .Message.Result.GoType }}, {{ end }}_err error) {
	_method := "{{ .Class.Name }}.{{ .Message.Name }}"{{ range .Message.Params }}
	_{{ .Name|internal }}Arg, _err := {{ .Type|convertToXen }}(fmt.Sprintf("%s(%s)", _method, {{ printf "%q" .Name }}), {{ .Name|internal }})
	if _err != nil {
		return
	}{{ end }}
	{{ if .Message.Result.IsVoid }}_, _err = {{ else }}_result, _err :={{ end }} _class.client.APICall(_method{{ range .Message.Params }}, _{{ .Name|internal }}Arg{{ end }}){{ if not .Message.Result.IsVoid }}
	if _err != nil {
		return
	}
	_retval, _err = {{ .Message.Result.Type|convertToGo }}(_method + " -> ", _result.Value){{ end }}
	return
}
`

// clientStructTemplate renders the single top-level Client struct (one
// field per XenAPI class) and its constructor helper prepClient, called
// from the hand-written NewClient in client.go. This is the only
// template whose comment (right below) ends up as real, user-facing
// GoDoc rather than being generated per-class.
const clientStructTemplate string = `
// Client is a XenAPI client. Create one with NewClient, then log in via
// Client.Session.LoginWithPassword before calling any other method - every
// generated method takes the resulting SessionRef as its first argument.
// Each exported field (Client.VM, Client.Host, Client.SR, ...) corresponds
// to one XenAPI class and exposes its Get*/Set*/etc. methods.
type Client struct {
	rpc *xmlrpc.Client{{ range .Classes }}
	{{ .Name|exported }} {{ .Name|exported }}Class{{ end }}
}

func prepClient(rpc *xmlrpc.Client) *Client {
	var client Client
	client.rpc = rpc{{ range .Classes }}
	client.{{ .Name|exported }} = {{ .Name|exported }}Class{&client}{{ end }}
	return &client
}
`

// convertSimpleTypeToGoFuncTemplate/convertSimpleTypeToXenFuncTemplate
// handle types that need no real conversion beyond a Go type assertion:
// string, bool, float64, time.Time, and the two opaque xmlrpc.Struct
// stand-ins ("an event batch", "<class> record" - see
// buildSimpleConverterFunc's callers in buildConverterFunc).
const convertSimpleTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (value {{ .GoType }}, err error) {
	if input == nil {
		return
	}
	value, ok := input.({{ .GoType }})
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", {{ printf "%q" .GoType }}, context, reflect.TypeOf(input), input)
	}
	return
}
`

const convertSimpleTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, value {{ .GoType }}) ({{ .GoType }}, error) {
	return value, nil
}
`

// convertIntToGoFuncTemplate/convertIntToXenFuncTemplate handle XenAPI's
// "int": XML-RPC has no native integer-as-string convention here, so
// XenAPI sends/expects ints as decimal strings, hence strconv rather than
// a plain type assertion.
const convertIntToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (value int, err error) {
	strValue, ok := input.(string)
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "string", context, reflect.TypeOf(input), input)
	} else {
  	value, err = strconv.Atoi(strValue)
	}
	return
}
`

const convertIntToXenFuncTemplate string = `
func {{ .FuncName }}(context string, value int) (string, error) {
	return strconv.Itoa(value), nil
}
`

// convertTimeToGoFuncTemplate handles XenAPI's "datetime": normally the
// xmlrpc decoder already produces a time.Time from a properly wire-tagged
// <dateTime.iso8601> element, but at least one deprecated field - Event's
// "timestamp" (the schema itself marks it Deprecated_s) - has been
// observed sent as an OCaml-style float-as-string Unix timestamp instead
// (e.g. "1784931839.", note the trailing dot from OCaml's string_of_float
// on a whole number), which the xmlrpc decoder faithfully passes through
// as a Go string rather than erroring on the wire format itself. This
// tolerates that as a fallback instead of hard-failing, mirroring the enum
// tolerance elsewhere in this generator for schema/wire divergence (see
// convertEnumTypeToGoFuncTemplate). ParseFloat (not ParseInt) is
// deliberate: it accepts both that trailing-dot float form and a plain
// integer string.
const convertTimeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (value time.Time, err error) {
	if input == nil {
		return
	}
	if t, ok := input.(time.Time); ok {
		value = t
		return
	}
	if s, ok := input.(string); ok {
		if seconds, perr := strconv.ParseFloat(s, 64); perr == nil {
			wholeSeconds := int64(seconds)
			nanoseconds := int64((seconds - float64(wholeSeconds)) * 1e9)
			value = time.Unix(wholeSeconds, nanoseconds)
			return
		}
	}
	err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "time.Time", context, reflect.TypeOf(input), input)
	return
}
`

// convertRefTypeToGoFuncTemplate/convertRefTypeToXenFuncTemplate handle
// "X ref" types (VMRef, HostRef, ...) - a plain string cast, since a ref
// is just an opaque handle string as far as the wire format is concerned.
const convertRefTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (ref {{ .GoType }}, err error) {
	value, ok := input.(string)
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "string", context, reflect.TypeOf(input), input)
	} else {
		ref = {{ .GoType }}(value)
	}
	return
}
`

const convertRefTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, ref {{ .GoType }}) (string, error) {
	return string(ref), nil
}
`

// convertSetTypeToGoFuncTemplate/convertSetTypeToXenFuncTemplate handle
// "X set" types ([]T in Go) by converting each element with the item
// type's own converter (ItemConverter, resolved by buildSetConverterFunc
// via getOrCreateConverterFunc) - this is what makes sets of any type,
// including sets of sets, work without a template per concrete type.
const convertSetTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (slice {{ .GoType }}, err error) {
	set, ok := input.([]interface{})
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "[]interface{}", context, reflect.TypeOf(input), input)
		return
	}
	slice = make({{ .GoType }}, len(set))
	for index, item := range set {
		itemContext := fmt.Sprintf("%s[%d]", context, index)
		itemValue, err := {{ .ItemConverter }}(itemContext, item)
		if err != nil {
			return slice, err
		}
		slice[index] = itemValue
	}
	return
}
`

const convertSetTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, slice {{ .GoType }}) (set []interface{}, err error) {
	set = make([]interface{}, len(slice))
	for index, item := range slice {
		itemContext := fmt.Sprintf("%s[%d]", context, index)
		itemValue, err := {{ .ItemConverter }}(itemContext, item)
		if err != nil {
			return set, err
		}
		set[index] = itemValue
	}
	return
}
`

// convertOptionTypeToGoFuncTemplate/convertOptionTypeToXenFuncTemplate
// handle "X option" types (a value that may be absent): the ToGo side
// short-circuits on a nil input rather than delegating to the inner
// converter, since the inner converter isn't guaranteed to accept nil
// itself; both sides otherwise just delegate to InnerConverter (see
// buildOptionConverterFunc).
const convertOptionTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (value {{ .GoType }}, err error) {
	if input == nil {
		return
	}
	value, err = {{ .InnerConverter }}(context, input)
	return
}
`

const convertOptionTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, value {{ .GoType }}) (interface{}, error) {
	return {{ .InnerConverter }}(context, value)
}
`

// convertRecordTypeToGoFuncTemplate/convertRecordTypeToXenFuncTemplate
// handle "X record" types (VMRecord, ...): one field at a time, using
// each field's own converter, generated from the same xapiField list
// used by recordTypeTemplate to build the struct in the first place (see
// buildRecordConverterFunc, which looks the class back up by name to get
// its Fields). A field absent from the XML-RPC struct (rather than
// present-but-nil) is simply left at its Go zero value on the ToGo side.
const convertRecordTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (record {{ .GoType }}, err error) {
	rpcStruct, ok := input.(xmlrpc.Struct)
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "xmlrpc.Struct", context, reflect.TypeOf(input), input)
		return
	}{{ range .Fields }}
  {{ .Name|internal }}Value, ok := rpcStruct[{{ printf "%q" .Name }}]
	if ok && {{ .Name|internal }}Value != nil {
  	record.{{ .Name|exported }}, err = {{ .Type|convertToGo }}(fmt.Sprintf("%s.%s", context, {{ printf "%q" .Name }}), {{ .Name|internal }}Value)
		if err != nil {
			return
		}
	}{{ end }}
	return
}
`

const convertRecordTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, record {{ .GoType }}) (rpcStruct xmlrpc.Struct, err error) {{ "{\n  rpcStruct = xmlrpc.Struct{}" }}{{ range .Fields }}
  rpcStruct[{{ printf "%q" .Name }}], err = {{ .Type|convertToXen }}(fmt.Sprintf("%s.%s", context, {{ printf "%q" .Name }}), record.{{ .Name|exported }})
  if err != nil {
		return
	}{{ end }}
	return
}
`

// convertMapTypeToGoFuncTemplate/convertMapTypeToXenFuncTemplate handle
// "(K -> V) map" types (map[K]V in Go), converting both keys and values
// via their own converters (KeyConverter/ValueConverter, see
// buildMapConverterFunc) - the XML-RPC wire representation is always an
// xmlrpc.Struct (string-keyed), so non-string K still round-trips through
// a string key, just converted back to K on the way in.
const convertMapTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (goMap {{ .GoType }}, err error) {
	xenMap, ok := input.(xmlrpc.Struct)
	if !ok {
		err = fmt.Errorf("Failed to parse XenAPI response: expected Go type %s at %s but got Go type %s with value %v", "xmlrpc.Struct", context, reflect.TypeOf(input), input)
		return
	}
	goMap = make({{ .GoType }}, len(xenMap))
	for xenKey, xenValue := range xenMap {
		keyContext := fmt.Sprintf("%s[%s]", context, xenKey)
		goKey, err := {{ .KeyConverter }}(keyContext, xenKey)
		if err != nil {
			return goMap, err
		}
		goValue, err := {{ .ValueConverter }}(keyContext, xenValue)
		if err != nil {
			return goMap, err
		}
		goMap[goKey] = goValue
	}
	return
}
`

const convertMapTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, goMap {{.GoType }}) (xenMap xmlrpc.Struct, err error) {
	xenMap = make(xmlrpc.Struct)
	for goKey, goValue := range goMap {
		keyContext := fmt.Sprintf("%s[%s]", context, goKey)
		xenKey, err := {{ .KeyConverter }}(keyContext, goKey)
		if err != nil {
			return xenMap, err
		}
		xenValue, err := {{ .ValueConverter }}(keyContext, goValue)
		if err != nil {
			return xenMap, err
		}
		xenMap[xenKey] = xenValue
	}
	return
}
`

// convertEnumTypeToGoFuncTemplate/convertEnumTypeToXenFuncTemplate handle
// "enum X" types: ToXen is just a string cast, but ToGo switches over
// every known value (from the same xapiEnumValue list enumTypeTemplate
// used to emit the constants - see buildEnumConverterFunc, which looks
// the class back up by enum name to get its Values) and, in the default
// case, passes an unrecognized value through as-is (as the same named Go
// type, just not one of its declared constants) instead of erroring.
// This is what lets this fork survive schema drift instead of
// hard-failing on any enum value newer than what it was generated
// against - which is exactly what upstream's original generator did
// (hard error), and exactly what broke against a modern XCP-ng host in
// the first place (the VM operation "sysprep", added after upstream's
// last schema snapshot). Baked directly into the template rather than
// patched into convert_gen.go after the fact (as it was through v0.2.0 -
// see patch_enums.go's git history) so every regeneration gets it for
// free with no separate step.
const convertEnumTypeToGoFuncTemplate string = `
func {{ .FuncName }}(context string, input interface{}) (value {{ .GoType }}, err error) {
	strValue, err := {{ "string"|convertToGo }}(context, input)
	if err != nil {
		return
	}
  switch strValue {{ "{" }}{{ range .Values }}
    case {{ printf "%q" .Name }}:
      value = {{ $.GoType }}{{ .Name|exported }}{{ end }}
    default:
      // Unknown value from a newer XAPI version than this was generated
      // against; pass it through as-is rather than failing the whole
      // record parse.
      value = {{ $.GoType }}(strValue)
	}
	return
}
`

const convertEnumTypeToXenFuncTemplate string = `
func {{ .FuncName }}(context string, value {{ .GoType }}) (string, error) {
	return string(value), nil
}
`

// converterFunc is one generated Go<->XenAPI conversion function: its
// name (for other generated code to call) and its full source definition
// (to be written into convert_gen.go).
type converterFunc struct {
	name       string
	definition string
}

// apiGenerator holds all the state threaded through one generation run:
// the parsed schema, the compiled template set, and two caches - one
// memoizing converter functions by (type, direction) so e.g. "VM ref" is
// only ever converted once no matter how many messages use it, one
// tracking which enum names have already been emitted so a shared enum
// (declared under more than one class) doesn't get a duplicate Go type.
type apiGenerator struct {
	classes      []*xapiClass
	templates    *template.Template
	converters   map[string]converterFunc
	emittedEnums map[string]bool
}

func newAPIGenerator() apiGenerator {
	return apiGenerator{
		converters:   make(map[string]converterFunc),
		emittedEnums: make(map[string]bool),
	}
}

// loadXenAPI reads and parses filename (xenapi.json) into generator.classes.
func (generator *apiGenerator) loadXenAPI(filename string) (err error) {
	xenAPI, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	return json.Unmarshal(xenAPI, &generator.classes)
}

// prepTemplates parses every template constant above into
// generator.templates and registers the pipe filters/functions
// ("godoc", "convertToGo", ...) they use. Must run once, after loadXenAPI
// and before any generate* method.
func (generator *apiGenerator) prepTemplates() (err error) {
	generator.templates = template.New("")

	generator.templates.Funcs(template.FuncMap{
		"godoc":      formatGoDoc,
		"singleLine": formatSingleLine,
		"exported":   exportedGoIdentifier,
		"internal":   internalGoIdentifier,
		"convertToGo": func(xenType string) (string, error) {
			converter, err := generator.getOrCreateConverterFunc(xenType, "ToGo")
			if err != nil {
				return "", err
			}
			return converter.name, nil
		},
		"convertToXen": func(xenType string) (string, error) {
			converter, err := generator.getOrCreateConverterFunc(xenType, "ToXen")
			if err != nil {
				return "", err
			}
			return converter.name, nil
		},
	})

	templateLedger := map[string]string{
		"FileHeader":                 fileHeaderTemplate,
		"EnumType":                   enumTypeTemplate,
		"RecordType":                 recordTypeTemplate,
		"ClassType":                  classTypeTemplate,
		"RefType":                    refTypeTemplate,
		"MessageFunc":                messageFuncTemplate,
		"ClientStruct":               clientStructTemplate,
		"convertSimpleTypeToGoFunc":  convertSimpleTypeToGoFuncTemplate,
		"convertSimpleTypeToXenFunc": convertSimpleTypeToXenFuncTemplate,
		"convertIntToGoFunc":         convertIntToGoFuncTemplate,
		"convertIntToXenFunc":        convertIntToXenFuncTemplate,
		"convertTimeToGoFunc":        convertTimeToGoFuncTemplate,
		"convertRefTypeToGoFunc":     convertRefTypeToGoFuncTemplate,
		"convertRefTypeToXenFunc":    convertRefTypeToXenFuncTemplate,
		"convertSetTypeToGoFunc":     convertSetTypeToGoFuncTemplate,
		"convertSetTypeToXenFunc":    convertSetTypeToXenFuncTemplate,
		"convertOptionTypeToGoFunc":  convertOptionTypeToGoFuncTemplate,
		"convertOptionTypeToXenFunc": convertOptionTypeToXenFuncTemplate,
		"convertRecordTypeToGoFunc":  convertRecordTypeToGoFuncTemplate,
		"convertRecordTypeToXenFunc": convertRecordTypeToXenFuncTemplate,
		"convertMapTypeToGoFunc":     convertMapTypeToGoFuncTemplate,
		"convertMapTypeToXenFunc":    convertMapTypeToXenFuncTemplate,
		"convertEnumTypeToGoFunc":    convertEnumTypeToGoFuncTemplate,
		"convertEnumTypeToXenFunc":   convertEnumTypeToXenFuncTemplate,
	}

	for name, value := range templateLedger {
		_, err = generator.templates.New(name).Parse(value)
		if err != nil {
			return
		}
	}

	return
}

// buildSimpleConverterFunc renders convertSimpleTypeTo{Go,Xen}FuncTemplate
// for a type that converts via a plain Go type assertion/cast.
func (generator *apiGenerator) buildSimpleConverterFunc(xenType string, direction string, funcName string, goType string) (string, error) {
	args := map[string]interface{}{
		"FuncName": funcName,
		"GoType":   goType,
	}

	return executeTemplateToString(generator.templates, "convertSimpleType"+direction+"Func", args)
}

// buildIntConverterFunc renders convertIntTo{Go,Xen}FuncTemplate for
// XenAPI's "int" (a decimal string on the wire).
func (generator *apiGenerator) buildIntConverterFunc(xenType string, direction string, funcName string) (string, error) {
	args := map[string]interface{}{
		"FuncName": funcName,
	}

	return executeTemplateToString(generator.templates, "convertInt"+direction+"Func", args)
}

// buildTimeConverterFunc renders the ToGo/ToXen converter for XenAPI's
// "datetime". ToXen behaves like any other simple type (the caller already
// holds a valid time.Time), but ToGo uses the dedicated
// convertTimeToGoFuncTemplate instead, which additionally tolerates a
// plain decimal-string Unix timestamp - see that template's doc comment.
func (generator *apiGenerator) buildTimeConverterFunc(xenType string, direction string, funcName string) (string, error) {
	if direction == "ToXen" {
		return generator.buildSimpleConverterFunc(xenType, direction, funcName, "time.Time")
	}

	args := map[string]interface{}{
		"FuncName": funcName,
	}

	return executeTemplateToString(generator.templates, "convertTimeToGoFunc", args)
}

// buildRefConverterFunc renders convertRefTypeTo{Go,Xen}FuncTemplate for
// an "X ref" type. baseType (the class name the ref points to) isn't
// actually needed by the template - goTypeForXenType derives the Go ref
// type name straight from xenType - but is accepted for symmetry with
// the other build*ConverterFunc signatures.
func (generator *apiGenerator) buildRefConverterFunc(xenType string, direction string, funcName string, baseType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	args := map[string]interface{}{
		"FuncName": funcName,
		"GoType":   goType,
	}

	return executeTemplateToString(generator.templates, "convertRefType"+direction+"Func", args)
}

// buildSetConverterFunc renders convertSetTypeTo{Go,Xen}FuncTemplate for
// an "X set" type, first resolving (or building) itemType's own converter
// via getOrCreateConverterFunc so nested sets/records/etc. work.
func (generator *apiGenerator) buildSetConverterFunc(xenType string, direction string, funcName string, itemType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	itemConverter, err := generator.getOrCreateConverterFunc(itemType, direction)
	if err != nil {
		return "", err
	}

	args := map[string]interface{}{
		"FuncName":      funcName,
		"GoType":        goType,
		"ItemConverter": itemConverter.name,
	}

	return executeTemplateToString(generator.templates, "convertSetType"+direction+"Func", args)
}

// buildOptionConverterFunc renders convertOptionTypeTo{Go,Xen}FuncTemplate
// for an "X option" type, delegating the non-nil case to innerType's own
// converter.
func (generator *apiGenerator) buildOptionConverterFunc(xenType string, direction string, funcName string, innerType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	innerConverter, err := generator.getOrCreateConverterFunc(innerType, direction)
	if err != nil {
		return "", err
	}

	args := map[string]interface{}{
		"FuncName":       funcName,
		"GoType":         goType,
		"InnerConverter": innerConverter.name,
	}

	return executeTemplateToString(generator.templates, "convertOptionType"+direction+"Func", args)
}

// buildRecordConverterFunc renders convertRecordTypeTo{Go,Xen}FuncTemplate
// for an "X record" type. Looks itemType (the class name) back up in
// generator.classes to get its field list - the same list
// recordTypeTemplate used to define the struct in the first place, so the
// two always stay in sync.
func (generator *apiGenerator) buildRecordConverterFunc(xenType string, direction string, funcName string, itemType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	var fields []*xapiField
	for _, class := range generator.classes {
		if class.Name+" record" == xenType {
			fields = class.Fields
			break
		}
	}
	if len(fields) == 0 {
		return "", fmt.Errorf("Unable to find definition for XenAPI %s", xenType)
	}

	args := map[string]interface{}{
		"FuncName": funcName,
		"GoType":   goType,
		"Fields":   fields,
	}

	return executeTemplateToString(generator.templates, "convertRecordType"+direction+"Func", args)
}

// buildMapConverterFunc renders convertMapTypeTo{Go,Xen}FuncTemplate for a
// "(K -> V) map" type, resolving both keyType's and valueType's own
// converters.
func (generator *apiGenerator) buildMapConverterFunc(xenType string, direction string, funcName string, keyType string, valueType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	keyConverter, err := generator.getOrCreateConverterFunc(keyType, direction)
	if err != nil {
		return "", err
	}

	valueConverter, err := generator.getOrCreateConverterFunc(valueType, direction)
	if err != nil {
		return "", err
	}

	args := map[string]interface{}{
		"FuncName":       funcName,
		"GoType":         goType,
		"KeyConverter":   keyConverter.name,
		"ValueConverter": valueConverter.name,
	}

	return executeTemplateToString(generator.templates, "convertMapType"+direction+"Func", args)
}

// buildEnumConverterFunc renders convertEnumTypeTo{Go,Xen}FuncTemplate for
// an "enum X" type. Searches every class's Enums for one named enumType -
// enums aren't indexed by name anywhere, so this is a linear scan, run at
// most once per enum since getOrCreateConverterFunc caches the result.
func (generator *apiGenerator) buildEnumConverterFunc(xenType string, direction string, funcName string, enumType string) (string, error) {
	goType, err := goTypeForXenType(xenType)
	if err != nil {
		return "", err
	}

	var values []*xapiEnumValue
classLoop:
	for _, class := range generator.classes {
		for _, enum := range class.Enums {
			if "enum "+enum.Name == xenType {
				values = enum.Values
				break classLoop
			}
		}
	}
	if len(values) == 0 {
		return "", fmt.Errorf("Unable to find definition for XenAPI %s", xenType)
	}

	args := map[string]interface{}{
		"FuncName": funcName,
		"GoType":   goType,
		"Values":   values,
	}

	return executeTemplateToString(generator.templates, "convertEnumType"+direction+"Func", args)
}

// buildConverterFunc is the type-shape dispatcher: given any XenAPI type
// string, it decides which build*ConverterFunc handles that shape and
// delegates to it. This is the function that panics with "Unsupported
// XenAPI type"/"Unable to build type conversion function" when the
// schema introduces a shape none of the case arms recognize - see the
// file-level comment for what to do when that happens. Order matters
// somewhat: more specific matches (exact string equality) are checked
// before the regexp-based compound-type matches.
func (generator *apiGenerator) buildConverterFunc(xenType string, direction string) (converter converterFunc, err error) {
	funcName, err := convertXenTypeFuncName(xenType, direction)
	if err != nil {
		return
	}

	var funcDefinition string
	if xenType == "string" {
		funcDefinition, err = generator.buildSimpleConverterFunc(xenType, direction, funcName, "string")
	} else if xenType == "bool" {
		funcDefinition, err = generator.buildSimpleConverterFunc(xenType, direction, funcName, "bool")
	} else if xenType == "int" {
		funcDefinition, err = generator.buildIntConverterFunc(xenType, direction, funcName)
	} else if xenType == "float" {
		funcDefinition, err = generator.buildSimpleConverterFunc(xenType, direction, funcName, "float64")
	} else if xenType == "an event batch" {
		funcDefinition, err = generator.buildSimpleConverterFunc(xenType, direction, funcName, "xmlrpc.Struct")
	} else if xenType == "<class> record" {
		funcDefinition, err = generator.buildSimpleConverterFunc(xenType, direction, funcName, "xmlrpc.Struct")
	} else if xenType == "datetime" {
		funcDefinition, err = generator.buildTimeConverterFunc(xenType, direction, funcName)
	} else if match := reXenRefType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildRefConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenOptionType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildOptionConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenSetType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildSetConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenRecordType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildRecordConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenMapType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildMapConverterFunc(xenType, direction, funcName, match[1], match[2])
	} else if match := reXenEnumType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildEnumConverterFunc(xenType, direction, funcName, match[1])
	} else {
		err = fmt.Errorf("Unable to build type conversion function for XenAPI: unsupported type %q", xenType)
	}
	if err != nil {
		return
	}

	converter = converterFunc{funcName, funcDefinition}
	return
}

// getOrCreateConverterFunc returns the (xenType, direction) converter,
// building it via buildConverterFunc on first request and caching the
// result in generator.converters thereafter - the single choke point
// every build*ConverterFunc goes through when it needs a nested type's
// converter (e.g. a set's item converter), which is what makes recursive
// types (sets of sets, records containing records, ...) terminate: each
// distinct (type, direction) pair is only ever built once.
func (generator *apiGenerator) getOrCreateConverterFunc(xenType string, direction string) (converter converterFunc, err error) {
	converterKey := xenType + direction
	converter, found := generator.converters[converterKey]
	if !found {
		converter, err = generator.buildConverterFunc(xenType, direction)
		if err != nil {
			return
		}
		generator.converters[converterKey] = converter
	}
	return
}

// generateClassAPI writes class's <name>_gen.go: the file header, its
// enums (skipping any already emitted under another class - see
// emittedEnums), its record type (if it has fields), its Class type, and
// one method per message.
func (generator *apiGenerator) generateClassAPI(class *xapiClass) (err error) {
	apiFilename := fmt.Sprintf("%s_gen.go", strings.ToLower(class.Name))

	fileHandle, err := os.Create(apiFilename)
	if err != nil {
		return
	}

	defer fileHandle.Close()

	err = generator.templates.ExecuteTemplate(fileHandle, "FileHeader", nil)
	if err != nil {
		return
	}

	for _, enum := range class.Enums {
		// The same enum can be declared under more than one class (e.g. a
		// shared type module); only emit its Go type once.
		if generator.emittedEnums[enum.Name] {
			continue
		}
		generator.emittedEnums[enum.Name] = true

		err = generator.templates.ExecuteTemplate(fileHandle, "EnumType", enum)
		if err != nil {
			return
		}
	}

	if len(class.Fields) > 0 {
		err = generator.templates.ExecuteTemplate(fileHandle, "RecordType", class)
		if err != nil {
			return
		}
	}

	err = generator.templates.ExecuteTemplate(fileHandle, "RefType", class)
	if err != nil {
		return
	}

	err = generator.templates.ExecuteTemplate(fileHandle, "ClassType", class)
	if err != nil {
		return
	}

	for _, message := range class.Messages {

		context := map[string]interface{}{
			"Class":   class,
			"Message": message,
		}

		err = generator.templates.ExecuteTemplate(fileHandle, "MessageFunc", context)
		if err != nil {
			return
		}

	}

	return
}

// generateConverters writes convert_gen.go: every converter function
// built over the course of generating all the classes (via
// getOrCreateConverterFunc), sorted by cache key for a deterministic
// diff between regenerations. Must run after every generateClassAPI call
// - it only knows about converters that were actually requested.
func (generator *apiGenerator) generateConverters() (err error) {
	fileHandle, err := os.Create("convert_gen.go")
	if err != nil {
		return
	}

	defer fileHandle.Close()

	err = generator.templates.ExecuteTemplate(fileHandle, "FileHeader", nil)
	if err != nil {
		return
	}

	var keys []string
	for key := range generator.converters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		converter := generator.converters[key]
		_, err = fileHandle.WriteString(converter.definition)
		if err != nil {
			return
		}
	}

	return
}

// generateClient writes client_gen.go: the top-level Client struct and
// its prepClient constructor helper, one field per class.
func (generator *apiGenerator) generateClient() (err error) {
	fileHandle, err := os.Create("client_gen.go")
	if err != nil {
		return
	}

	defer fileHandle.Close()

	err = generator.templates.ExecuteTemplate(fileHandle, "FileHeader", nil)
	if err != nil {
		return
	}

	err = generator.templates.ExecuteTemplate(fileHandle, "ClientStruct", map[string]interface{}{
		"Classes": generator.classes,
	})

	return
}

// run is the whole generation pipeline, in order: load xenapi.json,
// prepare templates, generate every class's file (which along the way
// populates generator.converters with whatever converters those classes
// turned out to need), then generate convert_gen.go and client_gen.go
// from what was accumulated.
func (generator *apiGenerator) run() (err error) {
	err = generator.loadXenAPI("xenapi.json")
	if err != nil {
		return
	}

	err = generator.prepTemplates()
	if err != nil {
		return
	}

	for _, class := range generator.classes {
		err = generator.generateClassAPI(class)
		if err != nil {
			return
		}
	}

	err = generator.generateConverters()
	if err != nil {
		return
	}

	err = generator.generateClient()
	return
}

// main runs the generator against xenapi.json in the current directory,
// writing *_gen.go/convert_gen.go/client_gen.go there too - run this via
// `go generate` (see the //go:generate directive in client.go) or
// `go run xenapi.go` from the repo root.
func main() {
	generator := newAPIGenerator()
	err := generator.run()
	if err != nil {
		panic(err)
	}
}
