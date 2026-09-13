TASK A1: database integrity — three critical bugs (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/db-transaction-driver-integrity/forge
(git worktree on branch fix/db-transaction-driver-integrity. Work ONLY there.)

Files you may modify — ONLY:
  db/transaction.go, db/db.go, db/migrations.go
  NEW file db/dsn.go
  tests: db/transaction_test.go, db/db_test.go (create if missing), NEW db/dsn_test.go,
         db/migrations_test.go (create if missing)
Do not touch anything outside forge/db. Do not edit go.mod / go.sum.

Write each failing test FIRST, run it, watch it fail, then fix.

=== BUG 1 (Critical): WithTx swallows commit failures ===
db/transaction.go:34

    func (db *DB) WithTx(ctx context.Context, fn func(*Tx) error) error {
        tx, err := db.BeginTx(ctx, nil)
        ...
        defer func() {
            if p := recover(); p != nil { _ = tx.Rollback(); panic(p) }
            else if err != nil { _ = tx.Rollback() }
            else { err = tx.Commit() }     // <-- assigns to a LOCAL; never returned
        }()
        err = fn(tx)
        return err
    }

The return value is unnamed, so the Commit error set inside the defer is discarded.
A failed COMMIT reports success to the caller -> silent data loss.

Fix: use a named return `(err error)` so the deferred Commit error is returned.
Keep rollback-on-error and rollback-then-repanic behaviour exactly.

Regression test (sqlite in-memory, the package already uses sqlite3 in tests — read
db/transaction_test.go first and reuse its setup helpers):
  - fn commits the tx itself (call the underlying commit inside fn, e.g. tx.Commit()),
    so the deferred Commit fails with sql.ErrTxDone. Assert WithTx returns a NON-nil
    error and errors.Is(err, sql.ErrTxDone).
  - fn returns an error -> WithTx returns that same error, and inserted rows are rolled back.
  - fn panics -> panic propagates (use defer/recover in the test), rows rolled back.
  - happy path -> nil error, rows persisted.

=== BUG 2 (Critical): a PostgreSQL failure silently falls back to SQLite ===
db/db.go NewDB (line ~166): it tries postgres; if Ping fails it opens the SAME dsn with
sqlite3. go-sqlite3 happily creates a local file literally named
"host=db port=5432 user=... password=..." and the app runs against an empty SQLite DB
(and writes the password into a filename). A production outage becomes silent data
divergence.

Fix:
1. Create db/dsn.go with:

     // DetectDriverFromDSN returns "postgres" or "sqlite3".
     func DetectDriverFromDSN(dsn string) string

   Rules, applied IN THIS ORDER:
     a. trimmed dsn starts with "postgres://" or "postgresql://"          -> "postgres"
     b. starts with "file:" OR equals ":memory:" OR starts with ":memory:" -> "sqlite3"
     c. strip any "?query" part; if the remaining path ends with
        ".db", ".sqlite" or ".sqlite3" (case-insensitive)                   -> "sqlite3"
     d. split on whitespace; if ANY whole field has prefix "host=", "dbname=",
        "user=", "sslmode=" or "port="                                      -> "postgres"
     e. otherwise                                                           -> "postgres"
   (Rule order matters: "/tmp/host=data.sqlite" must be sqlite3 via rule c, and
    "file:app.sqlite?sslmode=disable" must be sqlite3 via rule b.)

2. Add to db/db.go:

     // NewDBWithDriver opens exactly one driver and never falls back.
     func NewDBWithDriver(driver, dsn string, opts ...Option) (*DB, error)

   driver "postgres"/"postgresql" -> postgres + PostgreSQL dialect;
   "sqlite"/"sqlite3" -> sqlite3 + SQLite dialect; anything else -> error.
   Open, Ping; on Ping failure Close and return a wrapped error. Apply
   DefaultPoolConfig then opts, exactly as NewDB does today. Keep whatever extra
   sqlite setup NewDB/the DB type already does (read db.go around line 219).

3. Rewrite NewDB as:  return NewDBWithDriver(DetectDriverFromDSN(dsn), dsn, opts...)
   Signature unchanged. NO fallback.

4. NewDBFromConfig already knows the driver: make it call NewDBWithDriver(driver, dsn, ...).

Before finishing, check the callers compile and still make sense:
  ../examples/ecommerce/main.go:93, ../examples/ecommerce/scripts/seed.go:38-41
(read only — if seed.go relies on the fallback by passing a sqlite path that has no
.db/.sqlite extension, report it; do NOT edit files outside forge/db).

Tests (db/dsn_test.go, table-driven) — every row required:
  "postgres://u:p@h:5432/d?sslmode=disable"          postgres
  "postgresql://h/d"                                  postgres
  "host=localhost port=5432 user=u dbname=d sslmode=disable"  postgres
  "dbname=app.sqlite host=localhost sslmode=disable"  postgres  (rule c must not win: after
       stripping query the string does not END in .sqlite — confirm your impl handles it)
  "file:app.sqlite?sslmode=disable"                   sqlite3
  ":memory:"                                          sqlite3
  "file::memory:?cache=shared"                        sqlite3
  "/tmp/host=data.sqlite"                             sqlite3
  "./data/app.db"                                     sqlite3
  "app.SQLITE3"                                       sqlite3
  "app.db?_foreign_keys=on"                           sqlite3
And in db_test.go:
  - NewDBWithDriver("sqlite3", ":memory:") works; Driver == "sqlite3".
  - NewDBWithDriver("mysql", "x") returns an error.
  - NewDBWithDriver("postgres", "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable
    connect_timeout=1") returns an error, and NO file named like the DSN is created in
    the current dir (check with os.Stat after, using t.Chdir(t.TempDir()) first).

=== BUG 3 (High): migration runner picks the driver from global config ===
db/migrations.go NewMigrationRunner (line ~45):

    cfg := config.NewConfig()
    driverName := cfg.GetDriver()

It ignores the *DB it was given. A sqlite *DB with default config gets the postgres
migrate driver (or vice versa) -> migrations fail or corrupt.

Fix: use the passed connection: `driverName := db.Driver`. If db.Driver is empty,
fall back to DetectDriverFromDSN is NOT possible (no dsn) — return an error
"database driver is unknown". Then grep db/migrations.go for any OTHER
`config.NewConfig()` + `GetDriver()` used to pick a migrate driver for a passed *DB and
fix those the same way. Remove the now-unused config import only if unused.

Test: build a sqlite3 *DB via NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(),"m.db")),
with a temp migrations dir containing 000001_init.up.sql / .down.sql
(CREATE TABLE t(id INTEGER);). Ensure the process config would say postgres
(t.Setenv on whatever env var config.GetDriver reads — read config to find it; if
none, skip that part). NewMigrationRunner must succeed and applying Up must create
table t.

=== VERIFY (from the working directory) ===
  gofmt -l db                      (prints nothing)
  go vet ./db/...
  go test -race ./db/...
  go test ./...
  cd ../examples/ecommerce && go build ./... && cd -

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Do not change existing exported signatures (NewDB, WithTx, NewMigrationRunner, NewDBFromConfig).
- Minimal diffs, no unrelated refactors or reformatting.
- Final report: files changed, test names added, exact tail of `go test -race ./db/...`,
  and whether ecommerce seed.go depended on the removed fallback.
