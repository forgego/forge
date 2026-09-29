---
sidebar_position: 30
description: forge release notes and updates.
---

# Changelog

Forge is pre-1.0; see the [support contract](/docs/status/) for what a v0.x
minor or patch release may change. Entries cite the pull request that made
the change.

## Unreleased

## v0.2.0 (2026-09-29)

v0.2.0 is a minor release because it contains breaking changes. Read
**Upgrading from v0.1.1** below before you update.

### Upgrading from v0.1.1

**Who is affected:** every project. The changes that most often need action
are `server.trusted_proxies` behind a reverse proxy, `app.debug` now
defaulting to `false`, and the foreign keys the next `forge makemigrations`
adds.

1. Update the CLI and the library:
   `go install github.com/forgego/forge/cmd/forge@v0.2.0` and
   `go get github.com/forgego/forge@v0.2.0 && go mod tidy`.
   Projects created by v0.1.1 pin `v0.1.0` in `go.mod`, so check that the
   `require` line now says `v0.2.0`.
2. Regenerate each app: `forge generate --models ./app/<app> --output ./app/<app>`
   (plus `--api` if you use it). Expected
   changes in generated code: whitespace in `gen.go`; `api_gen.go` viewsets
   set `RejectUnknownRequestFields`.
3. Code changes:
   - Projects created by v0.1.0 or v0.1.1 do not compile until you delete the
     `//go:embed static templates` line, the `staticFiles` variable and the
     `embed` import from `cmd/server/main.go`.
   - An app created with `forge add app --example` declares
     `type Example struct { schema.BaseSchema }`, so the ORM and the admin
     read and write no columns. Change it to
     `type Example struct { ExampleGenerated }` and run `forge generate`.
   - Code that relied on a schema `Default` filling a field left at its zero
     value in `Manager.Create`, `Save` or `BulkCreate` must build the
     instance with `manager.New()` or call `orm.ApplyDefaults` first. Neither
     touches a default naming a database function (`now()`,
     `gen_random_uuid()`, ...): the database fills the column on insert, and a
     `Required` field with such a default may be left empty.
   - A `Default("...")` holding an SQL expression other than the known
     functions (`now()`, `CURRENT_TIMESTAMP`, `gen_random_uuid()`, ...)
     becomes `DBDefault("...")`.
   - `PUT` requests must send every required writable field; use `PATCH` for
     partial updates. Clients of generated viewsets must stop sending unknown
     keys, which are now a 400.
   - Clients of the REST API receive and send `TimeField` values as RFC 3339
     timestamps (`2026-09-19T14:30:05Z`) instead of `14:30:05`; a bare time
     of day is now a 400.
   - Anonymous requests that fail a permission answer 401 with a
     `WWW-Authenticate` challenge when the first authentication class can
     issue one (token, JWT, basic), instead of 403. Clients that treated 403
     as "log in" must accept 401.
   - `api.Router.Register` panics at startup when a viewset's queryset lacks
     an operation its routes need; set `ReadOnly` or complete the queryset.
   - The admin API answers 503 instead of 401 when its token or lockout store
     cannot be reached.
   - Optional: replace `srv.Start()` with `srv.StartWithGracefulShutdown()`
     to drain requests on SIGTERM.
4. Configuration changes:
   - `server.trusted_proxies` (`FORGE_SERVER_TRUSTED_PROXIES`): list your
     reverse proxy, or every request appears to come from it and forwarding
     headers are ignored.
   - `app.debug` defaults to `false`. Set `FORGE_APP_DEBUG=true` for local
     development; a production server refuses to start with it on.
   - `/info` is off unless `server.info_endpoint: true`.
   - New and optional: `server.stores` (`memory` by default). To share state
     between instances, set it to `database`, pass
     `server.WithDatabase(database)` to `server.NewServer`, call
     `adminSite.UseStores(ctx, settings.Server.Stores)` after `SetDB`, and run
     `forge migrate up` (see the [deployment guide](/docs/deployment/)).
