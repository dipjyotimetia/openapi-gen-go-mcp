# Troubleshooting

Common errors and what they mean.

## Generation errors

### `output file already exists; use -force to overwrite`

The generator refuses to overwrite by default. Either delete the file, or pass `-force`. This is a guardrail against accidentally clobbering hand-edited code.

### `-package is not allowed in batch mode`

Batch mode auto-derives package names from filename stems. Drop the `-package` flag.

### `-emit-v3 is not allowed in batch mode`

`-emit-v3` writes to a single path; it's incompatible with multi-spec input. Run it once per spec.

### `slug collision: <slug> from <pathA> and <pathB>`

Two specs in your batch reduce to the same slug (e.g. `v1/api.yaml` and `v2/api.yaml` both → `api`). Rename one of the files so their stems differ.

### `package name cannot begin with a digit`

The filename stem for batch mode starts with a digit, which isn't a valid Go identifier. Rename the spec file (e.g. `2024-billing.yaml` → `billing-2024.yaml`).

### `failed to convert Swagger 2.0 to OpenAPI 3`

The conversion via `kin-openapi/openapi2conv` failed. Common causes: malformed v2 spec, unresolved external `$ref`. Validate the v2 spec first with a tool like `swagger-cli`.

### `openapi3.Validate: ...`

The spec — after any v2→v3 conversion — failed validation. The error message points at the offending node.

## Build errors in generated code

### `cannot find package ".../gen/pet"`

Your `-client-import` value doesn't match where you actually put the `oapi-codegen` output. Either move the output, or fix the import path.

### `<Op>WithResponse undefined`

The `oapi-codegen` client doesn't expose the operation. Make sure you generated with `-generate types,client` (not just `types`), and that the operation isn't excluded by your `oapi-codegen` config.

### `ClientWithResponsesInterface is not implemented`

You passed a concrete client, but the operation needs the interface. Use `oapi-codegen`'s `*ClientWithResponses` (the generated type), or set `-client-type` to your custom interface name.

## Runtime errors

### Empty tool list when client connects

Most often: the binary started, but `RegisterXxxClient` was never called. Check your `main.go`.

Less often: `-exclude-by-default` was set and no operations were tagged `x-mcp: true`.

### `401 Unauthorized` from the upstream

- **Proxy mode:** check the env var matches what the spec's `securitySchemes` resolved to. The generated `README.md` lists the exact name per scheme.
- **Companion mode:** confirm your `WithRequestEditorFn` or `WithHTTPClient` actually adds the auth header.

### Tool call validates locally but the LLM keeps misformatting it

The LLM may need a clearer schema. If you're using a model that struggles with `$ref` or composition keywords, try `-openai-compat` to flatten the schema. See [Schema Modes](Schema-Modes).

## Spec quirks

### Multiple response content types break `oapi-codegen` v2.7.0

Use `-emit-v3` to prune non-JSON responses before passing the spec to `oapi-codegen`:

```bash
openapi-go-mcp -spec api.yaml -emit-v3 api-v3.yaml
oapi-codegen -generate types,client -package api -o gen/api/api.gen.go api-v3.yaml
```

### Request body declares multiple content types

The generator picks one deterministically: `application/json` → form → multipart → octet-stream → text/* → other. Override with `-prefer-content-type`.

## Still stuck?

- Run `openapi-go-mcp -list -spec your.yaml` to see what the generator sees.
- Check the [examples](Examples) for a working spec close to yours.
- Open an issue with the spec snippet + command + observed output: [issues](https://github.com/dipjyotimetia/openapi-go-mcp/issues).
