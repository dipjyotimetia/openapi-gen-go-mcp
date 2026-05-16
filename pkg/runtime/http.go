// Copyright 2026 Dipjyoti Metia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package runtime

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"
)

// DecodeField JSON-roundtrips args[key] into out. out must be a non-nil
// pointer to a struct or primitive. Returns a *ToolError on failure so
// HandleError can render a useful message.
func DecodeField(args map[string]any, key string, out any) error {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil
	}
	buf, err := json.Marshal(raw)
	if err != nil {
		return &ToolError{
			Status:  400,
			Code:    "invalid_argument",
			Message: fmt.Sprintf("marshal %q: %v", key, err),
		}
	}
	if err := json.Unmarshal(buf, out); err != nil {
		return &ToolError{
			Status:  400,
			Code:    "invalid_argument",
			Message: fmt.Sprintf("decode %q: %v", key, err),
		}
	}
	return nil
}

// DecodeBody is a convenience wrapper for the conventional "body" key.
func DecodeBody(args map[string]any, out any) error {
	return DecodeField(args, "body", out)
}

// DecodePathParam JSON-decodes args["path"][name] into out.
func DecodePathParam(args map[string]any, name string, out any) error {
	path, _ := args["path"].(map[string]any)
	if path == nil {
		if out != nil {
			return &ToolError{
				Status:  400,
				Code:    "missing_path_param",
				Message: fmt.Sprintf("missing path parameter %q", name),
			}
		}
		return nil
	}
	v, ok := path[name]
	if !ok {
		return &ToolError{
			Status:  400,
			Code:    "missing_path_param",
			Message: fmt.Sprintf("missing path parameter %q", name),
		}
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return &ToolError{Status: 400, Code: "invalid_path_param", Message: err.Error()}
	}
	if err := json.Unmarshal(buf, out); err != nil {
		return &ToolError{Status: 400, Code: "invalid_path_param", Message: fmt.Sprintf("decode path %q: %v", name, err)}
	}
	return nil
}

// DecodeQueryParams JSON-decodes args["query"] into out (typically a pointer
// to the oapi-codegen-generated <Op>Params struct).
func DecodeQueryParams(args map[string]any, out any) error {
	return DecodeField(args, "query", out)
}

// DecodeHeaderParams JSON-decodes args["header"] into out.
func DecodeHeaderParams(args map[string]any, out any) error {
	return DecodeField(args, "header", out)
}

// DecodeParamsCombined JSON-decodes the union of args["query"] and
// args["header"] into out. oapi-codegen emits a single <Op>Params struct
// covering both groups, so generated handlers can use this helper to populate
// it in one call.
func DecodeParamsCombined(args map[string]any, out any) error {
	merged := map[string]any{}
	if q, ok := args["query"].(map[string]any); ok {
		for k, v := range q {
			merged[k] = v
		}
	}
	if h, ok := args["header"].(map[string]any); ok {
		for k, v := range h {
			merged[k] = v
		}
	}
	if len(merged) == 0 {
		return nil
	}
	buf, err := json.Marshal(merged)
	if err != nil {
		return &ToolError{Status: 400, Code: "invalid_argument", Message: err.Error()}
	}
	if err := json.Unmarshal(buf, out); err != nil {
		return &ToolError{Status: 400, Code: "invalid_argument", Message: "decode params: " + err.Error()}
	}
	return nil
}

// multipartFilePartContentType is the per-part Content-Type written for every
// multipart file field. v1 always uses application/octet-stream; OpenAPI
// `encoding[field].contentType` overrides are a known gap (see
// docs/architecture.md).
const multipartFilePartContentType = "application/octet-stream"

// BuildMultipartBody encodes args["body"] (a JSON object) as a multipart/form-data
// payload. Properties whose JSON-pointer path is listed in fileFields are
// base64-decoded into a file part; all other properties become plain form
// fields (string values pass through, non-string values are JSON-encoded).
//
// Properties are emitted in sorted key order so generated tests can assert on
// the part list without relying on Go map iteration. The returned content type
// carries the boundary the multipart writer chose.
//
// A missing or empty body produces a valid empty multipart payload — callers
// rely on the MCP input schema to enforce required-ness.
func BuildMultipartBody(args map[string]any, fileFields []string) (string, io.Reader, error) {
	fileSet := make(map[string]struct{}, len(fileFields))
	for _, f := range fileFields {
		fileSet[f] = struct{}{}
	}

	body, err := bodyAsObject(args)
	if err != nil {
		return "", nil, err
	}

	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for _, k := range keys {
		v := body[k]
		if _, isFile := fileSet["/"+k]; isFile {
			if err := writeFilePart(mw, k, v); err != nil {
				return "", nil, err
			}
			continue
		}
		if err := writeFormField(mw, k, v); err != nil {
			return "", nil, err
		}
	}
	if err := mw.Close(); err != nil {
		return "", nil, &ToolError{Status: 500, Code: "multipart_close", Message: err.Error()}
	}
	return mw.FormDataContentType(), &buf, nil
}

