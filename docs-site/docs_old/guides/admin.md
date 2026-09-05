---
sidebar_position: 3
description: Configure the forge admin UI for your models.
keywords:
  - forge admin
  - admin ui
  - crud
image: /forge-social-card.svg
---

# Admin Guide

Configure the auto-generated admin UI for your models.

## 1. Register a model

```go
func init() {
    admincore.Register(&admincore.Config[Post]{
        ListDisplay: []string{"id", "title", "published"},
        SearchFields: []string{"title"},
    })
}
```

## 2. Mount admin routes

```go
adminRouter := adminhttp.NewRouter()
router.Mount("/admin", adminRouter)
```

## 3. Run

```bash
forge runserver
```

Open `/admin` and confirm CRUD screens.

## Next steps

- [Models guide](/docs/guides/models)
- [REST API guide](/docs/guides/rest-api)
