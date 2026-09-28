---
sidebar_position: 12
description: Automatic AST model change detection, migration workflows, checksum validation, and disaster recovery.
image: /social-card.png
---

# Migrations & Disaster Recovery

Forge generates SQL migrations from your models. It uses Go's Abstract Syntax Tree (`go/ast`) parser to read the declared models. It rebuilds the current schema from the migration files you already have, compares the two, and writes the difference as a numbered pair of SQL files.

---

## Core Commands

```bash
# Generate a migration from model changes (writes nothing if nothing changed)
forge makemigrations <name> --auto

# Preview, then apply pending migrations
forge migrate up --dry-run
forge migrate up

# Show the applied version and pending migrations
forge migrate status

# Roll back the last applied migration
forge migrate rollback

# Compare applied migration files with their recorded checksums, and report dirty state
forge migrate recover --verify

# Recover from a failed migration (see "Recovering from a failed migration")
forge migrate force <version>
forge migrate recover --clean --version <version>
```

---

## The Migration Workflow

1. **Modify your models:** update fields, indexes, constraints, or relations in `models/`.
2. **Generate a migration:** run `forge makemigrations add_status --auto`. Forge writes
   `migrations/000002_add_status.up.sql` and `.down.sql`, or prints
   `No changes detected` when the models match the existing migrations.
