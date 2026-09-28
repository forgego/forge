---
sidebar_position: 21
title: Support contract
description: What Forge supports, how each capability is verified, what is excluded, and the stability policy.
image: /social-card.png
---

# Support contract

This page is the canonical statement of what Forge supports. Other pages, the
[design notes](https://github.com/forgego/forge/blob/master/docs/DESIGN.md) and the
README link here instead of repeating it. When this page and another page
disagree, this page wins, and the other page is a documentation bug.

Forge is pre-1.0. The current release line is **v0.1.x**. See
[Stability and breaking changes](#stability-and-breaking-changes) for what that
means for your code.

## Support tiers

| Tier | Meaning |
| --- | --- |
| **Supported** | Implemented and exercised by automated tests in CI on every Go change. Where a database is involved, the tests run against PostgreSQL 15 and a missing database fails the run instead of skipping it. Defects are release blockers, and the [stability policy](#stability-and-breaking-changes) applies. |
| **Partially tested** | Implemented and covered by unit tests, SQL-string tests, SQLite tests or mock stores, but the end-to-end path (for example a live PostgreSQL database or a real browser) is not exercised in CI. Use it, but cover it with your own tests. |
| **Experimental** | Implemented but outside the release gate. It may change or be removed in any minor release. |
| **Not implemented** | Either no code exists, or the API exists and returns a `NotImplemented` error before doing any work. It never silently does nothing. |

Evidence links point at the test or CI job that backs each row. The CI jobs are
defined in [`.github/workflows/test.yml`](https://github.com/forgego/forge/blob/master/.github/workflows/test.yml).
The `release-gate` job runs every package serially against PostgreSQL with
`FORGE_REQUIRE_DB=1` and publishes the pass, fail and skip counts as the
`release-test-reports` artifact.

## Platforms

| Component | Tier | What is tested | Evidence |
| --- | --- | --- | --- |
| Go 1.26.x and 1.27.x | Supported | `go.mod` declares `go 1.26.0`. Unit tests run on Go 1.26.8 and 1.27.1. The other jobs use 1.27.1. | `unit-tests` matrix, all other jobs |
| Go older than 1.26 | Not supported | `go` refuses to build the module. | [`go.mod`](https://github.com/forgego/forge/blob/master/go.mod) |
| PostgreSQL 15 | Supported | The CI service image is `postgres:15` (major tag only). | `integration-tests`, `cli-e2e`, `release-gate`, `ecommerce-sample` |
| PostgreSQL 16 and later | Partially tested | Nothing in Forge is known to depend on 15. The admin browser journeys run against `postgres:16` (on pushes to master, and on pull requests that touch the admin or the ecommerce example); the ORM, migration and CLI suites do not run against 16, and no job runs 17 or later. | `Admin browser E2E result` job |
| PostgreSQL 14 and earlier | Not supported | Not tested. | None |
| SQLite (`github.com/mattn/go-sqlite3`, needs cgo) | Experimental | ORM queries, aggregates and some viewset paths have SQLite tests. Migration SQL for SQLite is generated, but applying it is not in the release gate: `TestMigrationApplySQLite` is skipped and the gate allows that one skip. | [`tests/pkg_migrations/migration_integration_test.go`](https://github.com/forgego/forge/blob/master/tests/pkg_migrations/migration_integration_test.go), `orm` SQLite tests |
| MySQL, SQL Server, other databases | Not implemented | `db.NewDBWithDriver` rejects other drivers. | [`db/db.go`](https://github.com/forgego/forge/blob/master/db/db.go) |
| Linux (amd64) | Supported | All CI jobs run on `ubuntu-latest`. | All jobs |
| macOS, Windows | Partially tested | Expected to work for development; not run in CI. | None |
| Admin UI browsers | Partially tested | The React admin is built with Vite's default browser target. CI runs its Vitest suite in jsdom and, in the `Admin browser E2E result` job, Playwright journeys in desktop Chromium (1280x720) and a 375x812 Chromium viewport against the built ecommerce example on PostgreSQL 16. Firefox and Safari are untested. | `frontend` job, `Admin browser E2E result` job, [`tests/e2e/admin`](https://github.com/forgego/forge/tree/master/tests/e2e/admin) |

## Models, keys and generated code

| Capability | Tier | Limits | Evidence |
| --- | --- | --- | --- |
| `int64` auto-increment primary key named `id` | Supported | The only key shape `Manager.Get` and the generated REST API accept. `forge generate --api` refuses any other key before writing files. | `TestValidateAPIModels_AcceptsOnlyInt64IDPrimaryKey`, `TestManager_WithTx_Commit` |
| UUID, string or composite primary keys | Not implemented | `Manager.Create` returns `NotImplemented` for non-integer keys; `Manager.Get` takes an `int64`. | [`orm/manager.go`](https://github.com/forgego/forge/blob/master/orm/manager.go) |
| Schema DSL and `forge generate` | Partially tested | Generation reads your `Fields()`, `Meta()` and relation definitions from source without running them, so it understands only a subset of Go expressions and reports the ones it cannot evaluate. The [models guide](/docs/models/) lists that subset. Fixtures pin the supported subset: generation is byte-identical across runs, strict mode reports unsupported expressions with file and line, the generated fixture compiles, and its migration SQL matches the schema. | `codegen` (`TestSupportedDSL*`), `unit-tests` |
| Custom field types (`schema.RegisterFieldType`) | Experimental | Registration works at runtime; generation and migrations do not know about custom types. | [`schema/registry.go`](https://github.com/forgego/forge/blob/master/schema/registry.go) |
| Projects created by `forge new` | Supported | Both templates compile and vet against the current checkout. | `TestNewProjectCompiles` |

## ORM operations

| Operation | Tier | Limits | Evidence |
| --- | --- | --- | --- |
| `Filter`, `Exclude`, `And`/`Or`/`Not`, `OrderBy`, `Limit`, `Offset`, `All`, `Get`, `First`, `Last`, `Count` | Supported | PostgreSQL. | `TestQuerySet_Integration_*`, `TestORMCRUDWithRelations` |
| `Manager.Create`, `BulkCreate`, `Update`, `Save`, `Delete`, `UpdateFields` | Supported | Integer primary keys only. | `TestORMCRUDWithRelations`, `TestManager_WithTx_*` |
| `SelectRelated` (foreign key and one-to-one joins), `PrefetchRelated` (many-to-many and reverse relations) | Supported | Nothing reports access to a relation that was not loaded. | `TestSelectRelated_Integration`, `TestPrefetchRelated_Integration` |
| Transactions: `db.WithTx`, `Manager.WithTx`, savepoints | Supported | One database. No distributed (2PC/XA) transactions. | `TestWithTx_*`, `TestManager_WithTx_*` |
| `AggregateValues` with `Count`, `Sum`, `Avg`, `Min`, `Max` | Supported | Ungrouped only. Also tested on SQLite. | `TestAggregateValues*` in `orm/aggregates_test.go` |
| `QuerySet.Update`, `UpdateBuilder`, `QuerySet.Delete`, `Exists` | Partially tested | `BulkUpdate` issues one `UPDATE` per entry. Mostly SQL-generation and SQLite tests. | `TestUpdateBuilder_Integration`, `orm/queryset_exists_test.go` |
| `Values`, `ValuesList`, `Only`, `Defer`, `Select`, `Distinct`, `Reverse`, `Annotate` | Partially tested | SQL-generation tests; not exercised against PostgreSQL in the gate. | `orm/queryset_test.go`, `orm/annotations_test.go` |
| Raw SQL | Supported | Use the `*sql.DB` embedded in `db.DB` directly (`database/sql`). There is no ORM raw-query API. | `database/sql` |
| Grouped aggregates, `STDDEV`/`VARIANCE`, aggregates registered with `orm.RegisterAggregate`, aggregates across many-to-many paths, `Aggregate` chained into a row query | Not implemented | Rejected with `NotImplemented` before any SQL runs. | `orm/aggregates_test.go` |
| `Union`, `Intersection`, `Difference` | Not implemented | Return `NotImplemented`. | `orm/queryset_not_implemented_test.go` |
| Window functions, recursive CTEs, full-text search | Not implemented | Not modeled in the query DSL. Use raw SQL. | None |

## Migrations, API, admin and identity

| Capability | Tier | Limits | Evidence |
| --- | --- | --- | --- |
| Migrations on PostgreSQL: generate from model diffs, apply, status, rollback, dirty-state recovery, checksum verification | Supported | See the [migrations guide](/docs/migrations/). `forge migrate squash` is not implemented. | `TestMigrationApplyPostgres`, `TestPostgresSchemaLifecycle` (evolution with seeded data, stable regeneration, failed-apply recovery, tamper detection), `tests/integration/migrate`, `TestCLIApplyMigration`; jobs `integration-tests`, `cli-e2e`, `release-gate` |
| Migration apply on SQLite | Experimental | See Platforms. | Skipped `TestMigrationApplySQLite` |
| Generated REST API (`forge generate --api`) | Partially tested | Dispatch, content negotiation, pagination and error mapping are tested with in-memory stores. A freshly scaffolded `forge new` application with related models is generated with `--api --strict`, migrated and exercised over HTTP against PostgreSQL in CI (CRUD, filters, validation, protected fields, authentication and ownership denials, transaction rollback, regeneration after a schema change); the tier moves to Supported after independent review per `docs/REVIEWING.md`. Integer `id` primary keys only. The field contract is in the [API docs](/docs/api/field-contract/). | `TestCLIPostgresAppJourney` in jobs `cli-e2e` and `release-gate`, `tests/integration/api`, `api` tests |
| Admin UI and admin REST API | Partially tested | Go handlers and the React app have unit tests. Browser journeys run in CI against PostgreSQL under a custom mount prefix: login/logout, list search/filter/sort/pagination, related-object selection, create/update/delete with field validation messages, read-only and auto-managed fields, object permission denials with stored-data checks, mixed-success bulk actions, and the 375px layout. Not covered: file uploads, inline related rows, saved views, export, history, plugins. Known defect: a boolean left unchecked on create stores the column default. Change history is kept in process memory (see [Deployment](/docs/deployment/)). | `admin` tests, `frontend` job, `Admin browser E2E result` job |
| Identity: users, password hashing, sessions, tokens, permissions | Supported | PostgreSQL. No OAuth/OIDC, social login or field-level permissions. | `identity/...` in `release-gate` with no skips allowed |
| API throttling | Partially tested | The default store is in process memory; supply a `throttling.Store` for anything shared. | `api/throttling` tests |
| Caching (`api/caching`) | Partially tested | In-memory only. | `TestMemoryCache_*` |
| Health, readiness and liveness endpoints | Supported | No database check is registered by default; see [Deployment](/docs/deployment/). | `TestHealthHandlers` |
| Graceful shutdown | Supported | `Server.StartWithGracefulShutdown` on SIGINT/SIGTERM; generated projects use it. | `TestServeUntilDrainsInFlightRequests` |
| Idempotency keys (`api/errors`) | Experimental | In-memory store only and no tests; `NewDatabaseStore` returns `NotImplemented`. | None |
| Log outputs `console` and `file` | Supported | The `remote` output returns `NotImplemented`. | `log` tests, `remote_output_not_implemented_test.go` |

## Extension points

| Extension point | Tier | Evidence |
| --- | --- | --- |
| HTTP middleware and routes (`Router.Use`, `Server.RegisterRoutes`) | Supported | `server` tests |
| Model lifecycle hooks (`BeforeCreate`, `AfterSave`, ...) | Supported | `orm/validation_context_test.go`, `api/viewset_validate_before_hooks_test.go` |
| API authentication and permission classes, renderers, parsers, versioning | Supported | `api/authentication`, `api/permissions`, `api/renderers`, `api/parsers` tests |
| Throttling store (`throttling.Store`) | Supported | `api/throttling/store_test.go` |
| Health checks (`server.RegisterHealthCheck`) | Supported | `TestHealthCheckRegistration` |
| Identity authentication backends (`identity/backends` registry) | Supported | `identity/backends/registry_test.go` |
| Admin login authenticator (`SetLoginAuthenticator`) | Supported | `TestHandleLogin_Authenticator*` |
| Admin actions and per-object permission hooks | Partially tested | `admin` tests |
| Admin change-history store (`core.HistoryManager`) | Partially tested | Interface only; Forge ships in-memory implementations. |
| Admin UI component overrides (`UIOverrides`) | Experimental | None |
| Custom field types, `orm.RegisterAnnotation`, `orm.RegisterQueryExpr` | Experimental | None end to end |
| Admin and API plugins (`registry.RegisterPlugin`) | Not implemented | `registry/plugin_unsupported_test.go` |

## Not in this release

These are out of scope for the v0.1 line and for the first v1 release. Where an
API exists, it fails with `NotImplemented`.

- GraphQL.
- Background jobs, task queues and schedulers.
- Distributed caching (Redis, Memcached) or any cache shared between processes.
- Social login, OAuth and OIDC providers.
- Built-in multi-tenancy.
- An advanced analytic query DSL: grouped aggregates, window functions, set operations, recursive CTEs.
- More than one database per application, read replicas and distributed transactions.
- Kubernetes manifests, Helm charts, service mesh integration.
- Running more than one application instance against shared in-memory state. See [Deployment](/docs/deployment/#multiple-instances).

## Framework readiness and the example application

The support tiers above describe the framework and are backed by the framework's
own tests. The [ecommerce example](https://github.com/forgego/forge/tree/master/examples/ecommerce)
is a separate Go module with its own CI job (`ecommerce-sample`). It shows how
the pieces fit, and its tests protect the example, but a passing example is not
evidence that a framework capability is supported, and a capability is not
promoted to Supported because the example uses it.

## Stability and breaking changes

**Public API** means the exported identifiers of packages under
`github.com/forgego/forge`, except `internal/...`, the Go packages under `cli/`,
`examples/...` and `tests/...`. The public surface also includes the `forge`
command's commands and flags, the configuration keys and `FORGE_*` environment
variables, the code `forge generate` writes, the migration file format and the
`schema_migrations` bookkeeping tables.

**While Forge is v0.x (now):**

- A minor release (v0.2.0, v0.3.0, ...) may break the public API, the generated
  code, CLI flags, configuration keys or migration bookkeeping. Each break is
  listed under **Breaking** in the [changelog](/docs/changelog/) with the steps
  to upgrade.
- A patch release (v0.1.2, ...) contains fixes. It does not break the public
  API on purpose; if a security fix has to, the changelog says so.
- Regenerate code (`forge generate`) after every upgrade. Hand edits to
  generated files are not preserved.
- Applied migrations are never rewritten by Forge. A release that changes
  migration bookkeeping ships a documented upgrade step.

**What v1 will promise:**

- Semantic import versioning: code that compiles against v1.x compiles against
  every later v1.y. Breaking changes wait for v2.
- Generated code: projects regenerated with a later v1.y keep compiling without
  changes to their own (non-generated) code.
- Migrations: migration files and bookkeeping written by v1.x apply, report
  status and verify with every later v1.y, and upgrading never requires
  regenerating or editing applied migrations.
- CLI commands, flags and configuration keys are removed only in a major
  release, after at least one minor release in which they are deprecated.
- Experimental and partially tested capabilities are exempt: they may change in
  a minor release, and the changelog says so.
- Raising the minimum Go or PostgreSQL version happens only in a minor release
  and is announced in the changelog.

The tags v1.0.0 and v1.0.1 exist but are retracted in `go.mod`; they contain no
Go packages. `go get github.com/forgego/forge@latest` ignores them. Because a
published version can never be reused, the first real v1 release will be
numbered above v1.0.1. The [release process](https://github.com/forgego/forge/blob/master/docs/RELEASING.md)
defines how releases are cut and verified.

## Related pages

- [Deployment](/docs/deployment/): running one production instance.
- [Changelog](/docs/changelog/): what changed in each release.
- [Installation](/docs/installation/).
