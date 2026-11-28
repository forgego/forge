---
name: forge-migrations-schema
description: Generate, apply, inspect, lint, roll back, and recover SQL migrations in Forge (`forge makemigrations`, `forge migrate up/status/show/lint/rollback/fake/force/recover/baseline`). Use when evolving database schemas, debugging dirty migration state, or running `db.NewMigrationRunner`. Do NOT use for writing model structs/fields (use `forge-models`) or ORM queries (use `forge-orm`).
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15 (SQLite migration apply is experimental)
metadata:
  author: forgego
  version: "0.2"
---

# Forge Migrations & Schema

Forge generates SQL migrations by diffing your Go model ASTs against schema state reconstructed from existing `migrations/*_*.up.sql` files — no running database is required for `makemigrations`.

## Golden-Path Workflow

```bash
forge makemigrations initial --auto --models ./app/blog
forge migrate show --sql --models ./app/blog
forge migrate lint --verbose
forge migrate up
forge migrate status
```

## Commands Reference

| Command | Purpose |
|---|---|
| `forge makemigrations <name> --auto` | Diff models vs migration history and write `.up.sql` / `.down.sql` |
| `forge migrate up` | Apply all pending migrations and framework store tables |
| `forge migrate rollback` | Roll back the last applied migration |
| `forge migrate status` | Print applied, pending, and dirty (`DIRTY`) versions |
| `forge migrate show` | Preview planned changes (`--sql`, `--verbose`) |
| `forge migrate lint` | Lint migration files (`--verbose` prints state-loader parse errors) |
| `forge migrate fake [ver]` | Mark pending migrations (or `--fake-initial` existing tables) as applied |
| `forge migrate force <ver>` | Force schema version in DB |
| `forge migrate recover` | Recover dirty state and verify checksums (`--clean`, `--verify`) |
| `forge migrate baseline` | Record a baseline version |

## Hard Guardrails

- **Always pass `<name>` to `makemigrations`** and **`up` to `forge migrate up`** (`forge migrate` alone does not apply migrations; there is no `migrate down` or target version).
- **`forge migrate squash` returns `NotImplemented`**.
- **PostgreSQL 15 is Supported; SQLite migration apply is Experimental**: `TestMigrationApplyPostgres` and `TestPostgresSchemaLifecycle` gate PostgreSQL; SQLite migration apply is experimental (`docs-site/docs/status.md`).
- **Framework store tables**: `forge migrate up` applies embedded framework migrations (`forge_framework_migrations`) before application migrations.
- **Checksums**: Editing an already-applied `.up.sql` fails verification (`forge migrate recover --verify`).

## References

- [Detailed migration & recovery guide](references/migrations.md)