// BuildBase64BytesBody reads args["body"] as a base64-encoded string and
// returns its decoded bytes as an io.Reader, suitable for an
// application/octet-stream request body.
func BuildBase64BytesBody(args map[string]any) (io.Reader, error) {
	s, ok, err := bodyAsString(args)
	if err != nil {
		return nil, err
	}
	if !ok {
		return bytes.NewReader(nil), nil
	}
	decoded, decodeErr := base64.StdEncoding.DecodeString(s)
	if decodeErr != nil {
		return nil, &ToolError{
			Status:  400,
			Code:    "invalid_body",
			Message: "decode body as base64: " + decodeErr.Error(),
		}
	}
	return bytes.NewReader(decoded), nil
}

// BuildStringBody reads args["body"] as a string and returns it as an
// io.Reader, suitable for text/* and other raw-string request bodies.
func BuildStringBody(args map[string]any) (io.Reader, error) {
	s, _, err := bodyAsString(args)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(s), nil
}

// bodyAsObject extracts args["body"] as a JSON object. A missing or nil body
// is reported as the empty map, not an error, matching DecodeBody semantics.
func bodyAsObject(args map[string]any) (map[string]any, error) {
	raw, ok := args["body"]
	if !ok || raw == nil {
		return map[string]any{}, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, &ToolError{
			Status:  400,
			Code:    "invalid_body",
			Message: fmt.Sprintf("body must be an object, got %T", raw),
		}
	}
	return obj, nil
}

// bodyAsString extracts args["body"] as a string. The bool return distinguishes
// "absent" (false) from "present and empty" (true) so callers can pick the
// right zero-value behaviour. A non-string body is rejected.
func bodyAsString(args map[string]any) (string, bool, error) {
	raw, ok := args["body"]
	if !ok || raw == nil {
		return "", false, nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", false, &ToolError{
			Status:  400,
			Code:    "invalid_body",
			Message: fmt.Sprintf("body must be a string, got %T", raw),
		}
	}
	return s, true, nil
}

// writeFilePart writes a single multipart file part. The value must be a
// base64-encoded string; arbitrary JSON types are rejected so that schema
// drift surfaces loudly rather than encoding the JSON literal as part bytes.
func writeFilePart(mw *multipart.Writer, name string, v any) error {
	s, ok := v.(string)
	if !ok {
		return &ToolError{
			Status:  400,
			Code:    "invalid_body",
			Message: fmt.Sprintf("file field %q must be a base64 string, got %T", name, v),
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return &ToolError{
			Status:  400,
			Code:    "invalid_body",
			Message: fmt.Sprintf("decode file field %q as base64: %v", name, err),
		}
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name=%q; filename=%q`, name, name))
	header.Set("Content-Type", multipartFilePartContentType)
	part, err := mw.CreatePart(header)
	if err != nil {
		return &ToolError{Status: 500, Code: "multipart_part", Message: err.Error()}
	}
	if _, err := part.Write(decoded); err != nil {
		return &ToolError{Status: 500, Code: "multipart_write", Message: err.Error()}
	}
	return nil
}

// writeFormField writes a single multipart form field. Strings pass through;
// everything else is JSON-encoded so structured arguments survive the form
// boundary.
func writeFormField(mw *multipart.Writer, name string, v any) error {
	var serialised string
	switch x := v.(type) {
	case nil:
		serialised = ""
	case string:
		serialised = x
	default:
		buf, err := json.Marshal(x)
		if err != nil {
			return &ToolError{
				Status:  400,
				Code:    "invalid_body",
				Message: fmt.Sprintf("marshal form field %q: %v", name, err),
			}
		}
		serialised = string(buf)
	}
	if err := mw.WriteField(name, serialised); err != nil {
		return &ToolError{Status: 500, Code: "multipart_field", Message: err.Error()}
	}
	return nil
}
