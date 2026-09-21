# Contributing to Loom

Focused issues and pull requests are welcome.

## Development

Loom uses the Go version declared in `go.mod`. Fork and clone the repository, create a branch, and keep changes scoped to one concern.

Before opening a pull request, run:

```bash
gofmt -w .
go mod tidy
git diff --check
go test ./...
go vet ./...
go test -race ./...
go build ./...
```

If your change affects release packaging, also install GoReleaser v2 and run:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

Describe the motivation, user-visible behavior, and verification performed in the pull request. Add or update tests when behavior changes.

## Security

Do not report vulnerabilities in public issues. Follow [SECURITY.md](SECURITY.md) instead.
