---
sidebar_position: 3
description: Admin actions and bulk operations.
image: /forge-social-card.svg
---

# Admin Actions

Define bulk actions that appear in list views.

## Example

```go
Actions: []admin.Action{
    actions.ExportCSV(),
}
```

You can also register custom actions with a handler function.

## Next steps

- [Admin Filters](/docs/admin/filters/)
- [Admin Config](/docs/admin/config/)
