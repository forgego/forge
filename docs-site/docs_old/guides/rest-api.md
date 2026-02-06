---
sidebar_position: 4
description: Build REST APIs with viewsets and serializers.
keywords:
  - forge rest api
  - viewsets
  - serializers
image: /forge-social-card.svg
---

# REST API Guide

Build REST APIs with viewsets and serializers.

## 1. Create a serializer

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

## 2. Create a viewset

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

## 3. Register routes

```go
apiRouter := apihttp.NewRouter()
apiRouter.Register("posts", NewPostViewSet())
router.Mount("/api", apiRouter)
```

## Next steps

- [Authentication guide](/docs/guides/authentication)
- [Validation guide](/docs/guides/validation)
