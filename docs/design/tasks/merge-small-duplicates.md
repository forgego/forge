Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Merge small duplicates: migration checksum, integration test helpers, settings loader durations, "user not found" errors

Module roots: /home/hamid/Other/projects/foreit-wt/wave1/forge (module github.com/forgego/forge) and /home/hamid/Other/projects/foreit-wt/wave1/tests (check tests/go.mod to see whether it is its own module).

## Principles (mandatory)
- Behaviour must not change except where a bug is named. `go build` after each step.
- Smallest change; gofmt; naming after behaviour; sentinel errors compared with `errors.Is`.
- Delete a duplicate only after every caller uses the kept copy.

## 1. Migration checksum is duplicated
forge/db/migrate/execute/checksum.go and forge/db/migrate/verify/checksum.go are identical except the package clause (91 lines). Callers of `verify.NewChecksumValidator`: forge/db/migrations.go:275, forge/cli/commands/migrations/fake.go:140,188. The execute copy is used inside package execute (e.g. recover.go `CompareChecksums`).
Keep verify. Make package execute use `verify` (check `go list -deps` for an import cycle first: if verify imports execute, move the file to forge/db/migrate/core instead, make both use core, and delete both copies). Delete forge/db/migrate/execute/checksum.go and move its tests if any.

## 2. Integration test helpers are duplicated
tests/helpers and tests/testhelpers both contain postgres_features.go, sql_assertions.go, doc.go (2 differing lines each) and cli_helper.go (59 differing lines). tests/testhelpers has 16 importers, tests/helpers 8 (tests/integration/schema/schema_integration_test.go, tests/integration/db/ecommerce_orm_integration_test.go, and six files in tests/integration/migrate).
Keep tests/testhelpers. For each file only in tests/helpers (assertions.go, migration_assertions.go, db.go, fixtures.go, ... list them), move it into tests/testhelpers (fix the package clause and any name clash: if both define the same function, keep one, and if their bodies differ, keep the testhelpers one unless helpers' callers need the other behaviour; report every clash). For the four shared files, keep the testhelpers version unless a helpers caller needs a function that only the helpers version has; then add that function. Update the 8 importers, then delete tests/helpers. Run `go vet ./...` in the tests module (integration tests may need Docker/Postgres to run; compiling with `go vet` and `go test -run XXX_NONE ./...` is enough).

## 3. Settings loader reads durations as integers (B36, verified)
forge/config/settings.go:109-110 reads `database.conn_max_lifetime` / `conn_max_idle_time` with `GetInt(..., 300/600)` while forge/config/config.go:49-50 sets defaults "5m" / "2m" and config.go:171-172 reads them with `GetDuration`. So `LoadSettings` reports 0 for both. Make settings.go use `GetDuration` with the same defaults as config.go (5m, 2m), changing the field types in the settings struct to `time.Duration` if they are ints; fix every user of those fields (grep). forge/config/settings_test.go:124-128 asserts the buggy 0: change it to assert 5m and 2m.

## 4. "user not found" is defined four times
forge/identity/repository/user.go:16 `ErrUserNotFound` is the sentinel. forge/identity/service/user.go:15 defines its own `ErrUserNotFound = fmt.Errorf("user not found")`; forge/identity/backends/registry.go:102 and backends/token.go:70 return new `fmt.Errorf("user not found")` values.
- service: `var ErrUserNotFound = repository.ErrUserNotFound` (keeps the exported name, now `errors.Is`-compatible).
- backends: return `repository.ErrUserNotFound` (wrapped with `%w` if context is added).
- grep for string comparisons on "user not found" and replace them with `errors.Is`.
Add one test in identity/backends: registry `GetUser` with no matching backend returns an error where `errors.Is(err, repository.ErrUserNotFound)`.

## Acceptance
From forge:
    gofmt -l db config identity                 # empty
    go build ./... && go vet ./db/... ./config/... ./identity/... ./cli/...
    staticcheck ./db/... ./config/... ./identity/...   # no output
    go test -race ./db/... ./config/... ./identity/... ./cli/... -count=1
From tests (if it is a module) or the repo root:
    go vet ./... && go test -count=1 -run XXX_NONE ./...

Report (max 30 lines): per item what moved/deleted, name clashes resolved, each command result.
