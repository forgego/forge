---
sidebar_position: 2
description: Admin model configuration and registration.
image: /forge-social-card.svg
---

# Admin Config

Configure how models are displayed and edited in the admin UI.

## Register a model

```go
admin.Register(Post{}, admin.ModelConfig{
    ListDisplay: []string{"id", "title", "published"},
    SearchFields: []string{"title"},
    Filters: []string{"published"},
})
```

## Common options

- ListDisplay, SearchFields, Filters
- Ordering, ListPerPage
- Actions (bulk actions)
- Field widgets and overrides

## Next steps

- [Actions](/docs/admin/actions/)
- [UI](/docs/admin/ui/)
