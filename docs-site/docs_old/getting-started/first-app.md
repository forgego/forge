---
sidebar_position: 4
description: Build a small app end-to-end.
keywords:
  - forge first app
  - forge tutorial
image: /forge-social-card.svg
---

# First App

Build a small app with models, admin, and API.

## 1. Create the project

```bash
forge new blog
cd blog
```

## 2. Model

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

## 3. Admin

```go
func init() {
    admincore.Register(&admincore.Config[Post]{
        ListDisplay: []string{"id", "title", "published"},
        SearchFields: []string{"title"},
    })
}
```

## 4. API

```go
type PostViewSet struct {
    *api.BaseViewSet
}

func NewPostViewSet() *PostViewSet {
    return &PostViewSet{
        BaseViewSet: api.NewBaseViewSet(
            func() api.Serializer { return NewPostSerializer() },
            Post.Objects,
            &Post{},
        ),
    }
}
```

## 5. Generate, migrate, run

```bash
forge generate
forge makemigrations blog --auto
forge migrate
forge runserver
```

## Next steps

- [Admin guide](/docs/guides/admin)
- [REST API guide](/docs/guides/rest-api)
