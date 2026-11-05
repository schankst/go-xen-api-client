// Command genapi generates the xenapi package's *_gen.go bindings from the
// XenAPI schema. Run it via `go generate` from the repo root, or directly
// with -schema (path to xenapi.json) and -out (output directory).
//
// If generation fails on an XenAPI type construct it doesn't know, add a
// case for it to goTypeForXenType, funcPartialForXenType and
// buildConverterFunc.
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

// XenAPI type strings nest (e.g. "VM ref set"), so these capture the inner
// type for recursion. They are anchored so leafConverter, which tests ref and
// enum first, doesn't treat e.g. "enum x set" as a leaf.
var (
	reXenRefType    = regexp.MustCompile("^(.+?) ref$")
	reXenSetType    = regexp.MustCompile("^(.+?) set$")
	reXenRecordType = regexp.MustCompile("^(.+?) record$")
	reXenEnumType   = regexp.MustCompile("^enum (\\w+)$")
	reXenMapType    = regexp.MustCompile("^\\((.+?) -> (.+?)\\) map$")
	reXenOptionType = regexp.MustCompile("^(.+?) option$")
)

// goTypeForXenType maps an XenAPI type string such as "VM ref" to its Go type.
func goTypeForXenType(xenType string) (goType string, err error) {
	var match []string
	if xenType == "bool" {
		goType = "bool"
	} else if xenType == "int" {
		goType = "int64"
	} else if xenType == "float" {
		goType = "float64"
	} else if xenType == "string" {
		goType = "string"
	} else if xenType == "datetime" {
		goType = "time.Time"
	} else if xenType == "an event batch" {
		// Not modeled as a record in the schema; passed through untyped.
		goType = "xmlrpc.Struct"
	} else if xenType == "<class> record" {
		// Event.snapshot: the record's class is only known at runtime.
		goType = "xmlrpc.Struct"
	} else if match = reXenOptionType.FindStringSubmatch(xenType); match != nil {
		// An absent option is represented by the zero value.
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

// funcPartialForXenType returns the converter-name fragment for an XenAPI
// type, e.g. "VM ref set" -> "VMRefSet", since Go type syntax can't appear
// in identifiers.
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

// convertXenTypeFuncName returns e.g. "convertVMRefSetToGo" for ("VM ref set", "ToGo").
func convertXenTypeFuncName(xenType string, direction string) (funcName string, err error) {
	funcPartial, err := funcPartialForXenType(xenType)
	if err != nil {
		return
	}

	funcName = "convert" + funcPartial + direction
	return
}

var reBeginningOfLine = regexp.MustCompile("(?m)^")

func formatGoDoc(input string) string {
	return reBeginningOfLine.ReplaceAllString(input, "// ")
}

func formatSingleLine(input string) string {
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(input)
}

func exportedGoIdentifier(input string) string {
	input = strings.ReplaceAll(input, "-", "_")
	return naming.SnakeToCamel(input)
}

// internalGoIdentifier returns an unexported Go name for an XenAPI identifier,
// renaming those that are Go keywords.
func internalGoIdentifier(input string) (ident string) {
	input = strings.ReplaceAll(input, "-", "_")

	_index := strings.IndexRune(input, '_')
	if _index == -1 {
		ident = strings.ToLower(input)
	} else {
		ident = strings.ToLower(input[:_index]) + naming.SnakeToCamel(input[_index+1:])
	}

	switch ident {
	case "type":
		ident = "atype"
	case "interface":
		ident = "iface"
	}

	return
}

func executeTemplateToString(templates *template.Template, name string, data any) (text string, err error) {
	var buf bytes.Buffer

	err = templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return
	}

	text = buf.String()
	return
}

// The xapi* types mirror the structure of xenapi.json.
type xapiLifecycle struct {
	Description string `json:"description"`
	Release     string `json:"release"`
	Transition  string `json:"transition"`
}

type xapiLifecycleInfo struct {
	State       string           `json:"state"`
	Transitions []*xapiLifecycle `json:"transitions"`
}

// lifecycleNotes returns doc comment paragraphs marking removed or deprecated
// items as Deprecated and prototyped ones as Experimental. Removed items are
// kept because older hosts still support them.
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

type xapiEnumValue struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
}

type xapiEnum struct {
	Values []*xapiEnumValue `json:"values"`
	Name   string           `json:"name"`
}

type xapiField struct {
	Default     string             `json:"default,omitempty"`
	Lifecycle   *xapiLifecycleInfo `json:"lifecycle"`
	Tag         string             `json:"tag"`
	Qualifier   string             `json:"qualifier"`
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Name        string             `json:"name"`
}

func (field *xapiField) GoType() (string, error) {
	return goTypeForXenType(field.Type)
}

type xapiParam struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (param *xapiParam) GoType() (string, error) {
	return goTypeForXenType(param.Type)
}

// xapiResult is a message's [type, doc] pair.
type xapiResult []string

// Type returns the result's XenAPI type, "void" if there is none.
func (result *xapiResult) Type() string {
	return (*result)[0]
}

func (result *xapiResult) GoType() (string, error) {
	return goTypeForXenType(result.Type())
}

func (result *xapiResult) IsVoid() bool {
	return result.Type() == "void"
}

type xapiError struct {
	Doc  string `json:"doc"`
	Name string `json:"name"`
}

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