5. Migrations: back up first. Then run
   `forge makemigrations upgrade --auto --models ./app/<app>`: it now emits
   the foreign keys and `Meta.Constraints` it used to skip, so the migration
   fails on rows that violate them; delete or fix orphan rows first. A column
   change it cannot express (on PostgreSQL: adding or removing `Unique`, the
   primary key, identity, a generated expression or `DBColumn`; on SQLite:
   any column change) is an error instead of an empty or comment-only
   migration: write that migration by hand (see the migrations guide). A
   SQLite project with an earlier comment-only migration for a column change
   must replace it with a hand-written table rebuild. v0.1.1 could leave an
   empty migration pair behind after `makemigrations --auto` found no
   changes; `forge migrate up` rejects a pending up file with `SQL is empty`,
   so delete each empty pair (both files) before applying. Then run
   `forge migrate up`.
   With `server.stores: database`, `forge migrate up` also creates the
   `forge_*` framework tables.
6. Verify: `forge migrate status` reports `Status: OK` (a failed migration
   leaves it dirty; follow `forge migrate recover`), then run the smoke check
   from the [deployment guide](/docs/deployment/).

**Rolling back:** reinstall and require `v0.1.1`, then regenerate with the
v0.1.1 CLI, for each app: `forge generate --models ./app/<app> --output
./app/<app>` (add `--api` for generated API code):
code generated by v0.2 refers to fields that v0.1.1 does not have. The generated foreign key
and constraint migrations have down migrations: undo them with
`forge migrate rollback` (one migration per run) before switching back. The framework tables are not touched by
`forge migrate rollback`; if you enabled database stores, set
`server.stores: memory` and drop `forge_sessions`, `forge_rate_limits`,
`forge_admin_tokens`, `forge_admin_saved_views`, `forge_admin_log` and
`forge_framework_migrations` by hand. Keep `forge_migration_checksums`: it
holds the recorded hashes of your applied migrations.

### Breaking

