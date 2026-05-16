// Copyright 2026 Dipjyoti Metia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

package runtime

import (
	"encoding/json"
	"errors"
)

// ToolError represents an MCP tool error that should be surfaced as a tool
// result (IsError=true) rather than a protocol-level error. Generated code
// returns this via HandleError when a typed validation or HTTP error occurs.
type ToolError struct {
	// Status is an optional HTTP-style status code, included in the JSON
	// payload when non-zero so callers can branch on it.
	Status int
	// Code is an optional short machine-readable label (e.g. "invalid_body").
	Code string
	// Message is the human-readable failure description.
	Message string
}

func (e *ToolError) Error() string { return e.Message }

// HandleError converts any error to a tool result. The error itself is never
// propagated as a protocol error — instead it is JSON-encoded into the result
// with IsError=true so the LLM can read and react to it.
//
// Returning (nil, nil) for a nil error matches the generated-handler pattern:
//
//	if err != nil { return runtime.HandleError(err) }
func HandleError(err error) (*CallToolResult, error) {
	if err == nil {
		return nil, nil
	}

	payload := map[string]any{"error": err.Error()}
	var te *ToolError
	if errors.As(err, &te) {
		if te.Status != 0 {
			payload["status"] = te.Status
		}
		if te.Code != "" {
			payload["code"] = te.Code
		}
	}

	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return NewToolResultError("error: " + err.Error()), nil
	}
	return NewToolResultError(string(body)), nil
}
