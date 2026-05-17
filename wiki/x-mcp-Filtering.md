# Filtering with `x-mcp`

Not every operation in a spec should be exposed as an MCP tool. The `x-mcp` extension lets spec authors control which operations are generated, at three levels of precedence.

## Precedence

```
operation  >  path-item  >  document  >  CLI default
```

The **most specific** declaration wins. If an operation has `x-mcp: true`, it's always generated, regardless of the path's or document's setting.

## Default behavior

By default, **every operation is generated** unless explicitly excluded with `x-mcp: false`.

## `-exclude-by-default` (curated exposure)

Pass this flag to **invert** the default: nothing is generated unless explicitly opted in with `x-mcp: true`. Useful when only a small curated subset of a large API should be exposed to an LLM.

## Examples

### Exclude an admin section, but keep one safe operation

```yaml
paths:
  /admin:
    x-mcp: false             # exclude every operation under /admin …
    delete:
      operationId: purgeAll
    get:
      operationId: listAdmins
      x-mcp: true            # … except this one
```

### Opt-in only mode

```bash
openapi-go-mcp -spec api.yaml -exclude-by-default \
    -out gen/apimcp -package apimcp \
    -client-import github.com/me/gen/api
```

```yaml
paths:
  /pets:
    get:
      operationId: listPets
      x-mcp: true            # generated
    post:
      operationId: createPet # not generated — no x-mcp: true
```

### Document-level default with selective overrides

```yaml
x-mcp: false                 # nothing by default
paths:
  /healthz:
    get:
      operationId: health
      x-mcp: true            # opted in
```

## Diagnostics

- **Excluded operations** are reported as info diagnostics during generation.
- **Typos** (`x-mcp: maybe`, `x-mcp: 1`) become warnings — they don't slip past review.
- Use `-warnings-as-errors` in CI to fail builds on typos.

## Related

- [CLI Reference](CLI-Reference#filtering)
