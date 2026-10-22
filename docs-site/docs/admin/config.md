---
sidebar_position: 2
description: Configure Forge admin models with typed fields, permissions, and actions.
image: /social-card.png
---

# Admin Model Configuration

`admin.Config[T]` controls a registered model. Pass a pointer to `admin.Register` for the default site or `admin.RegisterWithSite` for a specific site.

```go
_, err := admin.RegisterWithSite(site, &admin.Config[models.Product]{
    VerboseName: "Product",
    ListDisplay: []admin.Field{admin.Computed("name"), admin.Computed("price")},
    SearchFields: []admin.Field{admin.Computed("name")},
    ListFilter: []admin.Field{admin.Computed("active")},
    Ordering: []admin.Field{admin.Computed("-created_at")},
    ListPerPage: 25,
    ReadOnlyFields: []string{"created_at"},
    HasChangePermission: func(ctx context.Context, modelAdmin *admin.Admin[models.Product], user interface{}, obj *models.Product) bool {
        return user != nil && obj != nil
    },
    Actions: []admin.Action[models.Product]{
        {
            Name: "restock",
            Label: "Restock",
            Confirmation: "Restock selected products?",
            Handler: func(ctx context.Context, products []*models.Product) error {
                // Update selected model instances here.
                return nil
            },
        },
    },
})
```

Use generated field references such as `models.ProductFieldsInstance.Name` when available. `admin.Computed("field_path")` provides an `admin.Field` for a named path. A leading `-` in an ordering path requests descending order.

| Option | Type | Purpose |
| :--- | :--- | :--- |
| `ListDisplay` | `[]admin.Field` | Columns |
| `SearchFields` | `[]admin.Field` | Search paths |
| `ListFilter` | `[]admin.Field` | Filter paths |
| `Ordering` | `[]admin.Field` | Default sort paths |
| `ListPerPage` | `int` | Page size |
| `ReadOnlyFields` | `[]string` | Non-editable form fields |
| `Actions` | `[]admin.Action[T]` | Bulk actions |

`HasAddPermission` takes `(context.Context, *admin.Admin[T], interface{})`. `HasViewPermission`, `HasChangePermission`, and `HasDeletePermission` additionally take `*T`. Dashboard layout is configured separately with `core.SetDashboard`.
