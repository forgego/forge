---
sidebar_position: 3
description: Build your first REST API with a viewset.
keywords:
  - forge first api
  - forge rest api
image: /forge-social-card.svg
---

# First API

Create a small REST API with a model, serializer, and viewset.

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
    }
}
```

## 2. Serializer

```go
type PostSerializer struct {
    *api.Serializer
}

func NewPostSerializer() *PostSerializer {
    s := &PostSerializer{Serializer: api.NewSerializer()}
    s.AddField("id", api.IntegerField())
    s.AddField("title", api.StringField())
    s.AddField("content", api.StringField())
    return s
}
```

## 3. ViewSet

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

## 4. Routes

```go
apiRouter := apihttp.NewRouter()
apiRouter.Register("posts", NewPostViewSet())
router.Mount("/api", apiRouter)
```

## 5. Generate and run

```bash
forge generate
forge makemigrations posts --auto
forge migrate
forge runserver
```

## Next steps

- [REST API guide](/docs/guides/rest-api)
- [Models guide](/docs/guides/models)
