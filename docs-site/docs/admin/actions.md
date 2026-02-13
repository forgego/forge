---
sidebar_position: 3
description: Admin actions and bulk operations.
image: /forge-social-card.svg
---

# Admin Actions

Define bulk actions that appear in list views.

## Built-in actions

- Export CSV
- Custom action handlers

## Custom action

```go
Actions: []admin.Action{
    admin.NewAction("activate", "Activate", func(ctx context.Context, items []*Post) error {
        return nil
    }),
}
```

## Next steps

- [Admin Filters](/docs/admin/filters/)
- [Admin Config](/docs/admin/config/)
