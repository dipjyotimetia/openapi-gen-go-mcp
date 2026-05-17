# Deployment Patterns

Different ways to assemble the generated code into a deployable binary. The full set of patterns (with code) lives in [`docs/usage-patterns.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/usage-patterns.md) — this page is the index.

## Pattern 1 — Local stdio MCP server

One binary per upstream API, launched by the MCP host (Claude Desktop, VS Code, Cursor, Inspector) over stdio.

```go
raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")
petmcp.RegisterSwaggerPetstoreClient(s, client)
_ = raw.Run(context.Background(), &mcp.StdioTransport{})
```

→ [examples/petstore](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/petstore)

## Pattern 2 — Remote MCP server (HTTP / SSE)

Same generated code, different transport. Used when the upstream API isn't reachable from the user's laptop, or the MCP server should be deployed once and shared.

```go
handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return raw }, nil)
log.Fatal(http.ListenAndServe(":8080", handler))
```

Generated code is transport-agnostic — it only calls `runtime.MCPServer.AddTool`.

## Pattern 3 — Backend swap (go-sdk ↔ mark3labs)

Change one import and one constructor line. No regeneration. See [MCP Backends](MCP-Backends).

## Pattern 4 — Multi-tenant / multi-environment namespacing

Run the same API twice (e.g. staging vs prod) in one MCP server, distinguished by tool-name prefix:

```go
petmcp.RegisterSwaggerPetstoreClient(s, stagingClient, runtime.WithNamePrefix("staging"))
petmcp.RegisterSwaggerPetstoreClient(s, prodClient,    runtime.WithNamePrefix("prod"))
```

## Pattern 5 — Aggregating multiple APIs

Call several `Register*` functions from one `main`:

```go
petmcp.RegisterSwaggerPetstoreClient(s, petClient)
billmcp.RegisterBillingAPIClient(s, billClient)
auditmcp.RegisterAuditAPIClient(s, auditClient)
```

## Pattern 6 — Auth injection via the oapi-codegen client

Auth lives in the HTTP client you pass to the `Register*` function. Bearer tokens, mTLS, dynamic credentials — all handled at the `oapi-codegen` layer:

```go
client, _ := pet.NewClientWithResponses(url,
    pet.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+freshToken())
        return nil
    }),
)
```

## Pattern 7 — Per-call context via `WithExtraProperties`

Inject tenant IDs, correlation tokens, or environment selectors that the LLM provides on every call. See [Runtime Options](Runtime-Options).

## Pattern 12 — Batch generation from a folder of specs

Point `-spec` at a directory and emit one tool set per spec. See [Batch Mode](Batch-Mode).

## More patterns

The rest (custom retry loops, OpenTelemetry tracing, request signing, audit logging) live in [`docs/usage-patterns.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/usage-patterns.md).
