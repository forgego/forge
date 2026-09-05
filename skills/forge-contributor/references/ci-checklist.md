# Forge Contributor CI Checklist

Use this as the authoritative "everything must happen" list. Run the checks that match your changes.

## Versions and Tooling
- Go versions in CI: 1.24 and 1.25 for unit tests.
- Node version in CI: 22 for docs site and admin UI.
- Docs site requires Node >= 20 (see `docs-site/package.json`).

## Core Go (Forge framework)
- `cd forge`
- `go mod download`
- `go test -v -race -coverprofile=coverage.out ./...`
- `go build ./...`
- `golangci-lint run --timeout=5m --config=../.golangci.yml`
- `go install golang.org/x/vuln/cmd/govulncheck@latest`
- `govulncheck ./...`

## Integration Tests (Tests module)
- Requires PostgreSQL on localhost.
- Env vars: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`.
- `cd tests`
- `go mod download`
- `go test -v -timeout 15m ./integration/...`

## CLI E2E Tests
- `cd forge`
- `go mod download`
- `go build -o <bin>/forge ./cli/cmd`
- Add `<bin>` to PATH for the test run.
- `cd ..\\tests`
- `$env:RUN_POSTGRES_TESTS="1"`
- `go test -v ./e2e/cli/...`

## Admin UI (forge/admin/ui/web)
- `cd forge/admin/ui/web`
- `npm ci`
- `npm run build`
- `npm run lint`
- `npm run test -- --run`

## Docs Site (docs-site)
- `cd docs-site`
- `npm ci`
- `npm run build`
- Optional fast check: `npm run build:fast`

## Examples: Ecommerce (examples/ecommerce)
- `cd examples/ecommerce`
- `go mod download`
- `go build -o ecommerce .`
- `go test -v ./...`
- Docker build: `docker build -t ecommerce-sample:<sha> -f Dockerfile .`
- CI runs a full stack integration check (admin page reachable + health endpoint).

## Security and CI-only Checks
- PR security review: govulncheck, Semgrep, TruffleHog, dependency review, npm audit.
- Scheduled security scan: Trivy (repo and docker), OSV scanner, Snyk, gosec, CodeQL.
- If you change workflows or security-sensitive code, ensure these pass in CI or run locally where practical.

## Docs and Test Updates
- Update `docs-site/docs/contributing/development.md` when workflow changes.
- Update `tests/README.md` and `tests/TESTING.md` when test strategy changes.
