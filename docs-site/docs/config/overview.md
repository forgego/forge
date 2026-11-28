---
title: Config Overview
description: Configuration overview, YAML layout, and FORGE_* environment variable overrides.
image: /social-card.png
---

# Configuration Overview

Forge loads configuration via Viper (`config.NewConfig()`, `config.LoadSettings(cfg)`) from `config/config.yaml` (searching `.`, `./config`, and `../config`) and overrides any key with `FORGE_*` environment variables.

## Precedence

1. **Environment variables** (`FORGE_<SECTION>_<KEY>`, uppercase with dots replaced by underscores)
2. **`config/config.yaml`** (or `.env` loaded for local development)
3. **Built-in defaults** in `config.LoadSettings`

## Configuration Sections

| Section | Guide | Covers |
|---|---|---|
| `app` | [App Settings](/docs/config/app/) | Name, environment, version, debug flag |
| `server` | [Server Settings](/docs/config/server/) | Host, port, timeouts, health/metrics, trusted proxies, `server.stores` |
| `database` | [Database Settings](/docs/config/database/) | Driver, connection details, SSL mode, pool limits |
| `security` | [Security Settings](/docs/config/security/) | Secret key, CSRF secret, session secret, CSRF exemptions |
| `logging` | [Logging Settings](/docs/config/logging/) | Level, format, console/file outputs |
| `errors` | [Error Settings](/docs/config/errors/) | Problem details, request ID, sanitization, idempotency |
| `admin` | [Admin Configuration](/docs/admin/config/) | Admin mount path, title, enabled flag |

## Example `config/config.yaml`

```yaml
app:
  name: myapp
  env: development
  debug: true

server:
  host: localhost
  port: "8000"
  stores: database

database:
  driver: postgres
  host: localhost
  port: 5432
  name: myapp_db
  user: postgres
  sslmode: disable
```

Never commit secrets to `config.yaml`; supply `FORGE_SECURITY_SECRET_KEY`, `FORGE_SECURITY_CSRF_SECRET_KEY`, `FORGE_SECURITY_SESSION_SECRET`, and `FORGE_DATABASE_PASSWORD` via environment variables.
