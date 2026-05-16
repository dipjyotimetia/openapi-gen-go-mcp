# Usage patterns

`openapi-gen-go-mcp` doesn't ship a server — it generates one. The generated `*.mcp.go` registers every OpenAPI operation as an MCP tool that forwards to an `oapi-codegen` typed HTTP client. Each pattern below is a different way of assembling those pieces into a deployable binary.

> All examples assume you've already run the two-step codegen for your spec:
> ```bash
> oapi-codegen -generate types,client -package pet -o gen/pet/pet.gen.go petstore.yaml
> openapi-gen-go-mcp -spec petstore.yaml -out gen/petmcp -package petmcp \
>     -client-import github.com/me/myrepo/gen/pet
> ```

## Pattern 1 — Local stdio MCP server (Claude Desktop, IDEs)

The default. One binary per upstream API, launched by the MCP host over stdio.

```go
// main.go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"

    "github.com/me/myrepo/gen/pet"
    "github.com/me/myrepo/gen/petmcp"
    "github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/runtime/gosdk"
)

func main() {
    client, err := pet.NewClientWithResponses(os.Getenv("PETSTORE_BASE_URL"))
    if err != nil {
        log.Fatal(err)
    }

    raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")
    petmcp.RegisterSwaggerPetstoreClient(s, client)

    if err := raw.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "petstore": {
      "command": "/usr/local/bin/petstore-mcp",
      "env": { "PETSTORE_BASE_URL": "https://api.example.com" }
    }
  }
}
```

See [`examples/petstore`](../examples/petstore) for the working version.

## Pattern 2 — Remote MCP server (HTTP / SSE)

Same generated code, different transport. Useful when the upstream API can't be reached from the user's laptop, or when the MCP server should be deployed once and shared.

```go
raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")
petmcp.RegisterSwaggerPetstoreClient(s, client)

// Replace StdioTransport with the go-sdk's HTTP/SSE transport:
handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return raw }, nil)
log.Fatal(http.ListenAndServe(":8080", handler))
```

Generated code is transport-agnostic — it only calls `runtime.MCPServer.AddTool`. The transport lives entirely in `main`.

## Pattern 3 — Backend swap (go-sdk ↔ mark3labs)

Change one import and one constructor line:

```go
// Was:
//   "github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/runtime/gosdk"
//   raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")

import "github.com/dipjyotimetia/openapi-gen-go-mcp/pkg/runtime/mark3labs"
import mcpserver "github.com/mark3labs/mcp-go/server"

raw, s := mark3labs.NewServer("petstore-mcp", "1.0.0")
petmcp.RegisterSwaggerPetstoreClient(s, client)   // unchanged
mcpserver.ServeStdio(raw)
```

No regeneration needed. See [`examples/petstore-mark3labs/main.go`](../examples/petstore-mark3labs/main.go).

## Pattern 4 — Multi-tenant / multi-environment namespacing

Run two instances of the same API (e.g., staging vs prod) inside one MCP server, distinguished by tool-name prefix:

```go
prodClient, _ := pet.NewClientWithResponses("https://api.example.com")
stagingClient, _ := pet.NewClientWithResponses("https://staging.api.example.com")

raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")

petmcp.RegisterSwaggerPetstoreClient(s, prodClient,
    runtime.WithNamePrefix("prod"))     // tools: prod_addPet, prod_findPetById, ...
petmcp.RegisterSwaggerPetstoreClient(s, stagingClient,
    runtime.WithNamePrefix("staging"))  // tools: staging_addPet, staging_findPetById, ...
```

The prefix can also be baked in at generation time with `-name-prefix` if it never changes per-deployment.

## Pattern 5 — Per-call auth injection (multi-tenant API tokens)

Use `WithExtraProperties` to add a schema field that the LLM must supply on every tool call. The value is removed from the args and placed on the request context:

```go
type ctxKey string
const tokenKey ctxKey = "tenant-token"

petmcp.RegisterSwaggerPetstoreClient(s, client,
    runtime.WithExtraProperties(runtime.ExtraProperty{
        Name:        "tenant_token",
        Description: "Bearer token for the calling tenant",
        Required:    true,
        ContextKey:  tokenKey,
    }),
)
```

