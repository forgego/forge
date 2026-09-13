Your workspace is /home/hamid/Other/projects/foreit-wt/w0-admin. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-2: admin create/update only write writable fields

Module root: /home/hamid/Other/projects/foreit-wt/w0-admin/forge (run go commands from there).
Only edit: forge/admin/core/admin.go, and create forge/admin/core/write_fields.go (helper) and forge/admin/core/write_fields_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run `go test ./admin/core -run WriteFields -count=1`, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- New functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state.
- Table-driven tests. Look at forge/admin/core/readonly_fields_test.go and admin_test.go for how an `Admin[T]` with a schema and config is built in tests and reuse that setup.

## Bug (verified on master)

- `CreateObject` (forge/admin/core/admin.go ~548) validates only fields present in the payload (`validateData`) and then decodes the whole payload into the instance.
- `UpdateObject` (~575-605) copies every request key except "id" into `orm.UpdateMap` and calls `a.manager.UpdateFields`.
- `Config.Fields`, `Config.Exclude`, `Config.ReadOnlyFields`, `Config.GetFields` and `Config.GetReadOnlyFields` (forge/admin/core/config.go ~53-64) are never consulted on these write paths.
- Neither are non-editable or auto-managed schema fields. Metadata already marks those read-only for the UI (`ReadOnly: !field.Editable || isAutoManaged(field)`, forge/admin/core/metadata_builder.go ~79; `isAutoManaged` at ~122).
- Result: a PATCH or POST can overwrite `created_at`, auto PKs, or fields excluded or marked read-only (e.g. `password`, `is_superuser` on a users admin).

## Required behaviour (Django admin / DRF semantics)

1. Add in write_fields.go:
   `func (a *Admin[T]) writableFields(ctx context.Context, instance *T, isNew bool) map[string]bool`
   - Start set = `a.config.GetFields(ctx, instance, isNew)` if that func is non-nil, else `a.config.Fields` if non-empty, else every schema field name.
   - Remove `a.config.Exclude`.
   - Remove `a.config.GetReadOnlyFields(ctx, instance, isNew)` if non-nil, else `a.config.ReadOnlyFields`.
   - Remove schema fields where `!field.Editable || isAutoManaged(field)`.
   - Handle a nil `a.config` / nil schema safely (nil schema: return nil and let callers skip filtering, keeping today's behaviour).
   Field names are the schema field names (the same keys the UI sends; `validateData` uses `field.Name`).

2. Add `func (a *Admin[T]) filterWritable(ctx context.Context, instance *T, isNew bool, data map[string]interface{}) (map[string]interface{}, error)`:
   - returns a NEW map (do not mutate `data`) containing only writable keys;
   - silently drops keys that are schema fields but not writable (the UI sends whole forms back, so this must not error);
   - returns a `*validation.ValidationErrors` (same type `validateData` uses) with message "unknown field" for keys that are not schema fields at all, ignoring "id".

3. Use it:
   - `CreateObject`: filter first (instance nil, isNew true), then `validateData` and `decodeData` on the filtered map. The history log (`LogAction`) should record the filtered map.
   - `UpdateObject`: after loading `instance`, filter (isNew false), then build `UpdateMap` from the filtered map. Keep the early return when nothing is left to update.

## Tests (write_fields_test.go)
- Update with `{"name": "new", "created_at": "...", "unknown": 1}` on a model with an AutoNowAdd `created_at`: returns a validation error mentioning "unknown".
- Same payload without "unknown": `name` updated, `created_at` unchanged in the DB.
- `Config.ReadOnlyFields = []string{"email"}`: update of email is ignored, other fields updated.
- `Config.Exclude = []string{"notes"}`: create ignores "notes".
- `GetReadOnlyFields` callback takes precedence over `ReadOnlyFields`.
- Pure unit test of `writableFields` covering Fields/Exclude/ReadOnly/auto-managed combinations (table-driven).

## Acceptance (run from /home/hamid/Other/projects/foreit-wt/w0-admin/forge)
    gofmt -l admin                       # empty
    go vet ./admin/...
    go test ./admin/core -run WriteFields -count=1 -v
    go test -race ./admin/... -count=1

Report (max 30 lines): failing output before the fix, files changed, and each acceptance command with its result.
