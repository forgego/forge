---
sidebar_position: 4
description: Configure typed admin list filters and saved views.
image: /social-card.png
---

# Admin List Filters & Saved Views

`ListFilter` accepts `[]admin.Field`, not field name strings. Prefer the generated field references for your model; `admin.Computed` also supplies a field path.

```go
_, err := admin.RegisterWithSite(site, &admin.Config[models.Order]{
    ListFilter: []admin.Field{
        admin.Computed("status"),
        admin.Computed("created_at"),
    },
})
```

The UI builds filters from the registered model metadata. Saved views can store a user's search, filter, and sorting choices. The server exposes `GET/POST /admin/api/saved-views/{model}` and `DELETE /admin/api/saved-views/{model}/{id}` when mounted under `/admin`.
