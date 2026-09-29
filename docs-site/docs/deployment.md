---
sidebar_position: 22
title: Deployment
description: The minimum production deployment contract for Forge on PostgreSQL, with one or more instances.
image: /social-card.png
---

# Deployment

This guide covers the deployment Forge supports today: **one or more
application instances** built from a `forge new` project, talking to **one
PostgreSQL 15 database**, behind a reverse proxy that terminates TLS. More
than one instance needs `server.stores: database` (the `forge new` default);
see [multiple instances](#multiple-instances). The
[support contract](/docs/status/) defines the tiers used below.

Every statement here was checked against the code at the time of writing, and
the runbook at the end was rehearsed against PostgreSQL. Where Forge lacks
something a production service normally needs, this guide says so instead of
describing a workaround as a feature.

## The release artifact

A deployable release of your application is:

1. **The server binary**, built from your project with the Forge version pinned
   in `go.mod`:

   ```bash
   CGO_ENABLED=0 go build -trimpath -o server ./cmd/server
   ```

   `CGO_ENABLED=0` works for PostgreSQL. SQLite needs cgo and is experimental.
2. **`config/config.yaml`** with non-secret settings. Forge looks for
   `config.yaml` in `.`, `./config` and `../config` relative to the working
   directory, so start the binary from the directory that contains `config/`.
3. **The `migrations/` directory**, applied with the `forge` CLI. With
   `server.stores: database`, `forge migrate up` also creates Forge's own
   store tables (see [shared state](#shared-state)).
4. **The `forge` CLI at the same version as the library**, so migrations are
   applied by the code that generated them:

   ```bash
   go install github.com/forgego/forge/cmd/forge@$(go list -m -f '{{.Version}}' github.com/forgego/forge)
   ```

`forge new --docker` writes a `Dockerfile` and `compose.yaml` you can start
from. The image contains only the binary and `config/`. It does not include
`migrations/` or the `forge` CLI, runs as root on `alpine:latest`, and has no
`HEALTHCHECK`; the compose file uses fixed development passwords. Treat both as
a development starting point.

## PostgreSQL

- Version: 15 is the tested version. Newer majors are untested.
- Create a dedicated database and role for the application. The role needs to
  create tables for `forge migrate up`; if you run migrations with a separate
  owner role, the application role only needs data privileges.
- Set `database.sslmode` (`FORGE_DATABASE_SSLMODE`) to `require` or
  `verify-full` for any database reached over a network. The default is
  `disable`.
- Pool settings: `database.max_open_conns` (default 25),
  `database.max_idle_conns` (10), `database.conn_max_lifetime` (5m),
  `database.conn_max_idle_time` (2m). Keep `max_open_conns` below the
  server's `max_connections`, minus what migrations and administration need.
- Forge builds a `key=value` connection string with every value quoted, so a
  password may contain spaces, quotes and backslashes. An empty
  `database.password` is left out, so `PGPASSWORD` or a pgpass file can
  supply it.
- There is no `DATABASE_URL` setting.

## Settings and secrets

Every key can be set from the environment as `FORGE_` plus the key in upper
case with dots replaced by underscores. Environment variables win over
`config.yaml`. Supply secrets only through the environment or your secrets
manager; never ship the `.env` file that `forge new` generates for local use.

| Variable | Production value | Why |
| --- | --- | --- |
| `FORGE_APP_ENV` | `production` | Turns on secret validation and `Secure` cookies. Case-insensitive. |
| `FORGE_APP_DEBUG` | `false` (the default) | `true` allows the profiling routes (with `server.enable_profiling`) and is rejected in production; log level and format come from `logging.*`. `forge new` sets it to `true` only in the local `.env`. |
| `FORGE_LOGGING_LEVEL` | `info` | The `config.yaml` from `forge new` sets `info`; its local `.env` sets `debug`. |
| `FORGE_LOGGING_FORMAT` | `json` | The `config.yaml` from `forge new` sets `json`; its local `.env` sets `console`. |
| `FORGE_SERVER_HOST` | `0.0.0.0` or the proxy-facing address | The default `localhost` is unreachable from outside a container. |
| `FORGE_SERVER_PORT` | your port | Default `8000`. |
| `FORGE_SERVER_TRUSTED_PROXIES` | your proxy's address or CIDR | Forwarding headers are honored only from these peers, for rate limiting and the admin login lockout. Empty by default: the TCP peer address is used. |
| `FORGE_SERVER_STORES` | `database` | Keeps sessions, API throttling counters and the admin's tokens, login lockout, saved views and change history in PostgreSQL. `forge new` writes `stores: database` to `config.yaml`; the library default is `memory`. |
| `FORGE_DATABASE_HOST`, `_PORT`, `_NAME`, `_USER`, `_PASSWORD`, `_SSLMODE` | your database | |
| `FORGE_SECURITY_SECRET_KEY`, `FORGE_SECURITY_SESSION_SECRET`, `FORGE_SECURITY_CSRF_SECRET_KEY` | three independent random values, for example `openssl rand -hex 32` | Required in production. Rotating the session or CSRF secret invalidates existing session and CSRF cookies. |
| `FORGE_ADMIN_USERNAME`, `FORGE_ADMIN_PASSWORD` | only if you use environment-based admin login | Without them, and without a login authenticator, admin login answers `503 admin_login_disabled`. |

**What is checked before the server listens** (in a project created by
`forge new`):

- The database: `db.NewDBFromConfig` pings it, and the process exits with
  status 1 if it cannot connect.
- The secrets: with `app.env` set to production, `Server.Start` refuses to
  listen unless all three secrets are set explicitly (not empty, not a
  placeholder such as `change-me`, not generated at startup) and the process
  exits with status 1.
- Debug mode: with `app.env` set to production, `Server.Start` also refuses to
  listen while `app.debug` is true.
- The store tables: with `server.stores: database`, `server.NewServer` and
  `admin.Site.UseStores` fail, and the process exits with status 1, until
  `forge migrate up` has created the framework store tables.

**What is not checked**: `server.host`, `database.sslmode`, admin
credentials, and whether TLS is in front of the server. The server speaks
plain HTTP only; terminate TLS at the proxy. `Secure` cookies are only useful
if clients reach the proxy over HTTPS.

## Health and readiness

The server registers these routes (`server.health_check_path` defaults to
`/health`):

| Route | Behavior |
| --- | --- |
| `/health` | Runs every registered check. 200 if all pass, 503 otherwise. |
| `/health/ready` | Same checks, reported as readiness. |
| `/health/live` | Always 200 while the process is serving. |
| `/info` | Only when `server.info_endpoint` is true (default false). Public: app name, version, environment, debug flag, uptime. |

**No check is registered by default**, so `/health` and `/health/ready` return
200 even when the database is down. Register one in `main.go` after connecting:

```go
server.RegisterHealthCheckFunc("database", func(ctx context.Context) error {
	if err := database.PingContext(ctx); err != nil {
		return err
	}
	return nil
})
```

A failing check is reported publicly only as `unhealthy` or `not ready`; the
error text goes to the server log (standard `log` output), quoted.

## Database loss

- **At startup**: the process exits with status 1 (see above). Let your
  supervisor restart it with a backoff.
- **While running**: the process keeps running. Requests that need the
  database fail with an error response; `database/sql` opens new connections
  when the database comes back, without a restart. Readiness reports the outage
  only if you registered a database check. This recovery path is not covered
  by an automated test.

## Graceful shutdown

Projects created by `forge new` start the server with
`Server.StartWithGracefulShutdown`. On SIGINT or SIGTERM the server stops
accepting connections, waits up to `server.graceful_timeout` seconds (default
30) for in-flight requests, closes the database and exits 0. If requests are
still running when the timeout expires, they are cut off and the process
exits 1: `StartWithGracefulShutdown` returns `graceful shutdown: context
deadline exceeded` and the generated `main.go` passes it to `log.Fatal`,
which exits without running deferred calls such as closing the database. A
second SIGINT or SIGTERM during the wait ends the process at once with the
default signal behavior, so a wrapper that forwards both SIGINT and SIGTERM
skips the drain; send one signal. Set your supervisor's stop timeout
above `graceful_timeout` (for example `docker stop -t 40`, or
`terminationGracePeriodSeconds` if you run it on a cluster yourself).

Projects created before this change call `srv.Start()`, which has no signal
handling; switch them to `StartWithGracefulShutdown`. The ecommerce example
also still calls `ListenAndServe` directly.

## Logs and redaction

Each request is logged with method, path, query string, status, duration,
client address, user agent and request ID.

- The values of these query parameters are replaced with `REDACTED`:
  `api_key`, `apikey`, `key`, `token`, `access_token`, `refresh_token`,
  `password`, `secret`, `signature`, `session`, `session_key`.
- Request and response headers, cookies and bodies are not logged.
- Configuration warnings name missing secrets but never print values.
- The API error handler logs a PostgreSQL or SQLite driver error, returned
  or panicked, by its Go type, SQLSTATE (or SQLite result code) and the
  constraint, table and column it names; the driver message and detail,
  which can contain row values, are left out.
- **Not redacted**: URL paths are logged verbatim, so keep secrets out of
  paths. Other error text is logged as-is, and non-error panic values are
  logged as-is. Treat logs as containing personal data.
- The `main.go` from `forge new` builds its logger from the `logging.*` keys
  with `log.NewLoggerFromSettings(settings.Logging)`; see
  [logging](/docs/config/logging/).

## Shared state

`server.stores` decides where the state below lives. With `database` (what
`forge new` writes to `config.yaml`) it is kept in PostgreSQL, survives a
restart and is shared by every instance. With `memory` (the library default
when the key is absent) each process has its own copy and a restart loses it.

| State | `server.stores: database` | `server.stores: memory` |
| --- | --- | --- |
| Server sessions (`forge_session` cookie) | `forge_sessions` | In-memory `scs` store; a restart ends all sessions. |
| API throttling counters (throttles without their own `WithStore`) | `forge_rate_limits` | Per process; reset on restart. |
| Admin bearer tokens (24 h lifetime) | `forge_admin_tokens` (SHA-256 hashes only) | Every admin user is signed out on restart. |
| Admin login attempt limiter, keyed by client IP and by username | `forge_rate_limits` | Counters reset on restart. |
| Admin saved views | `forge_admin_saved_views` | Lost on restart. |
| Admin change history | `forge_admin_log` | Last 1000 entries per registered model, lost on restart. |

The tables are created by framework migrations embedded in Forge.
`forge migrate up` applies them before your own migrations when
`server.stores` is `database`, and records them in
`forge_framework_migrations`, separately from your `schema_migrations`, so
they never appear in `migrations/` or in `forge makemigrations` output.
`forge migrate status` shows their version. `forge migrate rollback` does not
touch them; to remove them, drop the five `forge_*` tables and
`forge_framework_migrations`. To apply them from Go instead, call
`stores.Migrate(ctx, database)`.

The generated `main.go` wires the setting in with
`server.NewServer(cfg, settings, logger, server.WithDatabase(database))` and
`adminSite.UseStores(ctx, settings.Server.Stores)`. Older projects that call
`server.NewServer(cfg, settings, logger)` keep in-memory stores; to switch,
add both calls, set `server.stores: database` and run `forge migrate up`.

Expired sessions, tokens and counter rows are deleted by the instances
themselves, at most once a minute per table, on the write path; no cron job is
needed. Behind a proxy, list it in `server.trusted_proxies`
(`FORGE_SERVER_TRUSTED_PROXIES`); otherwise every client shares the proxy's
address, and failed logins from one client can lock out everyone. Throttle
counts use one atomic `INSERT ... ON CONFLICT` per request.
If the database is unreachable or does not answer within two seconds,
throttled requests are allowed (and logged), while admin logins answer 503
because the lockout cannot be checked.

This state stays in process memory whatever `server.stores` says:

| State | Effect of a restart or a second instance |
| --- | --- |
| `server.RateLimitByIP` / `RateLimitByUser` middleware | Per process; reset on restart. |
| A throttle given its own store with `WithStore` | Whatever that store does. |
| `api/caching` cache, idempotency keys | Per process; lost on restart. |
| Secrets generated at startup (non-production only) | New values each start, so existing cookies stop working. |

These survive a restart in any configuration because they live in
PostgreSQL: your models, the `identity` package's users, sessions
(`user_sessions`) and tokens, and the migration bookkeeping. CSRF tokens are
signed cookies and survive as long as the CSRF secret does.

## Audit history

With `server.stores: database` the admin's change history is durable: every
add, change and delete is a row in `forge_admin_log`, kept until you delete
it (the history view shows the newest 1000 entries). It is not a compliance
audit log: anyone with write access to the database can change it. Models
whose `admin.Config` sets its own `HistoryManager` keep it; the compatibility
`admin.HistoryManager` writes to the shared table. With `memory` it is capped
at 1000 entries per registered model and lost on restart.

## Multiple instances

Running several instances behind a load balancer is supported with
`server.stores: database`: a session or admin token issued by one instance is
valid on every other, change history and saved views are shared, and API
throttling limits and login lockouts count requests across all instances. A
test starts two instances on one PostgreSQL database and checks each of these
(`tests/integration/stores`).

Every instance must use the same database and the same
`FORGE_SECURITY_*` secrets. The per-process state listed under
[shared state](#shared-state) (the `RateLimitByIP` middleware, caches,
idempotency keys) stays per instance. With `server.stores: memory`, more than
one instance is not supported: admin users are signed out when a request lands
on the other instance, limits are per instance and history is split.

Run migrations from one place only. `forge migrate up` holds a PostgreSQL
advisory lock while it applies migrations, so an accidental second run waits
instead of applying twice, but do not run migrations from every instance at
start-up.

## Deploy, back up, roll back

Run these from the project directory with the production `FORGE_DATABASE_*`
variables set and the `forge` CLI at the library's version.

**1. Back up and inspect.**

```bash
pg_dump -Fc -h "$FORGE_DATABASE_HOST" -U "$FORGE_DATABASE_USER" -f pre-deploy.dump "$FORGE_DATABASE_NAME"
forge migrate status            # must not say DIRTY
forge migrate recover --verify  # applied files still match their recorded checksums
forge migrate up --dry-run      # review the SQL that will run
```

**2. Migrate and switch.**

```bash
forge migrate up
# stop the old binary (SIGTERM), start the new one
./smoke-check.sh https://app.example.com /admin
```

Keep migrations backward compatible with the running binary where you can
(add columns and tables first, remove them in a later release), so the old
binary keeps working if you have to switch back.

**3. Roll back, choosing the least destructive option.**

- *The new binary is bad, the schema is fine*: start the previous binary.
- *A migration must be undone*: `forge migrate rollback` runs the newest
  migration's `.down.sql`. Run it once per migration to undo, then start the
  previous binary.
- *A migration failed half-way*: `forge migrate status` reports `DIRTY`. Fix
  the database by hand, then mark it clean with `forge migrate recover --clean`
  or set the version with `forge migrate force <version>`. See the
  [migrations guide](/docs/migrations/).
- *Data is damaged*: restore the backup. Everything written after the backup
  is lost.

  ```bash
  pg_restore --clean --if-exists -h "$FORGE_DATABASE_HOST" -U "$FORGE_DATABASE_USER" -d "$FORGE_DATABASE_NAME" pre-deploy.dump
  forge migrate status
  ```

## Smoke check

Save this as `smoke-check.sh` in your project and run it after every deploy
and every rollback. It exits non-zero on the first failure.

```bash
#!/usr/bin/env bash
# smoke-check.sh BASE_URL [ADMIN_PATH]
# Run from the project directory after a deploy. Exits non-zero on the first failure.
set -euo pipefail

base="${1:?usage: smoke-check.sh https://app.example.com [/admin]}"
base="${base%/}"
admin="${2-/admin}"

fail() { echo "FAIL: $*" >&2; exit 1; }

expect_status() { # path wanted-status
  local code
  code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 5 "$base$1") || fail "$1 unreachable"
  [ "$code" = "$2" ] || fail "$1 returned $code, want $2"
  echo "ok   $1 -> $code"
}

expect_status /health/live 200
expect_status /health/ready 200

body=$(curl -sS --max-time 5 "$base/health") || fail "/health unreachable"
echo "$body" | grep -q '"status":"healthy"' || fail "/health reported: $body"
echo "ok   /health reports healthy"

if [ -n "$admin" ]; then
  expect_status "$admin/" 200
fi

# Behind HTTPS every cookie must be Secure (app.env=production).
case "$base" in
  https://*)
    if curl -sS -D - -o /dev/null --max-time 5 "$base/health/live" | grep -i '^set-cookie:' | grep -viq 'secure'; then
      fail "a cookie is missing the Secure attribute; is FORGE_APP_ENV=production?"
    fi
    echo "ok   cookies are Secure"
    ;;
esac

# Migration state, when the forge CLI and the migrations directory are here.
if [ -d migrations ] && command -v forge >/dev/null 2>&1; then
  status=$(forge migrate status 2>&1) || fail "forge migrate status failed: $status"
  if echo "$status" | grep -q 'DIRTY'; then fail "migrations are dirty"; fi
  if echo "$status" | grep -q 'Pending Migrations'; then fail "migrations are pending"; fi
  echo "ok   migrations applied and clean"
fi

echo "smoke check passed"
```

The readiness checks are only as good as the health checks you register; with
none registered, a database outage still passes. Pass an empty second argument
(`''`) if the admin is disabled.
