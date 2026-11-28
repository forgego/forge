# Forge Contributor CI Checklist

Run the checks that match your changes.

## Versions and Tooling

- Go versions in CI: `1.26.10` and `1.27.3` (`unit-tests`); other Go jobs use `1.27.3`. Minimum supported: `1.26.10` (see `go.mod`).
- PostgreSQL in CI: `postgres:15` (integration, CLI E2E, release-gate, ecommerce-sample); `postgres:16` (admin browser E2E).
- Node version in CI: `26.8.1` (`docs-site` requires Node >= 20).

## Core Go (Repository Root)

```bash
gofmt -l .
go build ./...
go test -v -race -coverprofile=coverage.out ./...
golangci-lint run --timeout=5m --config=.golangci.yml
staticcheck ./...
govulncheck ./...
```

## Documentation Links & Samples

```bash
go run ./internal/tools/doclinks -root .
cd tests && go test -v ./docsamples/...
```

## Integration, Release Gate & CLI E2E (`tests/`)

Requires PostgreSQL 15 (`POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`):

```bash
cd tests
go mod download
go test -v -timeout 15m ./integration/...
```

For `release-gate` parity (root module against PostgreSQL with no DB skips allowed):

```bash
FORGE_REQUIRE_DB=1 go test -json ./... | go run ./internal/tools/testreport
```

For CLI E2E:

```bash
go build -o /tmp/forge-bin/forge ./cmd/forge
export PATH="/tmp/forge-bin:$PATH"
export RUN_POSTGRES_TESTS=1
cd tests && go test -v ./e2e/cli/...
```

## Admin UI (`admin/ui/web`) & Browser E2E

```bash
cd admin/ui/web
npm ci
npm run build
npm run lint
npm run test -- --run
```

Playwright browser journeys live in `tests/e2e/admin` (run by `.github/workflows/admin-e2e.yml`).

## Docs Site (`docs-site`)

```bash
cd docs-site
npm ci
npm run build
```

## Example App (`examples/ecommerce`)

```bash
cd examples/ecommerce
go mod download
go build -o ecommerce .
go test -v ./...
```
