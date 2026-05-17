# Examples

All examples live in [`examples/`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples). Each is a complete, runnable demo.

## Canonical end-to-end demo

| Directory | What it shows |
|---|---|
| [`examples/todos`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/todos) | Two real binaries: `todos-server` (standalone HTTP backend with graceful shutdown, request logging, `/healthz`) and `todos-mcp` (MCP proxy forwarding every tool call over HTTP). Ships a [README](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/examples/todos/README.md) with **MCP client configs for Claude Desktop, Claude Code, Cursor, VS Code, and Inspector**. Start here. |

## Per-backend / per-dialect

| Directory | What it shows |
|---|---|
| [`examples/petstore`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/petstore) | OpenAPI 3.0, JSON bodies, official `go-sdk` backend. |
| [`examples/petstore-mark3labs`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/petstore-mark3labs) | Same spec on the `mark3labs/mcp-go` backend — diff is one import + one constructor line. |
| [`examples/swagger2-petstore`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/swagger2-petstore) | Swagger 2.0 input via `-emit-v3` conversion. |

## Edge cases worth seeing

| Directory | What it shows |
|---|---|
| [`examples/users-api`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/users-api) | UUID path params, required headers, PUT / PATCH / DELETE methods. |
| [`examples/library`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/library) | Swagger 2.0 end-to-end (load → convert → generate). |
| [`examples/complex`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/complex) | Recursive `$ref`, `oneOf` / `allOf`, enums, `date-time` / `uuid` formats. |
| [`examples/non-json-bodies`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/non-json-bodies) | Form-urlencoded, multipart (with base64 file fields), octet-stream, `text/plain`, XML. |

## Regenerating examples

```bash
make regen-examples     # regenerates oapi-codegen + mcp output for every example
make smoke              # boots petstore over stdio, calls initialize + tools/list
```

`regen-examples` requires `oapi-codegen` on `PATH`.