Then read the token in an `oapi-codegen` request editor (configured on the client) and add it to the outbound `Authorization` header. The token never lives in process memory between calls.

## Pattern 6 — Static API key (single-tenant)

For single-tenant cases, use `oapi-codegen`'s standard request editor instead of an extra property — the LLM never sees the key:

```go
client, err := pet.NewClientWithResponses("https://api.example.com",
    pet.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+os.Getenv("API_KEY"))
        return nil
    }),
)
```

This is the right pattern when the deployment itself owns the credential.

## Pattern 7 — Strict-mode schema for OpenAI tool calls

OpenAI's tool-call validator rejects `$ref`, `oneOf`/`anyOf`/`allOf`, and open-ended objects. Generate with `-openai-compat`:

```bash
openapi-gen-go-mcp \
    -spec petstore.yaml -out gen/petmcp -package petmcp \
    -client-import github.com/me/myrepo/gen/pet \
    -openai-compat
```

The generator inlines `$defs`, collapses `oneOf`/`anyOf` to the first branch, shallow-merges `allOf`, and adds `additionalProperties: false` to every object. See [`architecture.md`](architecture.md#openai-compatibility-mode--openai-compat) for the full set of transforms.

> Lossy by design — composition keywords lose alternatives. Use only when targeting strict-schema validators.

## Pattern 8 — Sidecar to an internal service

Deploy the generated MCP server in the same pod as the upstream service and point it at `http://localhost:<port>`. The MCP server is what the LLM sees; the service itself stays internal.

```go
client, _ := internal.NewClientWithResponses("http://127.0.0.1:8080")
raw, s := gosdk.NewServer("internal-api-mcp", "1.0.0")
internalmcp.RegisterClient(s, client)
raw.Run(ctx, &mcp.StdioTransport{}) // or HTTP transport on a different port
```

This avoids exposing the upstream service to the network while still letting an LLM reach it.

## Pattern 9 — Swagger 2.0 input

`oapi-codegen` rejects Swagger 2.0. Convert in-process first, then run both codegens against the converted v3:

```bash
openapi-gen-go-mcp -spec petstore-v2.json -emit-v3 petstore-v3.yaml
oapi-codegen -generate types,client -package pet -o gen/pet/pet.gen.go petstore-v3.yaml
openapi-gen-go-mcp -spec petstore-v3.yaml -out gen/petmcp -package petmcp \
    -client-import github.com/me/myrepo/gen/pet
```

`-emit-v3` also prunes non-JSON content types on a deep clone — works around an `oapi-codegen` v2.7.0 bug with responses exposed under multiple content types. See [`examples/swagger2-petstore`](../examples/swagger2-petstore).

## Pattern 10 — Aggregating multiple APIs in one MCP server

Register more than one client against the same server. Use prefixes to keep tool names unambiguous:

```go
raw, s := gosdk.NewServer("aggregated-mcp", "1.0.0")

petmcp.RegisterSwaggerPetstoreClient(s, petClient,
    runtime.WithNamePrefix("pet"))
ordersmcp.RegisterOrdersClient(s, orderClient,
    runtime.WithNamePrefix("orders"))
billingmcp.RegisterBillingClient(s, billingClient,
    runtime.WithNamePrefix("billing"))

raw.Run(ctx, &mcp.StdioTransport{})
```

One binary, many upstream APIs, one MCP endpoint for the LLM.

## Choosing a pattern

| If you want… | Use |
|---|---|
| LLM in Claude Desktop calls a third-party API | Pattern 1 + Pattern 6 |
| Remote MCP server shared by a team | Pattern 2 |
| Same API at staging + prod in one binary | Pattern 4 |
| LLM acts on behalf of different end-users | Pattern 5 |
| Target OpenAI's tool-call validator | Pattern 7 |
| Expose an internal service to an LLM | Pattern 8 |
| Run a Swagger 2.0 spec | Pattern 9 |
| Aggregate several APIs behind one MCP endpoint | Pattern 10 |

For the reasoning behind the architectural choices these patterns rely on, see [`design-decisions.md`](design-decisions.md).
