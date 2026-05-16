# GoReleaser builds the binary and copies it into this image — there is no
# `go build` step here. To build a development image locally without
# GoReleaser, run: `goreleaser release --snapshot --clean --skip=publish`.
FROM gcr.io/distroless/static-debian12:nonroot

COPY openapi-gen-go-mcp /usr/local/bin/openapi-gen-go-mcp

USER nonroot:nonroot
WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/openapi-gen-go-mcp"]
