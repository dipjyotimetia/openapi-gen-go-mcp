# openapi-go-mcp Wiki

Turn any OpenAPI 3.x or Swagger 2.0 spec into a [Model Context Protocol](https://modelcontextprotocol.io) server in Go. Every operation becomes an MCP tool; HTTP work is delegated to an [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) typed client.

This wiki is the navigable, task-oriented companion to the in-repo [`docs/`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/docs). Use the sidebar to jump around.

## Where do I start?

| If you want to… | Go to |
|---|---|
| Run an MCP server in one command | [Getting Started → Proxy mode](Getting-Started#proxy-mode) |
| Embed MCP into an existing Go service | [Companion mode](Companion-Mode) |
| Understand every CLI flag | [CLI Reference](CLI-Reference) |
| See real working examples | [Examples](Examples) |
| Deploy over stdio, HTTP, SSE, or with auth | [Deployment Patterns](Deployment-Patterns) |
| Curate which operations are exposed | [Filtering with x-mcp](x-mcp-Filtering) |
| Generate from a whole folder of specs | [Batch Mode](Batch-Mode) |
| Use Swagger 2.0 input | [Swagger 2.0 Workflow](Swagger-2.0-Workflow) |
| Switch MCP libraries | [MCP Backends](MCP-Backends) |
| Understand the generator internals | [Architecture](Architecture) |
| Know why a choice was made | [Design Decisions](Design-Decisions) |

## What is `openapi-go-mcp` in one paragraph?

A CLI code generator. You point it at an OpenAPI spec; it writes a single Go file that registers every operation as an MCP tool. The generated code targets a thin `runtime.MCPServer` interface so you can pick between the official [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) and [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) by changing one import. Tool input schemas are derived from the spec (with optional OpenAI-strict mode). It does not own the HTTP transport — that's the `oapi-codegen` client's job.

## Two emission modes

- **Proxy mode** (`-mode=proxy`) — a complete, runnable Go module: `main.go` + `go.mod` + `<pkg>.mcp.go` + `README.md`. Auth from env vars derived from the spec's `securitySchemes`. Use when you just want to run a server.
- **Companion mode** (default) — emits one `<pkg>.mcp.go` file you import into your own `main`. Use when MCP is one feature of a larger service binary.

See [Proxy vs Companion](Proxy-vs-Companion-Mode) for the trade-offs.

## Status

Pre-1.0. APIs may change between minor versions. Apache 2.0 licensed.
