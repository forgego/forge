---
sidebar_position: 22
title: Deployment
description: The minimum production deployment contract for one Forge instance on PostgreSQL.
image: /social-card.png
---

# Deployment

This guide covers the deployment Forge supports today: **one application
instance** built from a `forge new` project, talking to **one PostgreSQL 15
database**, behind a reverse proxy that terminates TLS. Anything beyond that is
listed under [limits](#multiple-instances). The [support contract](/docs/status/)
defines the tiers used below.

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
3. **The `migrations/` directory**, applied with the `forge` CLI.
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
| `FORGE_APP_DEBUG` | `false` (the default) | `true` selects the debug logger and is rejected in production. `forge new` sets it to `true` only in the local `.env`. |
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
- **Not redacted**: URL paths are logged verbatim, so keep secrets out of
  paths. Error logs include the internal error text, which for a database
  error can contain row values (for example the duplicate value in a
  unique-constraint violation). Panic values are logged as-is. Treat logs as
  containing personal data.
- The `main.go` from `forge new` builds its logger with
  `log.NewLogger(settings.App.Debug)` and ignores the `logging.*` keys; see
  [logging](/docs/config/logging/).

## State kept in process memory

A restart loses the following, and each instance has its own copy:

| State | Where it lives | Effect of a restart |
| --- | --- | --- |
| Server sessions (`forge_session` cookie) | In-memory `scs` store | All sessions end. |
| Admin bearer tokens (24 h lifetime) | In-memory token store | Every admin user is signed out. |
| Admin login attempt limiter | In memory, keyed by client IP and by username | Counters reset. Behind a proxy, list it in `server.trusted_proxies` (`FORGE_SERVER_TRUSTED_PROXIES`); otherwise every client shares the proxy's address and failed logins from one client can lock out everyone. |
| Admin saved views | In memory | Lost. |
| Admin change history | In memory, last 1000 entries per registered model | Lost. See below. |
| API throttling counters | Default store is in memory | Reset; limits are per instance. |
| `api/caching` cache, idempotency keys | In memory | Lost. |
| Secrets generated at startup (non-production only) | Process | New values each start, so existing cookies stop working. |

These survive a restart because they live in PostgreSQL: your models, the
`identity` package's users, sessions (`user_sessions`) and tokens, and the
migration bookkeeping. CSRF tokens are signed cookies and survive as long as
the CSRF secret does.

## Audit history

The admin's change history is **not durable**. It is kept in memory, capped at
1000 entries per registered model, and lost on restart; it is not a compliance
audit log. To keep it, implement `core.HistoryManager` (`LogAction`,
`GetHistory`) against your database and set it as `HistoryManager` in each
`admin.Config`. Forge does not ship a durable implementation.

## Multiple instances

Running more than one instance is **not supported**. Because of the state
listed above, a second instance behind a load balancer means admin users are
signed out when a request lands on the other instance, rate limits and login
lockouts are per instance, and change history is split. Sticky sessions hide
some of this but not all of it. Scale vertically, or replace the in-memory
stores with shared ones you implement, and test that yourself.

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
