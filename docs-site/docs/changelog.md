---
sidebar_position: 30
description: forge release notes and updates.
---

# Changelog

Forge is pre-1.0; see the [support contract](/docs/status/) for what a v0.x
minor or patch release may change. Entries cite the pull request that made
the change.

## Unreleased

### Fixed

- `forge version` prints the version the binary was installed from, and
  `forge new` pins that version in the new project's `go.mod` instead of
  `v0.1.0` (#286).
- `forge generate` warns about struct fields that have no schema entry (#287).
- Projects created by `forge new` compile again: the generated `main.go` no
  longer embeds `static` and `templates` directories that do not exist next
  to it.
- `Server.StartWithGracefulShutdown` now handles SIGINT and SIGTERM: it stops
  accepting connections and waits up to `server.graceful_timeout` for
  in-flight requests. Projects created by `forge new` use it.
- `forge new --docker` builds with `golang:1.26-alpine` (the 1.25 image could
  not build a module that requires Go 1.26) and its compose file sets
  `FORGE_SERVER_HOST=0.0.0.0` so the published port reaches the server.
- Session and CSRF cookies are marked `Secure` for any spelling of
  `app.env: production`, matching the production secret check.

### Added

- The [support contract](/docs/status/), the [deployment guide](/docs/deployment/)
  and the release process (`docs/RELEASING.md` in the repository).
- An install smoke test that installs the CLI from the public module proxy on
  every tag and weekly.

### Known issues

- `forge makemigrations --auto` with no model changes writes an empty
  migration pair, which `forge migrate up` rejects with `SQL is empty`.
  Delete the empty files.

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
