# Roadmap

Tracked here so design-decisions.md and architecture.md can point at a real file. No timelines; items move into a release when picked up.

## Generator

- [ ] **Dynamic (no-codegen) registration path** — accept a spec at runtime and register tools via reflection / a `Register(spec, client)` library entry point, in addition to the current AOT codegen. See `docs/design-decisions.md` decision #1 for the trade-off.
- [ ] **`discriminator` mapping** — currently dropped during JSON-Schema conversion. Either translate to `oneOf` + `if/then/else` or emit a `description` hint.
- [ ] **Per-operation `--prefer-content-type` flag** — today the JSON → form → multipart → octet → text → xml → other priority is fixed. A CLI override would help specs where the preferred upload path isn't the priority-first one.
- [ ] **Multipart `encoding[field]` metadata** — propagate per-part `contentType`, custom headers, and explode style into `RequestFilePart` so the runtime emits them. v1 always writes `application/octet-stream`.
- [ ] **Nested multipart binary fields** — `rewriteMultipartBinaryFields` only walks top-level properties. Add recursion for nested objects / arrays.

## Runtime

- [ ] **Non-JSON response decoding** — today every response routes through `NewToolResultJSON`. Add `NewToolResultText` and `NewToolResultBinary` (base64) and pick by the operation's response content type at generation time.
- [ ] **Streaming responses (SSE, chunked)** — first-class support instead of returning bytes.
- [ ] **Header parameter `Content-Type` collision** — when a spec declares a `Content-Type` header param alongside a non-JSON body, `<Op>WithBodyWithResponse` silently overrides it. Emit a generation-time warning.

## Distribution

- [ ] **GoReleaser `dockers_v2` / `homebrew_casks` migration** — current config uses `dockers` + `brews`. GoReleaser flags them as future-deprecated; migrate before they become hard errors.
- [ ] **`HOMEBREW_TAP_GITHUB_TOKEN` repo secret + `dipjyotimetia/homebrew-tap` repo** — until both exist, the brew step in the release workflow falls back to `GITHUB_TOKEN` and the formula push fails silently. The install line in README + release notes won't resolve until this is fixed.

## CI / tooling

- [ ] **Generated-code compile check across every example fixture** — `go build ./...` already covers committed examples, but `make regen-examples` requires `oapi-codegen` on PATH and isn't part of CI. Add a CI job that runs it.
