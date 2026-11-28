# Multi-Instance Deployment & Shared Stores (`stores/`)

Set `server.stores: database` in `config/config.yaml` (the default for `forge new`) so sessions, API rate-limit counters, admin bearer tokens, admin login lockout attempts, saved views, and change history (`forge_admin_log`) are stored in PostgreSQL instead of process memory.

## Required Wiring

1. Apply framework store migrations (`forge migrate up` runs embedded `forge_framework_migrations` automatically when `server.stores: database`).
2. Pass `server.WithDatabase(database)` to `server.NewServer(...)`.
3. Call `adminSite.UseDatabaseStores(ctx)` before `adminSite.Handler()`.
4. Configure `server.trusted_proxies` if running behind a reverse proxy so client IP resolution for rate limiting and admin login lockout honors `X-Forwarded-For` / `X-Real-IP` only from trusted peers.
