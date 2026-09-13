Your workspace is /home/hamid/Other/projects/foreit-wt/w0-stubs. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-4: unfinished features must fail loudly instead of silently succeeding

Module root: /home/hamid/Other/projects/foreit-wt/w0-stubs/forge (run go commands from there).
Only edit: forge/api/errors/idempotency_stores.go, forge/log/logger.go, forge/db/migrate/generate/squash.go, forge/cli/commands/migrations/squash.go, and create one test file per package named w0_not_implemented_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Use the existing `github.com/forgego/forge/errors` package: `errors.NewNotImplementedError("<feature>")` and `errors.IsNotImplemented(err)` (see forge/errors/errors.go). Alias the import as `forgeerrors` where the file already uses stdlib `errors`.
- Errors wrapped with %w when adding context; no new package-level mutable state.

## Stubs (verified on master)

1. **DB-backed idempotency store is a no-op.**
   - `forge/api/errors/idempotency_stores.go`: `DatabaseStore.Get`, `Set` and `Delete` (~187-231) build SQL strings and discard them (`_ = query`). `Set` returns nil and `Get` always returns "not found", so any endpoint using it re-executes side effects.
   - `NewDatabaseStore` (~148) currently succeeds. Nothing in the repo calls it outside its own file.
   - Fix: `NewDatabaseStore` returns `nil, forgeerrors.NewNotImplementedError("idempotency DatabaseStore")`. Also make `Get`/`Set`/`Delete`/`Cleanup` on a zero-value `DatabaseStore` return the same error instead of fake success. Do not delete the type.
   - Test: `NewDatabaseStore(nil, "")` returns an error with `IsNotImplemented` true; `(&DatabaseStore{}).Set(...)` returns a NotImplemented error.

2. **Remote log output silently drops everything.**
   - `forge/log/logger.go`: `createRemoteCore` (~126) builds a `remoteExporter` whose writer is a `noOpWriter` (~138-156).
   - It is selected when an output has `Type == OutputRemote` in the constructor's switch (~58-67). That constructor returns `(*Logger, error)`.
   - Fix: in that switch, `case OutputRemote:` returns `nil, forgeerrors.NewNotImplementedError("log remote output")`. Delete `createRemoteCore`, `remoteExporter`, `newRemoteExporter` and `noOpWriter` if nothing else references them (grep first); keep `RemoteOutputConfig` (it is part of the config schema).
   - Test: building a logger with one output of Type OutputRemote returns an IsNotImplemented error. Find the constructor name and config types at the top of logger.go / config.go.

3. **Migration squash writes a duplicate migration.**
   - `forge/db/migrate/generate/squash.go`: `Squasher.SquashMigrations` combines the SQL of a version range into a new, higher-version migration but leaves the originals in place ("Old migrations should be archived" comment). `migrate up` would then apply every statement twice.
   - The CLI `forge/cli/commands/migrations/squash.go` `Execute` (~33-55) prints "Successfully squashed" and a false claim about a 'replaces' field.
   - Fix: `SquashMigrations` returns `forgeerrors.NewNotImplementedError("migrations squash")` as its FIRST statement, before reading or writing any file. Keep `getNextVersion` and the helpers (tests in write_safety_test.go use `getNextVersion`). In the CLI, return the error unchanged and remove both success print lines.
   - Test: `NewSquasher(t.TempDir()).SquashMigrations("000001", "000002", "x")` returns IsNotImplemented, and the temp dir contains no new files afterwards.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-stubs/forge)
    gofmt -l api log db cli                     # empty
    go vet ./api/... ./log/... ./db/... ./cli/...
    go test ./api/errors ./log ./db/migrate/generate ./cli/commands/migrations -count=1
    go build ./...

Report (max 30 lines): failing output before the fix, files changed, deleted identifiers, each acceptance command with its result.
