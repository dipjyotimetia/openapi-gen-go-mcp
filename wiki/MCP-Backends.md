# MCP Backends

Generated code targets a thin `runtime.MCPServer` interface — a single `AddTool(Tool, ToolHandler)` method. Concrete MCP libraries live behind adapter packages. **Switching backends is a one-line import change in your `main`. No regeneration.**

## Supported backends

| Backend | Adapter | Server construction |
|---|---|---|
| [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) (official) | `pkg/runtime/gosdk` | `raw, s := gosdk.NewServer(name, version)` |
| [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) | `pkg/runtime/mark3labs` | `raw, s := mark3labs.NewServer(name, version)` |

## go-sdk (official)

```go
import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/dipjyotimetia/openapi-go-mcp/pkg/runtime/gosdk"
)

raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")
petmcp.RegisterSwaggerPetstoreClient(s, client)
_ = raw.Run(context.Background(), &mcp.StdioTransport{})
```

Working example: [`examples/petstore`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/petstore).

## mark3labs/mcp-go

```go
import (
    mcpserver "github.com/mark3labs/mcp-go/server"
    "github.com/dipjyotimetia/openapi-go-mcp/pkg/runtime/mark3labs"
)

raw, s := mark3labs.NewServer("petstore-mcp", "1.0.0")
petmcp.RegisterSwaggerPetstoreClient(s, client)   // unchanged
mcpserver.ServeStdio(raw)
```

Working example: [`examples/petstore-mark3labs`](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/petstore-mark3labs).

## Adding a new backend

Create `pkg/runtime/<libname>/` exporting two symbols:

- `NewServer(name, version string) (Raw, MCPServer)` — constructs the underlying server and an adapter that satisfies `runtime.MCPServer`.
- `Wrap(...)` — wraps an existing server instance with the adapter.

The generator never changes. Only the adapter package and your `main.go` know which library you're using.

The interface is small on purpose — see [Design Decisions §2](Design-Decisions) for the rationale.

## Why the abstraction

The MCP ecosystem is young. Library APIs are still in flux. Generated code stays stable while underlying libraries change because the surface area between them is exactly one method: `AddTool`.

## Related

- [Architecture](Architecture)
- [Design Decisions](Design-Decisions)
- [Runtime Options](Runtime-Options)
