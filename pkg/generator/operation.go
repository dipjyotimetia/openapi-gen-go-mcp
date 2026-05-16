// Copyright 2026 Dipjyoti Metia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/loader"
)

// OpenAPI parameter "in" values. The spec defines these as enum strings; we
// alias them as constants so generator code can compare without string typos.
const (
	inPath   = "path"
	inQuery  = "query"
	inHeader = "header"
)

// BodyKind classifies how an operation's request body is encoded on the wire.
// The generator dispatches on this to choose which oapi-codegen helper to call
// and how to materialise the body inside the MCP handler.
type BodyKind string

const (
	// BodyNone marks an operation with no request body.
	BodyNone BodyKind = ""
	// BodyJSON marks application/json (or any *+json) — handler decodes into
	// the typed <Op>JSONRequestBody and calls <Op>WithResponse.
	BodyJSON BodyKind = "json"
	// BodyForm marks application/x-www-form-urlencoded — handler decodes into
	// the typed <Op>FormdataRequestBody and calls <Op>WithFormdataBodyWithResponse.
	BodyForm BodyKind = "form"
	// BodyMultipart marks multipart/form-data — handler builds the body via
	// runtime.BuildMultipartBody and calls <Op>WithBodyWithResponse.
	BodyMultipart BodyKind = "multipart"
	// BodyOctet marks application/octet-stream — handler base64-decodes a
	// single body string and calls <Op>WithBodyWithResponse.
	BodyOctet BodyKind = "octet"
	// BodyText marks any text/* content type — handler takes the body as a
	// raw string and calls <Op>WithBodyWithResponse.
	BodyText BodyKind = "text"
	// BodyRaw marks application/xml and any other content type — handler
	// takes the body as a raw string and calls <Op>WithBodyWithResponse.
	BodyRaw BodyKind = "raw"
)

// RequestFilePart describes one binary field inside a multipart body. v1 only
// fills Path; FieldName/ContentType are reserved for OpenAPI `encoding[field]`
// metadata in a future change and default to sensible values when empty.
type RequestFilePart struct {
	// Path is the JSON-pointer into the body object (e.g. "/attachment").
	Path string
	// FieldName overrides the form field name. When empty, the runtime derives
	// it from the last segment of Path.
	FieldName string
	// ContentType overrides the part's Content-Type. When empty, the runtime
	// uses "application/octet-stream".
	ContentType string
}

// Operation is the generator's internal view of one OpenAPI operation,
// pre-resolved into the values the template needs.
type Operation struct {
	// ToolName is the MCP-visible tool name (mangled to fit MaxToolNameLen).
	ToolName string
	// GoName is the Go method name on the oapi-codegen client interface
	// (e.g. FindPetByIDWithResponse). It is the canonical "<Base>WithResponse"
	// name used by paramsTypeName / bodyTypeName helpers; CallMethod records
	// which actual method the handler invokes (may differ for non-JSON bodies).
	GoName string
	// CallMethod is the oapi-codegen client method the generated handler calls.
	// JSON         → <Base>WithResponse
	// Form         → <Base>WithFormdataBodyWithResponse
	// Multipart/…  → <Base>WithBodyWithResponse
	CallMethod string
	// Description is the operation's summary/description, used as the tool description.
	Description string
	// Method/Path are the HTTP verb and templated path; retained for comments
	// and debug output.
	Method string
	Path   string
	// PathParams are the path-template parameters, in declaration order.
	PathParams []ParamField
	// QueryParams + HeaderParams together populate the oapi-codegen <Op>Params struct.
	QueryParams  []ParamField
	HeaderParams []ParamField
	// HasParamsStruct is true when at least one query or header param exists,
	// meaning oapi-codegen emitted a <Op>Params struct that the typed method
	// expects as an additional argument.
	HasParamsStruct bool
	// RequestBody is the body schema for typed kinds (JSON, Form, Multipart)
	// and nil for raw kinds (Octet, Text, Raw).
	RequestBody         *openapi3.SchemaRef
	RequestBodyRequired bool
	HasRequestBody      bool
	// RequestBodyKind classifies how the body is encoded on the wire.
	RequestBodyKind BodyKind
	// RequestContentType is the spec-declared content-type string. Emitted as
	// a literal into the generated call for raw fallback kinds.
	RequestContentType string
	// RequestFileFields lists JSON-pointer paths into the body object that
	// must be base64-decoded into multipart file parts. Populated only when
	// RequestBodyKind == BodyMultipart; sorted by Path for determinism.
	RequestFileFields []RequestFilePart
	// InputSchemaJSON is the encoded JSON Schema for the tool's input.
	InputSchemaJSON string
}

