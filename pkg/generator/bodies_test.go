// Copyright 2026 Dipjyoti Metia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package generator

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/loader"
)

// renderNonJSONFixture loads testdata/non-json-bodies-v3.yaml and renders the
// MCP wrapper for it. The fixture grows over the body-kind rollout, so any
// step that fails to render the current state of the file indicates a
// regression in a previously-enabled kind.
func renderNonJSONFixture(t *testing.T) string {
	t.Helper()
	doc, err := loader.Load(context.Background(),
		filepath.Join("..", "..", "testdata", "non-json-bodies-v3.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	src, err := Render(doc, Options{
		PackageName:  "nonjsonbodiesmcp",
		ClientImport: "github.com/example/nonjsonbodies",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return string(src)
}

func TestRender_BodyForm(t *testing.T) {
	src := renderNonJSONFixture(t)

	want := []string{
		// Typed form body, decoded via DecodeBody just like JSON.
		"var body nonjsonbodies.SubmitLoginFormdataRequestBody",
		"runtime.DecodeBody(req.Arguments, &body)",
		// Dispatches to the Formdata variant of the oapi-codegen client.
		"c.SubmitLoginWithFormdataBodyWithResponse(ctx, body)",
		// Input schema still presents `body` as the form-object shape.
		`"body":`,
		`"username":`,
	}
	for _, fragment := range want {
		if !strings.Contains(src, fragment) {
			t.Errorf("expected generated source to contain %q\n--- got ---\n%s", fragment, src)
		}
	}

	// Sanity: the JSON-only call shape must NOT leak in for a form op.
	if strings.Contains(src, "SubmitLoginJSONRequestBody") {
		t.Errorf("form op should not reference JSONRequestBody")
	}
}

func TestRender_BodyMultipart(t *testing.T) {
	src := renderNonJSONFixture(t)

	want := []string{
		// Handler builds the body via the multipart runtime helper and passes
		// the JSON-pointer list of binary fields.
		`runtime.BuildMultipartBody(req.Arguments, []string{"/attachment"})`,
		// Dispatches to the generic raw-body variant of the typed client.
		"c.UploadFileWithBodyWithResponse(ctx, contentType, body)",
		// Schema rewrite must replace format:binary with contentEncoding:base64.
		`"contentEncoding": "base64"`,
	}
	for _, fragment := range want {
		if !strings.Contains(src, fragment) {
			t.Errorf("expected generated source to contain %q\n--- got ---\n%s", fragment, src)
		}
	}

	// The binary field's original format:binary keyword must be gone.
	for _, attachmentBlock := range extractInputSchemas(src) {
		if !strings.Contains(attachmentBlock, "uploadFile") &&
			!strings.Contains(attachmentBlock, "attachment") {
			continue
		}
		if strings.Contains(attachmentBlock, `"format": "binary"`) {
			t.Errorf("multipart binary field still has format:binary\n%s", attachmentBlock)
		}
	}
}

func TestRender_BodyOctet(t *testing.T) {
	src := renderNonJSONFixture(t)

	want := []string{
		"runtime.BuildBase64BytesBody(req.Arguments)",
		`c.UploadBlobWithBodyWithResponse(ctx, "application/octet-stream", body)`,
		`"contentEncoding": "base64"`,
	}
	for _, fragment := range want {
		if !strings.Contains(src, fragment) {
			t.Errorf("expected generated source to contain %q", fragment)
		}
	}
}

func TestRender_BodyText(t *testing.T) {
	src := renderNonJSONFixture(t)

	want := []string{
		"runtime.BuildStringBody(req.Arguments)",
		`c.PostNoteWithBodyWithResponse(ctx, "text/plain", body)`,
		`"request body (text/plain)"`,
	}
	for _, fragment := range want {
		if !strings.Contains(src, fragment) {
			t.Errorf("expected generated source to contain %q", fragment)
		}
	}
}

func TestRender_BodyRaw_XML(t *testing.T) {
	src := renderNonJSONFixture(t)

	want := []string{
		"runtime.BuildStringBody(req.Arguments)",
		`c.ImportXMLWithBodyWithResponse(ctx, "application/xml", body)`,
		`"request body (application/xml)"`,
	}
	for _, fragment := range want {
		if !strings.Contains(src, fragment) {
			t.Errorf("expected generated source to contain %q", fragment)
		}
	}
}

// TestRender_Swagger2_FormData confirms that a Swagger 2.0 spec with formData
// parameters survives openapi2conv's lowering to application/x-www-form-urlencoded
// and is rendered as a BodyForm operation by the generator. This is the
// regression cover the previous JSON-only loader pruning was incidentally
// providing.
func TestRender_Swagger2_FormData(t *testing.T) {
	doc, err := loader.Load(context.Background(),
		filepath.Join("..", "..", "testdata", "form-swagger-v2.json"))
	if err != nil {
		t.Fatalf("load swagger 2.0: %v", err)
	}
	src, err := Render(doc, Options{
		PackageName:  "formloginmcp",
		ClientImport: "github.com/example/formlogin",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	want := []string{
		"var body formlogin.SwaggerLoginFormdataRequestBody",
		"runtime.DecodeBody(req.Arguments, &body)",
		"c.SwaggerLoginWithFormdataBodyWithResponse(ctx, body)",
	}
	for _, fragment := range want {
		if !strings.Contains(string(src), fragment) {
			t.Errorf("expected generated source to contain %q\n--- got ---\n%s", fragment, src)
		}
	}
}
