# Runtime Options

Options passed to the generated `RegisterXxxClient` function to customize tool registration at runtime. No regeneration needed.

```go
import "github.com/dipjyotimetia/openapi-go-mcp/pkg/runtime"

petmcp.RegisterSwaggerPetstoreClient(s, client,
    runtime.WithNamePrefix("staging"),
    runtime.WithExtraProperties(...),
)
```

## `WithNamePrefix(prefix string)`

Prepend a static prefix to every tool name registered by this call. Useful when:

- The same API is registered twice in one server (e.g. staging + prod).
- You're aggregating multiple APIs and want a namespace per upstream.

```go
petmcp.RegisterSwaggerPetstoreClient(s, stagingClient, runtime.WithNamePrefix("staging"))
petmcp.RegisterSwaggerPetstoreClient(s, prodClient,    runtime.WithNamePrefix("prod"))
```

Tool `getPet` becomes `staging_getPet` and `prod_getPet`.

## `WithExtraProperties(props ...ExtraProperty)`

Inject **per-call** context properties into every tool's input schema. The LLM is asked to provide them on every call; your code reads them from the decoded args.

Typical uses: tenant ID, request correlation token, environment selector.

```go
petmcp.RegisterSwaggerPetstoreClient(s, client,
    runtime.WithExtraProperties(
        runtime.ExtraProperty{
            Name:        "tenant",
            Description: "Tenant ID",
            Required:    true,
        },
    ),
)
```

The property appears at the top of the input schema (sibling to `path`/`query`/`header`/`body`) and is available to tool handlers via the decoded arguments. If `Required: true` is set, calls without it are rejected before the upstream HTTP request fires.

## CLI equivalents

Some options are also expressible at generation time:

| Runtime option | CLI flag | Difference |
|---|---|---|
| `WithNamePrefix` | `-name-prefix` | CLI bakes the prefix into the generated code (less flexible). Runtime lets you choose per-registration. |

There is no CLI equivalent for `WithExtraProperties` — extra properties are inherently a runtime concern.

## Adding your own options

Options follow the standard Go functional-options pattern. Adding one means:

1. Define `WithFoo(...) Option` in `pkg/runtime/`.
2. Update `ApplyConfig` to handle the new field.
3. Use it from your `main` — the generated code calls `ApplyConfig` and the option flows through.

The generated `Register*` function does not change.

## Related

- [Companion Mode](Companion-Mode)
- [Deployment Patterns](Deployment-Patterns)