// ParamField is a single OpenAPI parameter described enough to render Go code
// and the matching JSON Schema entry.
type ParamField struct {
	Name         string // OpenAPI name, e.g. "petId"
	GoVar        string // Go local variable name, e.g. "petId"
	GoType       string // Go type, e.g. "int64", "openapi_types.UUID"
	GoTypeImport string // import path required for GoType (empty for builtins)
	Required     bool
	Schema       *openapi3.SchemaRef // original parameter schema, used to build the input schema
}

// CollectOperations walks the spec and returns the Operations to generate, in
// a deterministic order. Each operation is rendered with its own schema
// converter so $defs are self-contained per tool. openAICompat selects the
// JSON-Schema dialect.
//
// Returns an error if any operation cannot be lowered (e.g. non-JSON request
// body) so the caller can fail fast.
func CollectOperations(doc *openapi3.T, openAICompat bool) ([]Operation, error) {
	var ops []Operation
	if doc.Paths == nil {
		return ops, nil
	}

	// Pre-compute the component-schema name map once; every per-operation
	// converter reuses it via Adopt.
	template := NewSchemaConverter(openAICompat)
	template.Bind(doc)
	nameByPtr := template.NameByPtr()

	paths := doc.Paths.Map()
	pathKeys := make([]string, 0, len(paths))
	for path := range paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := paths[path]
		opByMethod := item.Operations()
		methods := make([]string, 0, len(opByMethod))
		for m := range opByMethod {
			methods = append(methods, m)
		}
		sort.Strings(methods)
		for _, method := range methods {
			conv := NewSchemaConverter(openAICompat)
			conv.Adopt(nameByPtr)
			op, err := buildOperation(item, opByMethod[method], method, path, conv)
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", method, path, err)
			}
			ops = append(ops, op)
		}
	}
	return ops, nil
}

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

func buildOperation(item *openapi3.PathItem, op *openapi3.Operation, method, path string, conv *SchemaConverter) (Operation, error) {
	goName := goMethodName(op.OperationID, method, path)
	out := Operation{
		ToolName:    ToolName(op.OperationID, method, path),
		GoName:      goName,
		CallMethod:  goName,
		Description: chooseDescription(op),
		Method:      method,
		Path:        path,
	}

	paramByIn := groupParameters(append(append(openapi3.Parameters{}, item.Parameters...), op.Parameters...))

	for _, m := range pathParamRe.FindAllStringSubmatch(path, -1) {
		name := m[1]
		p := paramByIn[inPath][name]
		out.PathParams = append(out.PathParams, paramFieldFromSpec(name, p, true))
	}
	out.QueryParams = collectParams(paramByIn[inQuery])
	out.HeaderParams = collectParams(paramByIn[inHeader])
	out.HasParamsStruct = len(out.QueryParams)+len(out.HeaderParams) > 0

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		body := op.RequestBody.Value
		out.RequestBodyRequired = body.Required
		if len(body.Content) > 0 {
			kind, ct, schema := pickRequestContent(body.Content)
			// Kinds whose input-schema lowering (raw-string bodies) or
			// multipart binary-field rewrite is not yet implemented continue
			// to error so each rollout step is independently verifiable.
			out.HasRequestBody = true
			out.RequestBodyKind = kind
			out.RequestContentType = ct
			out.CallMethod = callMethodFor(goName, kind)
			// Typed kinds keep the schema for input-schema lowering and
			// (multipart) binary-field rewriting. Raw kinds intentionally
			// drop the spec schema — the MCP input is a single base64 /
			// plain-text string regardless of what the body looks like on
			// the wire.
			switch kind {
			case BodyJSON, BodyForm, BodyMultipart:
				out.RequestBody = schema
			case BodyOctet, BodyText, BodyRaw:
				out.RequestBody = nil
			default:
				return out, fmt.Errorf("unhandled body kind %q for content types %v", kind, contentKeys(body.Content))
			}
		}
	}

	schema, fileFields, err := buildInputSchema(out, conv)
	if err != nil {
		return out, err
	}
	out.InputSchemaJSON = schema
	out.RequestFileFields = fileFields
	return out, nil
}

