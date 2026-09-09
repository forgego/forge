---
sidebar_position: 2
description: Declarative Admin ModelConfig, ListDisplay, search, filters, fieldsets, and permissions.
image: /forge-social-card.svg
---

# Admin Model Configuration

The `admin.ModelConfig[T]` struct defines how a model is presented, filtered, edited, and audited in the React Admin SPA.

---

## Full Configuration Example

```go
package catalog

import (
    "context"
    "github.com/forgego/forge/admin"
    "myapp/models"
)

func RegisterProductAdmin(site *admin.Site) {
    admin.Register[models.Product](site, admin.ModelConfig[models.Product]{
        // Human-friendly labels
        VerboseName:       "Product",
        VerboseNamePlural: "Products",

        // List View Table Columns
        ListDisplay: []string{
            "SKU",
            "Name",
            "Category",
            "Price",
            "PriceWithTax",
            "InStock",
            "CreatedAt",
        },

        // Searchable fields (debounced ILIKE search)
        SearchFields: []string{"SKU", "Name", "Description"},

        // Sidebar faceted filters
        ListFilter: []string{
            "InStock",
            "Category",
            "Status",
            "CreatedAt",
        },

        // Default sorting
        Ordering: []string{"-created_at"},

        // Pagination size
        ListPerPage: 25,

        // Non-editable fields in form view
        ReadonlyFields: []string{"PriceWithTax", "CreatedAt", "UpdatedAt"},

        // Granular RBAC Permissions
        HasViewPermission: func(ctx context.Context, u *admin.User) bool {
            return true
        },
        HasChangePermission: func(ctx context.Context, u *admin.User) bool {
            return u.HasRole("admin", "manager")
        },
        HasDeletePermission: func(ctx context.Context, u *admin.User) bool {
            return u.HasRole("admin")
        },

        // Custom Bulk Actions
        Actions: []admin.Action{
            {
                ID:             "restock",
                Label:          "Restock (+50 Units)",
                ConfirmMessage: "Are you sure you want to add 50 units to the selected products?",
                Handler: func(ctx context.Context, ids []any) error {
                    return models.ProductManager.Restock(ctx, ids, 50)
                },
            },
        },
    })
}
```

---

## Configuration Options Reference

| Option | Type | Description |
| :--- | :--- | :--- |
| `VerboseName` | `string` | Singular human-readable name shown in breadcrumbs and titles. |
| `VerboseNamePlural` | `string` | Plural name shown in the sidebar navigation. |
| `ListDisplay` | `[]string` | Struct field names rendered as columns in the data table. |
| `SearchFields` | `[]string` | Fields queried when typing into the global table search box. |
| `ListFilter` | `[]string` | Fields rendered as filter chips or faceted dropdowns. |
| `Ordering` | `[]string` | Default SQL `ORDER BY` fields. Prefix with `-` for descending. |
| `ListPerPage` | `int` | Number of items per page (defaults to 25). |
| `ReadonlyFields` | `[]string` | Fields visible but non-editable in the detail/edit modal. |
| `Actions` | `[]admin.Action` | Custom batch actions shown when items are checked. |
| `Widgets` | `[]admin.Widget` | Dashboard widgets pinned to this model or main dashboard. |

---

## Next Steps

- **[Custom Actions](/docs/admin/actions/)**: Build batch processing actions.
- **[Faceted Filters](/docs/admin/filters/)**: Configure smart filter chips and saved views.
- **[Dashboard Widgets](/docs/admin/ui/)**: Attach charts and KPI metrics to models.

