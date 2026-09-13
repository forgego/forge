Your workspace is /home/hamid/Other/projects/foreit-wt/w0-orm-writes. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-8: ORM writes use real column names and the schema primary key

Module root: /home/hamid/Other/projects/foreit-wt/w0-orm-writes/forge (run go commands from there).
Only edit: forge/orm/queryset.go (method Update only), forge/orm/update_builder.go (Increment/Decrement only), forge/orm/manager.go (Create, setID), forge/orm/manager_helpers.go (BuildInsertSQL, BuildBulkInsertSQL, ExecuteInsert), forge/orm/manager_helpers_test.go (update call sites only), and create forge/orm/write_column_names_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming: descriptive names after the behaviour, never after tickets (no "W0" anywhere).
- Reuse the SQLite test helpers in forge/orm/relation_path_join_test.go (setupRelationTestDB pattern) for DB-backed tests.

## Bug 1 (verified on master): update ignores custom column names
A schema field can have a database column different from its name (`schema.DBColumn("...")`, forge/schema/fields_functional.go ~142; FieldInfo.DBColumn in forge/orm/schema.go ~150).
- `BaseQuerySet.Update` (forge/orm/queryset.go, loop `for fieldName, value := range updates` with `escapedField := EscapeIdentifier(fieldName)` ~1512) writes the caller's key as the column.
- `UpdateBuilder.Increment`/`Decrement` (forge/orm/update_builder.go, `fieldSQL := EscapeIdentifier(fieldName)` ~196) do the same.
- `Manager.UpdateFields` validates keys with `schema.GetField`, which matches a field name OR column, case-insensitively, so key "Price" or a field named "price" with column "unit_price" passes validation and then produces SQL for a column that does not exist.
Fix: add `func (qs *BaseQuerySet[T]) columnFor(key string) (string, error)` returning `field.DBColumn` from `qs.schema.GetField(key)` (error "field %s not found on %s" when missing, or when schema is nil return key unchanged to keep current behaviour for schema-less querysets). Use it in Update (sort the keys first so SET order and placeholders are deterministic) and in Increment/Decrement.
Tests: a model with field "price" stored in column "unit_price" (schema.DBColumn); `UpdateFields(ctx, id, UpdateMap{"price": 9.5})` updates the row (read back with raw SQL); `Update` with key "PRICE" also works; unknown key returns an error mentioning the key.

## Bug 2 (verified on master): Create assumes an int64 column named id
- `BuildInsertSQL` (forge/orm/manager_helpers.go ~160-200) always appends `RETURNING id`; `BuildBulkInsertSQL` (~419-475) too.
- `ExecuteInsert` scans into int64.
- `Manager.setID` (forge/orm/manager.go) only sets an int64 struct field literally named ID/Id/id, ignoring `m.schema.PrimaryKey` and the field's StructFieldName; its result is discarded.
- A model whose primary key column is `product_id` gets a failing INSERT.
Fix:
- Add a `pkColumn string` parameter to BuildInsertSQL and BuildBulkInsertSQL (empty string → "id") and use it in RETURNING; update the callers in manager.go (use `m.primaryKeyColumn()`) and in manager_helpers_test.go.
- In `setID`, when the instance does not implement ModelWithID, locate the struct field for the schema primary key (FieldInfo.StructFieldName, falling back to ID/Id/id) and set it if it is an int kind; return an error `fmt.Errorf("cannot set primary key %s on %T", ...)` when no settable integer field exists, and make `Create` return that error.
- Non-integer primary keys remain unsupported: if the schema primary key field's Go type is not an integer kind, `Create` returns `errors.NewNotImplementedError("non-integer primary keys")` from forge/errors before executing SQL.
Tests: model with primary key column "product_id" (int64 field ProductID): Create inserts and sets ProductID; model with string primary key: Create returns a NotImplemented error (`errors.IsNotImplemented`); existing manager_helpers_test expectations updated for the new parameter.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-orm-writes/forge)
    gofmt -l orm                       # empty
    go vet ./orm
    staticcheck ./orm                  # no output
    go test ./orm -count=1 -v -run 'ColumnName|PrimaryKey|UpdateFields|Create'
    go test -race ./orm ./admin/... -count=1

Report (max 30 lines): failing output before the fix, files changed, each acceptance command with its result.
