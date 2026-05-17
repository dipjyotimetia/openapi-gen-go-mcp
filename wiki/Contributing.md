# Contributing

A summary. The authoritative document is [`docs/contributing.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/contributing.md) in the repo.

## Dev loop

```bash
make build           # builds CLI into ./bin/openapi-go-mcp
make test            # go test ./...
make test-race       # go test ./... -race -count=1
make vet             # go vet ./...
make lint            # golangci-lint run (config in .golangci.yml)
make fmt             # gofmt -s -w .
make regen-examples  # regenerates every example's oapi-codegen + mcp output
make smoke           # boots petstore over stdio, calls initialize + tools/list
```

Run a single test:

```bash
go test ./pkg/generator/ -run TestRender_OpenAICompat_PetstoreSchemas -v
```

## Updating the golden test

The golden test in `pkg/generator/golden_test.go` guards generator output. When you change the generator and the diff is intentional:

```bash
UPDATE_GOLDEN=1 go test ./pkg/generator/...
git diff testdata/   # review carefully before committing
```

If you can't justify every byte of the diff, the change isn't ready.

## Conventions

- **Lint config** (`.golangci.yml`) enables `errcheck`, `govet`, `staticcheck`, `revive`, `gocritic`, `gosec`. `gocritic`'s `ifElseChain` is intentionally disabled.
- **Imports** — `goimports` `local-prefixes` is set to the module path. Groups are: stdlib, third-party, then this module.
- **Determinism is mandatory.** Sorted iteration everywhere, `gofmt` output, golden coverage. Reviews depend on this.
- **Update `docs/changelog.md`** under `## Unreleased` for user-visible changes.
- **Don't break the runtime interface lightly.** Both backend adapters depend on it; new methods need both `gosdk` and `mark3labs` implementations.

## Tests

- **Unit tests** live next to each package.
- **Golden tests** live in `pkg/generator/`, with fixtures in `testdata/`.
- **E2E tests** live in `tests/e2e/`. They exercise example servers over MCP stdio and require examples to already be generated (`make regen-examples` if you changed the generator).

## CI

CI runs on Go 1.26.x. Generated code itself targets Go 1.23+ — downstream users can still compile against older Go.

## Attribution

Portions of `pkg/runtime` and `pkg/generator/naming.go` are adapted from [redpanda-data/protoc-gen-go-mcp](https://github.com/redpanda-data/protoc-gen-go-mcp) under Apache 2.0. Keep the attribution comment intact when editing those files.

## Code of conduct

By contributing you agree to the [Code of Conduct](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/code-of-conduct.md).

## Security issues

Don't open public issues for security problems. See [`docs/security.md`](https://github.com/dipjyotimetia/openapi-go-mcp/blob/main/docs/security.md) for the disclosure process.