type xapiClass struct {
	Tag         string             `json:"tag"`
	Lifecycle   *xapiLifecycleInfo `json:"lifecycle"`
	Enums       []*xapiEnum        `json:"enums"`
	Messages    []*xapiMessage     `json:"messages"`
	Fields      []*xapiField       `json:"fields"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
}

// Templates for the generated source; prepTemplates parses them and
// registers the functions they use.
//
// fileHeaderTemplate imports every package a generated file may use;
// pruneImports drops the unused ones.
const fileHeaderTemplate string = `// Code generated by internal/cmd/genapi; DO NOT EDIT.

package xenapi

import (
	"cmp"
	"context"
	"time"

	"github.com/schankst/go-xen-api-client/xmlrpc"
)
`

const enumTypeTemplate string = `
type {{ .Name|exported }} string

const ({{ range .Values }}
	{{ .Doc|godoc }}
	{{ (printf "%s_%s" $.Name .Name)|exported }} {{ $.Name|exported }} = {{ printf "%q" .Name }}{{ end }}
)
`

const recordTypeTemplate string = `
type {{ .Name|exported }}Record struct {{ "{" }}{{ range .Fields }}
	{{ .Description|godoc }}{{ lifecycle .Lifecycle }}
	{{ .Name|exported }} {{ .GoType }}{{ end }}
}
`

const classTypeTemplate string = `
{{ .Description|godoc }}{{ lifecycle .Lifecycle }}
type {{ .Name|exported }}Class struct {
	client *Client
}
`

const refTypeTemplate string = `
type {{ .Name|exported }}Ref string
`

const messageFuncTemplate string = `
{{ .Message.Name|exported|godoc }} {{ .Message.Description|singleLine }}{{ if .Message.Errors }}
//
// Errors:{{ range .Message.Errors }}
//  {{ .Name|singleLine }} - {{ .Doc|singleLine }}{{ end }}{{ end }}{{ lifecycle .Message.Lifecycle }}
func (c {{ .Class.Name|exported }}Class) {{ .Message.Name|exported }}(ctx context.Context{{ range .Message.Params }}, {{ .Name|internal }} {{ .GoType }}{{ end }}) {{ if .Message.Result.IsVoid }}error{{ else }}({{ .Message.Result.GoType }}, error){{ end }} {
{{- if .Message.Result.IsVoid }}
	_, err := c.client.APICall(ctx, "{{ .Class.Name }}.{{ .Message.Name }}"{{ range .Message.Params }}, {{ xenValue .Type (.Name|internal) }}{{ end }})
	return err
{{- else }}
	return callAndConvert(ctx, c.client, "{{ .Class.Name }}.{{ .Message.Name }}", {{ .Message.Result.Type|convertToGo }}{{ range .Message.Params }}, {{ xenValue .Type (.Name|internal) }}{{ end }})
{{- end }}
}
`

// prepClient is called by the hand-written NewClient in client.go.
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

// Converter templates for compound types. Leaf types use the generic helpers
// in convert.go instead (see leafConverter).
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

type converterFunc struct {
	name       string
	definition string
}

// apiGenerator holds one generation run's state. converters caches converters
// by type and direction; emittedEnums avoids duplicating an enum declared
// under several classes.
type apiGenerator struct {
	outDir        string
	lifecyclePath string // where to write the lifecycle doc; "" skips it
	classes       []*xapiClass
	templates     *template.Template
	converters    map[string]converterFunc
	emittedEnums  map[string]bool
}

func newAPIGenerator(outDir string) apiGenerator {
	return apiGenerator{
		outDir:       outDir,
		converters:   make(map[string]converterFunc),
		emittedEnums: make(map[string]bool),
	}
}

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

// buildConverterFunc returns the converter for xenType in direction ("ToGo"
// or "ToXen"), generating a function for compound types.
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

// getOrCreateConverterFunc memoizes buildConverterFunc, so each converter is
// generated once however many types nest it.
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

// writeGoFile renders a file and formats it before writing, so the templates
// needn't get whitespace right.
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

func (generator *apiGenerator) generateClassAPI(class *xapiClass) (err error) {
	apiFilename := fmt.Sprintf("%s_gen.go", strings.ToLower(class.Name))

	return writeGoFile(filepath.Join(generator.outDir, apiFilename), func(out io.Writer) (err error) {
		err = generator.templates.ExecuteTemplate(out, "FileHeader", nil)
		if err != nil {
			return
		}

		for _, enum := range class.Enums {
			// Shared enums are declared under several classes; emit them once.
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

// generateConverters writes the converters requested while generating the
// classes, so it must run after them. Sorting keeps the output stable.
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
	if err != nil || generator.lifecyclePath == "" {
		return
	}
	return generator.generateLifecycleDoc(generator.lifecyclePath)
}

func main() {
	schema := flag.String("schema", "xenapi.json", "path to the XenAPI schema")
	out := flag.String("out", ".", "directory to write the generated files to")
	lifecycle := flag.String("lifecycle", "docs/lifecycle.md", `where to write the API lifecycle doc ("" to skip)`)
	flag.Parse()

	generator := newAPIGenerator(*out)
	generator.lifecyclePath = *lifecycle
	if err := generator.run(*schema); err != nil {
		fmt.Fprintln(os.Stderr, "genapi:", err)
		os.Exit(1)
	}
}
