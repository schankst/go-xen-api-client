// Command genapi is the code generator behind every *_gen.go file in the
// xenapi package (run via `go generate` from the repo root). It reads
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
// that happens, generation fails with "Unsupported XenAPI type: ..." or
// similar, naming the offending type string; teaching the generator that
// new shape means adding a case to all three of goTypeForXenType,
// funcPartialForXenType, and buildConverterFunc (and usually a matching
// template constant).
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/schankst/go-xen-api-client/internal/naming"
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
	reXenEnumType   = regexp.MustCompile("^enum (\\w+)$")
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
		goType = naming.SnakeToCamel(match[1]) + "Ref"
	} else if match = reXenRecordType.FindStringSubmatch(xenType); match != nil {
		goType = naming.SnakeToCamel(match[1]) + "Record"
	} else if match = reXenEnumType.FindStringSubmatch(xenType); match != nil {
		goType = naming.SnakeToCamel(match[1])
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
		partial = naming.SnakeToCamel(match[1]) + "Ref"
	} else if match = reXenRecordType.FindStringSubmatch(xenType); match != nil {
		partial = naming.SnakeToCamel(match[1]) + "Record"
	} else if match = reXenEnumType.FindStringSubmatch(xenType); match != nil {
		partial = "Enum" + naming.SnakeToCamel(match[1])
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
// convertXenTypeFuncName("VM ref set", "ToGo") -> "convertVMRefSetToGo".
// Only compound types get a generated function; see buildConverterFunc.
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
	input = strings.ReplaceAll(input, "-", "_")
	return naming.SnakeToCamel(input)
}

