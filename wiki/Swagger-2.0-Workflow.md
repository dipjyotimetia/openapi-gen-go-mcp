# Swagger 2.0 Workflow

`openapi-go-mcp` accepts Swagger 2.0 input natively — it's auto-converted via [`kin-openapi/openapi2conv`](https://github.com/getkin/kin-openapi) before generation.

`oapi-codegen` does **not** accept Swagger 2.0. So when your spec is v2, you need a small extra step.

## Three-step workflow

```bash
# 1. Convert Swagger 2.0 → OpenAPI 3 YAML.
openapi-go-mcp -spec petstore-v2.json -emit-v3 petstore-v3.yaml

# 2. Generate the typed HTTP client from the v3 output.
oapi-codegen -generate types,client -package pet -o gen/pet/pet.gen.go petstore-v3.yaml

# 3. Generate the MCP companion from the v3 output (or the v2 — both work here).
openapi-go-mcp \
    -spec petstore-v3.yaml \
    -out gen/petmcp \
    -package petmcp \
    -client-import github.com/me/gen/pet
```

Use the v3 YAML for both `oapi-codegen` and `openapi-go-mcp` so they see exactly the same shape.

## What `-emit-v3` does

1. Loads the Swagger 2.0 spec.
2. Converts it via `openapi2conv` on a **deep clone** — the original file is not mutated.
3. **Prunes non-JSON content types from responses.** This is a workaround for [oapi-codegen v2.7.0](https://github.com/oapi-codegen/oapi-codegen) issues with responses exposed under multiple content types.
4. Writes the result as YAML to the path you pass.

## When you can skip `-emit-v3`

If you only need the MCP companion and not an `oapi-codegen` client, point `openapi-go-mcp` directly at the Swagger 2.0 file — internal conversion is automatic. `-emit-v3` only matters when something else downstream (like `oapi-codegen`) consumes the spec.

## Validation

Every loaded spec — v2 or v3 — passes `openapi3.Validate` after conversion. External `$ref`s resolve against the spec file's directory.

## Related

- [CLI Reference](CLI-Reference#utilities)
- [Getting Started](Getting-Started)
- [examples/swagger2-petstore](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/swagger2-petstore)
- [examples/library](https://github.com/dipjyotimetia/openapi-go-mcp/tree/main/examples/library)
