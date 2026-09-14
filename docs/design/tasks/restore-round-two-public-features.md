Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

ROLE: You are the delegated coder: edit the files yourself; do not call other agents. Use US spelling in comments.

# Restore more working public features that the cleanup removed

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge. A read-only copy of the code before the cleanup is at /home/hamid/Other/projects/foreit-wt/verify-master/forge (same module layout): read the original implementations there and restore them faithfully, adapted only where the current code requires it. Never edit files under verify-master.

## Principles (mandatory)
- Restore behaviour exactly; do not redesign. gofmt; naming unchanged from the original (these are public names users may already call).
- Add a small test for each restored item. Keep everything the cleanup merged (one viewset, one error writer, one limiter).

## 1. API default getters and one-call setup (forge/api)
Removed from forge/api/helpers.go and forge/api/integration.go: `GetDefaultAuthentication`, `GetDefaultPermissions`, `GetDefaultThrottles`, `GetDefaultRenderers`, `GetDefaultParsers` (getters paired with the setters that still exist) and `SetupCompleteAPI` (sets default renderers JSON/XML/HTML and parsers JSON/form/multipart). Restore them from verify-master (put `SetupCompleteAPI` in helpers.go; do not restore `CompleteExample`, `CreateProductionViewSet` or `RegisterAPIWithDefaults`, which built the removed broken viewset). Test: after `SetupCompleteAPI()`, `GetDefaultRenderers()` has 3 entries and `GetDefaultParsers()` has 3.

## 2. Saved filters and the filter query planner (forge/filter)
Removed: forge/filter/persistence.go (`SavedFilter`, `FilterStorage`, `InMemoryFilterStorage`, `SaveFilter`, `LoadFilter`, `PreviewFilter`) and the `FilterStorage`/`RBACFilterStorage` interfaces from shared.go; forge/filter/optimizer.go (`QueryOptimizer`, `NewQueryOptimizer`, `Optimize`, `QueryPlan`).
- Restore persistence.go and the interfaces. Fix its only bug: `InMemoryFilterStorage` gets a `sync.RWMutex` guarding every map access (Save, Load, Update, Delete, List). Test: 50 goroutines saving and listing concurrently pass under `-race`; Save then Load round-trips the AST.
- Restore optimizer.go unchanged (do not re-add the unused `optimizer` field on `FilterSet`). Test: `Optimize(nil)` returns strategy "none"; a two-condition AST returns a non-empty strategy.
- Do not restore filter/cache.go (owner does not want filter caching) or filter/dialect.go (it interpolates values into SQL).

## 3. identity password helpers (forge/identity)
Removed forge/identity/password.go (`IsHashed`, `HashPassword`, `HashPasswordWithCost`, `CheckPassword`, `CheckPasswordHash`, `NeedsRehash` re-exporting forge/identity/utils). `identity.HashPassword` is a natural public import path. Restore the file from verify-master unchanged. Test: HashPassword then CheckPassword is true.

## 4. Migration checksum import path (forge/db/migrate/execute)
The duplicate checksum code was removed in favor of forge/db/migrate/verify. Keep one implementation, but keep the old import path working: in package execute add
`type ChecksumValidator = verify.ChecksumValidator`, `func NewChecksumValidator(migrationsDir string) *ChecksumValidator { return verify.NewChecksumValidator(migrationsDir) }`, and `CalculateChecksum`/`ValidateChecksum` delegating to verify's package functions (match the original signatures in verify-master exactly). Mark each `// Deprecated: use verify.X.` Test: execute.CalculateChecksum equals verify.CalculateChecksum for the same input.

## 5. Multi-step rollback (forge/db)
The removed `execute.RollbackManager` could roll back N steps (`RollbackOptions{Steps: n}`); `db.MigrationRunner` only has `Rollback` (one step) and `RollbackTo(version)`. Add `func (mr *MigrationRunner) RollbackSteps(ctx context.Context, steps int) error` to forge/db/migrations.go using golang-migrate's `Steps(-n)` (read how `Rollback` is implemented and reuse its setup and error wrapping). `steps <= 0` returns an error. Test with a temp dir of 3 SQLite migrations: Migrate, RollbackSteps(2), Version == 1.

## Acceptance (from the module root)
    gofmt -l api filter identity db               # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./api/... ./filter/... ./identity/... ./db/...
    staticcheck ./api/... ./filter/... ./identity/... ./db/...   # no output
    go test -race ./api/... ./filter/... ./identity/... ./db/... -count=1
    ~/go/bin/golangci-lint run --timeout=5m --config=../.golangci.yml ./api/... ./filter/... ./identity/... ./db/...   # 0 issues

Report (max 25 lines): what was restored per item, tests added, each command result.
