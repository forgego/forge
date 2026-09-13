Your workspace is /home/hamid/Other/projects/foreit-wt/w0-orm. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-1: ORM stubs fail loudly, First/Last default ordering, Set(nil)

Module root: /home/hamid/Other/projects/foreit-wt/w0-orm/forge (run go commands from there).
Only edit: forge/orm/queryset.go, forge/orm/update_builder.go, and test files you create in forge/orm/ (name them w0_orm_test.go).

## Principles (mandatory)
- Reproduce first: write the tests below FIRST, run `go test ./orm -run W0 -count=1`, and keep the failing output for your report. Then fix.
- Smallest correct change; no unrelated edits or renames; gofmt.
- New functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state.
- Table-driven tests with descriptive names. Reuse the SQLite helpers already in forge/orm/relation_path_join_test.go (setupRelationTestDB, TestCustomer/TestCompany/TestOrder models) instead of inventing new fixtures.

## Bugs (verified on master)

1. `Union`, `Intersection`, `Difference` (forge/orm/queryset.go ~1621-1645) ignore `other` and return `qs.clone()`.
   Fix: return a clone whose `err` field is set to `errors.NewNotImplementedError("QuerySet.Union")` (resp. Intersection, Difference). Import `github.com/forgego/forge/errors` (orm/manager.go already imports it as `errors`; if the file uses stdlib errors, alias forge's as `forgeerrors`). Keep an existing non-nil err.
   Test: `qs.Union(other).All(ctx)` returns an error for which `forgeerrors.IsNotImplemented(err)` is true; same for the other two.

2. `Aggregate` (~line 435) appends to `clone.aggregates`, which is never rendered into SQL.
   Fix: same as 1 with feature name "QuerySet.Aggregate". Do not delete the `aggregates` field.
   Test: `qs.Aggregate(Count("id")).All(ctx)` returns a NotImplemented error.

3. `First` / `Last` without ordering. `Reverse` (~338) only flips existing `orderBy`, so on an unordered queryset `Last()` runs the same `LIMIT 1` query as `First()`.
   Fix (Django semantics): if the queryset has no `orderBy`, `First` orders by primary key ascending and `Last` by primary key descending. Primary key column: `qs.schema.PrimaryKey`, falling back to "id" (same fallback used in Count). Add one small helper, e.g. `func (qs *BaseQuerySet[T]) withDefaultPKOrder() *BaseQuerySet[T]`, used by both. Existing ordering must be respected unchanged.
   Test: insert customers with ids 1,2,3 (names in non-id order); unordered `First` returns id 1, unordered `Last` returns id 3; `OrderBy("name").Last` returns the last by name.

4. `UpdateBuilder.Set(field, nil)` (forge/orm/update_builder.go ~53-56) panics: `reflect.TypeOf(nil)` is nil and `.AssignableTo` is called on it.
   Fix: if `value == nil`: allowed when the field's `expectedType` kind is Ptr, Interface, Slice or Map, or its package path is "database/sql" and its name starts with "Null"; then store nil so SQL writes NULL. Otherwise record `fmt.Errorf("field %s cannot be set to nil", fieldName)` in `ub.err` (keep the first error, as the surrounding code does).
   Test: Set on a non-nullable field with nil returns an error and does not panic (use `assert.NotPanics`); Set on a pointer-typed field with nil records no error. If no existing test model has a pointer field, define a minimal model in w0_orm_test.go following the TestCustomer pattern.

## Acceptance (run from /home/hamid/Other/projects/foreit-wt/w0-orm/forge)
    gofmt -l orm                      # empty
    go vet ./orm
    go test ./orm -run W0 -count=1 -v # new tests pass
    go test -race ./orm -count=1
    go test ./admin/... ./filter/... -count=1

Report (max 30 lines): the failing output from before the fix (one line per test), the files changed, and each acceptance command with its result.
