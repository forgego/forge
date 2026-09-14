Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Remove unreferenced filter/migrate code; make remaining stubs fail loudly

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Another agent is editing _test.go files in orm, db, identity, api, admin, log at the same time: do not edit those test files, and ignore test failures in them.

## Principles (mandatory)
- Before deleting anything, grep the whole workspace (forge, tests, examples, forge/cli/templates, docs-site) for each exported identifier you remove. If something outside the deleted files uses it, do not delete it; list it in the report.
- `go build ./...` must pass after every deletion step. Smallest change; gofmt. Naming after behaviour.
- Use `forgeerrors.NewNotImplementedError("<feature>")` from forge/errors for loud failures (see how forge/orm/queryset.go uses it).

## 1. Delete unreferenced files in forge/filter
Nothing outside each file references its types (verified with grep on master):
- forge/filter/dialect.go (JSON path interpolated into SQL, a MySQL adapter for a driver the project does not have)
- forge/filter/cache.go
- forge/filter/relations.go
- forge/filter/typed_filter.go
Delete them and tests that only cover them. Then:
- forge/filter/optimizer.go: `FilterSet` stores a `QueryOptimizer` (filterset.go:17, 38, 164, 326) but never calls `Optimize`. Remove the field, its three assignments and optimizer.go.
- forge/filter/persistence.go (`InMemoryFilterStorage`, map without a mutex): only forge/filter/shared.go refers to its interfaces. If nothing else in the workspace uses `FilterStorage`/`RBACFilterStorage`/`SavedFilter`, delete persistence.go and those interfaces from shared.go; otherwise keep them and add a `sync.RWMutex` to `InMemoryFilterStorage`.
- forge/filter/metrics.go: keep (other packages reference similar names; do not touch).

## 2. Delete the duplicate migration executor
forge/db/migrate/execute/executor.go `Executor`/`NewExecutor` is used only by forge/db/migrate/execute/integration_test.go. Callers use `db.MigrationRunner` (forge/db/migrations.go). Delete `Executor`, `NewExecutor`, helpers only they use, and the parts of integration_test.go that only exercise `Executor` (keep any test of other functions in that file).

## 3. Delete the identity password shim
forge/identity/password.go re-exports forge/identity/utils/password.go (`IsHashed`, `HashPassword`, `HashPasswordWithCost`, `CheckPassword`, `CheckPasswordHash`, `NeedsRehash`). If grep finds no user of `identity.<Name>` for these anywhere (including docs-site and cli templates), delete the file; otherwise keep it and report the users.

## 4. Registry plugin extensions are silent no-ops
forge/registry/plugin.go `applyAdminExtensions` (~232) and `applyAPIExtensions` (~238) are empty placeholders, so `registry.RegisterPlugin` accepts admin/API plugins and does nothing with them. (The working admin plugin path is `admin.Site.RegisterPlugin`, used by examples/ecommerce/main.go:212: do not touch it.) Make `RegisterPlugin` return `forgeerrors.NewNotImplementedError("registry admin plugin extensions")` / `("registry API plugin extensions")` when a plugin implements `AdminPlugin` / `APIPlugin`, before registering it, and delete the two placeholder functions. Test: registering such a plugin returns an error where `forgeerrors.IsNotImplemented(err)` is true; a plain plugin still registers.

## 5. `forge test` only prints a hint
forge/cli/commands/development/test.go prints "To run tests, use: go test ./..." instead of running tests (Django's `manage.py test` runs them). Make it run `go test` in the target directory with `exec.CommandContext`, passing `-v` when `--verbose` and `-cover` when `--coverage`, the package pattern `./...`, stdout/stderr connected to the command's output, and returning the exit error. Keep the existing flags. Test the argument building in a small pure function (`buildGoTestArgs(verbose, coverage bool) []string`), not by running go test.

## 6. MySQL mentions
forge/db/dialect/dialect.go documents MySQL support that does not exist. Edit only comments so they say PostgreSQL and SQLite. Leave forge/db/migrate/parse/lexer.go and db_test.go alone.

## Acceptance (from the module root)
    gofmt -l filter db identity registry cli          # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./filter/... ./db/... ./identity/... ./registry/... ./cli/...
    staticcheck ./filter/... ./db/... ./registry/... ./cli/...    # no output
    go test -race ./filter/... ./db/migrate/... ./registry/... ./cli/... -count=1

Report (max 30 lines): files deleted, identifiers removed, anything kept because grep found users, each command result.