func groupParameters(params openapi3.Parameters) map[string]map[string]*openapi3.Parameter {
	out := make(map[string]map[string]*openapi3.Parameter)
	for _, ref := range params {
		if ref == nil || ref.Value == nil {
			continue
		}
		p := ref.Value
		if out[p.In] == nil {
			out[p.In] = map[string]*openapi3.Parameter{}
		}
		out[p.In][p.Name] = p
	}
	return out
}

func collectParams(in map[string]*openapi3.Parameter) []ParamField {
	if len(in) == 0 {
		return nil
	}
	out := make([]ParamField, 0, len(in))
	for name, p := range in {
		out = append(out, paramFieldFromSpec(name, p, p.Required))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func paramFieldFromSpec(name string, p *openapi3.Parameter, required bool) ParamField {
	f := ParamField{
		Name:     name,
		GoVar:    goSafeIdent(name),
		GoType:   "string",
		Required: required,
	}
	if p != nil {
		f.GoType, f.GoTypeImport = goTypeForSchema(p.Schema)
		f.Schema = p.Schema
	}
	return f
}

// pickRequestContent chooses the request content-type for an operation that
// declares one or more bodies. The priority is fixed and deterministic:
//
//  1. application/json (and any *+json variant) — preferred.
//  2. application/x-www-form-urlencoded
//  3. multipart/form-data
//  4. application/octet-stream
//  5. text/*
//  6. application/xml
//  7. anything else — first key in lexicographic order.
//
// Iterating with a sorted key slice in every bucket guarantees deterministic
// output even when a content map declares multiple JSON suffix variants or
// multiple text/* subtypes.
//
// Returns BodyNone with empty values when the content map is empty.
func pickRequestContent(c openapi3.Content) (BodyKind, string, *openapi3.SchemaRef) {
	if len(c) == 0 {
		return BodyNone, "", nil
	}
	keys := contentKeys(c) // sorted

	// 1. JSON family.
	for _, ct := range keys {
		if loader.IsJSONContentType(ct) {
			return BodyJSON, ct, schemaOf(c[ct])
		}
	}
	// 2. application/x-www-form-urlencoded.
	for _, ct := range keys {
		if ct == "application/x-www-form-urlencoded" {
			return BodyForm, ct, schemaOf(c[ct])
		}
	}
	// 3. multipart/form-data.
	for _, ct := range keys {
		if ct == "multipart/form-data" {
			return BodyMultipart, ct, schemaOf(c[ct])
		}
	}
	// 4. application/octet-stream.
	for _, ct := range keys {
		if ct == "application/octet-stream" {
			return BodyOctet, ct, schemaOf(c[ct])
		}
	}
	// 5. text/*
	for _, ct := range keys {
		if strings.HasPrefix(ct, "text/") {
			return BodyText, ct, schemaOf(c[ct])
		}
	}
	// 6. application/xml.
	for _, ct := range keys {
		if ct == "application/xml" {
			return BodyRaw, ct, schemaOf(c[ct])
		}
	}
	// 7. Catch-all: first key in sorted order.
	ct := keys[0]
	return BodyRaw, ct, schemaOf(c[ct])
}

// schemaOf returns the schema ref carried by a MediaType, or nil if the entry
// has no schema attached (e.g. an empty value placeholder).
func schemaOf(mt *openapi3.MediaType) *openapi3.SchemaRef {
	if mt == nil {
		return nil
	}
	return mt.Schema
}

// bodyInputSchema returns the JSON Schema map that describes the MCP `body`
// argument for the given operation, plus the list of multipart file fields
// the runtime must base64-decode at request time (nil for non-multipart
// kinds). Typed kinds (JSON/Form/Multipart) lower the spec body schema
// through the SchemaConverter; raw kinds present the body as a single string.
func bodyInputSchema(op Operation, conv *SchemaConverter) (map[string]any, []RequestFilePart) {
	switch op.RequestBodyKind {
	case BodyOctet:
		return map[string]any{
			"type":            "string",
			"contentEncoding": "base64",
			"description":     "request body (application/octet-stream), base64-encoded",
		}, nil
	case BodyText, BodyRaw:
		return map[string]any{
			"type":        "string",
			"description": "request body (" + op.RequestContentType + ")",
		}, nil
	}
	bodySchema := conv.Convert(op.RequestBody)
	var fileFields []RequestFilePart
	if op.RequestBodyKind == BodyMultipart {
		fileFields = rewriteMultipartBinaryFields(bodySchema)
	}
	return bodySchema, fileFields
}

// rewriteMultipartBinaryFields walks the top-level properties of a converted
// multipart body schema and rewrites every {type:"string", format:"binary"}
// leaf into a base64-encoded-string shape suitable for an MCP JSON argument.
// It returns one RequestFilePart per rewritten field so the generated handler
// can pass the correct JSON-pointer list to runtime.BuildMultipartBody.
//
// Only direct properties of the body object are inspected. Nested binary
// leaves inside sub-objects or arrays are intentionally not rewritten in v1;
// real-world multipart specs put file fields at the top level.
func rewriteMultipartBinaryFields(root map[string]any) []RequestFilePart {
	propsAny, ok := root["properties"]
	if !ok {
		return nil
	}
	props, ok := propsAny.(map[string]any)
	if !ok || len(props) == 0 {
		return nil
	}

	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)

	var parts []RequestFilePart
	for _, name := range names {
		sub, ok := props[name].(map[string]any)
		if !ok {
			continue
		}
		if !isBinaryStringLeaf(sub) {
			continue
		}
		delete(sub, "format")
		sub["contentEncoding"] = "base64"
		if _, has := sub["description"]; !has {
			sub["description"] = "base64-encoded binary"
		}
		parts = append(parts, RequestFilePart{Path: "/" + name})
	}
	return parts
}

// isBinaryStringLeaf reports whether m is a schema leaf of the form
// {type:"string", format:"binary"}. Other modifiers (description, title, …)
// are allowed; presence of "properties" disqualifies it (that would be an
// object, not a leaf).
func isBinaryStringLeaf(m map[string]any) bool {
	if !typeIs(m, "string") {
		return false
	}
	if f, _ := m["format"].(string); f != "binary" {
		return false
	}
	if _, hasProps := m["properties"]; hasProps {
		return false
	}
	return true
}

func contentKeys(c openapi3.Content) []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func buildInputSchema(op Operation, conv *SchemaConverter) (string, []RequestFilePart, error) {
	root := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	props := root["properties"].(map[string]any)
	var required []any
	var fileFields []RequestFilePart

	addGroup := func(group string, fields []ParamField) {
		if len(fields) == 0 {
			return
		}
		groupProps := make(map[string]any, len(fields))
		var groupRequired []any
		for _, f := range fields {
			if f.Schema == nil {
				groupProps[f.Name] = map[string]any{"type": "string"}
			} else {
				groupProps[f.Name] = conv.Convert(f.Schema)
			}
			if f.Required {
				groupRequired = append(groupRequired, f.Name)
			}
		}
		groupSchema := map[string]any{"type": "object", "properties": groupProps}
		if len(groupRequired) > 0 {
			groupSchema["required"] = groupRequired
		}
		if conv.OpenAICompat {
			groupSchema["additionalProperties"] = false
		}
		props[group] = groupSchema
		if len(groupRequired) > 0 {
			required = append(required, group)
		}
	}

	addGroup(inPath, op.PathParams)
	addGroup(inQuery, op.QueryParams)
	addGroup(inHeader, op.HeaderParams)

	if op.HasRequestBody {
		bodySchema, parts := bodyInputSchema(op, conv)
		props["body"] = bodySchema
		fileFields = parts
		if op.RequestBodyRequired {
			required = append(required, "body")
		}
	}

	if len(required) > 0 {
		root["required"] = required
	}
	if conv.OpenAICompat {
		root["additionalProperties"] = false
	}
	if defs := conv.Defs(); len(defs) > 0 {
		root["$defs"] = defs
	}

	buf, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("marshal input schema: %w", err)
	}
	return string(buf), fileFields, nil
}

