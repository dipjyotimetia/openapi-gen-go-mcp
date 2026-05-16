// Copyright 2026 Dipjyoti Metia.
// Portions copyright 2025 Redpanda Data, Inc. (adapter shape adapted from
// redpanda-data/protoc-gen-go-mcp, Apache-2.0).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0

// Package mark3labs adapts mark3labs/mcp-go to the MCP-library-agnostic
// runtime.MCPServer interface so generated code can target it interchangeably
// with the official go-sdk.
package mark3labs

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/runtime"
)

type adapter struct {
	s *mcpserver.MCPServer
}

// Wrap exposes an existing *mcpserver.MCPServer as a runtime.MCPServer.
func Wrap(s *mcpserver.MCPServer) runtime.MCPServer { return &adapter{s: s} }

// NewServer creates a new mark3labs MCPServer and returns it alongside the
// runtime.MCPServer adapter. Use the raw *mcpserver.MCPServer for transport
// setup (e.g. server.ServeStdio) and the adapter for tool registration.
func NewServer(name, version string, opts ...mcpserver.ServerOption) (*mcpserver.MCPServer, runtime.MCPServer) {
	s := mcpserver.NewMCPServer(name, version, opts...)
	return s, Wrap(s)
}

func (a *adapter) AddTool(tool runtime.Tool, handler runtime.ToolHandler) {
	mcpTool := mcp.Tool{
		Name:            tool.Name,
		Description:     tool.Description,
		RawInputSchema:  tool.RawInputSchema,
		RawOutputSchema: tool.RawOutputSchema,
	}
	a.s.AddTool(mcpTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, &runtime.CallToolRequest{
			Arguments: req.GetArguments(),
		})
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil
		}
		if result.IsError {
			return mcp.NewToolResultError(result.Text), nil
		}
		res := mcp.NewToolResultText(result.Text)
		res.StructuredContent = result.StructuredContent
		return res, nil
	})
}
