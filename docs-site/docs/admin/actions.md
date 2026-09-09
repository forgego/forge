---
sidebar_position: 3
description: Custom batch operations, bulk actions, confirmation dialogs, and data export.
image: /forge-social-card.svg
---

# Admin Actions & Batch Operations

Bulk actions allow administrators to select multiple records across table pages and apply batch workflows with confirmation dialogs, progress indicators, and toast feedback.

---

## Defining Custom Actions

Add `admin.Action` structs to your model's `ModelConfig.Actions` slice:

```go
package orders

import (
    "context"
    "fmt"
    "github.com/forgego/forge/admin"
    "myapp/models"
)

var DeliverOrdersAction = admin.Action{
    ID:             "mark_delivered",
    Label:          "Mark As Delivered",
    Icon:           "truck", // Lucide icon name
    ConfirmMessage: "Are you sure you want to mark the selected orders as delivered?",
    Handler: func(ctx context.Context, ids []any) error {
        // Execute batch update in a single transaction
        affected, err := models.OrderManager.
            Filter(models.OrderExpr.ID.In(ids...)).
            Filter(models.OrderExpr.Status.Eq("shipped")).
            Update(ctx, orm.UpdateMap{
                "status":       "delivered",
                "delivered_at": time.Now(),
            })
        if err != nil {
            return fmt.Errorf("failed to update orders: %w", err)
        }
        return nil
    },
}
```

---

## Action Anatomy

| Field | Type | Description |
| :--- | :--- | :--- |
| `ID` | `string` | Unique action slug sent in the API request payload. |
| `Label` | `string` | Human-readable button label in the action dropdown. |
| `Icon` | `string` | Optional Lucide icon identifier (e.g. `check-circle`, `truck`, `archive`). |
| `ConfirmMessage` | `string` | If set, triggers a Radix UI confirmation modal before invoking the handler. |
| `Handler` | `func(ctx context.Context, ids []any) error` | Backend function that receives selected record primary keys. |
| `RequiresPermission` | `string` | Optional role or permission requirement for invoking this action. |

---

## Built-in Actions

Forge provides several built-in actions automatically on all registered models:

1. **Delete Selected**: Prompts with count of selected rows and cascades deletion safely.
2. **Export to CSV**: Generates a streamed CSV file respecting current filters and column visibility.
3. **Export to JSON**: Dumps formatted JSON records matching the current selection.

---

## User Experience & Feedback

When an administrator triggers an action from the UI:
1. If `ConfirmMessage` is provided, a modal appears with "Cancel" and "Confirm" buttons.
2. During execution, the action button enters a disabled loading state with a spinner.
3. On completion, a success toast notification confirms the operation (e.g. *"Action executed successfully"*), and the data table reloads automatically to show updated states.
4. If the handler returns an error, an error toast displays the message without deselecting items.

---

## Next Steps

- **[Admin Filters](/docs/admin/filters/)**: Faceted filtering and saved views.
- **[Admin UI & Widgets](/docs/admin/ui/)**: Pinned KPI widgets and custom plugins.

