TASK C3a: year/month/day filters generate invalid SQL (Go backend, orm).

Working directory: /home/hamid/Other/projects/foreit-wt/filter-expression-correctness/forge
(git worktree on branch fix/filter-expression-correctness; C1 and C2 already landed here.)

Files you may modify — ONLY:
  orm/sql_builder.go, orm/expression.go, orm/queryset.go, orm/projection.go
  tests: NEW orm/date_parts_test.go
Do NOT edit go.mod / go.sum, filter/*, orm/query_expr.go.

Write failing tests FIRST.

=== BUG ===
HTTP filters like `created_at__year=2024` go through filter/expression_converter.go, which builds an
orm.ComparisonExpression with Op = orm.OpYear / OpMonth / OpDay. But
orm/expression.go `ComparisonExpression[T].ToSQL(builder *SQLBuilder)` has NO case for these
operators, so it falls through to the generic branch and emits something like
    "created_at" EXTRACT(YEAR FROM $1)
which is invalid SQL on every database. (Only the legacy orm/query_expr.go path emits EXTRACT, and
even that is PostgreSQL-only — SQLite has no EXTRACT.)

SQLBuilder (orm/sql_builder.go) currently has no idea which dialect it is building for:
    type SQLBuilder struct { paramIndex int; args []interface{} }
    func NewSQLBuilder() *SQLBuilder
The db/dialect package has `type Dialect interface { Name() string; ... }` (names like
"postgres", "sqlite"; check db/dialect for the exact strings returned by the SQLite and PostgreSQL
dialects). orm/types.go has `func GetDialect(conn interface{}) (dialect.Dialect, error)`.

=== FIX ===
1. orm/sql_builder.go:
     - add field `dialect dialect.Dialect`
     - add `func NewSQLBuilderWithDialect(d dialect.Dialect) *SQLBuilder` (same init as NewSQLBuilder, plus d)
     - add `func (b *SQLBuilder) isSQLite() bool` — true only when b.dialect != nil and its Name()
       is the SQLite dialect's name (compare case-insensitively; accept "sqlite" and "sqlite3").
   NewSQLBuilder() keeps its behaviour (nil dialect = PostgreSQL SQL, as today).

2. orm/queryset.go and orm/projection.go: at each `NewSQLBuilder()` call site (queryset.go ~505,
   ~1196, ~1237, ~1318; projection.go ~81), use the queryset's database connection's dialect when
   one is available: read how BaseQuerySet stores its connection (field name), call
   GetDialect(thatConn); on success use NewSQLBuilderWithDialect(d), otherwise NewSQLBuilder().
   Put this in ONE small unexported helper, e.g. `func (qs *BaseQuerySet[T]) newSQLBuilder() *SQLBuilder`,
   and use it at every call site. If projection.go has no access to the queryset connection, leave it.

3. orm/expression.go ComparisonExpression[T].ToSQL: add explicit handling BEFORE the generic
   fallback for OpYear, OpMonth, OpDay:
       part := "YEAR"/"MONTH"/"DAY"; fmtCode := "%Y"/"%m"/"%d"
       placeholder := builder.AddArg(c.Value)
       if builder.isSQLite():
           sql = fmt.Sprintf("CAST(strftime('%s', %s) AS INTEGER) = %s", fmtCode, fieldSQL, placeholder)
       else:
           sql = fmt.Sprintf("EXTRACT(%s FROM %s) = %s", part, fieldSQL, placeholder)
   Careful: fmtCode contains '%' — build that fragment so fmt does not interpret it (e.g. concatenate
   "'" + fmtCode + "'" instead of passing it through a format verb that re-formats it). Test the
   exact output string.
   Operator constants compare by value; OpYear/OpMonth/OpDay have unique string values — put the new
   branch in the existing if/else chain or switch, wherever the operator dispatch lives.

=== TESTS (orm/date_parts_test.go) ===
  - ToSQL with NewSQLBuilder() for OpYear on a time field -> exactly `EXTRACT(YEAR FROM <fieldSQL>) = $1`
    and args [2024]; OpMonth -> MONTH; OpDay -> DAY.
  - ToSQL with NewSQLBuilderWithDialect(<the SQLite dialect constructor from db/dialect>) ->
    exactly `CAST(strftime('%Y', <fieldSQL>) AS INTEGER) = $1` (and %m, %d).
  - If the orm tests already have a sqlite in-memory helper (grep orm/*_test.go for "sqlite3"),
    add one integration test: table with created_at values in 2023 and 2024; filter year==2024
    returns only the 2024 row; month and day filters likewise. If no helper exists, skip this test
    with a comment — do not build a helper.

=== VERIFY ===
  gofmt -l on files you changed/created   (only pre-existing unformatted files may appear)
  go vet ./orm/... ./filter/...
  go test -race ./orm/ ./filter/
  go test ./...   (TestPrefetchRelated_Integration is known-flaky)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Only additive exported API: NewSQLBuilderWithDialect. No other signature changes.
- Minimal diff.
- Final report: files changed, tests added, exact SQL strings produced, tail of the -race run.
