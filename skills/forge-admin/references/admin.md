# Admin Interface Reference

## Registration & Site Wiring

`admin.Register(&admin.Config[T]{...})` constructs an `orm.Manager[T]` from `T.Meta().TableName` automatically and registers it on `admin.DefaultSite` (or use `admin.RegisterWithSite(site, cfg)`).

```go
adminSite := admin.DefaultSite
adminSite.Title = "Store Admin"
uiCfg := adminSite.GetUIConfig()
uiCfg.Prefix = "/admin"
// Optional: serve prebuilt static assets instead of embedded UI
// uiCfg.Source = admin.UISourceStatic
// uiCfg.StaticDir = "./admin/ui/dist"
adminSite.WithUIConfig(uiCfg)
adminSite.SetDB(database)

if err := adminSite.UseDatabaseStores(ctx); err != nil {
	return err
}

srv.RegisterRoutes(func(r *server.Router) {
	r.Mount("/admin", adminSite.Handler())
})
```

`UseDatabaseStores` persists admin bearer tokens, login lockout attempts, saved views, and change history (`forge_admin_log`) in the database instead of process memory.

## `admin.Config[T]` Options (`admin/core/config.go`)

### List View
- `ListDisplay []admin.Field`
- `ListDisplayLinks []admin.Field`
- `ListEditable []admin.Field`
- `ListFilter []admin.Field`
- `SearchFields []admin.Field`
- `Ordering []admin.Field`
- `SortableBy []admin.Field`
- `Filters []admin.Filter[T]` (custom filter implementations)
- `DateHierarchy string`
- `ListPerPage int`, `ListMaxShowAll int`
- `ListSelectRelated []string`, `ListPrefetchRelated []string`

### Forms & Fieldsets
- `Fields []string`, `Exclude []string`, `ReadOnlyFields []string`
- `AutocompleteFields []string`, `RawIDFields []string`
- `Fieldsets []admin.Fieldset[T]` — `{Name, Fields, Collapsed, Description}` (or `admin.NewFieldset[T]("Name", "f1", "f2").WithCollapsed(true)`)
- `InlineRelations map[string]admin.InlineRelationConfig` (note: inline related rows are not covered by Playwright E2E per `docs-site/docs/status.md`)

## Custom Actions

```go
publishAction := admin.NewAction[blog.Article](
	"publish",
	"Publish selected articles",
	func(ctx context.Context, instances []*blog.Article) error {
		for _, a := range instances {
			a.Published = true
			if err := blog.ArticleObjects.Update(ctx, a); err != nil {
				return err
			}
		}
		return nil
	},
)
```

`ExecuteAction` checks per-object `HasChangePermission` and returns `201` when all succeed or `207` with `BulkActionResponse.Errors` (`invalid_id`, `not_found`, `permission_denied`) on partial success.

## Per-Object Permissions & Hooks

```go
cfg := &admin.Config[orders.Order]{
	HasAddPermission: func(ctx context.Context, a *admin.Admin[orders.Order], user interface{}) bool {
		return true
	},
	HasChangePermission: func(ctx context.Context, a *admin.Admin[orders.Order], user interface{}, obj *orders.Order) bool {
		return obj == nil || obj.Status != "finalized"
	},
	HasDeletePermission: func(ctx context.Context, a *admin.Admin[orders.Order], user interface{}, obj *orders.Order) bool {
		return obj == nil || obj.Status == "draft"
	},
	GetQueryset: func(ctx context.Context, a *admin.Admin[orders.Order], qs orm.QuerySet[orders.Order]) (orm.QuerySet[orders.Order], error) {
		return qs, nil
	},
}
```

## Custom Login Authenticator

Override default `FORGE_ADMIN_USERNAME` / `FORGE_ADMIN_PASSWORD` env auth with `adminSite.SetLoginAuthenticator(...)` (implementing `rest.LoginAuthenticator`).
