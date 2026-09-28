---
sidebar_position: 30
description: forge release notes and updates.
---

# Changelog

Forge is pre-1.0; see the [support contract](/docs/status/) for what a v0.x
minor or patch release may change. Entries cite the pull request that made
the change.

## Unreleased

### Breaking

- `Manager.Create`, `Save` and `BulkCreate` write explicit zero values
  (`false`, `0`, `""`) instead of leaving the column out, so a column default
  no longer replaces them. Code that relied on a `Default` filling a field the
  struct left at zero must build the instance with `manager.New()` (or call
  `orm.ApplyDefaults`) first. The admin create endpoint now stores the zero
  value for a field the request omits; the public REST API still applies the
  schema `Default`. Zero foreign keys, `nil` pointers, zero `time.Time`
  values, zero unique optional fields and zero `DBDefault` fields are still
  left to the database (#291).
- `api.Router.Register` and `RegisterRoutes` panic when a viewset's data
  access is misconfigured, naming the resource and each missing operation.
  Before, the route answered 500 per request. Set `ReadOnly` for list and
  retrieve only, or complete the queryset (#248).
- Generated and `forge add api` viewsets reject unknown request keys with a
  400 (`RejectUnknownRequestFields`). Hand-written viewsets keep ignoring
  them unless they opt in (#245).
- `PUT` is a full update: a body missing a required, request-writable field
  without a Go-side `Default` is a 400. A field with only a `DBDefault` is
  still required, and a required write-only field such as a password must be
  sent again on every `PUT`. Read-only fields (including the serializer's
  `ReadonlyFields()`) and fields a request cannot write are not required.
  `PATCH` is unchanged.
- A field tagged `json:",omitempty"` is accepted in requests under its Go
  name, the name responses use, so a response can be sent back as a `PUT`
  body. Such a field whose schema is neither serialized nor editable stays
  unwritable.
- `forge makemigrations` now emits the foreign keys declared with
  `schema.ForeignKeyField` and `Meta.Constraints`, which it silently skipped.
  In an existing project the next migration adds them and fails if rows
  violate them; clean orphan rows first.
- `/info` is no longer registered by default. Set `server.info_endpoint:
  true` to expose it (#290).
- `app.debug` defaults to `false`, and `Server.Start` refuses to listen with
  `app.debug` true when `app.env` is production. Projects from `forge new`
  turn it on in the local `.env`; older projects that relied on the default
  set `FORGE_APP_DEBUG=true` for development (#290).
- An unauthenticated request that fails a permission answers 401 Not
  Authenticated with a `WWW-Authenticate` challenge when the first
  authentication class can issue one (`Token` for `TokenAuthentication`,
  `Bearer` for JWT, `Basic realm="api"` for basic auth), as in Django REST
  framework. With session or API key authentication first, or none, it stays
  403. Authenticated requests that fail a permission stay 403 (#294).

### Fixed

- Creating a record with a boolean unchecked, a number set to 0 or a text
  left empty stores that value instead of the column default, in the ORM,
  the admin and the REST API (#291).
- `AutoNow` fields such as `updated_at` are refreshed on every
  `Manager.Update`, `Save` and `UpdateFields` (including admin edits), in the
  database and on the struct. Updates no longer write generated columns or
  clear a zero `AutoNowAdd` timestamp (#291).
- The warnings about generated ephemeral secrets are printed when the server
  starts, not by every CLI command (`forge generate`, `forge version`, ...)
  that loads the config. `config.Config.SecretWarnings` returns them (#294).
- The `forge new` scaffold builds its logger from the `logging.*` keys with
  the new `log.NewLoggerFromSettings`, and its `config.yaml` has a `logging`
  section (`level: debug`, `format: console`). `config.LoadSettings` now reads
  `logging.outputs`. Before, `main.go` ignored every `logging.*` key (#294).
- The ecommerce example shuts down gracefully with
  `StartWithGracefulShutdown`, draining in-flight requests on SIGINT or
  SIGTERM and closing its database, instead of `log.Fatal(ListenAndServe())`
  (#294).
- The API docs describe only what exists: list responses use page-number
  pagination (there is no limit/offset or cursor pagination), and the
  OpenAPI document has the `info` block only, with no Swagger UI and no
  `forge routes` command (#294).
- `forge add api blog-posts` emits `RegisterBlogPostsAPI` instead of the
  invalid `RegisterBlog-PostsAPI`; the URL segment stays `blog-posts`. Names
  that cannot form a Go identifier are rejected (#294).
- `forge version` prints the version the binary was installed from, and
  `forge new` pins that version in the new project's `go.mod` instead of
  `v0.1.0` (#286).
- `forge generate` warns about struct fields that have no schema entry (#287).
- Generated ID methods follow Go's promotion rules, and pointer-embedded ID
  holders are rejected at generate time instead of panicking in `Create`
  (#259).
- Projects created by `forge new` compile again: the generated `main.go` no
  longer embeds `static` and `templates` directories that do not exist next
  to it.
- `Server.StartWithGracefulShutdown` handles SIGINT and SIGTERM: it stops
  accepting connections and waits up to `server.graceful_timeout` for
  in-flight requests; a second signal during that wait ends the process at
  once. Projects created by `forge new` use it.
- `forge new --docker` builds with `golang:1.26-alpine`, and its compose
  file sets `FORGE_SERVER_HOST=0.0.0.0` so the published port reaches the
  server.
- Session and CSRF cookies are marked `Secure` for any spelling of
  `app.env: production`.
- `forge makemigrations` on unchanged models writes nothing and prints
  `No changes detected`; it no longer re-types columns, re-adds foreign keys
  or drops `created_at` defaults, and it prints the files it actually wrote.
- `forge makemigrations` on PostgreSQL no longer fails when a model drops a
  relation or a `Meta` constraint; the down migration re-adds it. A
  constraint whose CHECK condition or UNIQUE fields change is dropped and
  re-added, and a string default that changes only in case is migrated.
- `forge makemigrations` reads dropped and altered columns, array types such
  as `TEXT[]`, and defaults containing spaces, commas or quotes back from
  migration files, so it no longer repeats those changes on every run.
- SQLite migrations declare a new table's constraints inside `CREATE TABLE`
  and default auto timestamps to `CURRENT_TIMESTAMP`, so they apply. Adding
  a constraint to, or changing a foreign key of, an existing SQLite table
  is an error instead of invalid SQL.
- `forge migrate recover` no longer advises marking a rolled-back failed
  migration clean.
- `gen.go` for models with relations compiles.
- `forge add api` emits a viewset that compiles and registers.
- `IsOwnerOrReadOnly` finds owner fields on the embedded generated struct.
- List filters parse `?field=1` as a number, not a boolean.
- Admin: configured read-only fields are shown read-only; constraint,
  validation and invalid-value (PostgreSQL data exception) failures return
  4xx errors without raw database text, in single and bulk create and update
  alike, and other database errors are a generic 500;
  deleting a referenced record is a 409, as is a bulk delete in which every
  record is referenced. Bulk result items carry the classified code
  (`validation_error`, `conflict`, `invalid_reference`) instead of
  `create_failed` or `update_failed`. Edits send only changed fields;
  partially applied bulk actions list each skipped record; the foreign-key
  picker is labelled; fonts load under a custom mount prefix.
- ORM: `Filter(Or(a, b)).Filter(c)` keeps the OR group intact.
- `/health` and `/health/ready` report a failing check as `unhealthy` or
  `not ready` without its error text; the cause is logged instead (#290).
- The admin login lockout keys on the client IP resolved through the new
  `server.trusted_proxies` setting, so clients behind a trusted reverse proxy
  no longer share one lockout; forwarding headers from other peers are
  ignored (#290).
- `db.NewDBFromConfig` quotes the PostgreSQL connection values, so a
  password with spaces, quotes or backslashes connects, and an empty
  password no longer swallows the database name (#290).
- The API error handler logs PostgreSQL and SQLite driver errors by type,
  SQLSTATE or SQLite code and constraint name instead of their message,
  which can contain row values (#290).
- `forge makemigrations` no longer re-proposes a change on every run for
  cast and operator DB defaults (`DBDefault("'{}'::jsonb")`,
  `DBDefault("'x' || 'y'")`), a changed or added `DBDefault` on PostgreSQL,
  a table dropped and later created again, or a hand-written
  `ALTER TABLE .. RENAME COLUMN` or `RENAME TO` (#296, #292).
- `forge makemigrations` writes no migration when the changes render no SQL,
  and fails instead of writing a comment for a SQLite column change or an
  empty migration for a PostgreSQL column change it cannot express. The
  migrations guide covers renames and hand-written SQLite table rebuilds
  (#296).
- Migrations: a `Default` string with parentheses, such as
  `Default("x(1)")`, is quoted, and the down migration of a foreign key whose
  target table changed restores the old target (#296).

### Added

- The [support contract](/docs/status/), the [deployment guide](/docs/deployment/),
  the [API field contract](/docs/api/field-contract/),
  [data-access requirements](/docs/api/data-access/) and the release
  process (`docs/RELEASING.md` in the repository).
- CI gates: a fresh `forge new` PostgreSQL application exercised over HTTP,
  a PostgreSQL schema lifecycle test, schema DSL fixtures, admin browser
  journeys against PostgreSQL, and an install smoke test from the public
  module proxy on every tag and weekly.

## v0.1.1 (2026-09-28)

### Fixed

- `forge generate` writes models with constructors the `schema` package
  exports, so generated models compile (#285).
- Dependency updates: `go-playground/validator` 10.30.5 (#281), admin UI
  packages (#280, #282, #283) and GitHub Actions (#276).

### Known issues

- `forge version` prints `v0.1.0`, and `forge new` pins
  `github.com/forgego/forge v0.1.0` in new projects. Fixed after v0.1.1 (#286).
- Projects created by `forge new` fail to build with `pattern static: no
  matching files found`. Remove the `//go:embed` line, the `staticFiles`
  variable and the `embed` import from `cmd/server/main.go`. Fixed after
  v0.1.1.

## v0.1.0 (2026-09-28)

The first release of `github.com/forgego/forge` as a Go module at the
repository root, installable with
`go install github.com/forgego/forge/cmd/forge@v0.1.0`.

### Breaking

- The module moved from a `forge` subdirectory to the repository root and
  the CLI entry point from `cli/cmd` to `cmd/forge`, so `go install` produces
  a binary named `forge`. Import paths are unchanged (#279).
- The stray v1.0.0 tag, which contained no Go packages, is retracted (#279).
- Unfinished features fail with `NotImplemented` instead of silently doing
  nothing: queryset set operations (#217); migration squash, remote log
  output and the idempotency database store (#220); admin and API plugins
  (#230); grouped and custom aggregates (#240).
- Generated projects no longer ship a signing key: `forge new` writes random
  secrets to a git-ignored `.env`, and production refuses to start without
  explicit secrets (#241).

### Added

- `forge generate --api` writes serializers, viewsets and routes; generated
  files are written atomically (#235, #258).
- `forge migrate recover --verify` checks applied migrations against the
  checksums recorded when they were applied, and `forge migrate baseline
  --adopt` records a baseline for databases migrated before checksums
  existed (#238).
- `orm.AggregateValues` computes ungrouped Count, Sum, Avg, Min and Max (#240).
- Code generation reports model expressions it cannot evaluate, and
  `--strict` makes that an error (#242).
- The release evidence job fails when the test database is unavailable
  instead of skipping database tests (#236).
- Redesigned admin UI (#204).

### Fixed

- Security: secrets are redacted from access logs, admin login is rate
  limited and expired sessions are purged (#212); the password backend no
  longer reveals whether an account exists (#219); owner and admin API
  permissions match real users and unknown tokens are rejected (#222).
- ORM: writes use schema column names and the real primary key (#225);
  to-many relation filters no longer duplicate rows (#215); lookup SQL, date
  parts and relation-path joins (#211); `First`/`Last` default to primary-key
  order (#217).
- Migrations: SQLite migrations emit SQLite DDL or fail (#224); status reports
  real versions (#223); removing a model or field no longer aborts generation
  (#227); safe SQL statement splitting and atomic migration pairs (#213).
- API: a JSON `null` no longer panics the viewset decoder; pagination links
  are valid (#221).
- Admin: create and update write only writable fields (#218); the SPA works
  under a custom mount prefix (#209).
- Logging: outputs are flushed and closed (#234).

## v1.0.0 and v1.0.1 (retracted)

Tagged by mistake before the module lived at the repository root; v1.0.0
contains no Go packages. Both are retracted in `go.mod`, and
`go get github.com/forgego/forge@latest` skips them. They are not
Forge 1.0.
