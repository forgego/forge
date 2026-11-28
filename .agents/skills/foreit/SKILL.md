---
name: foreit
description: Contributor conventions, architecture map, and verification workflows for developing the Forge (foreit) framework repository itself. Use when working inside the Forge repo on core Go packages, tests, admin UI, docs, or examples. Do NOT use for building end-user applications with Forge; use skills/forge-* instead.
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15, Node 20+
metadata:
  author: forgego
  version: "0.2"
---

# Foreit / Forge Contributor Conventions

Forge (`github.com/forgego/forge`) is a schema-first, batteries-included Go web framework (ORM, migrations, React admin, DRF-style REST API, identity, and server).

## Core Directory Layout

- **Framework root (`github.com/forgego/forge`)**:
  - `schema/`: Model schema DSL (`schema.Int64Field`, `ForeignKeyField`, `Meta`, `ModelHooks`)
  - `codegen/`: AST-based code generation (`gen.go`, `api_gen.go`)
  - `orm/`: Generic `Manager[T]`, immutable `QuerySet[T]`, typed expressions, `SelectRelated`/`PrefetchRelated`
  - `filter/`: Shared filter AST, parser, security whitelist, optimizer
  - `db/`: PostgreSQL/SQLite connections, pooling, `WithTx`, savepoints, `db/migrate/`
  - `admin/`: Admin registry, REST API (`admin/api/rest/`), and React 19 SPA (`admin/ui/web/`)
  - `api/`: ViewSets, serializers, auth, permissions, throttling, content negotiation, OpenAPI
  - `identity/`: Users, groups, permissions, bcrypt password hashing, sessions, tokens
  - `server/`, `config/`, `stores/`, `log/`, `netutil/`, `validate/`, `errors/`, `registry/`, `utils/`
  - `cli/`, `cmd/forge/`: Cobra CLI (`new`, `add`, `auth`, `generate`, `makemigrations`, `migrate`, `runserver`, `createsuperuser`, `check`, `test`, `shell`, `version`)
- **Cross-module & Examples**:
  - `tests/`: Integration (`tests/integration/`), CLI E2E (`tests/e2e/cli/`), doc samples (`tests/docsamples/`), Playwright admin E2E (`tests/e2e/admin/`)
  - `examples/ecommerce/`: 57-model reference app across 10 apps
- **Docs**:
  - `docs/`: `DESIGN.md`, `PRD.md`, `ROADMAP.md`, `TECH-DEBT.md`, `BUGS.md`, `REVIEWING.md`, `RELEASING.md`, `DOCUMENTATION-POLICY.md`, `design/admin-ui-system.md`
  - `docs-site/`: Docusaurus user documentation website (`docs-site/docs/status.md` is the canonical support contract)

## Verification Commands

- Framework Go suite: `go test -race ./...`
- Doc links check: `go run ./internal/tools/doclinks -root .`
- Integration suite (requires PostgreSQL 15): `cd tests && go test -v -timeout 15m ./integration/...`
- Doc samples suite: `cd tests && go test -v ./docsamples/...`
- Admin UI: `cd admin/ui/web && npm run build && npm run lint && npm run test -- --run`
- Docs site: `cd docs-site && npm run build`
- Ecommerce example: `cd examples/ecommerce && go test ./...`

## Git & Collaboration Rules

- Follow `AGENTS.md`: never run state-changing Git commands (`git add`, `git commit`, `git push`, `git reset`, `git clean`, branch creation) without current explicit user authorization.
- Leave changes in the working tree unless explicitly asked to commit.
- When authorized to commit, use Conventional Commits (`feat(subsystem): ...`, `fix(subsystem): ...`, `docs(subsystem): ...`, `test(subsystem): ...`) with no AI attribution trailers.
