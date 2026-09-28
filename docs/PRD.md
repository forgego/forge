# forge PRD (comprehensive, implementation-backed)

## Table of contents

- [1. Product summary](#1-product-summary)
- [2. Goals](#2-goals)
- [3. Non-goals](#3-non-goals)
- [4. Personas](#4-personas)
- [5. Key user journeys](#5-key-user-journeys)
- [6. Functional requirements](#6-functional-requirements)
- [7. Non-functional requirements](#7-non-functional-requirements)
- [8. Compatibility policy](#8-compatibility-policy)
- [9. Observability and operability](#9-observability-and-operability)
- [10. Risks and mitigations](#10-risks-and-mitigations)
- [11. Acceptance criteria and quality gates](#11-acceptance-criteria-and-quality-gates)
- [12. Open questions](#12-open-questions)

## 1. Product summary

forge is a Django-inspired Go framework intended to reduce boilerplate in CRUD web apps while increasing safety.

The value proposition, consistent with archived docs, is:

- Schema-first model definition in Go (`schema`).
- AST-based code generation for type-safe managers/querysets (`codegen`).
- ORM/query engine (`orm`) and reusable filtering (`filter`).
- A built-in admin system (`admin`) that can be extended/overridden.
- A DRF-like REST API layer (`api`) with serializers/viewsets/auth/permissions/throttling.
- A migration engine (`db/migrate`) with CLI orchestration (`cli`).
- An identity system (`identity`) with sessions/tokens/passwords.

The product is validated by a multi-module test suite (unit tests + integration tests in `tests/`).

## 2. Goals

### 2.1 Primary goals

1. Enable developers to build CRUD apps quickly without sacrificing correctness.
2. Provide predictable, repeatable migrations.
3. Offer a powerful, secure admin interface out of the box.
4. Provide a consistent REST API framework with first-class errors, auth, and filtering.
5. Provide identity as a built-in system that can be overridden.

### 2.2 Secondary goals

1. Support both server-rendered and SPA admin experiences.
2. Allow plugin systems for admin widgets and API enhancements.
3. Keep code generation deterministic and safe.

## 3. Non-goals

- Being an unopinionated micro-library.
- Supporting every database equally; Postgres is the primary target.
- Guaranteeing zero breaking changes between major versions.
- Using reflection for the default ORM API.

## 4. Personas

### 4.1 Backend engineer

- Wants fast model definition.
- Needs safe migrations.
- Expects type-safe queries.

### 4.2 Platform engineer

- Wants predictable CLI workflows.
- Needs auth/security primitives.
- Wants the ability to constrain behavior (e.g., filter whitelists).

### 4.3 Admin/product operator

- Uses admin UI daily.
- Needs robust filtering/search/export.
- Needs permissioning and audit trails.

### 4.4 API consumer developer

- Uses REST endpoints.
- Needs consistent errors, pagination, filtering.

## 5. Key user journeys

This section is intentionally concrete. Every journey should be supported by code and tests.

### 5.1 Bootstrap a project

1. Developer runs `forge new myapp`.
2. CLI scaffolds directory structure and config.
3. Developer runs `forge runserver`.

Success:

- Server starts without manual wiring.
- Sample routes and admin URL respond.

Evidence:

- CLI template references exist in `cli/templates/templates/*`.
- `cli/commands/server/runserver.go` is implemented.

### 5.2 Define a model and generate code

1. Developer writes `app/blog/models.go` defining schema types.
2. Developer runs `forge generate`.
3. Generated files compile and expose typed managers and fields.

Success:

- Generated QuerySet supports `Filter`, `OrderBy`, `Limit`, `Values`, etc.
- Field expressions allow typed conditions.

Evidence:

- `codegen/ast_parser.go`, templates under `codegen/templates`.

### 5.3 Create migrations and apply them

1. Developer runs `forge makemigrations initial --auto`.
2. Up/down SQL is generated, including bookkeeping.
3. Developer runs `forge migrate up`.

Success:

- DB schema matches model definitions.
- Status is reported accurately.

Evidence:

- `db/migrate/*`, `db/migrations.go`, and `tests/integration/migrate/*`.

### 5.4 Use admin for CRUD

1. Developer registers models via admin registry or auto-discovery.
2. Operator visits `/admin/`.
3. Operator lists objects, filters, edits, exports.

Success:

- List view supports filtering/search/pagination.
- Forms are generated from schema metadata.
- Permissions are enforced.

Evidence:

- `admin/*` plus ecommerce example admin code.

### 5.5 Expose REST APIs

1. Developer defines serializer and viewset.
2. Developer registers routes.
3. Consumers use endpoints.

Success:

- CRUD endpoints include pagination/filtering.
- Errors are consistent problem details.
- Auth and permissions enforce access.

Evidence:

- `api/*` plus tests (`api/*_test.go`).

### 5.6 Authenticate users

1. User signs in via identity endpoints.
2. Session/token is created.
3. Subsequent API/admin requests are authenticated.

Success:

- Password strength and reset flows work.
- Tokens expire correctly.
- Middleware sets request identity.

Evidence:

- `identity/*` and tests (`identity/service/*_test.go`).

## 6. Functional requirements

Functional requirements are grouped by subsystem.

### 6.1 Schema system

FR-SCHEMA-1: The framework MUST provide a schema DSL for defining models in Go.

- Implemented by: `schema/*`.
- Test expectation: schema builders must return chainable Field/Relation values.

FR-SCHEMA-2: The schema DSL MUST support field options required by migrations and admin.

- Required, unique, primary key, auto increment.
- DB column/type/default.
- Temporal automation (AutoNow, AutoNowAdd).
- Generated columns.

FR-SCHEMA-3: The schema DSL MUST support relations with cascade settings.

- ForeignKey, OneToOne, OneToMany (reverse helper), ManyToMany.
- OnDelete/OnUpdate cascade.

FR-SCHEMA-4: The schema DSL MUST provide meta options.

- table name
- indexes and constraints

### 6.2 Code generation

FR-CODEGEN-1: The framework MUST parse schema definitions via Go AST.

- Implemented by: `codegen/ast_parser.go`.

FR-CODEGEN-2: The generator MUST emit type-safe code artifacts.

- fields expressions, manager, queryset.
- generated code must compile.

FR-CODEGEN-3: Code generation MUST be deterministic.

- changes to models should produce predictable diffs.

FR-CODEGEN-4: Codegen MUST not rewrite user-authored non-generated files.

### 6.3 ORM and QuerySet

FR-ORM-1: Provide a chainable, lazy QuerySet API.

- Filter, Exclude, OrderBy, Limit, Offset, Distinct.

FR-ORM-2: Provide typed field expressions.

- Eq/Ne/Gt/Gte/Lt/Lte
- Contains/StartsWith/EndsWith
- IsNull/IsNotNull

FR-ORM-3: Provide projection and value extraction.

- Select/Only/Defer
- Values/ValuesList

FR-ORM-4: Provide relation helpers.

- SelectRelated/PrefetchRelated surfaces exist and must be safe.

FR-ORM-5: Provide manager CRUD with hook support.

### 6.4 Filtering

FR-FILTER-1: Provide a shared filter AST.

FR-FILTER-2: Support deep relation filtering.

FR-FILTER-3: Enforce security constraints.

- whitelist fields
- whitelist lookups
- cost scoring

FR-FILTER-4: Integrate with admin and API.

### 6.5 Migrations

FR-MIG-1: Generate up/down migrations based on schema diffs.

FR-MIG-2: Maintain previous state from migration files.

FR-MIG-3: Execute migrations via CLI and golang-migrate.

FR-MIG-4: Provide status and recovery tools.

### 6.6 Admin

FR-ADMIN-1: Provide CRUD pages for registered models.

FR-ADMIN-2: Provide filtering/search/pagination.

FR-ADMIN-3: Provide exports and bulk actions.

FR-ADMIN-4: Provide extensibility.

- registry
- plugin hooks

### 6.7 API framework

FR-API-1: Provide serializer system.

FR-API-2: Provide viewsets.

FR-API-3: Provide auth and permissions.

FR-API-4: Provide throttling.

FR-API-5: Provide content negotiation.

FR-API-6: Provide OpenAPI generator.

### 6.8 Identity

FR-ID-1: Provide user and session models.

FR-ID-2: Provide password hashing and validation.

FR-ID-3: Provide token flows (verification, reset).

FR-ID-4: Provide middleware for auth.

FR-ID-5: Provide extensible auth backends.

### 6.9 CLI

FR-CLI-1: Provide commands for generation and migrations.

FR-CLI-2: Provide runserver and developer tools.

FR-CLI-3: Be testable in CI.

## 7. Non-functional requirements

### 7.1 Security

NFR-SEC-1: SQL injection protections

- Parameterized queries; identifiers escaped.
- Filter system must not allow arbitrary field access unless whitelisted.

NFR-SEC-2: CSRF protections

- Admin and session-auth flows must be protected.

NFR-SEC-3: Password safety

- bcrypt hashing.
- password strength validation.

NFR-SEC-4: Permission system

- API permissions must be enforced.
- Admin must enforce access control.

### 7.2 Performance

NFR-PERF-1: QuerySets should be lazy

- avoid accidental DB calls

NFR-PERF-2: Filtering should be optimized

- avoid unnecessary joins
- cost scoring

NFR-PERF-3: Migration generation should not require a running DB

- state reconstruction is file-based

### 7.3 Reliability

NFR-REL-1: Integration tests must validate end-to-end flows

- migrations
- CLI
- schema

NFR-REL-2: Recovery tools must exist for migration failures

- dirty state is detected and recovered/forced explicitly

### 7.4 Developer experience

NFR-DX-1: CLI scaffolding must produce compilable projects.

NFR-DX-2: Generated code must be readable.

NFR-DX-3: Documentation must be authoritative and code-aligned.

### 7.5 Maintainability

NFR-MAINT-1: Keep extension points explicit.

NFR-MAINT-2: Avoid global side effects at import time when possible.

NFR-MAINT-3: Prefer registries and interfaces.

## 8. Compatibility policy

The archived docs referenced v1/v2 transitions. The published policy for
v0.x and v1, including what counts as public API, generated code and
migration compatibility, is in the
[support contract](../docs-site/docs/status.md#stability-and-breaking-changes).
Requirements on the process:

- Major version changes may break APIs.
- Any breaking change must update:
  - code
  - tests
  - these docs
  - examples

## 9. Observability and operability

### 9.1 Logging

- Use structured logging (`log/*`).
- Log request IDs and relevant error codes.

### 9.2 Metrics

- Throttling and filter cost scoring are a first step.
- Future: Prometheus/OTel integration.

### 9.3 Health checks

- Server health endpoints should exist and be used by deployments.

### 9.4 Operational workflows

- `forge migrate status` to inspect status.
- `forge migrate force` or recovery commands for dirty state.

## 10. Risks and mitigations

### 10.1 Risk: codegen/migration drift

If AST parser and schema builder options diverge, migrations become invalid.

Mitigations:

- maintain tests that cover option mapping (e.g., generated columns)
- keep state loader robust

### 10.2 Risk: admin override instability

If overriding admin requires editing core `main.go`, it becomes unmaintainable.

Mitigations:

- registry-centric design
- plugin hooks

### 10.3 Risk: filter system security

Filters can allow data exfiltration or expensive queries.

Mitigations:

- whitelist
- cost scoring
- query planner

### 10.4 Risk: identity coupling

If identity is too coupled, apps cannot replace it.

Mitigations:

- backend interfaces
- middleware boundaries

## 11. Acceptance criteria and quality gates

Acceptance criteria are primarily test-based.

### 11.1 Must-pass tests

- `go test ./...`
- `go test ./examples/ecommerce/...`
- `go test ./tests/...`

### 11.2 Migration acceptance

- migration generator produces deterministic filenames.
- incremental migrations pass in ecommerce tests.
- recovery/status tests pass.

### 11.3 CLI acceptance

- CLI e2e tests can execute `forge makemigrations` and see output.
- CLI helper can build `forge` binary when missing.

### 11.4 Security acceptance

- identity tests cover password/token expiry.
- filter security tests cover whitelist and parser behavior.

## 12. Open questions

These are deliberate unknowns that should be decided before major expansions:

1. Admin UI direction: SPA-first vs server-rendered as default.
2. Multi-tenancy strategy.
3. Database support beyond Postgres/SQLite.
4. Caching strategy (app-level vs framework-managed).
5. How far to take auto-generated API surface (OpenAPI + client generation vs minimal).

## Appendix A: CLI surface (current)

This appendix is a practical snapshot of the command surface. It intentionally mirrors the commands registered in `cli/commands/registry.go`.

### A.1 Project scaffolding

- `forge new <project>`
- `forge add app <name>`
- `forge add model <name>`
- `forge auth` (scaffold identity integration)

### A.2 Generation

- `forge generate`

### A.3 Migrations

- `forge makemigrations <name> [--auto] [--models <dir>] [--path <dir>]`
- `forge migrate up [--dry-run]` (applies every pending migration)
- `forge migrate rollback` (rolls back the last applied migration; there is no `migrate down`)
- `forge migrate status`
- `forge migrate show`, `forge migrate lint`
- `forge migrate force <version>`, `forge migrate recover [--verify|--clean]`,
  `forge migrate baseline [--adopt]`, `forge migrate fake`
- `forge migrate squash` (registered, returns NotImplemented)

### A.4 Development

- `forge runserver`
- `forge shell`
- `forge test`
- `forge check`

### A.5 Admin utilities

- `forge createsuperuser`

## Appendix B: Test matrix

This appendix ties acceptance gates to packages.

- CLI e2e: `tests/cmd_forge/cli_e2e_test.go`, `tests/e2e/cli/cli_e2e_test.go`
- Migrations integration: `tests/integration/migrate/*`
- Schema integration: `tests/integration/schema/*`
- ORM integration: `tests/integration/orm/*`
- API unit tests: `api/*_test.go`
- Identity unit tests: `identity/*/*_test.go`
- Filter tests: `filter/*_test.go`

## Appendix C: Glossary

- **Schema**: declarative model definition via `schema`.
- **Generated code**: `*.gen.go` files produced by `codegen`.
- **QuerySet**: chainable query builder API (`orm/queryset.go`).
- **Filter AST**: serialized boolean expression tree for filters (`filter/ast.go`).
- **Migration state**: in-memory schema reconstructed from migration SQL (`db/migrate/state`).
- **Bookkeeping table**: `schema_migrations` used to track versions.

