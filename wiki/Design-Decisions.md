# Design Decisions

Summaries of the non-obvious choices. For each, see the full entry in [`docs/design-decisions.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/design-decisions.md) — it includes the alternatives we rejected and the cost of each choice.

## 1. Companion codegen, not runtime introspection

We generate a `*.mcp.go` file at build time. Alternative was a runtime library that parses an OpenAPI spec on startup.

**Why codegen wins:** compile-time safety, reviewable diffs, zero cold-start cost, symmetry with the existing `oapi-codegen` flow.

**Cost:** spec changes require regeneration.

## 2. `runtime.MCPServer` interface, not vendor lock-in

Generated code targets a small interface. Concrete MCP libraries live behind thin adapter packages.

**Why the abstraction:** the MCP ecosystem is young and unsettled; pinning to one library would force every user to rev when it breaks API. The interface shifts that cost into a ~50-line adapter file.

**Cost:** one tiny indirection layer. Generated code can't use library-specific features.

## 3. Grouped input schema (`path` / `query` / `header` / `body`), not flat

Every tool's input is an object with up to four sub-objects, one per OpenAPI parameter location.

**Why grouped:** prevents name collisions between a path param and a body field; mirrors the HTTP wire shape (improves LLM tool-use accuracy); maps 1:1 onto the runtime decoders.

**Cost:** one extra level of nesting. Empty groups are omitted.

## 4. Per-operation `$defs`, not a single shared pool

Each tool's schema carries its own `$defs` with only the components reachable from that operation.

**Why per-operation:** tool schemas are independent units the LLM sees in isolation; supports different dialects per tool (`-openai-compat` can flatten one without touching another).

**Cost:** repetition. Mitigated by a shared `nameByPtr` cache so it's bytes-only, not CPU.

## 5. JSON Schema with `$ref` by default; OpenAI-compat opt-in

Default emits draft-07 JSON Schema. `-openai-compat` is a lossy dialect that collapses `oneOf`/`anyOf` to the first branch and shallow-merges `allOf`.

**Why the richer default:** loss is one-way. Most MCP clients accept `$ref`. Explicit opt-in matches the contract — if you ask for a lossy dialect, you get one.

## 6. JSON-only response decoding

Only `application/json` response bodies are decoded into structured tool results. Other content types are surfaced as raw bytes.

**Why:** LLMs benefit from structured responses; non-JSON responses are the exception, not the rule.

**Cost:** users who need typed binary/XML response handling have to write a handler wrapper.

## See also

- [Architecture](Architecture) — the *how* behind these decisions
- [Schema Modes](Schema-Modes) — the user-facing surface of decision 4 & 5
- [MCP Backends](MCP-Backends) — decision 2 in practice
- Full doc with alternatives and costs: [`docs/design-decisions.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/design-decisions.md)
