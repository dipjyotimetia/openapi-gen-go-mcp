# FAQ

### Why two codegen steps? Why not generate the HTTP client too?

`oapi-codegen` is already the de-facto standard for OpenAPI → Go HTTP clients, with active community support. Reinventing it would mean owning request signing, retries, mocking, generics, error types, and a thousand spec edge cases. By delegating, the generated MCP code stays small and benefits from every `oapi-codegen` improvement. See [Design Decisions](Design-Decisions).

### Can I avoid both codegen steps with proxy mode?

Yes. `-mode=proxy` produces a complete Go module — `main.go`, `go.mod`, the generated package, and a `README.md`. One command, then `go build`. See [Proxy Mode](Proxy-Mode).

### Does it work with Swagger 2.0?

Yes — directly. The loader detects Swagger 2.0 and converts via `kin-openapi/openapi2conv` internally. You only need `-emit-v3` when something downstream (like `oapi-codegen`) needs the v3 form. See [Swagger 2.0 Workflow](Swagger-2.0-Workflow).

### Why is my response body raw bytes instead of structured?

Only `application/json` responses are decoded into structured tool results. Other content types are surfaced as raw bytes so nothing is silently lost.

### Why are tool inputs grouped into `path` / `query` / `header` / `body`?

To avoid name collisions and mirror the HTTP wire shape — both help LLM tool-use accuracy. See [Schema Modes](Schema-Modes).

### Can I switch from go-sdk to mark3labs/mcp-go without regenerating?

Yes. Change one import and one constructor line in `main.go`. The generated `*.mcp.go` is identical. See [MCP Backends](MCP-Backends).

### Can I expose only some operations?

Yes. Use the `x-mcp` extension on operations, path-items, or the document. Pair with `-exclude-by-default` for curated opt-in mode. See [x-mcp Filtering](x-mcp-Filtering).

### How do I generate from many specs at once?

Point `-spec` at a directory, glob, or comma-separated list. Each spec lands in its own subdirectory under `-out`. See [Batch Mode](Batch-Mode).

### Why is my generated code re-emitting on every run?

By design — generation is reproducible. If you have a hand-edited `*.mcp.go`, it'll be a fatal error until you pass `-force`. Hand-editing isn't supported; if you need to change behavior, change the generator or use [Runtime Options](Runtime-Options).

### Can I use this with OpenAI tool-calling instead of MCP?

Use `-openai-compat` to emit the strict JSON Schema dialect OpenAI accepts. You'll still need an MCP server (or wrapper) to expose the tools — `openapi-go-mcp` only emits MCP tools.

### Is this related to `protoc-gen-go-mcp`?

Yes — it's the OpenAPI counterpart to [`redpanda-data/protoc-gen-go-mcp`](https://github.com/redpanda-data/protoc-gen-go-mcp). Portions of `pkg/runtime` and `pkg/generator/naming.go` are adapted from it under Apache 2.0.

### How do I report a bug or request a feature?

[Open an issue](https://github.com/dipjyotimetia/openapi-go-mcp/issues). Include the spec snippet, the command you ran, and the actual vs expected output.

### Where's the changelog?

[`docs/changelog.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/changelog.md). User-visible changes go under `## Unreleased`.
