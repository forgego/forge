---
name: forge-server
description: Configure Forge's HTTP server, middleware, security headers, CORS, rate limiting, Viper config (`config/`), `FORGE_*` env vars, logging (`log/`), and database-backed framework stores (`server.WithDatabase`). Use when wiring `server.NewServer`, graceful shutdown, health checks, or multi-instance deployment. Do NOT use for REST API ViewSet logic (use `forge-api`) or ORM queries (use `forge-orm`).
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15
metadata:
  author: forgego
  version: "0.2"
---

# Forge Server, Config & Deployment

`server.NewServer` wraps `chi/v5` and `http.Server` with request IDs, structured logging (`log/`), `scs` sessions, CSRF protection, health endpoints, and graceful shutdown.

## Golden Path

```go
cfg := config.NewConfig()
settings := config.LoadSettings(cfg)

logger, err := forgelog.NewLogger(settings.App.Debug)
if err != nil {
	log.Fatal(err)
}
defer logger.Sync()

srv, err := server.NewServer(cfg, settings, logger, server.WithDatabase(database))
if err != nil {
	log.Fatal(err)
}

srv.RegisterRoutes(func(r *server.Router) {
	// CORS, SecureHeaders, and RateLimit are opt-in:
	// r.Use(server.SecureHeaders(server.DefaultSecureHeadersOptions()))
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
})

if err := srv.StartWithGracefulShutdown(); err != nil {
	log.Fatal(err)
}
```

## Hard Guardrails

- **CORS, SecureHeaders, and Rate Limiting are opt-in**: `server.NewServer` does NOT install them automatically (`server/docs_reference_contract_test.go`); add them explicitly via `Router.Use(...)`.
- **Sessions & CSRF are secret-gated**: `SessionManager` is installed only when `security.session_secret` is non-empty; CSRF middleware is installed only when `security.csrf_secret_key` is non-empty.
- **Production secret validation is fail-closed**: When `app.env == "production"`, `secret_key`, `csrf_secret_key`, and `session_secret` must all be set, >= 32 characters, and distinct.
- **Framework stores always live in the database**: Pass `server.WithDatabase(database)` to `server.NewServer` and call `adminSite.UseDatabaseStores(ctx)` (and apply framework store migrations via `forge migrate up`). There is no `server.stores` setting or memory store mode.
- **`log` remote output returns `NotImplemented`**: Only `console` and `file` outputs are supported; YAML config does not configure log sampling (use `log.NewBuilder().Sampling(...)` in Go).

## References

- [Server middleware, config & FORGE_* env vars](references/server-config.md)
- [Multi-instance deployment & shared stores](references/deployment.md)
