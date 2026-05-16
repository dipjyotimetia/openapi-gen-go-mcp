# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) starting with the v1.0.0 release.

## [Unreleased]

### Added

- Distribution channels: pre-built binaries on every tagged release (darwin/linux/windows × amd64/arm64), Homebrew formula via `dipjyotimetia/homebrew-tap`, and a multi-arch container image at `ghcr.io/dipjyotimetia/openapi-gen-go-mcp`. Configured in `.goreleaser.yml`; release workflow runs on `v*` tag push.
- `-version` flag on the CLI prints the goreleaser-injected build metadata, falling back to `runtime/debug.BuildInfo` for `go install` builds.
- `examples/non-json-bodies/` — end-to-end example exercising every non-JSON request body kind (form, multipart, octet, text, xml). Provides a real `oapi-codegen` + generator compile check on every CI run.
- Non-JSON request bodies. The generator now lowers `application/x-www-form-urlencoded`, `multipart/form-data`, `application/octet-stream`, `text/*`, `application/xml`, and any other content type into MCP tool arguments alongside the existing `application/json` path. When an operation declares multiple content types the generator picks deterministically (JSON → form → multipart → octet → text → xml → first). Multipart `format:binary` fields are accepted as base64 strings via three new runtime helpers (`BuildMultipartBody`, `BuildBase64BytesBody`, `BuildStringBody`).
- Format-aware Go type resolution for path parameters. `format: uuid` / `email` / `date` produce typed wrappers (`openapi_types.UUID`, `openapi_types.Email`, `openapi_types.Date`) matching oapi-codegen's output; `format: date-time` produces `time.Time`. The generator emits any required extra imports automatically.
- End-to-end test suite in `internal/e2e` (36 tests across 4 spec fixtures):
  - **Petstore v3** — basic JSON body and primitive path/query (gosdk + mark3labs adapter parity).
  - **Users API v3** — UUID path params, multi-path params, required headers, PUT / PATCH / DELETE, no-param operations, bad-UUID error path.
  - **Library Swagger 2.0** — full v2 → v3 → oapi-codegen → MCP pipeline.
  - **Complex Schemas** — recursive `$ref` in `$defs`, oneOf, allOf, enums, date-time/uuid formats, nullable.
  - CLI integration: build sanity, list, missing-flag exit, `-emit-v3` round-trip, generated-file structural invariants.

### Changed

- `loader.WriteV3YAMLJSONOnly` (used by `-emit-v3`) now prunes non-JSON content from response bodies only. Request bodies are preserved verbatim so downstream `oapi-codegen` emits the matching Formdata / Multipart / WithBody helpers that the MCP wrapper now calls.

### Fixed

- Path parameters with non-string Go types (e.g. `format: uuid` → `openapi_types.UUID`) were generated as `string`, which failed to compile against oapi-codegen output.

## [0.1.0] — 2026-05-16

Initial public release.

### Added

- CLI `openapi-gen-go-mcp` reads OpenAPI 3.0 / 3.1 / Swagger 2.0 specs and generates a `*.mcp.go` file per spec. Each operation becomes an MCP tool whose handler forwards to an `oapi-codegen` `ClientWithResponsesInterface`.
- `pkg/loader` — spec ingestion with auto-conversion of Swagger 2.0 via `kin-openapi/openapi2conv`. Exports `Load`, `WriteV3YAMLJSONOnly`, `IsJSONContentType`.
- `pkg/generator` — operation walk, JSON-Schema conversion (draft-07 compatible, recursion-safe via `$defs`), `text/template` driven Go-source emission with gofmt post-pass.
- `pkg/runtime` — MCP-library-agnostic types (`MCPServer`, `Tool`, `CallToolRequest`, `CallToolResult`), JSON decode helpers (`DecodePathParam`, `DecodeBody`, `DecodeParamsCombined`), functional options (`WithNamePrefix`, `WithExtraProperties`).
- `pkg/runtime/gosdk` — adapter for the official `modelcontextprotocol/go-sdk`.
- `pkg/runtime/mark3labs` — adapter for `mark3labs/mcp-go`. Generated code is unchanged when switching between the two.
- `-openai-compat` flag — emits OpenAI-tool-compatible JSON Schema (no `$ref`, no `oneOf`/`anyOf`/`allOf`, `additionalProperties:false` on every object).
- `-emit-v3` flag — converts a Swagger 2.0 spec to OpenAPI 3 YAML, pruning non-JSON content types. Works around an oapi-codegen v2.7.0 quirk on responses exposed under multiple content types.
- `-list` flag — print operations in the spec and exit.
- Examples: `examples/petstore` (go-sdk), `examples/petstore-mark3labs` (mark3labs), `examples/swagger2-petstore` (Swagger 2.0 input).
- Golden test for generator output (`pkg/generator/golden_test.go`); refresh via `UPDATE_GOLDEN=1`.
- End-to-end OpenAI-compat invariants test (`pkg/generator/openai_compat_test.go`).
- Loader, schema-converter, and runtime helper unit tests.

### Known limitations

- Only `application/json` response bodies are decoded; non-JSON responses surface as raw bytes.
- Streaming responses are returned as raw bytes; no first-class SSE / chunked support.
- No dynamic (no-codegen, reflection-based) registration path yet.
- `discriminator` is dropped during schema conversion — JSON Schema has no direct equivalent.

[Unreleased]: https://github.com/dipjyotimetia/openapi-gen-go-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dipjyotimetia/openapi-gen-go-mcp/releases/tag/v0.1.0
