# forge Design (comprehensive, code-aligned)

## Table of contents

- [1. Design goals](#1-design-goals)
- [2. Non-goals](#2-non-goals)
- [3. Design principles](#3-design-principles)
- [4. Key invariants](#4-key-invariants)
- [5. API design conventions](#5-api-design-conventions)
- [6. Design patterns used](#6-design-patterns-used)
- [7. Schema DSL design](#7-schema-dsl-design)
- [8. Code generation design](#8-code-generation-design)
- [9. ORM/QuerySet design](#9-ormqueryset-design)
- [10. Filtering design](#10-filtering-design)
- [11. Migration design](#11-migration-design)
- [12. Admin design](#12-admin-design)
- [13. API framework design](#13-api-framework-design)
- [14. Identity design](#14-identity-design)
- [15. Error handling design](#15-error-handling-design)
- [16. Security design](#16-security-design)
- [17. Performance and scaling](#17-performance-and-scaling)
- [18. Testing strategy](#18-testing-strategy)
- [19. Documentation policy](#19-documentation-policy)
- [20. Recorded subsystem decisions](#20-recorded-subsystem-decisions)
- [21. Capability status and supported configurations](#21-capability-status-and-supported-configurations)

## 1. Design goals

Forge aims to give Go developers Django-like leverage without sacrificing compile-time safety.

Concrete goals:

- Express data models in Go (schema DSL) and generate a strongly typed API.
- Provide a full CRUD pipeline: ORM + migrations + admin + REST APIs.
- Deliver built-in identity and security primitives.
- Keep extension points explicit, stable, and test-covered.
- Keep runtime flexibility (dynamic filters and runtime field refs) without undermining the type-safe default.

## 2. Non-goals

- A fully dynamic runtime ORM (reflection-only) is not the primary path.
- Requiring manual per-model registration everywhere is not a goal.
- Hiding SQL entirely is not a goal; we prefer safe SQL generation and transparent behavior.
- Backward compatibility at any cost is not a goal for major revisions (breaking changes are acceptable if tests and docs are updated).

## 3. Design principles

These principles are adapted from archived architecture/design docs, but verified against the current codebase.

### 3.1 Type-safe first

- The default experience should be type-safe.
- Generated QuerySets embed `forge/orm.BaseQuerySet[T]` and expose typed `Fields` instances.
- Prefer compile-time constraints; only fall back to runtime interfaces when needed.

### 3.2 Dynamic when needed

- Dynamic queries exist for runtime-defined fields and filters.
- `forge/filter` and runtime field references in `forge/orm` support this.

### 3.3 Convention over configuration

- CLI scaffolding aims to create predictable structure.
- Admin and API should work with minimal wiring.

### 3.4 Extensibility over hard-coding

- Registries (`forge/admin/core/registry.go`, `forge/schema/registry.go`, `forge/identity/backends/registry.go`) are preferred over hard-coded lists.
- Interfaces shape extension points and allow mocking.

### 3.5 Security by default

- Default middleware should not be optional-by-accident.
- Identity and permissions should be first-class.

### 3.6 Testability

- Every major subsystem has at least unit tests.
- Cross-module integration is validated in `tests/`.

## 4. Key invariants

Invariants are the rules we preserve even during large refactors.

### 4.1 Schema -> AST -> migrations invariants

- Every schema builder option that affects SQL must be captured by `forge/codegen/ast_parser.go`.
- Every captured option must be reflected in the SQL builder (`forge/db/migrate/sql/*`) and preserved in state conversion if required.
- Primary keys are always treated as required (NOT NULL). This prevents invalid migrations against identity columns.

### 4.2 Registry invariants

- Model registration is centralized (admin registry, identity backend registry, schema registries).
- Projects should not need to wire everything in `main.go`. The design preference is: defaults exist, and overrides are opt-in.

### 4.3 CLI invariants

- CLI commands are the canonical way to run generation/migrations.
- Tests must be able to run CLI without external setup; helper builds `forge` when not found (`tests/testhelpers/cli_helper.go`).

### 4.4 Separation of concerns

- Schema definitions do not contain DB logic.
- Repositories access DB, services implement business logic, handlers manage HTTP.
- Admin/API should not directly embed DB SQL strings; they call into ORM/query primitives.

## 5. API design conventions

### 5.1 Naming

- Prefer explicit names over abbreviations.
- Use Go naming conventions (`ID` vs `Id` depends on generator conventions; generated code must be consistent).
- Avoid suffixes that force ceremony (for example, `Build()` patterns in schema were removed in favor of direct chaining).

### 5.2 Fluent builder style

The schema DSL uses fluent chains:

- Field: `schema.String("email").WithRequired().WithUnique().WithMaxLength(255)`
- Relation: `schema.ForeignKey("customer_id", "Customer").WithOnDelete(schema.CascadeCASCADE)`

Design reason: chainable methods map cleanly to AST extraction and maintain readability.

### 5.3 Immutability and chaining

- QuerySets are chainable and should behave like immutable builders: each call returns a new QuerySet value.
- Filters similarly build AST trees rather than mutating global state.

### 5.4 Errors

- Prefer typed, structured errors for API surfaces (`forge/api/errors/*`).
- Preserve root cause for debugging but expose safe, consistent messages externally.

## 6. Design patterns used

Patterns below are described against current packages.

### 6.1 Strategy pattern

Used when multiple behaviors exist behind a common interface.

- Identity auth backends: `forge/identity/backends/interface.go` and `forge/identity/backends/registry.go`.
- API auth mechanisms: `forge/api/authentication/*`.

### 6.2 Repository pattern

Used to isolate data access.

- Identity repositories: `forge/identity/repository/*`.

Design reason: services can be tested independently from DB drivers.

### 6.3 Service pattern

Used to encapsulate business logic.

- Identity services: `forge/identity/service/*`.

### 6.4 Factory pattern

Used to build composite systems with injected dependencies.

- Identity factory patterns exist in `forge/identity/factory.go`.

### 6.5 Builder pattern

Used for fluent construction of schema fields/relations and query expressions.

- Schema field chain methods: `forge/schema/field_methods.go`.
- Relations chain methods: `forge/schema/relation.go`.
- Query expression builders: `forge/orm/expression.go`.

### 6.6 Registry pattern

Used to centralize discovery and extension.

- Admin registry: `forge/admin/core/registry.go`.
- CLI registry: `forge/cli/core/registry.go`.
- Schema registries: `forge/schema/registry.go`, `forge/schema/relation_registry.go`.

### 6.7 Template method pattern

Used in codegen templates and in base viewsets.

- Codegen: `forge/codegen/templates/*`.
- API viewsets: base types expose common flows and override hooks.

### 6.8 Chain of responsibility

Used in middleware chains.

- Server middleware: `forge/server/middleware.go`.
- API middleware integration: `forge/api/middleware_integration.go`.

### 6.9 Decorator pattern

Used to wrap querysets and values querysets.

- Values/ValuesList wrappers: `forge/orm/queryset.go` defines `ValuesQuerySet` and `ValuesListQuerySet` wrappers.

### 6.10 Observer/event pattern

Used in admin notifications.

- Admin notifications and hooks: `forge/admin/core/notifications.go`.

## 7. Schema DSL design

### 7.1 Why a schema DSL instead of struct tags

Struct tags are not expressive enough for:

- relation options (on delete/on update, deferrable, match types)
- generated columns
- migration-specific metadata (db type, db default)
- admin metadata (editable, serialize/write-only)

A schema DSL expresses these directly and keeps AST extraction deterministic.

### 7.2 Field model

`forge/schema/field.go` defines the field model: name, type, required, unique, primary key, auto increment, default, and an options map for additional metadata.

Design notes:

- Options map is intentionally flexible to avoid exploding the struct with rarely used fields.
- Options keys must remain stable because they are persisted into migrations state and used by SQL builders.

### 7.3 Temporal traits

`forge/schema/field_traits.go` defines traits, including temporal fields with `AutoNow` and `AutoNowAdd`.

Design reason: these traits allow the SQL builder to set driver-specific defaults (e.g., `DEFAULT now()` in Postgres).

### 7.4 Generated columns

Generated columns are modeled via `WithGeneratedColumn(expression, stored)`.

- AST parser must capture both the expression and stored/virtual setting.
- SQL builders must implement driver rules:
  - Postgres: `GENERATED ALWAYS AS (...) STORED`
  - SQLite: `STORED` or `VIRTUAL` depending on engine support

### 7.5 Relations

`forge/schema/relation.go` defines relations and chain methods.

Important decisions:

- `ForeignKey(name, to)` uses the local column name as the relation key.
- `ManyToMany(name, to)` must support a through table to avoid magical naming.
- Cascade and constraint options must be preserved into migrations.

### 7.6 Meta

Meta exists so:

- table names remain stable (critical for migrations)
- indexes and constraints can be expressed declaratively
- admin and API can display verbose names

## 8. Code generation design

### 8.1 Why AST parsing

The archive emphasized AST parsing to avoid reflection at runtime. The current code follows that:

- `forge/codegen/ast_parser.go` parses source and extracts model definitions.
- This allows generation of strongly typed helpers and avoids runtime field lookups.

### 8.2 Option extraction model

The AST parser normalizes builder chains into a single options map, for both fields and relations.

Design reason:

- Downstream systems (migrations, admin metadata) read from a consistent representation.

### 8.3 Template boundaries

Templates in `forge/codegen/templates` are the only place where generated code structure should be defined.

Rules:

- Do not hand-write generated files.
- If you change templates, rerun generation and update tests.

### 8.4 Deterministic naming

Deterministic naming prevents migration churn.

- Prefer explicit table names via `Meta.TableName`.
- Prefer explicit index names where possible.

### 8.5 Regeneration safety

Writer should avoid clobbering user code.

- Generated files should have a consistent suffix (`*.gen.go`).
- Non-generated files should not be modified by the generator.

## 9. ORM/QuerySet design

Described against the current implementation in `forge/orm/*`.

### 9.1 QuerySet as a persistent builder

QuerySets are designed to be:

- chainable
- lazy
- mostly immutable

Design reason:

- consistent mental model (like Django)
- easy to compose filters across admin/API
- safe reuse in concurrent contexts

Implementation: `forge/orm/queryset.go` clones and returns new QuerySet values on most operations.

### 9.2 Field expressions

Field expressions represent typed accessors plus query operators.

- Generated fields expose `Eq`, `Gt`, `Contains`, etc.
- Runtime field references exist for dynamic use.

### 9.3 Values and ValuesList

These are projection tools:

- `Values("id", "name")` returns a list of maps (or map-like rows) for dynamic output.
- `ValuesList("id", "name")` returns `[][]interface{}`.

See tests: `forge/orm/queryset_test.go`.

### 9.4 SelectRelated and PrefetchRelated

These exist as API surfaces and are intended to prevent N+1.

- `SelectRelated(...)` implies JOIN-based eager loading.
- `PrefetchRelated(...)` implies separate query prefetch.

Guardrail: `forge/orm/preload.go` defines errors when relations are accessed without being preloaded.

### 9.5 Updates

- Update builders (`forge/orm/update_builder.go` and helpers) support safe partial updates.
- Manager CRUD (`forge/orm/manager.go`) is the main entry point for Create/Update/Delete with hooks.

### 9.6 Transactions and unit of work

`forge/db/transaction.go` is used by services/repositories. The ORM does not attempt to hide transaction boundaries; it exposes enough primitives for explicit, business-aligned transactions.

## 10. Filtering design

Described against current `forge/filter/*`.

### 10.1 Why a separate filtering subsystem

Admin and REST APIs need a common filtering language. Reusing a single filter AST avoids:

- different semantics between admin and API
- duplicated parsing rules
- inconsistent security policy

### 10.2 Filter AST as a contract

The filter AST is:

- serializable
- validated
- composable (boolean trees)

It must remain stable because:

- persisted filters can be stored and loaded
- UI components can share filter structures

### 10.3 Security model

Filtering is an attack surface.

Design requirements:

- Default deny: fields/lookups must be whitelisted.
- Prevent expensive queries by applying cost scoring.
- Prevent JOIN explosion via optimization strategies.

Implementation: `forge/filter/security.go`, `optimizer.go`, `metrics.go`.

### 10.4 Integration points

- Admin list view uses filter metadata to render sidebar widgets.
- API list endpoints parse query parameters and apply them to querysets.

## 11. Migration design

### 11.1 Why migrations are generated from model diffs

- Developers author models; migrations are derived.
- Diffing allows incremental updates and repeatable CI.

### 11.2 State is reconstructed from migration files

Rather than depending only on database introspection, the state manager (`forge/db/migrate/state/loader.go`) reconstructs schema from existing migration SQL. This enables:

- running generation without a database
- deterministic diffs in CI

Tradeoff:

- the parser must be robust against SQL variations emitted by the builder.

### 11.3 Change detection

The detector compares current model definitions vs. reconstructed state.

Key design concerns:

- ignore internal tables (bookkeeping)
- avoid changing identity columns in invalid ways
- avoid repeated add column operations due to incomplete parsing

### 11.4 Driver-specific SQL generation

SQL generation is separated by driver to avoid "lowest common denominator" output.

- Postgres: `forge/db/migrate/sql/postgres.go`
- SQLite: `forge/db/migrate/sql/sqlite.go`

### 11.5 Recovery and verification

Dirty-state recovery is built into the system.

- `forge/db/migrate/execute/recover.go` provides tools to inspect and recover.
- `verify/*` provides lint/safety checks.

The design is to fail loudly on dirty state by default, but provide force/recovery mechanisms that are test-covered.

## 12. Admin design

This section reconciles the archived admin redesign notes with the implemented admin system in `forge/admin/*`.

### 12.1 Design goals

- Provide CRUD for registered models.
- Provide filters, search, pagination, exports, bulk actions.
- Avoid per-route boilerplate.
- Allow overrides:
  - custom widgets
  - custom list/detail behavior
  - custom permission rules

### 12.2 Registry as the central override point

Admin is centered around a model registry.

- Default behavior should be good enough.
- Overrides should register configuration rather than forcing edits in `main.go`.

### 12.3 Metadata-driven UI

Admin UI is metadata-driven:

- Schema fields -> form inputs
- Field traits -> widget hints
- Relations -> foreign key selectors/autocomplete

The design reduces duplication: models are the single source of truth.

### 12.4 API-first vs server-rendered

Archived docs promoted a headless admin.

The current implementation includes REST routers (`forge/admin/api/rest/router.go`) and UI helpers (`forge/admin/ui/*`), allowing either:

- a JS SPA consuming admin APIs
- server-rendered admin pages

### 12.5 Audit/history

`forge/admin/core/history.go` provides an audit trail building block.

Design requirement: history must be opt-in and storage-agnostic.

## 13. API framework design

Derived from archived REST/API docs and the implemented `forge/api/*`.

### 13.1 ViewSet abstraction

- ViewSets centralize CRUD behavior with override hooks.
- Serializer is provided as a factory (`func() Serializer`) to avoid shared mutable state.

### 13.2 Serializer design

- Serializer fields provide validation, defaulting, and read/write constraints.
- Specialized serializer fields exist for core types and formatting.

### 13.3 Content negotiation

- Parsers deserialize request bodies.
- Renderers serialize responses.

Design goal: predictable behavior across JSON/XML/YAML/HTML/CSV.

### 13.4 Errors

- API errors return problem details and stable error codes.
- Exceptions are mapped via `forge/api/errors/mapper.go`.

### 13.5 OpenAPI

OpenAPI generation exists (`forge/api/docs/openapi.go`). It must remain consistent with actual route wiring.

## 14. Identity design

Identity follows repository + service + backend patterns.

### 14.1 Repositories

Repositories encapsulate persistence logic and isolate SQL or ORM calls.

### 14.2 Services

Services implement business rules:

- password validation
- token issuance/rotation
- session tracking
- email verification

### 14.3 Backends

Backends implement authentication strategies.

Design requirement: adding OAuth or external identity providers should be additive.

### 14.4 Middleware

Middleware populates request context and enforces permissions.

## 15. Error handling design

### 15.1 Structured errors

- Use typed errors where possible.
- Preserve root cause for logging.

### 15.2 API vs internal errors

- Internal errors can include stack info.
- External API errors must be safe.

## 16. Security design

Security is layered:

- request protection: CSRF, CORS, rate limiting (`forge/server/security.go`, `forge/server/ratelimit.go`).
- identity enforcement: auth middleware.
- query safety: parameterized SQL builder and filter security.

## 17. Performance and scaling

Performance design is pragmatic:

- QuerySets avoid work until execution.
- Migration generation does not require a running database.
- Filter optimizer reduces expensive query patterns.

Scaling notes:

- Prefer database-side filtering.
- Use pagination by default.
- Cache at API layer when needed (`forge/api/caching/*`).

## 18. Testing strategy

Testing mirrors the architecture:

- Unit tests for each package.
- Integration tests in `tests/` for migrations and CLI.
- Example module tests.

Quality gates:

- Migration tests must pass on Postgres.
- CLI e2e tests must pass without requiring global installation.

## 19. Documentation policy

This repository contains a large archive.

Rules:

- `docs/DESIGN.md`, `docs/PRD.md`, `docs/ROADMAP.md`, `docs/TECH-DEBT.md`,
  `docs/BUGS.md` are authoritative.
- When code changes, update the authoritative docs.


## 20. Recorded subsystem decisions

Decisions below were salvaged from ops run notes and the archive before
those sources were deleted. They document behavior the code implements.

### 20.1 Admin REST conventions

- Bulk endpoints use `201` (full success) / `207` (partial) / `400`-`500`
  (terminal) semantics.
- Bulk actions report per-item codes `invalid_id` / `not_found` /
  `permission_denied` in `BulkActionError.Code`.
- Single-object update/delete fetches the object first, returns 404 when
  missing, and enforces object-level `HasChange`/`HasDeletePermission`.
- `ExecuteAction` enforces per-object `HasChangePermission` with skip
  reporting in `BulkActionResponse.Errors`.
- Auth is bearer-token via an in-memory expiring session store;
  `/api/login` is public; credentials come from `FORGE_ADMIN_USERNAME` /
  `FORGE_ADMIN_PASSWORD` with constant-time compare and `401
  invalid_credentials` on failure.

### 20.2 Migrations

- Migration status infers applied history via `mergeAppliedVersions`;
  a dirty current version is reported `DIRTY` and excluded from the
  applied list.
- The status reporter and CLI degrade gracefully on a nil engine
  (explicit "Unknown" metadata, ASCII `[WARN]`/`[x]`/`[ ]`/`[!]` markers,
  empty-directory message) instead of panicking.
- SQL builders are split per driver (`postgres.go` / `sqlite.go`):
  never lowest-common-denominator SQL.

### 20.3 Identity

- `RequirePermission` is wired to `PermissionService.CheckPermission`
  with superuser bypass plus staff fallback.
- Password reset uses a notifier plus a fixed hashed-token flow.
- Serializer-as-factory (`func() Serializer`) avoids shared mutable state.

### 20.4 ORM

- `Field.Resolve` validates full nested relation paths (rejects
  non-relation traversal and terminal relations).
- Aggregate/annotation registries are thread-safe.
- Hook order on write paths is
  `BeforeSave` → `BeforeCreate` → `AfterCreate` → `AfterSave`
  (and the `BeforeUpdate`/`BeforeDelete` analogues).

### 20.5 Example authorization pattern

- The ecommerce example uses object-aware permissions: finalized
  payments/orders, primary warehouses, and used coupons are immutable.
  Mixed-outcome `207` tests verify the pattern.


## 21. Capability status and supported configurations

This table records which configurations are verified for release.

| Capability | Status | Evidence |
| --- | --- | --- |
| SQLite migration apply | unverified | no automated apply test: the migration generator has no driver selection; tests/pkg_migrations skips it |

