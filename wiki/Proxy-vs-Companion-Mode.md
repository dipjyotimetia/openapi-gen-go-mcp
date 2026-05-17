# Proxy vs Companion Mode

Pick a mode based on **who owns the binary** and **what auth looks like**.

| | Proxy | Companion |
|---|---|---|
| Output | full Go module (`main.go` + `go.mod` + pkg + README) | one `<pkg>.mcp.go` file |
| Build | `go mod tidy && go build` in the generated dir | you write `main.go`, you build |
| HTTP client | embedded, not customizable | yours — bring any `oapi-codegen` client |
| Auth | env-var → matched against spec's `securitySchemes` | your code; whatever you want |
| Aggregating APIs | one spec per server | many specs, one server |
| Best for | turnkey servers, mass deployment | embedded MCP, custom transport, complex auth |
| Time to running server | 30 seconds | minutes (you write the main) |

## Decision tree

```
Do you need custom HTTP behavior (retries, tracing, mTLS, dynamic auth)?
├── yes → Companion
└── no
    │
    ▼
Are you embedding MCP into an existing Go service?
├── yes → Companion
└── no
    │
    ▼
Do you need to expose multiple APIs from one MCP server?
├── yes → Companion (call multiple Register* functions in your main)
└── no  → Proxy
```

## Switching between them

The generated `*.mcp.go` is identical in both modes — proxy mode just adds `main.go` + `go.mod` + `README.md` around it. You can start in proxy mode and "graduate" to companion mode by deleting the wrapper files and writing your own `main.go` against the same `*.mcp.go`.

## Related

- [Proxy Mode](Proxy-Mode)
- [Companion Mode](Companion-Mode)
- [Deployment Patterns](Deployment-Patterns)