func goMethodName(operationID, method, path string) string {
	if operationID == "" {
		operationID = method + " " + path
	}
	return PascalCase(operationID) + "WithResponse"
}

// callMethodFor returns the oapi-codegen client method name the generated
// handler invokes for the given body kind. The Go name of an operation is
// always "<Base>WithResponse"; non-JSON kinds dispatch to differently-named
// helpers that oapi-codegen emits on the same client interface.
func callMethodFor(goName string, kind BodyKind) string {
	switch kind {
	case BodyForm:
		base := strings.TrimSuffix(goName, "WithResponse")
		return base + "WithFormdataBodyWithResponse"
	case BodyMultipart, BodyOctet, BodyText, BodyRaw:
		base := strings.TrimSuffix(goName, "WithResponse")
		return base + "WithBodyWithResponse"
	default:
		return goName
	}
}

func chooseDescription(op *openapi3.Operation) string {
	switch {
	case op.Summary != "" && op.Description != "":
		return op.Summary + "\n\n" + op.Description
	case op.Summary != "":
		return op.Summary
	default:
		return op.Description
	}
}

// goSafeIdent turns an OpenAPI parameter name into a valid Go identifier.
// Note: this is intentionally distinct from naming.sanitize, which preserves
// dot/dash for MCP tool names but produces invalid Go identifiers.
func goSafeIdent(s string) string {
	if s == "" {
		return "_"
	}
	var b strings.Builder
	for i, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_':
			b.WriteRune(r)
		case i > 0 && r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	id := b.String()
	if reservedGoWords[id] {
		id += "_"
	}
	return id
}

