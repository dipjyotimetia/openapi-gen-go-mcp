# Batch Mode

Generate MCP tools from many specs in a single invocation. Triggers automatically when `-spec` resolves to more than one file.

## What activates it

`-spec` accepts any of:

- A single file (`apis/billing.yaml`) — single-spec mode.
- A directory (`apis/`) — **recursively** walked; `.yaml`, `.yml`, `.json` files are matched.
- A `filepath.Glob` pattern (`apis/*.yaml`) — single-level glob; no `**` in v1, use a directory for recursion.
- A comma-separated list of any of the above (`core/,extras/audit.yaml`).

When the value resolves to more than one match, batch mode activates.

## Output layout

Each spec is rendered into its own subdirectory under `-out`:

```
gen/
├── billingmcp/
│   └── billingmcp.mcp.go
├── petstoremcp/
│   └── petstoremcp.mcp.go
└── auditmcp/
    └── auditmcp.mcp.go
```

The **slug** is derived from the filename stem (`billing-api.yaml` → `billingapi`). Slugs become:

- The output directory: `<out>/<slug>mcp/`
- The Go package name: `<slug>mcp`
- The appended segment for `-client-import` (see below)

Stems that start with a digit are rejected — Go package names can't begin with a digit.

## `-client-import` in batch mode

The value is treated as a **base path**. The slug is appended with forward slashes:

```bash
-spec apis/ -client-import github.com/acme/apis/gen
```

For `billing.yaml`:
```
import path → github.com/acme/apis/gen/billing
```

## Rejected flags

These are incompatible with batch mode and fail with exit code `2`:

- `-package` — packages are auto-derived from filename stems.
- `-emit-v3` — single-spec only (it writes one output path).

## Slug collisions

Two specs with the same filename stem (e.g. `v1/api.yaml` and `v2/api.yaml` both → `api`) are detected **before any file is written**. The run aborts with a clear error pointing at both source paths.

## Error handling

Failures don't stop the run:

- Each spec is processed independently.
- Every error is accumulated and reported at end.
- The process exits with code `3` if any spec failed.

This means one CI run sees every failing spec, not just the first.

## `-list` in batch mode

Output is grouped under headers per spec:

```
=== apis/billing.yaml ===
- billingmcp.GetInvoice            GET    /invoices/{id}
- billingmcp.CreateInvoice         POST   /invoices

=== apis/petstore.yaml ===
- petstoremcp.GetPet               GET    /pets/{petId}
...
```

## Examples

```bash
# Recursive directory
openapi-go-mcp \
    -spec apis/ \
    -out gen \
    -client-import github.com/acme/apis/gen \
    -force

# Glob
openapi-go-mcp \
    -spec 'apis/*.yaml' \
    -out gen \
    -client-import github.com/acme/apis/gen

# Mixed inputs, comma-separated
openapi-go-mcp \
    -spec 'core/,extras/audit.yaml' \
    -out gen \
    -client-import example.com/g
```

## Related

- [Usage patterns — Pattern 12](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/usage-patterns.md) (full walkthrough)
- [CLI Reference](CLI-Reference)
