---
name: forge-contributor
description: Contributor workflow for developing the Forge framework repository itself (core Go packages, integration/E2E tests, admin React UI, Docusaurus docs site, examples/ecommerce, and CI/security checks). Use when changing Forge internals or contributing to the framework repo. Do NOT use for building end-user applications with Forge (use `forge-cli`, `forge-models`, `forge-orm`, `forge-admin`, or `forge-api`).
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15, Node 20+
metadata:
  author: forgego
  version: "0.2"
---

# Forge Contributor

## Scope

- Use this skill when modifying the Forge framework repository itself: root Go packages, `tests/`, `admin/ui/web/`, `docs/`, `docs-site/`, `examples/ecommerce/`, `skills/`, or `.github/workflows/`.
- Do NOT use this skill for building end-user applications with Forge.

## Authoritative Repo Standards

- [Root agent rules](../../AGENTS.md) — repo map, collaboration, and explicit Git authorization rules (never run state-changing git commands without explicit authorization)
- [Review standard](../../docs/REVIEWING.md) — flow tracing, invariants, and evidence requirements
- [Documentation policy](../../docs/DOCUMENTATION-POLICY.md) — public vs local docs ownership
- [Admin UI visual system](../../docs/design/admin-ui-system.md) — required reading before touching `admin/ui/web/`
- [Support contract](../../docs-site/docs/status.md) — canonical tiers, Go/PostgreSQL versions, and stability policy

## Workflow

1. Identify all affected areas (`go` root packages, `tests/`, `admin/ui/web/`, `docs-site/`, `examples/ecommerce/`, `skills/`).
2. Run the required validations for every touched area (see [CI checklist](references/ci-checklist.md)).
3. If public APIs, CLI flags, config keys, or support tiers changed, update `docs-site/docs/` (especially `status.md`) and `skills/`.
4. Run `go run ./internal/tools/doclinks -root .` whenever any `.md` file changes.

## Required Validations Summary

- **Go core**: `gofmt`, `go build ./...`, `go test -race ./...`, `golangci-lint`, `staticcheck`, `govulncheck ./...`
- **Doc links & samples**: `go run ./internal/tools/doclinks -root .` and `cd tests && go test -v ./docsamples/...`
- **Integration & CLI E2E**: PostgreSQL 15 integration (`tests/integration/...`) and CLI E2E (`tests/e2e/cli/...`)
- **Admin UI (`admin/ui/web`)**: `npm run build`, `npm run lint`, `npm run test -- --run`, and Playwright E2E (`tests/e2e/admin`)
- **Docs site (`docs-site`)**: `npm run build`
- **Example app (`examples/ecommerce`)**: `go build -o ecommerce .` and `go test ./...`

## References

- [CI checklist](references/ci-checklist.md)
- [Contributor resources](references/contributor-resources.md)
