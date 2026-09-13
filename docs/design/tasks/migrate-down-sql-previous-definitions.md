Your workspace is /home/hamid/Other/projects/foreit-wt/w0-down-sql. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Removing a model or a field must not abort migration generation

Module root: /home/hamid/Other/projects/foreit-wt/w0-down-sql/forge (run go commands from there).
Only edit: forge/db/migrate/core/change.go, forge/db/migrate/generate/detector.go, forge/db/migrate/sql/postgres.go, forge/db/migrate/sql/sqlite.go, and create forge/db/migrate/sql/drop_down_sql_test.go and forge/db/migrate/generate/detector_drop_definitions_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming: descriptive names after the behaviour, never after tickets (no "W0"/wave ids anywhere).

## Bug (verified on master)
- The detector emits `&core.DropTable{Table: name}` (forge/db/migrate/generate/detector.go ~97) and `&core.DropColumn{Table, ColumnName}` (~167). Neither carries the previous definition, although `previousMap` holds it (`*generator.ModelDefinition` for tables, `generator.FieldDefinition` for columns).
- `buildChangeDownSQL` in postgres.go (~127-144) and sqlite.go (~176-193) returns an error for `DropTable` and `DropColumn` ("cannot generate down SQL ... without ... definition").
- `MigrationGenerator` (generate/generator.go ~145) aborts on that error, so deleting any model or any field makes `makemigrations` fail completely, while both types claim `Reversible() == true`.

## Fix
1. core/change.go: add `Definition *generator.ModelDefinition` to `DropTable` and `Column *generator.FieldDefinition` to `DropColumn`, each with a one-line doc comment ("previous definition; nil when unknown, e.g. parsed from SQL"). Other constructors (for example forge/db/migrate/parse) keep compiling with nil.
2. detector.go: fill them from `previousMap` (take the address of a local copy for the column).
3. postgres.go and sqlite.go `buildChangeDownSQL`:
   - `DropTable` with non-nil Definition → `b.BuildCreateTable(&core.CreateTable{Table: c.Definition})`.
   - `DropColumn` with non-nil Column → `b.BuildAddColumn(&core.AddColumn{Table: c.Table, Column: *c.Column})`.
   - nil → keep the existing error unchanged.
   - Foreign keys of a re-created table are not restored; say so in one comment line. Do not add FK handling.

## Tests
- sql/drop_down_sql_test.go, table-driven over both builders (`NewPostgreSQLBuilder`, `NewSQLiteBuilder`):
  - DropColumn with a Column → down SQL contains an ADD COLUMN for that column.
  - DropTable with a Definition → down SQL contains CREATE TABLE with its columns.
  - nil definitions still return an error.
  - SQLite only: execute on `sql.Open("sqlite3", ":memory:")` (see sql/sqlite_ddl_test.go for the setup): create the table, run the up SQL of a DropColumn, then the down SQL, and assert `PRAGMA table_info` lists the column again. Same round trip for DropTable, checking `sqlite_master`.
  - Build ModelDefinition/FieldDefinition values the way sql/sqlite_ddl_test.go does.
- generate/detector_drop_definitions_test.go: previous defs with a table and a field that current defs lack → the DropTable has a non-nil Definition, and the DropColumn has a non-nil Column with the right name. Read the existing detector tests first for how to call the detector.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-down-sql/forge)
    gofmt -l db                                  # empty
    go vet ./db/...
    staticcheck ./db/migrate/...                 # no output
    go test ./db/migrate/sql ./db/migrate/generate -count=1 -run 'Drop'
    go test -race ./db/... -count=1

Report (max 25 lines): failing output before the fix, files changed, each acceptance command with its result.
