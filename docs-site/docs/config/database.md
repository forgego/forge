---
title: Database Settings
description: Database driver, connection details, SSL mode, and connection pool settings.
image: /social-card.png
---

# Database Settings

`DatabaseSettings` (`config/settings.go`) configures `db.NewDBFromConfig(cfg)`.

## Keys & Defaults

| Key | Env Override | Default | Description |
|---|---|---|---|
| `database.driver` | `FORGE_DATABASE_DRIVER` | `"postgres"` | `"postgres"` (supported) or `"sqlite"` / `"sqlite3"` (partially tested) |
| `database.host` | `FORGE_DATABASE_HOST` | `"localhost"` | PostgreSQL host |
| `database.port` | `FORGE_DATABASE_PORT` | `5432` | PostgreSQL port |
| `database.user` | `FORGE_DATABASE_USER` | `"postgres"` | Database user |
| `database.password` | `FORGE_DATABASE_PASSWORD` | `""` | Database password (prefer env var or `.pgpass`) |
| `database.name` | `FORGE_DATABASE_NAME` | `"forge"` | Database name (or file path for SQLite) |
| `database.sslmode` | `FORGE_DATABASE_SSLMODE` | `"disable"` | PostgreSQL SSL mode (`require` or `verify-full` over network) |
| `database.max_open_conns` | `FORGE_DATABASE_MAX_OPEN_CONNS` | `25` | Max open pool connections (`>= 2` for Postgres migrations) |
| `database.max_idle_conns` | `FORGE_DATABASE_MAX_IDLE_CONNS` | `5` | Max idle pool connections |
| `database.conn_max_lifetime` | `FORGE_DATABASE_CONN_MAX_LIFETIME` | `5m` | Max connection lifetime |
| `database.conn_max_idle_time` | `FORGE_DATABASE_CONN_MAX_IDLE_TIME` | `2m` | Max idle connection time |

Forge does not read a `DATABASE_URL` environment variable; set `FORGE_DATABASE_*` keys instead.

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  name: myapp_db
  user: myapp
  sslmode: require
  max_open_conns: 25
  max_idle_conns: 5
```
