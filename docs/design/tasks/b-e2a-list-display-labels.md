TASK E2a (BH-2 / VB-02 phase 2): human-readable labels for foreign keys in admin list responses (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/admin-metadata/forge
(git worktree on branch fix/admin-metadata-readonly-display; task E1 already landed here.)

Files you may modify — ONLY:
  admin/core/metadata.go        (PaginatedResponse)
  admin/core/admin.go           (new ObjectLabels method)
  admin/api/rest/router.go      (handleList)
  tests: NEW admin/core/object_labels_test.go, NEW admin/api/rest/list_display_test.go
Do NOT edit go.mod / go.sum. Do NOT touch metadata_builder.go (E1) or the React UI.

=== PROBLEM ===
The admin list endpoint returns FK columns as bare ids; the UI can only show "#4182". The only relation
endpoint is autocomplete (text search), so the UI cannot batch-resolve ids. Labels must come from the
list response. Resolve them with ONE query per relation per page — never one query per row.

=== CONTEXT (verified) ===
- admin/core/metadata.go:  type PaginatedResponse struct { Count int64; Next, Previous string; PageSize, Page,
  TotalPages int; Results interface{} `json:"results"` }
- admin/core/admin.go:
    func (a *Admin[T]) GetQueryset(ctx) (orm.QuerySet[T], error)
    func (a *Admin[T]) getObjectID(obj *T) interface{}
    func (a *Admin[T]) getObjectLabel(obj *T) string     // same label autocomplete uses
    func (a *Admin[T]) Autocomplete(...)                  // reference for querying + labeling
- admin/core/registry.go:  func (r *Registry) Get(modelName string) (AdminInterface, error)
- admin/api/rest/router.go: Router has `registry *core.Registry`; handleList(admin core.AdminInterface)
  calls admin.ListObjects(ctx, params) and respondJSON(w, 200, response). admin.GetMetadata(ctx, user)
  returns *core.Metadata with Relations []RelationMetadata{Name, Type, RelatedModel, ...}.
  Relation Type strings come from schema.RelationType.String() — check the exact strings for foreign
  key and one-to-one in forge/schema/relation.go.

=== DESIGN (do exactly this; additive only) ===
1. PaginatedResponse gains:
       // Display maps relation field name -> related object id (as string) -> human label.
       Display map[string]map[string]string `json:"display,omitempty"`

2. admin/core/admin.go — an OPTIONAL capability (do NOT add it to AdminInterface):
       // LabelResolver is implemented by admins that can label objects by id in bulk.
       type LabelResolver interface {
           ObjectLabels(ctx context.Context, ids []interface{}) (map[string]string, error)
       }
   Put the interface in admin/core (next to AdminInterface or in admin.go) and implement it on *Admin[T]:
     - if len(ids)==0 return empty map
     - cap: if more than 1000 ids, only resolve the first 1000
     - qs := GetQueryset(ctx); filter primary key IN ids (use the same pk expression style the admin
       uses elsewhere, e.g. orm.F("id").In(ids...) — check how getObjectID/safeGetObjectByID address the pk)
     - All(ctx); map key = fmt.Sprint(getObjectID(obj)), value = getObjectLabel(obj)

3. router.go handleList — after ListObjects succeeds and before respondJSON:
     - meta, err := admin.GetMetadata(ctx, user); on error just skip labels (never fail the list).
     - Convert response.Results to generic rows ONCE: json.Marshal then json.Unmarshal into
       []map[string]interface{}; on error skip labels.
     - For each relation in meta.Relations whose Type is foreign key or one-to-one:
         * value key: use row[rel.Name]; if absent, try row[rel.Name+"_id"]; skip nil values.
         * collect distinct ids (string form -> original value) across rows.
         * related, err := r.registry.Get(rel.RelatedModel); skip if error; resolver, ok :=
           related.(core.LabelResolver); skip if !ok.
         * labels, err := resolver.ObjectLabels(ctx, ids); skip on error.
         * if len(labels) > 0: response.Display[rel.Name] = labels  (init the map lazily).
     - Do not change response.Results.
   Extract this into an unexported method `func (r *Router) attachDisplayLabels(ctx context.Context,
   admin core.AdminInterface, user interface{}, response *core.PaginatedResponse)` — adapt `user`'s
   type to whatever GetMetadata takes. Keep handleList readable.

=== TESTS ===
admin/core/object_labels_test.go: using the existing admin/core test fixtures (read existing tests for
how an Admin[T] with a sqlite DB is built), insert 3 objects, ObjectLabels(ctx, [id1, id3]) returns
exactly those 2 keys with getObjectLabel values; empty ids -> empty map.
admin/api/rest/list_display_test.go: using the router test harness already in router_test.go (read it),
register a parent model and a child model with a foreign key to it; list the child model; assert JSON
has display[<fk field>][<parent id>] == parent label; a child with a NULL FK contributes nothing; and
(count queries if the harness allows, otherwise skip) labels are resolved in one call per relation.
If the existing harness has no model with a relation and building one is disproportionate, test
attachDisplayLabels directly with a fake AdminInterface + fake LabelResolver instead.

=== VERIFY ===
  gofmt -l admin    (your files must not appear)
  go vet ./admin/...
  go test -race ./admin/...
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky)
  cd ../examples/ecommerce && go build ./... && cd -

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Additive only: no changes to AdminInterface or existing JSON fields.
- Final report: files changed, relation type strings used, tests, tail of the -race run.
