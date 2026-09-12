---
sidebar_position: 4
description: Admin faceted list filters, smart chips, date range pickers, and saved views.
image: /forge-social-card.svg
---

# Admin List Filters & Saved Views

The Admin UI features an interactive filtering panel on every model list view. Filters automatically adapt to field types declared in your schema, providing intuitive filter chips, date range selectors, and foreign key dropdowns.

---

## Configuring `ListFilter`

Specify field names in `admin.ModelConfig[T].ListFilter`:

```go
admin.Register[models.Order](site, admin.ModelConfig[models.Order]{
    ListFilter: []string{
        "Status",         // Choice field -> Multi-select dropdown
        "PaymentMethod",  // String/Choice -> Filter chips
        "CreatedAt",      // DateTime -> Date range calendar picker
        "IsPaid",         // Bool -> Tri-state toggle (All / Yes / No)
        "Customer",       // ForeignKey -> Searchable entity dropdown
    },
})
```

---

## Filter Widget Types

Forge inspects the underlying `schema.Field` to automatically render the appropriate UI widget:

| Field Schema Type | Rendered Widget in React Admin | Interaction |
| :--- | :--- | :--- |
| `TypeBool` | **Tri-State Segmented Control** | Select *All*, *Active (true)*, or *Inactive (false)*. |
| `TypeTime` / `TypeDate` | **Date Range Picker** | Presets (*Today*, *Last 7 Days*, *This Month*) or custom calendar date range. |
| `Choices` | **Faceted Badges / Checkboxes** | Counts of matching items next to each choice value. |
| `TypeForeignKey` | **Searchable Combobox** | Async debounced search matching related record names. |
| `TypeDecimal` / `TypeInt` | **Numeric Range Inputs** | Minimum and maximum numerical bounds. |

---

## Saved Views

Administrators can save frequently used combinations of search keywords,
filters, and sort order as **Saved Views**. Views are stored per user on the
server (`GET/POST /admin/api/saved-views/{model}`):

1. **Creating a view**: apply any combination of search, filters, and sorting,
   click **Save view**, and name it (e.g. *"Pending High-Value Orders"*).
   Saving a name that already exists updates that view (`200 OK`);
   new names return `201 Created`.
2. **Applying a view**: pick it from the **Saved views** dropdown; search,
   filters, and sorting are restored.
3. **Deleting a view**: with a view selected, click the trash button
   (`DELETE /admin/api/saved-views/{model}/{id}` → `204`).

---

## Next Steps

- **[UI Customization & Widgets](/docs/admin/ui/)**: Attach KPI charts and cards.
- **[Admin Actions](/docs/admin/actions/)**: Execute batch operations on filtered datasets.

