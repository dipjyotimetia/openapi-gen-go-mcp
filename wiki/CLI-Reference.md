# CLI Reference

```
openapi-go-mcp [flags]
```

## Required

| Flag | Description |
|---|---|
| `-spec PATH` | OpenAPI 3.x / Swagger 2.0 source. Accepts a single file, `http(s)://` URL, directory (recursively walked), `filepath.Glob` pattern, or comma-separated list of any of those. When the value matches multiple specs, **batch mode** activates. |
| `-client-import PATH` | Import path of the `oapi-codegen` output package. In batch mode this is a base path; the slug is appended with forward slashes. Not required with `-list` or `-emit-v3`. Not used by `-mode=proxy`. |

## Output

| Flag | Default | Description |
|---|---|---|
| `-out DIR` | `./mcp` | Output directory. In batch mode this is the base; each spec lands in `<out>/<slug>mcp/`. |
| `-package NAME` | derived from spec title | Go package name. **Rejected in batch mode** — package names are auto-derived from filename stems. |
| `-client-type NAME` | `ClientWithResponsesInterface` | The interface name exported by your `oapi-codegen` package. |
| `-force` | off | Overwrite the generated `*.mcp.go` file if it exists. Without this, an existing file is a fatal error. |

## Modes

| Flag | Description |
|---|---|
| `-mode=proxy` | Emit a full runnable Go module (`main.go` + `go.mod` + `<pkg>/<pkg>.mcp.go` + `README.md`). See [Proxy Mode](Proxy-Mode). |
| `-module PATH` | Go module path for proxy mode. Required when `-mode=proxy`. |

## Schema

| Flag | Description |
|---|---|
| `-openai-compat` | Emit OpenAI-tool-compatible JSON Schema: `$ref`-free, `oneOf`/`anyOf` collapsed to first branch, `allOf` shallow-merged, every object gets `additionalProperties: false`. See [Schema Modes](Schema-Modes). |
| `-name-prefix PREFIX` | Static prefix added to every tool name. Useful when registering the same API more than once. |
| `-prefer-content-type CT` | Pick this content type for the request body when an operation declares multiple. Overrides the default JSON → form → multipart → octet → text → xml priority. |

## Filtering

| Flag | Description |
|---|---|
| `-exclude-by-default` | Invert `x-mcp` filtering: only operations explicitly tagged `x-mcp: true` are generated. See [x-mcp Filtering](x-mcp-Filtering). |

## Utilities

| Flag | Description |
|---|---|
| `-list` | Print the operations found in the spec and exit. In batch mode, output is grouped under `=== <path> ===` headers per spec. |
| `-emit-v3 PATH` | Write the spec as OpenAPI 3 YAML to `PATH`. Helper for Swagger 2.0 → 3 conversion. Also prunes non-JSON content types (workaround for `oapi-codegen` v2.7.0). |
| `-warnings-as-errors` | Exit non-zero when any warning-level diagnostic fires. |
| `-version` | Print version information and exit. |

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Generic error |
| `2` | Usage error (bad flags, conflicting flags) |
| `3` | Generation failure for one or more specs (batch mode reports all, then exits 3) |

## Examples

```bash
# Companion mode — single spec
openapi-go-mcp \
    -spec petstore.yaml \
    -out gen/petmcp \
    -package petmcp \
    -client-import github.com/me/myrepo/gen/pet

# Proxy mode — single spec
openapi-go-mcp \
    -mode=proxy \
    -spec petstore.yaml \
    -out gen/petstore-mcp \
    -module github.com/me/petstore-mcp

# Batch mode — every spec under apis/
openapi-go-mcp \
    -spec apis/ \
    -out gen \
    -client-import github.com/acme/apis/gen \
    -force

# OpenAI-strict schema
openapi-go-mcp -spec petstore.yaml -out gen/petmcp -package petmcp \
    -client-import github.com/me/gen/pet -openai-compat

# Just list the operations
openapi-go-mcp -spec petstore.yaml -list

# Convert Swagger 2.0 to OpenAPI 3 YAML
openapi-go-mcp -spec petstore-v2.json -emit-v3 petstore-v3.yaml
```

## Flag compatibility matrix

|  | Single spec | Batch | `-list` | `-emit-v3` | `-mode=proxy` |
|---|---|---|---|---|---|
| `-package` | ✅ | ❌ | ✅ | n/a | ✅ |
| `-client-import` | required | required (base) | optional | n/a | n/a |
| `-out` | ✅ | ✅ | n/a | n/a | ✅ |
| `-emit-v3` | ✅ | ❌ | n/a | — | ❌ |
| `-openai-compat` | ✅ | ✅ | n/a | n/a | ✅ |
