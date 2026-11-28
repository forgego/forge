---
title: Server Settings
description: HTTP server, timeouts, health checks, trusted proxies, and framework store behavior.
image: /social-card.png
---

# Server Settings

`ServerSettings` (`config/settings.go`) configures `server.NewServer`.

## Keys & Defaults

| Key | Env Override | Default | Description |
|---|---|---|---|
| `server.host` | `FORGE_SERVER_HOST` | `"localhost"` | Bind host (`0.0.0.0` in containers) |
| `server.port` | `FORGE_SERVER_PORT` | `"8000"` | Listen port |
| `server.read_timeout` | `FORGE_SERVER_READ_TIMEOUT` | `30` | Read timeout in seconds |
| `server.write_timeout` | `FORGE_SERVER_WRITE_TIMEOUT` | `30` | Write timeout in seconds |
| `server.graceful_timeout` | `FORGE_SERVER_GRACEFUL_TIMEOUT` | `30` | Graceful shutdown wait in seconds |
| `server.max_request_size` | `FORGE_SERVER_MAX_REQUEST_SIZE` | `10485760` (10 MB) | Max request body size in bytes |
| `server.static_files_path` | `FORGE_SERVER_STATIC_FILES_PATH` | `""` | Local directory mounted at `/static` when non-empty |
| `server.health_check_path` | `FORGE_SERVER_HEALTH_CHECK_PATH` | `"/health"` | Base path for `/health`, `/health/ready`, `/health/live` |
| `server.metrics_enabled` | `FORGE_SERVER_METRICS_ENABLED` | `false` | Enable `/metrics` uptime endpoint |
| `server.metrics_path` | `FORGE_SERVER_METRICS_PATH` | `"/metrics"` | Path for metrics endpoint |
| `server.enable_profiling` | `FORGE_SERVER_ENABLE_PROFILING` | `false` | Mount `/debug` pprof routes (only when `app.debug: true`) |
| `server.info_endpoint` | `FORGE_SERVER_INFO_ENDPOINT` | `false` | Expose `/info` endpoint (off by default) |
| `server.trusted_proxies` | `FORGE_SERVER_TRUSTED_PROXIES` | `nil` | Proxy IPs/CIDRs trusted for forwarding headers |

Note: `server.static_files_url`, `server.uploads_path`, and `server.uploads_url` exist on `ServerSettings` in `config/settings.go`, but `server.NewServer` currently mounts only `server.static_files_path` at `/static`.

## Framework stores

Sessions, API throttling counters, admin API tokens, admin login lockout, saved views, and change history always live in the application database. Pass `server.WithDatabase(database)` to `server.NewServer` and run `forge migrate up`. There is no configuration setting or memory option. See [Deployment](/docs/deployment/#shared-state).

```yaml
server:
  host: 0.0.0.0
  port: "8000"
  graceful_timeout: 30
  trusted_proxies:
    - 10.0.0.0/8
```
