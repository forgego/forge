Your workspace is /home/hamid/Other/projects/foreit-wt/w0-mig-status. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-7a: migration status reports real versions; generator stops creating golang-migrate's table

Module root: /home/hamid/Other/projects/foreit-wt/w0-mig-status/forge (run go commands from there).
Only edit: forge/db/migrate/execute/status.go, forge/db/migrate/execute/status_test.go, forge/db/migrate/generate/generator.go, and create forge/db/migrate/generate/generator_bookkeeping_test.go.

## Principles (mandatory)
- Reproduce first: write/adjust the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming: descriptive names after the behaviour. Never name files or tests after tickets (no "W0" anywhere).

## Bug 1 (verified on master): status marks versions that do not exist as applied
- `mergeAppliedVersions(currentVersion uint, dirty bool, explicit map[uint]bool)` (forge/db/migrate/execute/status.go ~258) adds every integer from 1 to `currentVersion` (minus one if dirty).
- With timestamp versions (for example 20240101120000) that loop runs ~2×10¹³ iterations and exhausts memory.
- With sequential versions that have gaps, it reports versions that have no file.
- Caller: `GetDetailedStatus` (~53-100). It already loads `allMigrations` (via `getAllMigrations`) before calling the merge.

Fix:
- Change the signature to `mergeAppliedVersions(currentVersion uint, dirty bool, explicit map[uint]bool, fileVersions []uint) map[uint]bool`.
- Result: every version in `explicit`, except the current version when dirty, plus every version in `fileVersions` that is `<= currentVersion` (strictly `<` when dirty).
- No loop over the numeric range.
- In `GetDetailedStatus`, build `fileVersions` from `allMigrations` using the existing `parseVersion`, skipping parse errors.

Tests (status_test.go):
- Update the existing four cases to pass file versions, keeping their intent.
- Add a case with current version 20240101120000 and file versions [20231201000000, 20240101120000, 20240201000000]. Expect exactly the first two, and the call must return immediately.
- Add a gap case: files [1, 3, 5], current 3 → {1, 3}.
- Add a dirty case: files [1, 2, 3], current 3, dirty → {1, 2}.

## Bug 2 (verified on master): the first generated migration creates and drops golang-migrate's table
- `MigrationGenerator.GenerateMigrations` (forge/db/migrate/generate/generator.go ~135-172) detects the first migration.
- It prepends `generateBookkeepingTable(g.driver)` (~348-367, `CREATE TABLE IF NOT EXISTS schema_migrations (id, name, checksum, applied_at)`) to the up SQL.
- It appends `DROP TABLE IF EXISTS schema_migrations;` to the down SQL.
- Migrations are applied with golang-migrate, which owns `schema_migrations(version, dirty)` and creates it itself; forge/db/migrate/execute/recover.go reads `version, dirty` from it.
- So the CREATE silently no-ops, and rolling back the first migration drops golang-migrate's own tracking table, breaking every later `migrate` call.

Fix: remove the first-migration detection block, the bookkeeping prepend and append, and the `generateBookkeepingTable` function. Do not change anything else in GenerateMigrations.

Test (generator_bookkeeping_test.go):
- Generate the first migration into an empty temp migrations dir from a minimal models dir.
- Look at forge/db/migrate/generate's existing tests and forge/tests/integration/migrate/testdata for how a models directory is set up, and reuse that.
- Assert the written .up.sql and .down.sql do not contain "schema_migrations".
- If a full GenerateMigrations run is impractical in a unit test, assert instead that `generateBookkeepingTable` no longer exists by testing the up/down SQL assembly through the smallest existing seam, and explain in the report.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-mig-status/forge)
    gofmt -l db                                   # empty
    go vet ./db/...
    staticcheck ./db/migrate/...                  # no output
    go test ./db/migrate/execute ./db/migrate/generate -count=1 -v -run 'Merge|Status|Bookkeeping|Generate'
    go test -race ./db/... -count=1

Report (max 30 lines): failing output before the fix, files changed, removed identifiers, each acceptance command with its result.
