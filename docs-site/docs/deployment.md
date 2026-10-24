---
sidebar_position: 22
title: Deployment
description: The minimum production deployment contract for Forge on PostgreSQL, with one or more instances.
image: /social-card.png
---

# Deployment

This guide covers the deployment Forge supports today: **one or more
application instances** built from a `forge new` project, talking to **one
PostgreSQL 15 database**, behind a reverse proxy that terminates TLS.
Framework state always lives in the database; see
[multiple instances](#multiple-instances). The
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
3. **The `migrations/` directory**, applied with the `forge` CLI. `forge
   migrate up` also creates Forge's own store tables (see
   [shared state](#shared-state)).
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
- The store tables: `server.NewServer` fails, and the process exits with
  status 1, until `forge migrate up` has created the framework store tables.

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

Framework state always lives in the application database (PostgreSQL or
SQLite). It survives restarts and is shared by every instance. There is no
memory option.

| State | Table |
| --- | --- |
| Server sessions (`forge_session` cookie) | `forge_sessions` |
| API throttling counters (throttles without their own `WithStore`) | `forge_rate_limits` |
| Admin bearer tokens (24 h lifetime) | `forge_admin_tokens` (SHA-256 hashes only) |
| Admin login attempt limiter, keyed by client IP and by username | `forge_rate_limits` |
| Admin saved views | `forge_admin_saved_views` |
| Admin change history | `forge_admin_log` |

The tables are created by framework migrations embedded in Forge.
`forge migrate up` always applies them before your own migrations, and records
them in `forge_framework_migrations`, separately from your
`schema_migrations`, so they never appear in `migrations/` or in `forge
makemigrations` output. Saved filters (`forge_saved_filters`) and idempotency
keys (`idempotency_keys`) also use framework tables. `forge migrate status`
shows their version. `forge migrate rollback` does not touch them; to remove
them, drop the framework tables and `forge_framework_migrations`. To apply
them from Go instead, call `stores.Migrate(ctx, database)`.

`server.NewServer` requires `server.WithDatabase(database)`. It fails to start
if framework tables are missing, reporting that you must run `forge migrate
up`. In the admin, `adminSite.SetDB(database)` applies the database stores
automatically. If the tables are missing, the admin answers 503 "admin stores
unavailable: run forge migrate up".

Expired sessions, tokens and counter rows are deleted by the instances
themselves, at most once a minute per table, on the write path; no cron job is
needed. Behind a proxy, list it in `server.trusted_proxies`
(`FORGE_SERVER_TRUSTED_PROXIES`); otherwise every client shares the proxy's
address, and failed logins from one client can lock out everyone. Throttle
counts use one atomic `INSERT ... ON CONFLICT` per request.
Each throttle and lockout query waits at most two seconds for the database.
If it is unreachable or that runs out, throttled requests are allowed (and
logged), while admin logins answer 503 because the lockout cannot be checked.

This state stays in process memory:

| State | Effect of a restart or a second instance |
| --- | --- |
| `server.RateLimitByIP` / `RateLimitByUser` middleware | Per process; reset on restart. |
| A throttle given its own store with `WithStore` | Whatever that store does. |
| Response cache (`api/caching` MemoryCache) | Per process; lost on restart. Shared or query caching is planned, not available. |
| Secrets generated at startup (non-production only) | New values each start, so existing cookies stop working. |

These survive a restart in any configuration because they live in
PostgreSQL: your models, the `identity` package's users, sessions
(`user_sessions`) and tokens, and the migration bookkeeping. CSRF tokens are
signed cookies and survive as long as the CSRF secret does.

## Audit history

The admin's change history is durable: every add, change and delete is a row in
`forge_admin_log`, kept until you delete it (the history view shows the newest
1000 entries). It is not a compliance audit log: anyone with write access to
the database can change it. Models whose `admin.Config` sets its own
`HistoryManager` keep it; the compatibility `admin.HistoryManager` writes to
the shared table. Accessing admin history without a database returns
`admin.ErrNoHistoryBackend`.

## Multiple instances

Running several instances behind a load balancer is supported: a session or
admin token issued by one instance is valid on every other, change history and
saved views are shared, and API throttling limits and login lockouts count
requests across all instances. A test starts two instances on one PostgreSQL
database and checks each of these (`tests/integration/stores`).

Every instance must use the same database and the same
`FORGE_SECURITY_*` secrets. The per-process state listed under
[shared state](#shared-state) (such as the `RateLimitByIP` middleware and
in-memory response cache) stays per instance.

Run migrations from one place only. `forge migrate up` holds a PostgreSQL
advisory lock while it applies migrations, so an accidental second run waits
instead of applying twice, but do not run migrations from every instance at
start-up.

## Production deployment path

This numbered production deployment path brings a release into service,
verifies its health, and provides the exact rollback procedures.

### 1. Configuration

Configure the release using non-secret values in `config/config.yaml` or
environment variables prefixed with `FORGE_`. Environment variables override
configuration file keys:

- `app.env` (`FORGE_APP_ENV`): set to `production` (case-insensitive).
- `app.debug` (`FORGE_APP_DEBUG`): set to `false`. Production refuses startup if `app.debug` is true.
- `server.host` (`FORGE_SERVER_HOST`): bind address (`0.0.0.0` or container address; default `localhost`).
- `server.port` (`FORGE_SERVER_PORT`): listening port (default `8000`).
- `server.health_check_path` (`FORGE_SERVER_HEALTH_CHECK_PATH`): base path for health endpoints (default `/health`).
- `server.trusted_proxies` (`FORGE_SERVER_TRUSTED_PROXIES`): proxy IPs/CIDRs for client IP resolution.
- `server.graceful_timeout` (`FORGE_SERVER_GRACEFUL_TIMEOUT`): seconds to drain requests on SIGTERM (default `30`).
- `database.driver` (`FORGE_DATABASE_DRIVER`): database engine (`postgres` or `sqlite3`).
- `database.host`, `database.port`, `database.name`, `database.user`, `database.password`: database connection settings.
- `database.sslmode` (`FORGE_DATABASE_SSLMODE`): set to `require` or `verify-full` for networked databases.
- `database.max_open_conns`, `database.max_idle_conns`, `database.conn_max_lifetime`, `database.conn_max_idle_time`: connection pool limits.
- `logging.level` (`FORGE_LOGGING_LEVEL`), `logging.format` (`FORGE_LOGGING_FORMAT`): logging level (`info`) and format (`json`).

Compile the application server binary with the pinned Forge library version:

```bash
CGO_ENABLED=0 go build -trimpath -o server ./cmd/server
```

### 2. Secrets

Supply secrets **exclusively through environment variables** (or your platform's
secrets manager mapped to environment variables). Never commit secrets into
`config.yaml` or include development `.env` files in production images.

Required production secrets:

- `FORGE_SECURITY_SECRET_KEY`: signing key for JWTs and tokens.
- `FORGE_SECURITY_CSRF_SECRET_KEY`: signing key for CSRF protection cookies.
- `FORGE_SECURITY_SESSION_SECRET`: signing key for session cookies.

**Production validation (`validateProductionSecrets` in `server/server.go`):**
When `app.env` is set to `production`, `Server.Start` and `Server.NewServer`
refuse default and insecure secrets before the server begins listening:

- **No empty secrets**: all three keys must be explicitly set.
- **No placeholder values**: values starting with `change-me` (case-insensitive) or equal to `secret` or `default` are rejected.
- **No generated ephemeral secrets**: keys generated dynamically in process memory are rejected.
- **Minimum length**: each secret must be at least 32 bytes (`len >= 32`).
- **Distinct values**: all three secrets must be distinct from one another (for example, generated with `openssl rand -hex 32`).
- **Debug disabled**: `app.debug` must be `false`; if `app.debug` is `true`, startup is aborted.

If any secret check fails, the process exits with status 1.

### 3. Schema changes

Run migration commands from the project directory using the matching `forge` CLI version.
Framework store tables need `forge migrate up` before the application server starts:
`server.NewServer` checks for framework tables (`forge_sessions`, `forge_rate_limits`,
`forge_admin_tokens`, `forge_admin_saved_views`, `forge_admin_log`) and fails with exit
status 1 if they have not been created.

1. **Take a database backup**:
   ```bash
   pg_dump -Fc -h "$FORGE_DATABASE_HOST" -U "$FORGE_DATABASE_USER" -f pre-deploy.dump "$FORGE_DATABASE_NAME"
   ```
2. **Verify migration integrity**:
   ```bash
   forge migrate recover --verify
   ```
   Compares applied migration files against recorded SHA-256 baselines in
   `forge_migration_checksums` and checks for dirty state (`cli/commands/migrations/recover.go`).
3. **Inspect current migration status**:
   ```bash
   forge migrate status
   ```
   Ensures the database is not in a `DIRTY` state and lists pending migrations
   (`cli/commands/migrations/status.go`).
4. **Preview pending migrations (plan)**:
   ```bash
   forge migrate up --dry-run
   ```
   Previews the SQL statements that will execute without altering the database
   (`cli/commands/migrations/up.go`).
5. **Apply migrations**:
   ```bash
   forge migrate up
   ```
   Applies framework store tables first, then pending application migrations in order
   (`cli/commands/migrations/up.go`).
   If any pending migration contains destructive SQL (`DROP TABLE`, `DROP COLUMN`,
   `TRUNCATE`, `DELETE FROM`, `DROP SCHEMA`, `DROP DATABASE`), `forge migrate up`
   refuses to run and lists the destructive operations. To proceed after reviewing the
   SQL and backup, run:
   ```bash
   forge migrate up --allow-destructive
   ```

### 4. Health checks and server startup

Start the application binary:

```bash
./server
```

Verify service readiness and health using the endpoints registered by `server/server.go`
and implemented in `server/health.go` (prefixed by `server.health_check_path`, default `/health`):

1. **Readiness probe (`GET /health/ready`)**:
   Checks database connectivity (`database.Ping(ctx)`) and any registered health
   functions via `readinessHandler`. Returns HTTP 200 `{"status":"ready", ...}` when
   ready, or HTTP 503 `{"status":"not ready", ...}` if any check fails. Point load
   balancer and container readiness probes here before directing traffic to the instance.
2. **Health check (`GET /health`)**:
   Runs all health checks registered in `server.HealthRegistry` via `healthHandler`.
   Returns HTTP 200 `{"status":"healthy", ...}` when all pass, or HTTP 503
   `{"status":"unhealthy", ...}` if any fail.
3. **Liveness probe (`GET /health/live`)**:
   Simple liveness check via `LivenessHandler`. Always returns HTTP 200
   `{"status":"alive", ...}` while the HTTP process is serving requests.

You can also run the automated `./smoke-check.sh` script to verify endpoints:

```bash
./smoke-check.sh https://app.example.com /admin
```

### 5. Rollback

When a deployment fails, choose the appropriate rollback procedure:

1. **Schema rollback**:
   **Forge does not have a `forge migrate down` command.** The real command to roll back
   schema migrations is `forge migrate rollback`:
   ```bash
   forge migrate rollback
   ```
   This executes the `.down.sql` file of the last applied migration
   (`cli/commands/migrations/rollback.go`). To step back multiple migrations, execute
   `forge migrate rollback` once for each version.
2. **Recovering from dirty migration state**:
   If a migration fails midway, `forge migrate status` reports `DIRTY`.
   - If the migration aborted without committing partial changes (PostgreSQL rolls back
     failed DDL transactions), fix the migration SQL and reset the version pointer:
     ```bash
     forge migrate force <version>
     ```
     (`cli/commands/migrations/force.go`), then re-run `forge migrate up`.
   - If the schema changes were completed manually, mark the version clean:
     ```bash
     forge migrate recover --clean --version <version>
     ```
     (`cli/commands/migrations/recover.go`).
3. **Binary rollback**:
   Send `SIGTERM` to the failing binary to initiate graceful shutdown
   (`server.graceful_timeout`), then start the previous release binary.
4. **Data restore**:
   If data was corrupted or destructive changes cannot be undone with down migrations,
   restore from the pre-deploy backup:
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
