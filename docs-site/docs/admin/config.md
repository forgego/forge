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
    VerboseName: "Post",
    ListDisplay: []string{"id", "title", "published"},
    SearchFields: []string{"title"},
    Filters: []string{"published"},
    Ordering: []admin.Ordering{admin.Desc("created_at")},
    ListPerPage: 25,
})
```

## Common config fields

- VerboseName, VerboseNamePlural
- ListDisplay, SearchFields, Filters
- Ordering, ListPerPage
- Actions
- UI overrides

## Next steps

- [Actions](/docs/admin/actions/)
- [UI](/docs/admin/ui/)
