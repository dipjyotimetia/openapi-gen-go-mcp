# Schema Modes

Each MCP tool ships with a JSON Schema describing its inputs. `openapi-go-mcp` emits this schema from the OpenAPI spec, with two dialects.

## Default — draft-07 JSON Schema with `$defs`

The richer of the two. Preserves spec semantics: `$ref`, `oneOf`, `anyOf`, `allOf`, recursion. Each tool's `$defs` is **self-contained** — the components reachable from that operation, nothing else.

Use this with Claude, the official MCP go-sdk, mark3labs/mcp-go, and any modern validator that understands draft-07.

```json
{
  "type": "object",
  "properties": {
    "path":   { "type": "object", "properties": { "petId": { "type": "integer" } }, "required": ["petId"] },
    "query":  { "type": "object", "properties": { "limit": { "type": "integer" } } },
    "header": { "type": "object", "properties": { "X-Trace-Id": { "type": "string" } } },
    "body":   { "$ref": "#/$defs/NewPet" }
  },
  "required": ["path", "body"],
  "$defs": { "NewPet": { ... } }
}
```

Empty groups (`path`, `query`, `header`, `body`) are omitted so the model isn't asked to fill them.

## `-openai-compat` — flattened, `$ref`-free

OpenAI's strict tool-call schema validator rejects `$ref` and most composition keywords. `-openai-compat` lowers the spec into a shape that validator accepts:

- `$ref` inlined everywhere
- `oneOf` / `anyOf` collapsed to the **first branch**
- `allOf` shallow-merged
- every object carries `additionalProperties: false`

This loses information. If your spec relies on union types, you'll lose all but the first branch. That's why it's an opt-in flag, not the default.

## Per-operation `$defs`

Each tool's schema is independent. Two tools that both reference `Pet` get their own copy of the `Pet` definition in `$defs`. The generator shares a `nameByPtr` map across operations during one `CollectOperations` call so this duplication is in **output bytes**, not in CPU cost.

This matters because:
- MCP tool schemas are isolated units — the LLM sees one schema at a time when picking a tool.
- `-openai-compat` can inline one tool's `$defs` without affecting another's.
- Tools that touch few components don't ship the entire spec's schema graph.

## Why this is grouped (`path` / `query` / `header` / `body`)

The four-group input structure mirrors the HTTP wire shape. Reasons:

- **Name collisions are real.** A path param `id` and a body field `id` would collide under a flat schema.
- **LLM tool-use accuracy improves** when the call structure matches an HTTP request — the model's reasoning maps cleanly onto the wire.
- **Decoder simplicity.** `runtime.DecodePathParam` / `DecodeBody` / `DecodeQueryParams` each handle one group; no merge logic.

## Adding a new schema mode

Threading a flag through:

1. Add to `generator.Options`.
2. Pass through `Render` → `CollectOperations` → `NewSchemaConverter`.
3. If it affects envelope objects (`root`, `path`/`query`/`header` groups), update `buildInputSchema` in `pkg/generator/operation.go`.

## Related

- [Design Decisions](Design-Decisions) — full rationale for the default-vs-strict split
- [CLI Reference](CLI-Reference#schema)
