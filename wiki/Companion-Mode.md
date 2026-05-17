# Companion Mode

Companion mode is the default. The generator emits **one Go source file** (`<pkg>.mcp.go`) that you import from your own `main`. You own the HTTP transport, the auth strategy, and the lifecycle.

## Two-step generation

```bash
# 1. Typed HTTP client from oapi-codegen.
oapi-codegen \
    -generate types,client \
    -package pet \
    -o gen/pet/pet.gen.go \
    petstore.yaml

# 2. MCP companion — imports the client above.
openapi-go-mcp \
    -spec petstore.yaml \
    -out gen/petmcp \
    -package petmcp \
    -client-import github.com/me/myrepo/gen/pet
```

Output: `gen/petmcp/petmcp.mcp.go`.

## A minimal `main.go`

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"

    "github.com/me/myrepo/gen/pet"
    "github.com/me/myrepo/gen/petmcp"
    "github.com/dipjyotimetia/openapi-go-mcp/pkg/runtime/gosdk"
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

The exported `RegisterSwaggerPetstoreClient` function name is derived from the spec's `info.title` — your spec will produce a different name. Run `openapi-go-mcp -list` to see operations and inferred names without writing files.

## Customizing the HTTP client

This is the whole point of companion mode. You build the `oapi-codegen` client with whatever middleware you need:

```go
hc := &http.Client{
    Timeout:   30 * time.Second,
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

client, err := pet.NewClientWithResponses(
    "https://api.example.com",
    pet.WithHTTPClient(hc),
    pet.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
        req.Header.Set("Authorization", "Bearer "+freshToken())
        return nil
    }),
)
```

Retries, tracing, mTLS, request signing, dynamic auth — all live in your HTTP client, transparent to the MCP layer.

## Aggregating multiple APIs

Register more than one client against the same MCP server:

```go
raw, s := gosdk.NewServer("acme-mcp", "1.0.0")

petClient, _ := pet.NewClientWithResponses(petURL)
petmcp.RegisterSwaggerPetstoreClient(s, petClient)

billClient, _ := bill.NewClientWithResponses(billURL)
billmcp.RegisterBillingAPIClient(s, billClient)

_ = raw.Run(context.Background(), &mcp.StdioTransport{})
```

If two APIs share an operation name, use [`runtime.WithNamePrefix`](Runtime-Options) to disambiguate.

## When to use

- You need custom HTTP behavior (retries, OpenTelemetry, mTLS, dynamic auth).
- MCP is one feature of a larger service binary.
- You want to register multiple APIs in one MCP server.
- You need to do something between argument decoding and the upstream call (rate limiting, audit logging).

## When to use [Proxy Mode](Proxy-Mode) instead

- You just want to run a server. No custom Go code.
- Auth is a static env-var credential per environment.

## Related

- [CLI Reference](CLI-Reference)
- [Runtime Options](Runtime-Options)
- [MCP Backends](MCP-Backends)
- [Deployment Patterns](Deployment-Patterns)
