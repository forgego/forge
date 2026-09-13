TASK C3b: filtering/ordering by related fields (`customer__name`) generates invalid SQL (Go backend, orm).

Working directory: /home/hamid/Other/projects/foreit-wt/filter-expression-correctness/forge
(git worktree on branch fix/filter-expression-correctness; C1, C2, C3a already landed here.)

Files you may modify — ONLY:
  orm/sql_builder.go, orm/expression.go, orm/queryset.go
  tests: NEW orm/relation_path_join_test.go
Do NOT edit go.mod / go.sum, filter/*, orm/query_expr.go.

Write failing tests FIRST.

=== THE BUG (verified) ===
- Expressions resolve relation paths for VALIDATION (orm/expression.go resolveNestedFieldPath walks
  schema.GetRelation(part) -> GetModelSchemaByName(rel.TargetModel)), but SQL generation ignores that:
  `Field[T].ToSQL` (expression.go ~50) and the FieldRef equivalent (find it: `func (f FieldRef) ToSQL`)
  emit EscapeIdentifier(fieldPath) => `"customer__name"`, a column that does not exist.
- Joins are only built for select_related: queryset.go `buildJoinClause` (~554) loops qs.selectRelated,
  resolves the FK column with a multi-strategy search over qs.schema.Fields, and records qs.joins /
  qs.joinMap. Nothing adds joins for relation paths used in conditions, excludes, or ordering.
So `qs.Filter(orm.F("customer__name").Eq("Acme"))` fails at the database.

=== DESIGN (implement this) ===
1. SQLBuilder gets an optional join resolver, set by the queryset:
       type JoinResolver func(relationPath []string) (alias string, err error)
       func (b *SQLBuilder) SetJoinResolver(r JoinResolver)       // exported only if needed by tests; else unexported
       func (b *SQLBuilder) resolveColumn(fieldPath string) (string, error)
   resolveColumn:
     - parts := splitFieldPath(fieldPath); if len(parts)==1 or no resolver: return EscapeIdentifier(fieldPath)
       (preserve today's single-field output EXACTLY, including the table-prefixed branch in Field.ToSQL).
     - else alias, err := resolver(parts[:len(parts)-1]); return EscapeIdentifier(alias) + "." +
       EscapeIdentifier(<db column of the last part in the target schema>) — the resolver must also give
       the column; design the resolver signature as needed, e.g.
           func(parts []string) (alias string, column string, err error)
       taking the FULL path.

2. Field[T].ToSQL and FieldRef.ToSQL call builder.resolveColumn for paths containing "__"
   (keep behaviour identical when builder is nil or path has no "__").

3. BaseQuerySet provides the resolver (queryset.go):
     - walks the path like resolveNestedFieldPath; for each relation hop computes the FK column using the
       SAME strategy buildJoinClause already uses (extract that FK-column search into a helper
       `fkColumnFor(schema, rel)` and use it from both places — do not duplicate the logic);
     - alias for a hop = the relation path joined with "__" (e.g. "customer", "customer__company");
     - records a LEFT JOIN `LEFT JOIN "<target_table>" AS "<alias>" ON "<alias>"."<target pk column>" =
       "<parent alias or main table>"."<fk column>"` once per alias (dedupe with qs.joinMap, which
       select_related also uses — if select_related already joined the same first-level relation under
       the same alias, reuse it);
     - returns alias + the db column of the final field (FieldInfo DBColumn, fallback to name).
   Use the target schema's primary key column (find how the pk is represented in ModelSchema; default "id").

4. SQL assembly order: joins required by WHERE/ORDER BY are only known after those clauses are built.
   In the All() builder path (queryset.go ~505) and Count() (~1203): build WHERE and ORDER BY strings
   first (with the resolver set on the builder), THEN assemble
       SELECT ... FROM main [select_related joins] [path joins] WHERE ... ORDER BY ... LIMIT/OFFSET
   Placeholders are numbered as clauses are built; joins add no args, so numbering stays correct —
   but ensure the final SQL text places arguments in the same order they were added (WHERE before
   ORDER BY is fine because ORDER BY has no args; verify).
   When a path join is used, SELECT must still select only main-table columns (qualify with the main
   table name if the current select list is unqualified, to avoid ambiguous column errors) and Count must
   use COUNT(DISTINCT main.pk) instead of COUNT(*) so one-to-many joins cannot inflate counts.

5. Update() and Delete(): if building their WHERE would require a join, return a clear error
   `fmt.Errorf("filtering by related fields is not supported in %s", "update"/"delete")` instead of
   emitting invalid SQL. (Detect via the resolver having been called.)

=== TESTS (orm/relation_path_join_test.go) ===
Use the orm package's existing sqlite in-memory test helpers and model registration pattern (grep
orm/*_test.go for how models/schemas are registered and a sqlite DB is created — prefetch_related_test.go
uses relations and is a good reference).
Models: Company{id, name}, Customer{id, name, company FK}, Order{id, total, customer FK}.
  - Filter Order by customer__name == "Acme" returns only Acme's orders (All) and Count matches.
  - Filter Order by customer__company__name (two hops) works.
  - OrderBy("customer__name") sorts correctly.
  - A customer with two orders: Count on Customer filtered by a one-to-many path (if reverse relations are
    supported by GetRelation; if not, skip) — or at minimum verify COUNT(DISTINCT ...) is emitted when a
    join is present.
  - Filter with no relation path produces SQL byte-identical to before (assert the SQL string for a simple
    `name = ?` filter did not change — capture it before your change in the test as a literal).
  - Update/Delete with a relation-path filter returns the "not supported" error.

=== VERIFY ===
  gofmt -l on files you changed/created   (only pre-existing unformatted files may appear)
  go vet ./orm/... ./filter/...
  go test -race ./orm/ ./filter/
  go test ./...   (TestPrefetchRelated_Integration is known-flaky; if it now fails consistently, investigate
                   whether your join/alias change caused it and fix)
  cd ../examples/ecommerce && go build ./... && cd -

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No changes to exported signatures other than additive helpers. Single-field SQL output must not change.
- If you cannot make a part work, STOP and report precisely which part and why — do not leave
  half-working join generation in place; revert your edits to that function.
- Final report: design as implemented, files changed, test names, exact SQL for the two-hop filter,
  tail of the -race run.
