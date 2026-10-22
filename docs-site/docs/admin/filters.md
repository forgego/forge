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

Boolean filters offer All, Yes, and No. Date and datetime filters use From and To date inputs; there are no date-range presets. Numeric filters use Min and Max inputs. Choice filters show the configured choices without per-choice matching counts. Relation filters currently use a plain text input on the list page; that input does not offer debounced related-object search.
