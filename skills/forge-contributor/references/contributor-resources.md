# Contributor Resources

## CI Workflows

- `.github/workflows/test.yml` (unit tests, integration tests, lint, build, frontend, CLI E2E, release-gate)
- `.github/workflows/admin-e2e.yml` (Playwright admin browser E2E against PostgreSQL 16)
- `.github/workflows/docs-links.yml` (`internal/tools/doclinks` markdown link & path validation)
- `.github/workflows/deploy-docs.yml` (Docusaurus build & GitHub Pages deploy)
- `.github/workflows/ecommerce-sample.yml` (`examples/ecommerce` build, tests, and Docker smoke)
- `.github/workflows/install-smoke.yml` (public `go install` smoke verification)
- `.github/workflows/pr-security-check.yml` (`gofmt`, `staticcheck`, `govulncheck`, Semgrep, TruffleHog, dependency review, npm audit)
- `.github/workflows/security.yml` (scheduled Trivy, OSV, Snyk, gosec)
- `.github/workflows/codeql.yml` (CodeQL analysis)

## Authoritative Docs & Test Guides

- `AGENTS.md`
- `docs/REVIEWING.md`
- `docs/DOCUMENTATION-POLICY.md`
- `docs/design/admin-ui-system.md`
- `docs-site/docs/status.md`
- `docs-site/docs/contributing/development.md`
- `docs-site/README.md`
- `tests/README.md`
