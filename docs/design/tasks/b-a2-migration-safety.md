TASK A2: migration file generation and recovery safety (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/db-transaction-driver-integrity/forge
(git worktree on branch fix/db-transaction-driver-integrity; A1 already landed here.)

Files you may modify — ONLY:
  db/migrate/execute/recover.go          (splitSQL)
  db/migrate/generate/generator.go       (file writes, next version)
  db/migrate/generate/squash.go          (file writes, getNextVersion, parseVersion)
  NEW db/migrate/generate/atomic_write.go
  tests: NEW db/migrate/execute/split_sql_test.go, NEW db/migrate/generate/write_safety_test.go
Do NOT edit go.mod / go.sum. Do NOT touch db/db.go, db/transaction.go, db/migrations.go.

Write failing tests FIRST.

=== BUG 1 (High, data corruption): recovery splits SQL on every semicolon ===
db/migrate/execute/recover.go splitSQL:
    statements := strings.Split(sql, ";")
Legitimate semicolons inside string literals ('a;b'), quoted identifiers ("x;y"), comments
(-- ...; and /* ...; */) and PostgreSQL dollar-quoted bodies ($$ ... ; ... $$, $tag$ ... $tag$)
are split, so recovery executes broken statement fragments.

Fix: rewrite splitSQL as a small single-pass scanner (no regex) that tracks state:
  normal | single-quoted '...' (with '' as an escaped quote) | double-quoted "..." ("" escape) |
  line comment -- to newline | block comment /* */ | dollar-quoted $tag$ ... $tag$ (tag may be
  empty: $$; tag chars are letters, digits, underscore; must not start with a digit)
Split only on ';' in normal state. Trim each statement; drop empty ones. Keep comments that are
inside a statement as part of that statement. Keep the function name and signature.

Tests (table-driven):
  "CREATE TABLE a(x int); CREATE TABLE b(y int);"          -> 2 statements
  "INSERT INTO t VALUES ('a;b'); SELECT 1"                  -> 2, first contains 'a;b'
  "INSERT INTO t VALUES ('it''s;ok');"                      -> 1
  `SELECT "we;ird" FROM t;`                                 -> 1
  "SELECT 1; -- trailing; comment\nSELECT 2;"               -> 2
  "/* a; b */ SELECT 1;"                                    -> 1
  "CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql; SELECT 2;" -> 2
  "DO $body$ BEGIN PERFORM 1; END $body$; SELECT 3"         -> 2
  "  ;  ; "                                                 -> 0
  "SELECT $1; SELECT 2"  (positional param, not dollar quote) -> 2

=== BUG 2 (Medium): migration pairs written non-atomically and can overwrite ===
generator.go and squash.go write `<version>_<name>.up.sql` then `.down.sql` with os.WriteFile:
  - a crash/error after the up file leaves a half pair on disk;
  - two concurrent generators pick the same next version and silently OVERWRITE each other.

Fix: create db/migrate/generate/atomic_write.go:

    // writeMigrationPair creates both files exclusively. If either file already exists it
    // returns an error wrapping os.ErrExist and writes nothing. If writing the second file
    // fails, the first is removed.
    func writeMigrationPair(upPath string, upContent []byte, downPath string, downContent []byte) error

  Implementation: for each file, write to a temp file in the SAME directory
  (os.CreateTemp(dir, ".migration-*.tmp")), fsync, close, then publish it with os.Link(tmp, final)
  — os.Link fails with an "exists" error if final exists — then remove the temp file. Before
  publishing the up file, check the down path does not exist either. On any failure after the up
  file was linked, remove the up file. Always remove temp files.
  File mode 0644.

Use writeMigrationPair in generator.go and squash.go in place of the two os.WriteFile calls,
keeping their existing error wrapping (core.NewMigrationError in generator.go, fmt.Errorf in
squash.go) around the returned error.

Tests:
  - writes both files with exact contents;
  - if the up file already exists: returns error with errors.Is(err, os.ErrExist), existing file
    content unchanged, down file NOT created;
  - if the down file already exists: error, up file NOT left behind;
  - no ".migration-*.tmp" files remain in the dir after success or failure.

=== BUG 3 (Low/Medium): next-version increment can overflow and parse silently ===
squash.go getNextVersion: `nextVersion := maxVersion + 1` with maxVersion uint64 from a hand-rolled
parseVersion (read it) — a version of 18446744073709551615 wraps to 0, and non-numeric prefixes
parse as garbage instead of being skipped. generator.go similar at `nextSeq = maxVersion + 1`.
Fix in both:
  - parse versions with strconv.ParseUint(s, 10, 64); skip files whose prefix does not parse.
  - if maxVersion == math.MaxUint64 return an error "migration version overflow".
  - squash.go: replace parseVersion's use with strconv.ParseUint (delete parseVersion only if
    nothing else uses it).
Tests: dir containing "18446744073709551615_x.up.sql" -> error mentioning overflow;
dir containing "abc_x.up.sql" and "000007_y.up.sql" -> next version is 8 (formatted as the existing
code formats it).

=== VERIFY ===
  gofmt -l on the files you changed/created   (must print nothing)
  go vet ./db/...
  go test -race ./db/migrate/... ./db/
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes. Minimal diff.
- Final report: files changed, tests added, exact tail of the -race run.
