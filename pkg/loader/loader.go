// Copyright 2026 Dipjyoti Metia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

// Package loader reads OpenAPI 3.x and Swagger 2.0 specifications and
// normalises them into the kin-openapi *openapi3.T type used by the rest
// of the generator.
package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// Load reads an OpenAPI spec from a file path and returns the kin-openapi v3
// representation. Swagger 2.0 specs are detected and converted automatically.
// The returned document is validated.
func Load(ctx context.Context, path string) (*openapi3.T, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	if isSwagger2(raw) {
		doc, err := convertSwagger2(raw)
		if err != nil {
			return nil, fmt.Errorf("convert swagger 2.0: %w", err)
		}
		if err := doc.Validate(ctx); err != nil {
			return nil, fmt.Errorf("validate converted v3: %w", err)
		}
		return doc, nil
	}

	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = true
	l.Context = ctx

	var doc *openapi3.T
	if base, err := filepath.Abs(path); err == nil {
		doc, err = l.LoadFromDataWithPath(raw, &url.URL{Scheme: "file", Path: base})
		if err != nil {
			return nil, fmt.Errorf("parse openapi: %w", err)
		}
	} else {
		doc, err = l.LoadFromData(raw)
		if err != nil {
			return nil, fmt.Errorf("parse openapi: %w", err)
		}
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate openapi: %w", err)
	}
	return doc, nil
}

// IsJSONContentType reports whether ct is a JSON media type — either the
// canonical "application/json" or any suffix variant such as
// "application/problem+json". Exported so other packages share a single
// predicate.
func IsJSONContentType(ct string) bool {
	return ct == "application/json" || strings.HasSuffix(ct, "+json")
}

// WriteV3YAMLJSONOnly serialises doc as OpenAPI 3.x YAML, with non-JSON
// content types pruned from response bodies. Request bodies are preserved
// verbatim so downstream oapi-codegen can emit the matching Formdata /
// Multipart / WithBody helpers. The input doc is not modified — pruning
// happens on an internal clone. Useful for piping a Swagger-2.0-converted
// spec into tools (like oapi-codegen) that only accept OpenAPI 3 and can
// mis-handle responses exposed under multiple content types.
func WriteV3YAMLJSONOnly(doc *openapi3.T, path string) error {
	if doc == nil {
		return fmt.Errorf("nil document")
	}
	cloned, err := cloneDoc(doc)
	if err != nil {
		return fmt.Errorf("clone openapi: %w", err)
	}
	pruneNonJSONContent(cloned)

	yamlBytes, err := yaml.Marshal(cloned)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	return os.WriteFile(path, yamlBytes, 0o644)
}

// isSwagger2 reports whether raw is a Swagger 2.0 document. It decodes only
// the top-level shape, so a "swagger: 2.0" reference inside a description
// string of an OpenAPI 3 file does not trigger a false positive.
func isSwagger2(raw []byte) bool {
	var top map[string]any
	if err := yaml.Unmarshal(raw, &top); err != nil {
		return false
	}
	v, ok := top["swagger"]
	if !ok {
		return false
	}
	switch s := v.(type) {
	case string:
		return strings.HasPrefix(s, "2.")
	case float64:
		return s >= 2 && s < 3
	}
	return false
}

func convertSwagger2(raw []byte) (*openapi3.T, error) {
	jsonBytes, err := yamlOrJSONToJSON(raw)
	if err != nil {
		return nil, err
	}
	var v2 openapi2.T
	if err := v2.UnmarshalJSON(jsonBytes); err != nil {
		return nil, fmt.Errorf("unmarshal swagger 2.0: %w", err)
	}
	v3, err := openapi2conv.ToV3(&v2)
	if err != nil {
		return nil, fmt.Errorf("convert v2 → v3: %w", err)
	}
	return v3, nil
}

// yamlOrJSONToJSON returns raw as JSON, decoding YAML if needed.
func yamlOrJSONToJSON(raw []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return raw, nil
	}
	var node any
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return json.Marshal(normaliseYAMLMaps(node))
}

// cloneDoc returns a deep copy of doc by round-tripping through JSON. This
// is the only stable way to clone *openapi3.T — the struct is graph-shaped
// with shared *SchemaRef pointers.
func cloneDoc(doc *openapi3.T) (*openapi3.T, error) {
	buf, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}
	out := &openapi3.T{}
	if err := out.UnmarshalJSON(buf); err != nil {
		return nil, err
	}
	return out, nil
}

// pruneNonJSONContent removes every non-JSON content type from response
// bodies. Request bodies are left intact — the generator now lowers
// form/multipart/octet/text/raw request bodies into MCP tool arguments and
// needs the original content map to pick a content type. JSON is recognised
// by IsJSONContentType.
func pruneNonJSONContent(doc *openapi3.T) {
	if doc.Paths == nil {
		return
	}
	for _, item := range doc.Paths.Map() {
		if item == nil {
			continue
		}
		for _, op := range item.Operations() {
			if op == nil || op.Responses == nil {
				continue
			}
			for _, respRef := range op.Responses.Map() {
				if respRef == nil || respRef.Value == nil {
					continue
				}
				keepJSONOnly(respRef.Value.Content)
			}
		}
	}
}

func keepJSONOnly(c openapi3.Content) {
	for ct := range c {
		if !IsJSONContentType(ct) {
			delete(c, ct)
		}
	}
}

// normaliseYAMLMaps converts map[interface{}]interface{} (legacy YAML) into
// map[string]interface{} for JSON compatibility.
func normaliseYAMLMaps(v any) any {
	switch x := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[fmt.Sprint(k)] = normaliseYAMLMaps(val)
		}
		return out
	case map[string]any:
		for k, val := range x {
			x[k] = normaliseYAMLMaps(val)
		}
		return x
	case []any:
		for i, val := range x {
			x[i] = normaliseYAMLMaps(val)
		}
		return x
	default:
		return v
	}
}