// internalGoIdentifier converts an XenAPI snake_case identifier into an
// unexported, camelCase Go identifier, e.g. "session_id" -> "sessionID"
// with initialisms kept in caps (see naming.SnakeToCamel). Used for local variable and
// method-parameter names. A couple of XenAPI names collide with Go
// keywords ("type", "interface") and get renamed outright.
func internalGoIdentifier(input string) (ident string) {
	input = strings.ReplaceAll(input, "-", "_")

	// The first component of the name should be all lowercase.
	_index := strings.IndexRune(input, '_')
	if _index == -1 {
		ident = strings.ToLower(input)
	} else {
		ident = strings.ToLower(input[:_index]) + naming.SnakeToCamel(input[_index+1:])
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
func executeTemplateToString(templates *template.Template, name string, data any) (text string, err error) {
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

// lifecycleNotes returns doc comment paragraphs for l: a Deprecated:
// paragraph for removed and deprecated items, naming the XAPI release, and
// a note for items that are only prototyped. Removed items stay in the
// bindings because older hosts still support them.
func lifecycleNotes(l *xapiLifecycleInfo) string {
	if l == nil {
		return ""
	}
	last := func(transition string) *xapiLifecycle {
		for _, t := range slices.Backward(l.Transitions) {
			if t.Transition == transition {
				return t
			}
		}
		return nil
	}
	note := func(text string, t *xapiLifecycle) string {
		text = fmt.Sprintf(text, t.Release)
		if d := strings.TrimSpace(formatSingleLine(t.Description)); d != "" {
			text += " " + strings.ToUpper(d[:1]) + d[1:]
			if !strings.HasSuffix(d, ".") {
				text += "."
			}
		}
		return "\n//\n// " + text
	}
	if t := last("removed"); t != nil {
		return note("Deprecated: Removed in XAPI release %s.", t)
	}
	if t := last("deprecated"); t != nil {
		return note("Deprecated: Deprecated in XAPI release %s.", t)
	}
	if t := last("prototyped"); t != nil && l.State == "Prototyped_s" {
		return fmt.Sprintf("\n//\n// Experimental: prototyped in XAPI release %s; it may change or be removed.", t.Release)
	}
	return ""
}

// unreleased reports whether l was introduced only in an unreleased
// "-next" version, i.e. no shipped host has it.
func unreleased(l *xapiLifecycleInfo) bool {
	introduced := false
	for _, t := range l.Transitions {
		if t.Transition == "published" || t.Transition == "prototyped" {
			if !strings.HasSuffix(t.Release, "-next") {
				return false
			}
			introduced = true
		}
	}
	return introduced
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
// generated Go method returns only an error (see messageFuncTemplate).
func (result *xapiResult) IsVoid() bool {
	return result.Type() == "void"
}

// xapiError is one of the named errors a message can raise (documented in
// XenAPI, but not represented as a distinct Go type - see error.go/
// internal/cmd/generrors instead, which cover the ErrorCode constants generically).
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
// fileHeaderTemplate is prepended to every generated file. It imports
// every package a generated file may use; pruneImports drops the unused
// ones.
const fileHeaderTemplate string = `// Code generated by internal/cmd/genapi; DO NOT EDIT.

package xenapi

import (
	"cmp"
	"context"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)
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
	{{ .Description|godoc }}{{ lifecycle .Lifecycle }}
	{{ .Name|exported }} {{ .GoType }}{{ end }}
}
`

// classTypeTemplate renders a class's method-wrapper type, e.g. VMClass -
// the receiver every one of that class's generated methods hangs off of
// (see messageFuncTemplate). Holds a back-reference to the owning Client
// so its methods can call client.APICall.
const classTypeTemplate string = `
{{ .Description|godoc }}{{ lifecycle .Lifecycle }}
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
// class's Class type, e.g. VMClass.Start: parameters are converted with
// xenValue, the call goes through Client.APICall, and a result is
// converted with callAndConvert.
const messageFuncTemplate string = `
{{ .Message.Name|exported|godoc }} {{ .Message.Description|singleLine }}{{ if .Message.Errors }}
//
// Errors:{{ range .Message.Errors }}
//  {{ .Name }} - {{ .Doc }}{{ end }}{{ end }}{{ lifecycle .Message.Lifecycle }}
func (c {{ .Class.Name|exported }}Class) {{ .Message.Name|exported }}(ctx context.Context{{ range .Message.Params }}, {{ .Name|internal }} {{ .GoType }}{{ end }}) {{ if .Message.Result.IsVoid }}error{{ else }}({{ .Message.Result.GoType }}, error){{ end }} {
{{- if .Message.Result.IsVoid }}
	_, err := c.client.APICall(ctx, "{{ .Class.Name }}.{{ .Message.Name }}"{{ range .Message.Params }}, {{ xenValue .Type (.Name|internal) }}{{ end }})
	return err
{{- else }}
	return callAndConvert(ctx, c.client, "{{ .Class.Name }}.{{ .Message.Name }}", {{ .Message.Result.Type|convertToGo }}{{ range .Message.Params }}, {{ xenValue .Type (.Name|internal) }}{{ end }})
{{- end }}
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
// generated method takes a context.Context and then the resulting
// SessionRef as its first arguments.
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

// Converter templates for compound types. Leaf types (strings, refs,
// enums, ints, ...) need no generated code: they map to the generic
// helpers in convert.go (see leafConverter). Every ToGo converter has the
// signature func(any) (T, error); ToXen converters are total.

const convertSetTypeToGoFuncTemplate string = `
func {{ .FuncName }}(input any) ({{ .GoType }}, error) {
	return toSlice(input, {{ .ItemConverter }})
}
`

const convertSetTypeToXenFuncTemplate string = `
func {{ .FuncName }}(v {{ .GoType }}) []any {
	return fromSlice(v, {{ .ItemConverter }})
}
`

const convertOptionTypeToGoFuncTemplate string = `
func {{ .FuncName }}(input any) ({{ .GoType }}, error) {
	return toOption(input, {{ .InnerConverter }})
}
`

const convertMapTypeToGoFuncTemplate string = `
func {{ .FuncName }}(input any) ({{ .GoType }}, error) {
	return toMap(input, {{ .KeyConverter }}, {{ .ValueConverter }})
}
`

const convertMapTypeToXenFuncTemplate string = `
func {{ .FuncName }}(v {{ .GoType }}) xmlrpc.Struct {
	return fromMap(v, {{ .KeyConverter }}, {{ .ValueConverter }})
}
`

// A field missing from the response, or nil, keeps its zero value.
const convertRecordTypeToGoFuncTemplate string = `
func {{ .FuncName }}(input any) ({{ .GoType }}, error) {
	var r {{ .GoType }}
	s, ok := input.(xmlrpc.Struct)
	if !ok {
		return r, &conversionError{want: "struct", got: input}
	}
	err := cmp.Or({{ range .Fields }}
		field(s, {{ printf "%q" .Name }}, &r.{{ .Name|exported }}, {{ .Type|convertToGo }}),{{ end }}
	)
	return r, err
}
`

const convertRecordTypeToXenFuncTemplate string = `
func {{ .FuncName }}(r {{ .GoType }}) xmlrpc.Struct {
	return xmlrpc.Struct{ {{- range .Fields }}
		{{ printf "%q" .Name }}: {{ xenValue .Type (printf "r.%s" (.Name|exported)) }},{{ end }}
	}
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
	outDir       string
	classes      []*xapiClass
	templates    *template.Template
	converters   map[string]converterFunc
	emittedEnums map[string]bool
}

func newAPIGenerator(outDir string) apiGenerator {
	return apiGenerator{
		outDir:       outDir,
		converters:   make(map[string]converterFunc),
		emittedEnums: make(map[string]bool),
	}
}

// loadXenAPI reads and parses filename (xenapi.json) into generator.classes.
func (generator *apiGenerator) loadXenAPI(filename string) (err error) {
	xenAPI, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	if err := json.Unmarshal(xenAPI, &generator.classes); err != nil {
		return err
	}

	// Skip what only an unreleased version has; it would fail on every host.
	generator.classes = slices.DeleteFunc(generator.classes, func(c *xapiClass) bool {
		return c.Lifecycle != nil && unreleased(c.Lifecycle)
	})
	for _, c := range generator.classes {
		c.Fields = slices.DeleteFunc(c.Fields, func(f *xapiField) bool {
			return f.Lifecycle != nil && unreleased(f.Lifecycle)
		})
		c.Messages = slices.DeleteFunc(c.Messages, func(m *xapiMessage) bool {
			return m.Lifecycle != nil && unreleased(m.Lifecycle)
		})
	}
	return nil
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
		"xenValue":  generator.xenValue,
		"lifecycle": lifecycleNotes,
	})

	templateLedger := map[string]string{
		"FileHeader":                 fileHeaderTemplate,
		"EnumType":                   enumTypeTemplate,
		"RecordType":                 recordTypeTemplate,
		"ClassType":                  classTypeTemplate,
		"RefType":                    refTypeTemplate,
		"MessageFunc":                messageFuncTemplate,
		"ClientStruct":               clientStructTemplate,
		"convertSetTypeToGoFunc":     convertSetTypeToGoFuncTemplate,
		"convertSetTypeToXenFunc":    convertSetTypeToXenFuncTemplate,
		"convertOptionTypeToGoFunc":  convertOptionTypeToGoFuncTemplate,
		"convertRecordTypeToGoFunc":  convertRecordTypeToGoFuncTemplate,
		"convertRecordTypeToXenFunc": convertRecordTypeToXenFuncTemplate,
		"convertMapTypeToGoFunc":     convertMapTypeToGoFuncTemplate,
		"convertMapTypeToXenFunc":    convertMapTypeToXenFuncTemplate,
	}

	for name, value := range templateLedger {
		_, err = generator.templates.New(name).Parse(value)
		if err != nil {
			return
		}
	}

	return
}

// leafConverter returns the convert.go helper that converts a leaf type,
// e.g. toString[VMRef] or fromInt, or "" if xenType isn't a leaf.
func leafConverter(xenType string, direction string) (string, error) {
	toGo := direction == "ToGo"
	switch xenType {
	case "string", "bool", "float", "an event batch", "<class> record":
		goType, err := goTypeForXenType(xenType)
		if err != nil {
			return "", err
		}
		if toGo {
			return "toValue[" + goType + "]", nil
		}
		return "fromValue[" + goType + "]", nil
	case "int":
		if toGo {
			return "toInt", nil
		}
		return "fromInt", nil
	case "datetime":
		if toGo {
			return "toTime", nil
		}
		return "fromValue[time.Time]", nil
	}
	if reXenRefType.MatchString(xenType) || reXenEnumType.MatchString(xenType) {
		goType, err := goTypeForXenType(xenType)
		if err != nil {
			return "", err
		}
		if toGo {
			return "toString[" + goType + "]", nil
		}
		return "fromString[" + goType + "]", nil
	}
	return "", nil
}

// xenValue returns the Go expression that converts expr, of XenAPI type
// xenType, to its wire form. Conversions that are a no-op or a plain
// string conversion are written inline.
func (generator *apiGenerator) xenValue(xenType string, expr string) (string, error) {
	converter, err := generator.getOrCreateConverterFunc(xenType, "ToXen")
	if err != nil {
		return "", err
	}
	switch {
	case strings.HasPrefix(converter.name, "fromValue["):
		return expr, nil
	case strings.HasPrefix(converter.name, "fromString["):
		return "string(" + expr + ")", nil
	}
	return converter.name + "(" + expr + ")", nil
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

	args := map[string]any{
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

	args := map[string]any{
		"FuncName":       funcName,
		"GoType":         goType,
		"InnerConverter": innerConverter.name,
	}

	return executeTemplateToString(generator.templates, "convertOptionTypeToGoFunc", args)
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

	args := map[string]any{
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

	args := map[string]any{
		"FuncName":       funcName,
		"GoType":         goType,
		"KeyConverter":   keyConverter.name,
		"ValueConverter": valueConverter.name,
	}

	return executeTemplateToString(generator.templates, "convertMapType"+direction+"Func", args)
}

// buildConverterFunc is the type-shape dispatcher: given any XenAPI type
// string, it decides which build*ConverterFunc handles that shape and
// delegates to it. This is the function that fails with "Unsupported
// XenAPI type"/"Unable to build type conversion function" when the
// schema introduces a shape none of the case arms recognize - see the
// file-level comment for what to do when that happens. Order matters
// somewhat: more specific matches (exact string equality) are checked
// before the regexp-based compound-type matches.
func (generator *apiGenerator) buildConverterFunc(xenType string, direction string) (converter converterFunc, err error) {
	leaf, err := leafConverter(xenType, direction)
	if err != nil || leaf != "" {
		return converterFunc{name: leaf}, err
	}

	funcName, err := convertXenTypeFuncName(xenType, direction)
	if err != nil {
		return
	}

	var funcDefinition string
	if match := reXenOptionType.FindStringSubmatch(xenType); match != nil {
		if direction == "ToXen" {
			// An absent option is the inner type's zero value, so the
			// inner converter handles both cases.
			return generator.getOrCreateConverterFunc(match[1], direction)
		}
		funcDefinition, err = generator.buildOptionConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenSetType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildSetConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenRecordType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildRecordConverterFunc(xenType, direction, funcName, match[1])
	} else if match := reXenMapType.FindStringSubmatch(xenType); match != nil {
		funcDefinition, err = generator.buildMapConverterFunc(xenType, direction, funcName, match[1], match[2])
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

// writeGoFile renders generated source into memory via render, runs it
// through go/format, and only then writes filename. Formatting here
// rather than hand-tuning whitespace in the templates keeps the
// templates readable and makes every generated file gofmt-clean by
// construction, so a formatting check can be enforced over the whole
// repository. A format error means the templates emitted source that
// doesn't parse, which is a generator bug - it's reported with the file
// name rather than silently writing something unbuildable.
func writeGoFile(filename string, render func(w io.Writer) error) (err error) {
	var buffer bytes.Buffer

	err = render(&buffer)
	if err != nil {
		return
	}

	formatted, err := pruneImports(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("generated %s does not parse: %w", filename, err)
	}

	return os.WriteFile(filename, formatted, 0644)
}

// pruneImports drops the imports src doesn't use and formats the result.
// The file header imports everything a generated file might need.
func pruneImports(src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	used := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})

	file.Decls = slices.DeleteFunc(file.Decls, func(decl ast.Decl) bool {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			return false
		}
		gen.Specs = slices.DeleteFunc(gen.Specs, func(spec ast.Spec) bool {
			importPath, _ := strconv.Unquote(spec.(*ast.ImportSpec).Path.Value)
			return !used[path.Base(importPath)]
		})
		return len(gen.Specs) == 0
	})

	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		return nil, err
	}
	return format.Source(out.Bytes())
}

// generateClassAPI writes class's <name>_gen.go: the file header, its
// enums (skipping any already emitted under another class - see
// emittedEnums), its record type (if it has fields), its Class type, and
// one method per message.
func (generator *apiGenerator) generateClassAPI(class *xapiClass) (err error) {
	apiFilename := fmt.Sprintf("%s_gen.go", strings.ToLower(class.Name))

	return writeGoFile(filepath.Join(generator.outDir, apiFilename), func(out io.Writer) (err error) {
		err = generator.templates.ExecuteTemplate(out, "FileHeader", nil)
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

			err = generator.templates.ExecuteTemplate(out, "EnumType", enum)
			if err != nil {
				return
			}
		}

		if len(class.Fields) > 0 {
			err = generator.templates.ExecuteTemplate(out, "RecordType", class)
			if err != nil {
				return
			}
		}

		err = generator.templates.ExecuteTemplate(out, "RefType", class)
		if err != nil {
			return
		}

		err = generator.templates.ExecuteTemplate(out, "ClassType", class)
		if err != nil {
			return
		}

		for _, message := range class.Messages {

			context := map[string]any{
				"Class":   class,
				"Message": message,
			}

			err = generator.templates.ExecuteTemplate(out, "MessageFunc", context)
			if err != nil {
				return
			}

		}

		return
	})
}

// generateConverters writes convert_gen.go: every converter function
// built over the course of generating all the classes (via
// getOrCreateConverterFunc), sorted by cache key for a deterministic
// diff between regenerations. Must run after every generateClassAPI call
// - it only knows about converters that were actually requested.
func (generator *apiGenerator) generateConverters() (err error) {
	return writeGoFile(filepath.Join(generator.outDir, "convert_gen.go"), func(out io.Writer) (err error) {
		err = generator.templates.ExecuteTemplate(out, "FileHeader", nil)
		if err != nil {
			return
		}

		for _, key := range slices.Sorted(maps.Keys(generator.converters)) {
			converter := generator.converters[key]
			_, err = io.WriteString(out, converter.definition)
			if err != nil {
				return
			}
		}

		return
	})
}

// generateClient writes client_gen.go: the top-level Client struct and
// its prepClient constructor helper, one field per class.
func (generator *apiGenerator) generateClient() (err error) {
	return writeGoFile(filepath.Join(generator.outDir, "client_gen.go"), func(out io.Writer) (err error) {
		err = generator.templates.ExecuteTemplate(out, "FileHeader", nil)
		if err != nil {
			return
		}

		return generator.templates.ExecuteTemplate(out, "ClientStruct", map[string]any{
			"Classes": generator.classes,
		})
	})
}

// run is the whole generation pipeline, in order: load the schema,
// prepare templates, generate every class's file (which along the way
// populates generator.converters with whatever converters those classes
// turned out to need), then generate convert_gen.go and client_gen.go
// from what was accumulated.
func (generator *apiGenerator) run(schemaPath string) (err error) {
	err = generator.loadXenAPI(schemaPath)
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

// main generates the bindings from -schema into -out. The defaults match
// `go generate` (see the //go:generate directive in client.go), which runs
// it from the repo root.
func main() {
	schema := flag.String("schema", "xenapi.json", "path to the XenAPI schema")
	out := flag.String("out", ".", "directory to write the generated files to")
	flag.Parse()

	generator := newAPIGenerator(*out)
	if err := generator.run(*schema); err != nil {
		fmt.Fprintln(os.Stderr, "genapi:", err)
		os.Exit(1)
	}
}
