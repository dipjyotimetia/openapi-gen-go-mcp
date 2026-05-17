# Getting Started

Two paths. Pick **Proxy mode** if you just want a runnable MCP server. Pick **Companion mode** if you're embedding MCP into an existing Go binary.

## Install

| Channel | Command |
|---|---|
| Homebrew | `brew install dipjyotimetia/tap/openapi-go-mcp` |
| Pre-built binary | Download from [releases](https://github.com/dipjyotimetia/openapi-go-mcp/releases/latest), unpack, put on `PATH` |
| Container | `docker pull ghcr.io/dipjyotimetia/openapi-go-mcp:latest` |
| Go (1.26+) | `go install github.com/dipjyotimetia/openapi-go-mcp/cmd/openapi-go-mcp@latest` |

Verify: `openapi-go-mcp -version`. Generated code itself compiles against Go 1.23+.

## Proxy mode

One command. Get a full Go module with `main.go`, `go.mod`, the generated `<pkg>.mcp.go`, and a `README.md`.

```bash
openapi-go-mcp \
    -mode=proxy \
    -spec petstore.yaml \
    -out gen/petstore-mcp \
    -module github.com/me/petstore-mcp

cd gen/petstore-mcp
go mod tidy
go build
./petstore-mcp        # serves MCP on stdio
```

The proxy reads auth credentials from environment variables derived from the spec's `securitySchemes`. The generated `README.md` lists the exact env var per scheme. Override the upstream base URL with `API_BASE_URL`.

→ Full details: **[Proxy Mode](Proxy-Mode)**

## Companion mode

Two codegen steps; you own `main.go` and the HTTP transport.

```bash
# 1. Typed HTTP client from oapi-codegen.
oapi-codegen -generate types,client -package pet -o gen/pet/pet.gen.go petstore.yaml

# 2. MCP companion that imports the client.
openapi-go-mcp \
    -spec petstore.yaml \
    -out gen/petmcp \
    -package petmcp \
    -client-import github.com/me/myrepo/gen/pet
```

```go
package main

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/me/myrepo/gen/pet"
    "github.com/me/myrepo/gen/petmcp"
    "github.com/dipjyotimetia/openapi-go-mcp/pkg/runtime/gosdk"
)

func main() {
    client, _ := pet.NewClientWithResponses("https://api.example.com")
    raw, s := gosdk.NewServer("petstore-mcp", "1.0.0")
    petmcp.RegisterSwaggerPetstoreClient(s, client)
    _ = raw.Run(context.Background(), &mcp.StdioTransport{})
}
```

→ Full details: **[Companion Mode](Companion-Mode)**

## Connect a client

Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "petstore": {
      "command": "/absolute/path/to/petstore-mcp",
      "env": { "API_BASE_URL": "https://api.example.com" }
    }
  }
}
```

For Claude Code, Cursor, VS Code, and Inspector configs see the [todos example README](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/examples/todos/README.md).

## Next steps

- [CLI Reference](CLI-Reference) — every flag explained
- [Examples](Examples) — working repos you can clone
- [Deployment Patterns](Deployment-Patterns) — stdio, HTTP, SSE, multi-tenant, auth injection
- [Filtering with x-mcp](x-mcp-Filtering) — curate which operations are exposed