3. **Review the SQL:** migrations are plain SQL that you commit. Read every generated file
   before applying it (see [Safe schema changes](#safe-schema-changes)).
4. **Apply:** run `forge migrate up`. Forge records the applied version in
   `schema_migrations` (`version`, `dirty`) and the SHA-256 of each applied up and down file
   in `forge_migration_checksums`.

---

## Safe Schema Changes

Generated migrations are proposals, and an operator should review each one before it
reaches a database with data in it.

- **Renames are not detected.** Renaming a field or a table in a model produces a `DROP`
  of the old name and an `ADD` of the new one, which loses the data. Write the rename by
  hand instead (see [Renaming a column or table](#renaming-a-column-or-table)).
- **Destructive statements.** Treat `DROP TABLE`, `DROP COLUMN`, and type changes
  (`ALTER COLUMN ... TYPE`) as data loss until proven otherwise. `forge migrate lint` flags
  `DROP TABLE` and `TRUNCATE`, but not `DROP COLUMN`, so read the SQL yourself.
- **Constraints check existing rows.** Adding a foreign key, a `CHECK`, a `UNIQUE`
  constraint, or a `NOT NULL` column without a default fails if existing rows violate it.
  Clean up the data first, in its own migration.
- **Hand-written DDL becomes part of the recorded schema.** `makemigrations` reads every
  migration file, so a column or constraint that you add by hand but do not declare in a
  model shows up as a drop in the next generated migration.
- **Never edit an applied migration.** Write a new one. Editing an applied file changes its
  checksum, and `forge migrate up` then refuses to run (see below).

### Backups and restores

- Take a backup immediately before applying migrations to a production database, for
  example `pg_dump --format=custom --file=before-000007.dump "$DATABASE_URL"`.
- Down migrations are not backups. Rolling back a `DROP COLUMN` recreates the column
  empty, so the only way back to the data is a restore.
- Restore into a new database (`createdb app_restore && pg_restore --dbname=app_restore
  before-000007.dump`), then check it. `schema_migrations` and `forge_migration_checksums`
  are restored with the data, so `forge migrate status` shows the version the backup was
  taken at, and `forge migrate recover --verify` checks the files you have against it.
  Run `forge migrate up` to apply the migrations that came after the backup.

### Renaming a column or table

`makemigrations` reads a hand-written `ALTER TABLE .. RENAME COLUMN` and
`ALTER TABLE .. RENAME TO` back from the migration files. The indexes, constraints, and
foreign keys that refer to the renamed column or table follow it, as they do in the
database. Rename in one migration, and change the model to match before you generate
again:

1. Create an empty migration: `forge makemigrations rename_isbn --empty`.
2. Write the rename in the up file and its reverse in the down file:

   ```sql
   -- 000004_rename_isbn.up.sql
   ALTER TABLE books RENAME COLUMN isbn TO code;
   ALTER TABLE authors RENAME TO writers;
   ```

   ```sql
   -- 000004_rename_isbn.down.sql
   ALTER TABLE writers RENAME TO authors;
   ALTER TABLE books RENAME COLUMN code TO isbn;
   ```

3. Update the model: the field name, `Meta.TableName`, and the fields and conditions of any
   `Meta.Indexes` and `Meta.Constraints` that name the column. Index and constraint names
   do not change.
4. Run `forge makemigrations check --auto`. It prints `No changes detected`.

On PostgreSQL, a foreign key constraint keeps its name when its table or column is
renamed, but Forge names it `fk_<table>_<column>` from the current names, and drops it by
that name later. When the renamed table or column has its own foreign key, rename the
constraint in the same migration, for example
`ALTER TABLE books RENAME CONSTRAINT fk_books_author_id TO fk_books_writer_id;`. SQLite
(3.25 or later) runs the same `RENAME` statements; its foreign keys have no name.

For a rolling deploy, where old and new code run side by side, rename in steps instead:
add the new field, copy the data in a hand-written migration (for example
`UPDATE posts SET headline = title;`), deploy code that uses the new field, then remove
the old field.

### Changes makemigrations cannot generate

`makemigrations` stops with an error, and writes nothing, when a model change has no
generated SQL. It does not write an empty or comment-only migration that would be proposed
again on every run.

- **PostgreSQL column changes.** A column's type, `NOT NULL`, and default (`Default`,
  `DBDefault`, `AutoNow`, `AutoNowAdd`) are changed with `ALTER COLUMN`. A change to an
  existing column's `Unique()`, primary key, `AutoIncrement`, generated expression, or
  `DBColumn` name is not generated. For uniqueness on an existing table, declare a
  `UNIQUE` constraint in `Meta.Constraints` instead of adding `Unique()` to the field.
  Otherwise, revert the model change.
- **SQLite column, foreign key, and constraint changes.** SQLite's `ALTER TABLE` can add,
  drop, and rename a column, and rename a table, but it cannot change a column or add or
  drop a foreign key or constraint of an existing table. That takes a table rebuild, which
  `makemigrations` does not generate: it fails with
  `modify column <table>.<column> is not supported on SQLite without rebuilding the table`
  (or `add constraint`, `add foreign key`, and so on).

---

## Recovering from a Failed Migration

When a statement in a migration fails, `forge migrate up` stops with the database error and
marks that version dirty. Every further `up` or `rollback` refuses to run until you resolve
it.

1. **Inspect:** `forge migrate recover` prints the dirty version and both ways out.
2. **If the failed migration left no changes, re-apply it.** This is the usual case on
   PostgreSQL, where each migration file runs as one transaction, so a failure rolls back
   the statements that had already run. Confirm that the database does not contain the
   migration's changes, fix the file, and re-apply it:

   ```bash
   forge migrate force <previous version>   # records the last good version, clean
   forge migrate up                         # runs the fixed migration and records its checksum
   ```

   A file that contains its own `COMMIT` can leave partial changes. Undo them by hand
   before you force the previous version.
3. **If you completed the migration's changes by hand, mark it applied:**
   `forge migrate recover --clean --version <version>`. Then run
   `forge migrate baseline --adopt` to record its checksum.

Do not mark a rolled-back version clean. That records the migration as applied when none
of its changes are in the database.

### Checksum mismatches

`forge migrate recover --verify` compares each applied file with its recorded checksum and
lists a changed one as `<version>  mismatched`. It also exits non-zero. While any file is
mismatched or missing, `forge migrate up` refuses to run and names the version. Restore the
file from version control, and write a new migration for the change you intended. For
databases migrated before checksums were recorded, `forge migrate baseline --adopt` records
the current files. It never overwrites a mismatch.

---

## What Is Verified

The release gate runs `TestPostgresSchemaLifecycle` (in `tests/pkg_migrations`) against
PostgreSQL through the same commands used above. The test covers these steps:

- It generates and applies an initial schema, then seeds rows.
- It adds a column with a default, a related table with a foreign key, and `CHECK` and
  `UNIQUE` constraints. It checks the result in `information_schema` and confirms that the
  seeded rows are unchanged.
- It regenerates the unchanged models and confirms that nothing is written, and that
  `schema_migrations` and `forge_migration_checksums` match the applied files.
- It applies a migration that fails after its first statement ran, and confirms that the
  version is dirty and the partial change was rolled back. It then recovers with
  `force` + `up`.
- It edits an applied migration and confirms that the mismatch is reported and blocks
  `up`.

To reproduce, run the following against a PostgreSQL role that can create databases (the
test creates and drops a `lifecycle_<n>` database):

```bash
cd tests
FORGE_REQUIRE_DB=1 \
FORGE_TEST_DATABASE_URL='postgres://forge_test:forge_test@localhost:5432/forge_test?sslmode=disable' \
go test -count=1 -v -run 'TestPostgresSchemaLifecycle|TestChecksumBaseline' ./pkg_migrations/

# Generation stability and the DSL contract (no database needed)
cd ..
go test -count=1 -run 'Regeneration|InitialMigration|ChangedForeignKey' ./db/migrate/generate/
go test -count=1 -run 'DSL' ./codegen/
```

Without `FORGE_REQUIRE_DB=1` the PostgreSQL tests skip when no database is reachable. With
it set, they fail instead.

:::warning SQLite is experimental
Applying migrations to SQLite is experimental, and the release gate does not cover it.
`TestMigrationApplySQLite` is the one test the gate allows to skip. On SQLite, a new
table's foreign keys and `Meta.Constraints` are declared inside its `CREATE TABLE` and read
back from there, so regenerating unchanged models writes nothing. SQLite cannot add, drop
or change a foreign key or constraint of an existing table without rebuilding it, and
`makemigrations` does not write that rebuild: it stops with an error instead. The same
applies to a changed column, because SQLite has no `ALTER COLUMN`. See
[Changes makemigrations cannot generate](#changes-makemigrations-cannot-generate).
:::

---

## Best Practices for Production

- **Run the checks in CI/CD:** run `forge migrate status` and `forge migrate recover --verify`
  in your deployment pipeline before application pods start.
- **Review, back up, then apply:** generate and review migrations in development, commit
  them, back up production, and only then run `forge migrate up`.
- **One concern per migration:** keep data backfills in their own hand-written migrations,
  separate from schema changes. That keeps failures small and recovery clear.

---

## Next Steps

- **[Models & Schema DSL](/docs/models/)**: Define fields and constraints that power migrations.
- **[Quickstart Guide](/docs/quickstart/)**: Run your first migration in 60 seconds.
