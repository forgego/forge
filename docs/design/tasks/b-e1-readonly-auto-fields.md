TASK E1 (BH-1 / VB-06): auto-managed fields are editable in the admin create form (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/admin-metadata/forge
(git worktree on branch fix/admin-metadata-readonly-display, based on master.)

Files you may modify — ONLY:
  admin/core/metadata_builder.go
  tests: admin/core/*_test.go (existing or NEW admin/core/readonly_fields_test.go)
Do NOT edit go.mod / go.sum. Do NOT touch the React UI.

=== THE BUG ===
admin/core/metadata_builder.go buildFieldsMetadata (around line 79):
    ReadOnly: !field.Editable, // ReadOnly is inverse of Editable usually, or need to check field definition
It ignores whether the ORM/database manages the value. forge/schema/field.go has:
    PrimaryKey bool; AutoIncrement bool; AutoNow bool; AutoNowAdd bool; Generated bool
So created_at (AutoNowAdd), updated_at (AutoNow), generated columns and auto-increment primary keys
are emitted with read_only=false (if Editable is true) and the admin create form renders them as
editable inputs whose values the backend discards. The frontend already honours read_only correctly
(ModelUpsertPage hides read_only fields on create) — the flag is simply wrong.

=== FIX ===
Add an unexported helper in admin/core/metadata_builder.go:

    // isAutoManaged reports whether the database/ORM owns this field's value.
    func isAutoManaged(f schema.Field) bool {
        return f.AutoNow || f.AutoNowAdd || f.Generated || (f.PrimaryKey && f.AutoIncrement)
    }

(Use the actual field type that buildFieldsMetadata iterates — read `s.Fields()` return type; it may be
a pointer — adapt the signature accordingly.)
Then set:   ReadOnly: !field.Editable || isAutoManaged(field),
and replace the uncertain trailing comment with an accurate one.

Also check: if a field is auto-managed, `Required` must not force the user to fill it — set
`Required: field.Required && !isAutoManaged(field)`.

Before changing, check how `Editable` defaults (forge/schema/field_config.go and field constructors):
if Editable defaults to false for normal fields, the current code would already mark everything
read-only — report what you find, and do NOT change the Editable semantics; only add the auto-managed
rule.

=== TESTS ===
Build a schema with fields (use whatever schema/test helper admin/core tests already use — read
existing *_test.go in admin/core first):
  - name: string, Editable true                      -> read_only false, required as declared
  - created_at: AutoNowAdd, Editable true            -> read_only true, required false
  - updated_at: AutoNow, Editable true               -> read_only true
  - id: PrimaryKey + AutoIncrement, Editable true    -> read_only true
  - slug: Generated                                  -> read_only true
  - sku: PrimaryKey without AutoIncrement, Editable  -> read_only false (user-assigned key)
  - notes: Editable false                            -> read_only true (unchanged behaviour)

=== VERIFY ===
  gofmt -l admin/core     (your files must not appear)
  go vet ./admin/...
  go test -race ./admin/...
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes. Minimal diff.
- Final report: what Editable defaults to, files changed, tests, tail of the -race run.
