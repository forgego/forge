---
sidebar_position: 2
---

# Admin Interface Tutorial

Build a CRUD admin UI in a few steps.

## 1. Define a model

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.Required(), schema.MaxLength(200)),
        schema.TextField("content", schema.Required()),
        schema.BoolField("published", schema.Default(false)),
    }
}
```

## 2. Register the model with admin

```go
package blog

import admincore "github.com/forgego/forge/admin/core"

func init() {
    admincore.Register(&admincore.Config[Post]{
        ListDisplay: []string{"id", "title", "published"},
        SearchFields: []string{"title"},
    })
}
```

## 3. Generate and migrate

```bash
forge generate
forge makemigrations admin --auto
forge migrate
```

## 4. Mount admin routes

```go
adminRouter := adminhttp.NewRouter()
router.Mount("/admin", adminRouter)
```

## 5. Run

```bash
forge runserver
```

Open `/admin` and verify the CRUD screens.

## Next steps

- [Admin guide](/docs/guides/admin)
- [Models guide](/docs/guides/models)
