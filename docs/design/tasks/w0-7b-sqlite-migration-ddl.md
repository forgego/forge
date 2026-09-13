Your workspace is /home/hamid/Other/projects/foreit-wt/w0-sqlite-ddl. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-7b: SQLite migrations emit SQLite DDL or fail loudly, never PostgreSQL DDL or comments

Module root: /home/hamid/Other/projects/foreit-wt/w0-sqlite-ddl/forge (run go commands from there).
Only edit: forge/db/migrate/sql/sqlite.go, and create forge/db/migrate/sql/sqlite_ddl_test.go.
Do not change base.go or postgres.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming: descriptive names after the behaviour, never after tickets (no "W0" anywhere).
- Tests must EXECUTE the generated SQL on a real in-memory SQLite database (`sql.Open("sqlite3", ":memory:")` with `_ "github.com/mattn/go-sqlite3"`, already a dependency, v1.14.52 bundles SQLite >= 3.35), not just compare strings.

## Bugs (verified on master, forge/db/migrate/sql/sqlite.go)

1. **Foreign keys use PostgreSQL DDL.**
   - `SQLiteBuilder.BuildAddForeignKey` (~278) delegates to `baseBuilder.BuildAddForeignKey` (base.go ~184-220), which emits a PostgreSQL `DO $$ ... pg_constraint ... ::regclass` block. SQLite cannot run it.
   - The detector (forge/db/migrate/generate/detector.go ~59-82) emits `CreateTable` followed by one `AddForeignKey` per FK for every NEW table, so any SQLite migration for a model with a foreign key is broken.
2. **Drop column is a comment.** `buildChangeUpSQL` `case *core.DropColumn` (~84-86) returns a SQL comment, so the forward migration silently leaves the column in place.
3. **Rollback of add column is a comment.** `buildChangeDownSQL` `case *core.AddColumn` (~137-138) returns a comment, so the rollback "succeeds" while the column stays.
4. **Drop foreign key and drop constraint are comments.** `case *core.DropForeignKey` and `case *core.DropConstraint`, in both up (~99-100, ~105-106) and down (~174 and the matching DropConstraint case), return comments.

## Required behaviour
- `DropColumn` up: `ALTER TABLE <table> DROP COLUMN <column>;`
- `AddColumn` down: `ALTER TABLE <table> DROP COLUMN <column>;`
- Foreign keys in `BuildUpSQL`:
  - Before building statements, collect the `AddForeignKey` changes whose `Table` is created by a `CreateTable` in the same change list.
  - Render those tables with the foreign keys as table constraints inside CREATE TABLE: `FOREIGN KEY (<column>) REFERENCES <target_table> (id) ON DELETE <x> ON UPDATE <y>`. Mirror exactly how base.go BuildAddForeignKey derives the column name, target table and ON DELETE/ON UPDATE (reuse `mapCascadeType`).
  - Skip those `AddForeignKey` changes afterwards.
  - Add a small SQLite-specific create-table helper that reuses `BuildColumnDefinition`; do not modify base.go.
- `AddForeignKey` for a table NOT created in the same change list (up), and `DropForeignKey` / `DropConstraint` (up and down): return an error, a `core.NewMigrationError(core.ErrInvalidChange, "<what> is not supported on SQLite without rebuilding the table", nil)`, instead of SQL.
- `BuildDownSQL`: an `AddForeignKey` whose table is also created in the same change list produces no statement, because dropping the table removes it. Any other `AddForeignKey` down returns the same error.
- Keep every other case unchanged.

## Tests (sqlite_ddl_test.go), each executes the SQL on :memory:
- A change list with CreateTable categories, CreateTable products (with a category_id column), and AddForeignKey products.category_id → categories.
  - The up SQL contains no "DO $$", runs without error, and `PRAGMA foreign_key_list(products)` returns one row referencing categories.
  - Build the ModelDefinition / FieldDefinition / Relation values the way the existing tests in forge/db/migrate/sql and forge/db/migrate/generate do; read them first.
- DropColumn on an existing table: the SQL runs, and `PRAGMA table_info` no longer lists the column.
- AddColumn down: running the up SQL and then the down SQL leaves the column absent.
- AddForeignKey on a table not created in the change list, DropForeignKey and DropConstraint: `BuildUpSQL` returns an error (and `BuildDownSQL` too, where applicable).

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-sqlite-ddl/forge)
    gofmt -l db                                  # empty
    go vet ./db/...
    staticcheck ./db/migrate/...                 # no output
    go test ./db/migrate/sql -count=1 -v -run 'SQLite'
    go test -race ./db/... -count=1

Report (max 30 lines): failing output before the fix, files changed, each acceptance command with its result, and any existing test you had to change and why.
