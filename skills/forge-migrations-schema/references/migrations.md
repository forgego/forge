# Migrations Guide

Forge's migration engine (`db/migrate/` and `db/migrations.go`) compares your Go models against schema state reconstructed from `migrations/` and runs migrations via `golang-migrate`.

## Generating Migrations

```bash
forge makemigrations <name> --auto --models ./app/blog
```

1. Scans `--models` (defaults to `./models` if `--models` is omitted; pass `./app/<name>` for `forge new` projects).
2. Reconstructs prior schema state from `migrations/*_*.up.sql` (`db/migrate/state/loader.go`).
3. Emits driver-specific SQL (`db/migrate/sql/postgres.go` or `sqlite.go`) into `migrations/<version>_<name>.up.sql` and `.down.sql`.

Note: `--verbose` on `makemigrations` is reserved for future output; use `forge migrate show --verbose` or `forge migrate lint --verbose` to inspect state-loader diagnostics.

## Applying, Inspecting & Recovering

```bash
forge migrate up
forge migrate status
forge migrate rollback
forge migrate lint --verbose
forge migrate recover --verify
forge migrate recover --clean
```

- **Dirty state**: If a migration fails mid-apply, `forge migrate status` marks it `DIRTY` and excludes it from applied versions. Inspect the database, then use `forge migrate force <ver>` or `forge migrate recover --clean`.
- **Fake migrations**: Use `forge migrate fake --fake-initial` when tables already exist, or `forge migrate fake <ver>`.
- **SQLite**: Applying generated migrations on SQLite is experimental; use PostgreSQL 15+ for production and CI.

## Programmatic Execution (`db.NewMigrationRunner`)

```go
runner, err := db.NewMigrationRunner(database, "./migrations")
if err != nil {
	return err
}
defer runner.Close()

if err := runner.Migrate(); err != nil {
	return err
}
```

Note: PostgreSQL migrations require at least 2 open connections in the DB pool (`db/migrations.go:196`).