var reservedGoWords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// oapiTypesImport is the import path of the oapi-codegen helper types package.
// It's pulled in whenever a path parameter has a format (uuid/email/date)
// that oapi-codegen maps to a typed wrapper rather than a plain string.
const oapiTypesImport = "github.com/oapi-codegen/runtime/types"

// goTypeForSchema returns the Go type oapi-codegen emits for a primitive
// parameter schema, along with the import path required to reference it
// (empty when only stdlib types are needed). Anything non-primitive falls
// back to string.
func goTypeForSchema(ref *openapi3.SchemaRef) (goType, importPath string) {
	if ref == nil || ref.Value == nil {
		return "string", ""
	}
	s := ref.Value
	types := normaliseTypes(s.Type)
	if len(types) == 0 {
		return "string", ""
	}
	switch types[0] {
	case "string":
		switch s.Format {
		case "uuid":
			return "openapi_types.UUID", oapiTypesImport
		case "email":
			return "openapi_types.Email", oapiTypesImport
		case "date":
			return "openapi_types.Date", oapiTypesImport
		case "date-time":
			return "time.Time", "time"
		}
		return "string", ""
	case "boolean":
		return "bool", ""
	case "integer":
		switch s.Format {
		case "int64":
			return "int64", ""
		case "int32":
			return "int32", ""
		}
		return "int", ""
	case "number":
		if s.Format == "float" {
			return "float32", ""
		}
		return "float64", ""
	}
	return "string", ""
}
