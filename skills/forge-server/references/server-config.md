# Server, Middleware & Configuration

## Default Middleware Installed by `server.NewServer`

1. `RequestIDMiddleware`
2. `ErrorHandler` (when `logger != nil`)
3. `log.Middleware` (when `logger != nil`)
4. `RequestSizeLimit` (when `server.max_request_size > 0`)
5. `SessionManager` (when `security.session_secret != ""`, `Cookie.Secure = true` in production)
6. `CSRF` (when `security.csrf_secret_key != ""`, honoring `security.csrf_exempt_paths`)

Built-in endpoints:
- Health: `server.health_check_path` (`/health` by default) plus `/ready` and `/live` (register custom checks via `server.RegisterHealthCheck`).
- Metrics / Info / Profiling: opt-in via `server.metrics_enabled`, `server.info_endpoint`, and `server.enable_profiling` (when `app.debug: true`).

## Environment Overrides (`FORGE_*`)

Viper maps nested keys with underscores and `FORGE_` prefix:
- `FORGE_APP_ENV`, `FORGE_APP_DEBUG`
- `FORGE_SERVER_HOST`, `FORGE_SERVER_PORT`
- `FORGE_DATABASE_DRIVER`, `FORGE_DATABASE_HOST`, `FORGE_DATABASE_PORT`, `FORGE_DATABASE_USER`, `FORGE_DATABASE_PASSWORD`, `FORGE_DATABASE_NAME`
- `FORGE_SECURITY_SECRET_KEY`, `FORGE_SECURITY_CSRF_SECRET_KEY`, `FORGE_SECURITY_SESSION_SECRET`
