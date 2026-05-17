# Architecture

A concise overview. For the full version (data-flow diagrams, extension points, where to add a new backend), read [`docs/architecture.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/architecture.md) in the repo.

## Package layout

```
cmd/openapi-go-mcp/   CLI entry point + batch orchestration
pkg/loader/           Spec ingestion: OpenAPI 3.x direct, Swagger 2.0 via openapi2conv
pkg/batch/            Per-spec option derivation (slug → PackageName/OutDir/ClientImport), collision detection
pkg/generator/        Operation collection, JSON Schema conversion, text/template → gofmt
pkg/runtime/          MCPServer interface + decoders + ApplyConfig (library-agnostic)
pkg/runtime/gosdk/    Adapter for modelcontextprotocol/go-sdk
pkg/runtime/mark3labs/ Adapter for mark3labs/mcp-go
examples/             One end-to-end demo per backend / per spec dialect
testdata/             Spec fixtures + golden generator output
tests/e2e/            Black-box tests over MCP stdio; CLI integration tests
```

## Generator pipeline

```
loader.Load
    ↓
generator.CollectOperations   (walks paths × methods, sorted)
    ↓
generator.Render              (text/template → gofmt)
    ↓
write <out>/<pkg>.mcp.go
```

Determinism — sorted iteration, `gofmt`, golden test — is a hard requirement. Diff reviews depend on it.

## Two decoupling boundaries

These do most of the architectural work.

### 1. `runtime.MCPServer` interface

Generated code only calls `AddTool(Tool, ToolHandler)`. The choice of MCP library is a one-line import swap. See [MCP Backends](MCP-Backends).

### 2. Per-operation `SchemaConverter`

Each operation gets its own converter so each tool's `$defs` is self-contained. A `nameByPtr` map is shared across converters within one `CollectOperations` call to avoid O(P·S) rebuild cost. See [Schema Modes](Schema-Modes).

## Batch mode

Batch mode sits in front of the single-spec pipeline rather than changing it:

```
loader.ExpandSpecArg   →   sorted, deduped list of SpecRef
batch.PlanFor          →   derives generator.Options per spec
batch.DetectCollisions →   fails up front on slug collisions
    ↓ for each plan
single-spec pipeline runs
```

Per-spec failures are accumulated; the process exits with code `3` at the end. See [Batch Mode](Batch-Mode).

## Generated handler shape

For every operation the generator emits an `AddTool` call wrapping a closure that:

1. Decodes path/query/header/body args via `runtime.DecodeBody`, `DecodePathParam`, etc.
2. Calls the typed `<Op>WithResponse(ctx, ...)` method on the `oapi-codegen` client.
3. Returns the response body via `runtime.NewToolResultJSON`.

Argument order to the typed client follows `oapi-codegen`'s deterministic convention:

```
ctx, positional path params, *<Op>Params (only when query/header present), body (only when present), reqEditors...
```

## Spec ingestion

- `loader.Load` detects Swagger 2.0 by the top-level `swagger: 2.0` and converts via `kin-openapi/openapi2conv`.
- All loaded specs pass `openapi3.Validate`.
- External `$ref`s resolve against the spec file's directory.
- `WriteV3YAMLJSONOnly` (the `-emit-v3` flag) prunes non-JSON content types on a deep clone — the original document is never mutated.

Request bodies support `application/json`, `application/x-www-form-urlencoded`, `multipart/form-data`, `application/octet-stream`, `text/*`, and any other content type as a raw-string / base64 fallback. When an operation declares multiple, the generator picks deterministically in priority order. Only `application/json` response bodies are decoded; non-JSON responses are surfaced as raw bytes.

## Related

- [Design Decisions](Design-Decisions) — the *why* behind these choices
- [MCP Backends](MCP-Backends)
- [Schema Modes](Schema-Modes)
- [Contributing](Contributing)
