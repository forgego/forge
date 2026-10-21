---
sidebar_position: 3
description: Define typed bulk actions for Forge admin models.
image: /social-card.png
---

# Admin Actions & Batch Operations

Add `admin.Action[T]` values to `admin.Config[T].Actions`. The action handler receives the selected model instances after the admin loads them and checks permissions.

```go
action := admin.Action[models.Order]{
    Name: "mark_delivered",
    Label: "Mark as delivered",
    Icon: "truck",
    Confirmation: "Mark selected orders as delivered?",
    Permissions: []string{"orders.change"},
    Handler: func(ctx context.Context, orders []*models.Order) error {
        for _, order := range orders {
            // Update and persist each order using your application service.
            _ = order
        }
        return nil
    },
}
```

| Field | Type | Purpose |
| :--- | :--- | :--- |
| `Name` | `string` | Unique action name sent to the API |
| `Label` | `string` | Display label |
| `Icon` | `string` | Optional icon name |
| `Confirmation` | `string` | Confirmation message |
| `Handler` | `func(context.Context, []*T) error` | Processes permitted selected instances |
| `Permissions` | `[]string` | Permissions checked before calling the handler |

`admin.NewAction[T](name, label, handler)` is a shorthand. Its `WithPermissions` and `WithConfirmation` methods set the corresponding fields. The API reports records it skipped because they were missing or denied.

On the list page, row selections remain selected when you move between pages. Select all applies to the current page; clear selection removes all selected rows. Delete shows the current selection count in its confirmation dialog. While a bulk action runs, its button is disabled and shows a spinner. A successful action shows a success toast and clears the selection. A failed request shows an error toast and keeps the selected rows for retry.

The CSV and JSON export controls export rows matching the current search, filters, and ordering. They do not export only the selected rows. CSV uses the model's configured list columns, excluding server-hidden fields; client-side column visibility does not change the export. The server fetches all matching pages before writing the CSV response, so the response is not streamed row by row.
