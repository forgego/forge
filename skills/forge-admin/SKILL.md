---
name: forge-admin
description: Register and customize Forge's React admin panel (`github.com/forgego/forge/admin`). Use when configuring `admin.Config[T]`, `ListDisplay`, `ListFilter`, `SearchFields`, `Fieldsets`, `Actions`, per-object permission hooks, `DefaultSite`, `UseDatabaseStores`, or mounting `adminSite.Handler()`. Do NOT use for REST API ViewSets/Serializers (use `forge-api`) or model schema definitions (use `forge-models`).
license: MIT
compatibility: Go 1.26.10+, github.com/forgego/forge v0.2.x
metadata:
  author: forgego
  version: "0.2"
---

# Forge Admin

Forge provides a registry-driven admin backend (`admin/`, `admin/core/`, `admin/api/rest/`) paired with an embedded React 19 + TypeScript + Vite + Tailwind CSS 4 SPA (`admin/ui/web/`).

## Golden Path

```go
p := catalog.ProductFieldsInstance

if _, err := admin.Register(&admin.Config[catalog.Product]{
	VerboseName:    "Product",
	ListDisplay:    []admin.Field{p.Name, p.Sku, p.Price, p.IsActive},
	ListFilter:     []admin.Field{p.IsActive},
	SearchFields:   []admin.Field{p.Name, p.Sku},
	Ordering:       []admin.Field{p.Name},
	ReadOnlyFields: []string{"created_at"},
	Fieldsets: []admin.Fieldset[catalog.Product]{
		{Name: "General", Fields: []string{"name", "sku", "price", "is_active"}},
		{Name: "Timestamps", Fields: []string{"created_at"}, Collapsed: true},
	},
}); err != nil {
	return err
}
```

Mount on `server.Router`:

```go
adminSite := admin.DefaultSite
adminSite.Title = "My Admin"
uiCfg := adminSite.GetUIConfig()
uiCfg.Prefix = settings.Admin.Path // e.g. "/admin"
adminSite.WithUIConfig(uiCfg)
adminSite.SetDB(database)

// Persist tokens, lockout, saved views & history in the database
if err := adminSite.UseDatabaseStores(ctx); err != nil {
	return err
}

srv.RegisterRoutes(func(r *server.Router) {
	r.Mount(settings.Admin.Path, adminSite.Handler())
})
```

## Hard Guardrails

- **No `admin/http`, `admin/filters`, `admin/widgets`, or Go HTML templates**: Those do not exist. The UI is a React SPA (`UISourceEmbedded` by default, or `UISourceStatic` / `UISourceExternal`).
- **Field type split in `admin.Config[T]`**:
  - `ListDisplay`, `ListDisplayLinks`, `ListEditable`, `ListFilter`, `SearchFields`, `Ordering`, `SortableBy` take `[]admin.Field` (pass generated `<Model>FieldsInstance.<Field>` or `admin.Computed("MethodName")`).
  - `Fields`, `Exclude`, `ReadOnlyFields`, `RawIDFields`, `AutocompleteFields` take `[]string`.
- **`Fieldset[T]` fields**: Uses `Name`, `Fields`, `Collapsed`, `Description` (not `Title` or `Collapsible`).
- **Permission hook signatures**: Include `admin *admin.Admin[T]` as the second parameter (`HasAddPermission(ctx, a, user)`, `HasChangePermission(ctx, a, user, obj)`).
- **Bulk actions**: Return `201` (all succeeded) or `207` (partial success with `BulkActionResponse.Errors` codes `invalid_id`, `not_found`, `permission_denied`).
- **Auth**: Admin REST login (`POST /api/login`) is bearer-token based, validating against `FORGE_ADMIN_USERNAME` / `FORGE_ADMIN_PASSWORD` (or `Site.SetLoginAuthenticator`) with rate-limited lockout.

## References

- [Admin configuration, permissions, actions & UI](references/admin.md)
