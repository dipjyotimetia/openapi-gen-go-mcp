# Proxy Mode

`-mode=proxy` emits a **complete, runnable Go module**. No second codegen step, no hand-written `main.go`. The result is a turnkey MCP server that forwards every tool call to the upstream HTTP API described by your spec.

## What gets generated

```
gen/petstore-mcp/
├── go.mod                     # module = -module flag value
├── main.go                    # builds & runs the MCP server on stdio
├── README.md                  # lists the required env vars per security scheme
└── petstoremcp/
    └── petstoremcp.mcp.go     # the generated tool registrations
```

## Generate

```bash
openapi-go-mcp \
    -mode=proxy \
    -spec petstore.yaml \
    -out gen/petstore-mcp \
    -module github.com/me/petstore-mcp

cd gen/petstore-mcp
go mod tidy
go build
./petstore-mcp        # MCP on stdio
```

## Authentication via env vars

The generator inspects the spec's `securitySchemes` and wires each one to a deterministic environment variable. The generated `README.md` lists the exact variables for **your** spec, but the pattern is:

| Scheme type | Env var |
|---|---|
| `apiKey` | `API_KEY_<SCHEMENAME>` |
| `http` bearer | `BEARER_TOKEN_<SCHEMENAME>` |
| `http` basic | `BASIC_AUTH_<SCHEMENAME>` *(or)* `BASIC_AUTH_USERNAME_<SCHEMENAME>` + `BASIC_AUTH_PASSWORD_<SCHEMENAME>` |
| `oauth2` / `openIdConnect` | `OAUTH2_ACCESS_TOKEN_<SCHEMENAME>` |

Names are uppercased and non-alphanumerics become `_`. Leading/trailing underscores are trimmed.

## Upstream base URL

```
API_BASE_URL=https://api.example.com ./petstore-mcp
```

If unset, the proxy uses `servers[0].url` from the spec. Set it explicitly to point at a different environment (staging vs prod, mock server, etc.).

## What the proxy does at request time

1. Receives the MCP `tools/call`.
2. Validates the arguments against the JSON Schema derived from the spec.
3. Looks up the env-var credentials for whichever `securitySchemes` the operation requires.
4. Builds the HTTP request via the embedded `oapi-codegen` client.
5. Returns the response — body, status, and headers — as the MCP tool result. Non-2xx HTTP responses are surfaced rather than swallowed.

## When to use this mode

- You want an MCP server but you do **not** want to maintain a Go module.
- You're standing up many APIs the same way and don't want to write per-API `main.go` files.
- Auth is a static credential per environment (token in CI secret, etc.) — env vars are enough.

## When to use [Companion Mode](Companion-Mode) instead

- You need custom HTTP behavior (retries, OpenTelemetry, mTLS, request signing beyond basic schemes).
- You're embedding MCP into a larger service binary.
- Auth requires app-level logic (token refresh, per-tenant credentials from a database).

## Limitations

- The proxy only knows the auth schemes encoded in the spec. Custom or non-standard auth needs companion mode.
- Only `application/json` response bodies are decoded; other content types are surfaced as raw bytes.
- One MCP server per spec. To aggregate multiple APIs in one server, use companion mode and call multiple `RegisterXxxClient` functions from your `main`.

## Related

- [CLI Reference](CLI-Reference)
- [Companion Mode](Companion-Mode)
- [Deployment Patterns](Deployment-Patterns)
