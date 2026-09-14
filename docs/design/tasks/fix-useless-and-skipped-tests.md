Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Remove tautological tests, give empty tests real assertions, un-skip skipped tests

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit _test.go files (and test-only helpers). Do not change non-test code; if a test can only pass by changing non-test code, keep it skipped, and list it with the reason in the report.

## Principles (mandatory)
- A test must be able to fail. Assert the behaviour the test name promises.
- Table-driven where it helps; `require.NoError` before using a result.
- Naming after behaviour, never after tickets.

## 1. Tests that only re-check what the compiler checks: replace with a compile-time assertion
- forge/db/dialect/dialect_test.go:383 `TestPostgreSQLDialect_ImplementsDialect`, :387 `TestSQLiteDialect_ImplementsDialect` → delete both; add `var _ Dialect = (*<PostgresType>)(nil)` and the SQLite equivalent in the test file (use the real type names).
- forge/db/pool_test.go:268 `TestOptionType`, forge/admin/site_test.go:179 `TestTypeAliases` → read them; if they only assert that a type exists or aliases compile, delete them.

## 2. Tests without assertions: add the assertion the name promises
- forge/orm/expression_test.go:253 `TestCombinedExpression_WithValues`
- forge/api/serializer_test.go:54 `TestBaseSerializer_Validate_Invalid` (must assert a validation error)
- forge/api/parsers/parser_test.go:81 `TestXMLParser_Parse` (assert the parsed value)
- forge/log/logger_test.go:48 `TestLoggerTrace` (use an observer core or buffer and assert the entry was written, if the logger allows it; otherwise assert no panic and that the level is enabled)
- forge/orm/schema_test.go:9 `TestGetModelSchema`, :19 `TestNewFieldAccessor`: assertions only run `if err == nil`, so setup failures pass. Use `require.NoError` first.
If an added assertion fails, the test found a real bug: keep the assertion, mark the test `t.Skip("known bug: <one line>")`, and list it in the report. Do not weaken the assertion.

## 3. Skipped tests: un-skip where the reason is gone
Skips: forge/orm/update_builder_test.go (7), orm/schema_test.go (4), orm/date_parts_test.go (1), orm/expression_test.go (1), orm/select_related_test.go (1), identity/service/password_test.go (1), identity/service/user_test.go (1), db/transaction_test.go (1). Leave forge/db/pool_test.go alone.
- "schema not registered": register the test model's schema the way passing orm tests do (grep for `RegisterSchema` / `MustRegister` in forge/orm/*_test.go and forge/orm/testing helpers).
- "no sqlite helper": sqlite helpers exist; grep forge/db and forge/internal/testutils for an in-memory SQLite helper.
- Anything else: fix it if the fix is test-only, otherwise leave the skip with an accurate reason.

## Acceptance (from the module root)
    gofmt -l .                                         # empty
    go vet ./...
    go test -race ./orm/... ./db/... ./identity/... ./api/... ./admin/... ./log/... -count=1
    grep -rn 't\.Skip' --include=*_test.go orm identity db/transaction_test.go | wc -l    # report before/after

Report (max 30 lines): deleted tests, fixed tests, skips removed (before/after count), skips kept with reason, real bugs found.