- The REST API reads and writes `TimeField` as an RFC 3339 timestamp, like
  `DateTimeField`, and the admin edits it as a date and time. `TimeField` is
  a timestamp column; the API formatted it as `15:04:05`, so a response sent
  back as a `PUT` body stored the time with a zero date (#299).
- `forge makemigrations` fails instead of writing a comment-only migration
  for a SQLite column change, and instead of an empty migration for a
  PostgreSQL column change it cannot express (`Unique`, primary key,
  identity, generated expression, `DBColumn`). A SQLite project whose history
  already holds such a comment-only migration must hand-write the table
  rebuild before `makemigrations` succeeds again; the migrations guide covers
  renames and hand-written SQLite table rebuilds (#296).
- A `Default` string is quoted as a literal unless it is one of the known
  SQL functions (`now()`, `CURRENT_TIMESTAMP`, `gen_random_uuid()` and the
  like). A model with an expression default such as
  `Default("uuid_generate_v7()")` must switch to `DBDefault(...)` before the
  next `makemigrations`, which would otherwise propose changing the column
  default to the quoted text (#296).
- `server.RealIP`, part of `DefaultMiddlewares`, honors `X-Forwarded-For`
  and `X-Real-IP` only from peers listed in `server.trusted_proxies`. Before,
  it trusted the headers from any client, so a request could set its own
  `RemoteAddr`. Behind a reverse proxy, list the proxy there to keep seeing
  client addresses (#298).
- `Manager.Create`, `Save` and `BulkCreate` write explicit zero values
  (`false`, `0`, `""`) on fields with a schema `Default`, instead of leaving
  the column out, so the column's `DEFAULT` clause no longer replaces them
  (required fields were already written). Code that relied on a `Default`
  filling a field the struct left at zero must build the instance with
  `manager.New()`
  (or call `orm.ApplyDefaults`) first. The admin create endpoint and the
  public REST API apply the schema `Default` to a field the request omits.
  Optional fields without a `Default`, zero foreign keys, `nil` pointers,
  zero `time.Time` values, zero unique optional fields, zero `DBDefault`
  fields and `""` on non-text columns (UUID, JSON, decimal, date and time,
  non-text custom `DBType`) are still left to the database, so they store
  `NULL` or the column default as before (#291).
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

- A `Default` naming a database function (`now()`, `CURRENT_TIMESTAMP`,
  `gen_random_uuid()`, ... as the migration writes it unquoted) is left for
  the database: `Manager.New`, `orm.ApplyDefaults`, the REST API and the
  admin no longer assign the function name as text (a UUID column stored
  `gen_random_uuid()`) or fail converting it to `time.Time` (every admin
  create returned 500), and `Create` leaves such a column out of the INSERT.
  The new `schema.IsDatabaseFunctionDefault` reports these defaults (#299).
- A malformed `logging.outputs` value, which `config.LoadSettings` ignores so
  logs go to the console, is reported as a warning when the server starts.
  The new `config.Config.SettingsWarnings` returns it. Before, it was
  silently ignored.
- The API field contract documents which zero values `POST` still leaves to
  the database: `DBDefault` fields, optional foreign keys and optional unique
  fields, whether the key is omitted or sent as zero.
- The REST API overview and ViewSets pages use the real API
  (`api.NewBaseViewSet`, `api.NewRouter`, `Router.Register`, `Router.Action`
  and `throttling.NewUserRateThrottle`) instead of `api.ModelViewSet[T]`,
  `api.RegisterViewSet`, `CustomActions` and a `GetQuerySet` override, which
  do not exist. Detail routes are documented without a trailing slash.
- The ecommerce example (server and seed script) builds its PostgreSQL
  connection string with the new exported `db.PostgresKeywordDSN`, and the
  test database helper builds an escaped URL, so a password or database name
  with spaces or quotes connects. Before, both formatted unquoted
  keyword/value strings.
- The admin create form starts with each field's schema `Default` (from the
  metadata `default_value`), and the admin create endpoint applies the
  `Default` of a field the request omits, as the public REST API does. Before,
  a `Default(true)` checkbox started unchecked and an omitted field stored
  the zero value. Admin metadata no longer fails to encode for a model with
  a callable `Default` such as `time.Now`; that field reports no
  `default_value` (#298).
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
- `api.Router.Register` also checks the types a queryset's `Filter`,
  `OrderBy`, `Offset` and `Limit` return, so a chain result without the
  `Count` or `All` that `list` calls on it fails at startup. A result declared
  as an interface is checked per request: a mismatch answers 500 and logs the
  resource and the problem instead of panicking in reflection (#294).
- `forge version` prints the version the binary was installed from, and
  `forge new` pins that version in the new project's `go.mod` instead of
  `v0.1.0` (#286).
- `forge generate` warns about struct fields that have no schema entry (#287).
- `forge add app --example` writes a model that embeds `ExampleGenerated`,
  so the ORM and the admin see its columns. Before, `Example` declared only
  `schema.BaseSchema`: lists returned empty objects and creates wrote no
  columns.
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
  cast and operator DB defaults (`DBDefault("'{}'::jsonb")` on a `JSONB`
  column,
  `DBDefault("'x' || 'y'")`), a changed or added `DBDefault` on PostgreSQL,
  a table dropped and later created again, or a hand-written
  `ALTER TABLE .. RENAME COLUMN` or `RENAME TO` (#296, #292).
- `forge makemigrations` writes no migration when the changes render no SQL.
  The down migration of a foreign key whose target table changed restores
  the old target (#296).

### Added

- The [support contract](/docs/status/), the [deployment guide](/docs/deployment/),
  the [API field contract](/docs/api/field-contract/),
  [data-access requirements](/docs/api/data-access/) and the release
  process (`docs/RELEASING.md` in the repository).
- CI gates: a fresh `forge new` PostgreSQL application exercised over HTTP,
  a PostgreSQL schema lifecycle test, schema DSL fixtures, admin browser
  journeys against PostgreSQL, and an install smoke test from the public
  module proxy on every tag and weekly.
- Shared stores for running more than one instance (#293, #298). With the new
  `server.stores: database` setting, sessions (`forge_session`), API
  throttling counters, admin bearer tokens, the admin login lockout, saved
  views and admin change history are kept in PostgreSQL or SQLite instead of
  process memory, so they survive restarts and are shared by every instance
  on the database. Throttle counts are one atomic upsert per request and
  expired rows are deleted by the instances without a cron job. The tables
  are framework migrations embedded in Forge: `forge migrate up` applies
  them before the application's migrations and tracks them in
  `forge_framework_migrations` (`stores.Migrate` does the same from Go), and
  `forge migrate status` shows their version. `forge new` projects set
  `stores: database` and pass the database to the server and the admin
  (`server.WithDatabase`, `admin.Site.UseStores`). The library default stays
  `memory`, so existing projects are unchanged; to switch, add those two
  calls to `main.go`, set `server.stores: database` and run
  `forge migrate up`. With database stores, `server.NewServer` and
  `UseStores` fail at startup until the tables exist. New APIs: package
  `stores` and `stores/adminstore`, `server.Option`, `Server.SessionManager`,
  `throttling.SetDefaultStoreFactory`, `throttling.ContextStore`, `rest.TokenStore`,
  `rest.SavedViewStore`, `rest.LoginAttemptStore` and `Router.SetStores`.
  The admin API answers 503 instead of 401 when its token or lockout store
  cannot be reached, and admin history records a logged-in admin by username
  instead of the printed user map.
  Database throttle and admin lockout queries are bound to the request's
  context and to `stores.DefaultQueryTimeout` (two seconds, changed with
  `WithTimeout`), so a stalled database cannot hold requests: a throttle then
  allows the request and logs the error, and admin login answers 503. A server created with
  `server.stores: memory` resets throttles to process memory even after an
  earlier server in the process used the database, and
  `Site.UseStores(ctx, "memory")` (or the new `Site.UseMemoryStores`) undoes
  an earlier `UseDatabaseStores`. `forge migrate up --dry-run` rejects an
  invalid `server.stores` like the real run does.
- The admin create endpoint no longer replaces an explicit `null` with the
  field's schema `Default`: a nullable (pointer) field stores `NULL`, and a
  non-pointer field stores its zero value (`false`, `0`, `""`), as it did
  before defaults were applied on create. A sent list or map replaces the
  default instead of being merged into it.
- Admin field metadata carries `has_default`, set for any schema `Default`
  including a callable one such as `time.Now`. The create form no longer
  requires such a field in the browser, since the server fills it.

### Known issues

- `forge makemigrations` skips `schema.Check(...)` and `schema.UniqueOn(...)`
  helper calls in `Meta` and prints `No changes detected`; only
  `forge generate` warns about them. Write constraints as
  `schema.Constraint{...}` literals.
- `forge add api <name> --model <Model>` followed by `forge generate --api`
  declares `<Model>Serializer` twice, so the package does not compile;
  delete one of the two.
- `forge add app` does not validate the name as a Go package name.
- The `forge new` next-step hint prints `forge makemigrations` without the
  migration name and `--models` argument an `app/` project needs.
- A schema `Default` given as a Go function (`Default(time.Now)`) is applied by
  the ORM and the admin but not by the REST API: a `POST` that omits the
  field answers 400. Send the field, or set the value in a `BeforeCreate` hook.
- `forge createsuperuser` fails in a fresh project until a `users` table
  exists.
- `/info` reports `app.version` (default `0.1.0`), not the Forge version.

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
contains no Go packages and v1.0.1 points at the v0.1.0 commit. Both are retracted in `go.mod`, and
`go get github.com/forgego/forge@latest` skips them. They are not
Forge 1.0.
