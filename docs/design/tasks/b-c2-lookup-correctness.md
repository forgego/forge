TASK C2: filter lookups generate wrong or empty SQL (Go backend, filter + orm).

Working directory: /home/hamid/Other/projects/foreit-wt/filter-expression-correctness/forge
(git worktree on branch fix/filter-expression-correctness; C1 already landed here.)

Files you may modify — ONLY:
  filter/parser.go, filter/expression_converter.go
  orm/query_expr.go, orm/expression.go
  tests: NEW filter/lookup_correctness_test.go, NEW orm/lookup_sql_test.go
Do NOT edit go.mod / go.sum. Do NOT touch filter/filterset.go or filter/cache.go.

There are TWO SQL builders for comparisons:
  (1) orm/query_expr.go  QueryExpr.buildSingle   (around line 150)
  (2) orm/expression.go  ComparisonExpression SQL generation (switch around lines 400-470)
Read both first. Every fix below must be applied to BOTH builders where the operator exists.
Write failing tests FIRST.

=== BUG 1: IN / NOT IN / range silently emit EMPTY SQL for typed slices ===
query_expr.go:
    values, ok := q.value.([]interface{})
    if !ok { return "", nil, currentIndex }      // same for OpNotIn and OpRange
filter/parser.go parseValue returns []string for "in" and "range", so the assertion fails and
the predicate becomes "" — the filter is DROPPED and the query returns unfiltered rows.

Fix: add an unexported helper in package orm
    func toInterfaceSlice(v interface{}) ([]interface{}, bool)
using reflect: nil -> (nil,false); a slice or array of ANY element type -> converted, true;
anything else -> (nil,false). Use it for OpIn, OpNotIn, OpRange in both builders.
If conversion fails, do NOT return empty SQL silently: return SQL that matches nothing for IN
and range ("1=0"), and everything for NOT IN ("1=1")... EXCEPT keep the function signatures;
if the builder has an error return path, prefer returning an error there instead — read the code.

=== BUG 2: empty IN produces invalid SQL "field IN ()" ===
After conversion, if len(values)==0:
  OpIn    -> "1=0"   (no args)
  OpNotIn -> "1=1"   (no args)
Range with len != 2 -> "1=0".

=== BUG 3: HTTP `field__in=` (empty value) becomes nil ===
filter/parser.go parseValue: `if valueStr == "" { return nil, nil }` runs before the switch, so an
explicit empty IN list turns into nil and skips BUG 2 handling.
Fix: for lookup "in", an empty valueStr must return `[]string{}` (empty slice, not nil).
Keep returning nil for other lookups.

=== BUG 4: isnull=false still generates IS NULL ===
parser returns false for `field__isnull=false`, but OpIsNull ignores the value and always emits
"IS NULL". Fix in both builders: if op is OpIsNull and the value is the bool `false`, emit
"<field> IS NOT NULL". (value true, or no value, keeps IS NULL.)

=== BUG 5: case-insensitive prefix/suffix are case-SENSITIVE, and ILIKE breaks SQLite ===
filter/expression_converter.go lookupToOperator maps "istartswith" -> orm.OpStartsWith and
"iendswith" -> orm.OpEndsWith (comment claims "dialect adapter will handle case"; nothing does).
Also OpIContains / OpIExact emit `ILIKE`, which does not exist in SQLite (supported dialect).

Fix:
  - In orm/query_expr.go add operators  OpIStartsWith Operator = "ISTARTSWITH"  and
    OpIEndsWith Operator = "IENDSWITH"  (next to the existing operator constants; make sure the
    string values are unique among Operator constants — the buildSingle if/else chain compares by
    value).
  - Map istartswith -> OpIStartsWith, iendswith -> OpIEndsWith in lookupToOperator.
  - Generate portable case-insensitive SQL for ALL four i-lookups in both builders:
        LOWER(<field>) LIKE LOWER($n)
    with args: iexact -> value, icontains -> "%"+v+"%", istartswith -> v+"%", iendswith -> "%"+v.
    Keep the case-sensitive ops (contains/startswith/endswith) as plain LIKE.
  - Add FieldExpression methods only if orm.Field[string] already exposes IContains/IExact-style
    methods and a matching IStartsWith/IEndsWith is needed for the converter to compile; otherwise
    don't add public API.

Out of scope (do NOT attempt): EXTRACT(YEAR/MONTH/DAY) on SQLite, relation joins, dialect plumbing.

=== TESTS ===
orm/lookup_sql_test.go — build SQL through each builder and assert exact SQL + args:
  IN with []string{"a","b"}         -> "... IN ($1, $2)", args [a b]
  IN with []int{1,2,3}              -> 3 placeholders
  IN with []string{}                -> "1=0", no args
  NOT IN with []string{}            -> "1=1"
  range with []string{"1","9"}      -> "BETWEEN $1 AND $2"
  range with []int{1}               -> "1=0"
  isnull true -> "IS NULL"; isnull false -> "IS NOT NULL"
  iexact "Bob" -> "LOWER(<f>) LIKE LOWER($1)", args [Bob]
  icontains/istartswith/iendswith   -> LOWER(...) LIKE LOWER(...) with correct % placement
  startswith stays "LIKE $1" (no LOWER)
  No generated SQL contains "ILIKE".
filter/lookup_correctness_test.go:
  parseValue("", "in") returns an empty non-nil []string
  parseValue("", "exact") returns nil
  lookupToOperator("istartswith", string type) == orm.OpIStartsWith; iendswith likewise.
Also add one test that runs a real query on sqlite in-memory if the orm tests already have a
sqlite helper (grep orm/*_test.go for sqlite3): name__istartswith=AL matches "alice";
id__in= (empty) returns zero rows. Skip this test if no helper exists — do not build one.

=== VERIFY ===
  gofmt -l filter orm        (only pre-existing files may appear; your files must not)
  go vet ./filter/... ./orm/...
  go test -race ./filter/... ./orm/...
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky; report, don't fix)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes except the two new Operator constants.
- Minimal diff; no unrelated refactors.
- Final report: files changed, tests added, exact tail of the -race run.
