---
title: CLI Reference
description: Complete command and flag reference for the forge CLI.
image: /social-card.png
---

# CLI Reference

Install the `forge` CLI:

```bash
go install github.com/forgego/forge/cmd/forge@latest
forge version
```

## Project & Scaffolding

| Command | Description | Key Flags |
|---|---|---|
| `forge new <name>` | Scaffold a new Forge project | `--database` (`postgres`/`sqlite`), `--template` (`simple`/`full`), `--docker` |
| `forge add app <name>` | Add `app/<name>/` (`models.go`, `admin.go`, `api.go`) | `--example` |
| `forge add model <name>` | Add a model to an app | `--app <name>` |
| `forge add api <name>` | Add an API ViewSet/serializer file | `--app <name>` |
| `forge add handler <name>` | Add an HTTP handler | `--app <name>` |
| `forge add service <name>` | Add a service interface and implementation | `--app <name>`, `--inject` |
| `forge auth` | Scaffold `app/auth` and wire auth routes in `cmd/server/main.go` | — |

## Code Generation

```bash
forge generate --models ./app/blog --output ./app/blog
forge generate --models ./app/blog --output ./app/blog --api --strict
```

| Flag | Default | Description |
|---|---|---|
| `--models` | `./app` (if exists) else `./models` | Directory containing model `.go` files |
| `--output` | `./models` | Output directory for `gen.go` and `api_gen.go` |
| `--api` | `false` | Generate REST serializers, ViewSets, and route registration (`api_gen.go`) |
| `--strict` | `false` | Fail on any schema expression the AST parser cannot evaluate |

See [Code Generation](/docs/codegen/) for full details and caveats.

## Migrations

| Command | Description |
|---|---|
| `forge makemigrations <name> --auto --models <dir>` | Diff models vs migration SQL history and write `.up.sql` / `.down.sql` |
| `forge migrate up` | Apply all pending migrations (and framework store tables when `server.stores: database`) |
| `forge migrate rollback` | Roll back the last applied migration |
| `forge migrate status` | Show applied, pending, and `DIRTY` versions |
| `forge migrate show [--sql] [--verbose]` | Preview planned migration changes from current models |
| `forge migrate lint [--verbose]` | Lint migration files and report state-loader warnings |
| `forge migrate fake [version]` | Mark pending migrations (or `--fake-initial` existing tables) as applied |
| `forge migrate force <version>` | Force migration version in the database |
| `forge migrate recover [--clean] [--verify]` | Inspect/clean dirty state and verify SHA-256 checksums |
| `forge migrate baseline [--adopt]` | Record a baseline migration version |
| `forge migrate squash` | Registered command; returns `NotImplemented` |

## Server, Admin & Development

| Command | Description |
|---|---|
| `forge runserver` | Run `cmd/server/main.go` |
| `forge createsuperuser` | Create an admin/superuser account |
| `forge check` | Validate project configuration |
| `forge test` | Run project Go tests |
| `forge shell` | Minimal command loop (`help`, `history`, `clear`, `models`, `dbinfo`) |
| `forge version` | Print CLI version |
