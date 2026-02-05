---
sidebar_position: 1
---

# Blog Example

A complete blog application demonstrating forge features.

## Overview

This example shows how to build a blog with:

- Post model with author relationship
- Category model with many-to-many relationship
- Admin interface for content management
- REST API for frontend consumption

## Models

### Post Model

```go
package models

import (
    "github.com/forgego/forge/schema"
)

type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.Required(), schema.MaxLength(200)),
        schema.StringField("slug", schema.Unique(), schema.MaxLength(200)),
        schema.TextField("content", schema.Required()),
        schema.TextField("excerpt", schema.MaxLength(500)),
        schema.BoolField("published", schema.Default(false)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
        schema.TimeField("updated_at", schema.AutoNow()),
        schema.TimeField("published_at"),
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("author", "User",
            schema.OnDelete(schema.CascadeCASCADE),
            schema.RelatedName("posts"),
        ),
        schema.ManyToManyField("categories", "Category",
            schema.RelatedName("posts"),
        ),
    }
}

func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "posts",
        VerboseName:      "Post",
        VerboseNamePlural: "Posts",
        OrderBy:          []string{"-created_at"},
    }
}

func (Post) Hooks() *schema.ModelHooks {
    return nil
}
```

### Category Model

```go
type Category struct {
    schema.BaseSchema
}

func (Category) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("name", schema.Required(), schema.Unique(), schema.MaxLength(100)),
        schema.StringField("slug", schema.Required(), schema.Unique(), schema.MaxLength(100)),
        schema.TextField("description"),
    }
}

func (Category) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "categories",
        VerboseName:      "Category",
        VerboseNamePlural: "Categories",
    }
}

func (Category) Relations() []schema.Relation {
    return []schema.Relation{}
}

func (Category) Hooks() *schema.ModelHooks {
    return nil
}
```

## Views

### List Posts

```go
package views

import (
    "context"
    "encoding/json"
    "net/http"
    "myblog/models"
)

func ListPosts(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    posts, err := models.PostObjects.
        Filter(models.PostFieldsInstance.Published.Equals(true)).
        OrderBy("-created_at").
        PrefetchRelated("author", "categories").
        All(ctx)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
```

### Get Post

```go
func GetPost(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // Extract ID from URL
    id := extractID(r.URL.Path)
    
    post, err := models.PostObjects.
        Filter(models.PostFieldsInstance.ID.Equals(id)).
        Filter(models.PostFieldsInstance.Published.Equals(true)).
        SelectRelated("author").
        PrefetchRelated("categories").
        Get(ctx)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}
```

## Admin Setup

```go
admin.Register(&admin.Config[models.Post]{
    ListDisplay: []admin.Field{
        models.PostFieldsInstance.Title,
        models.PostFieldsInstance.Author,
        models.PostFieldsInstance.Published,
        models.PostFieldsInstance.CreatedAt,
    },
    SearchFields: []admin.Field{
        models.PostFieldsInstance.Title,
        models.PostFieldsInstance.Content,
    },
    ListFilter: []admin.Field{
        models.PostFieldsInstance.Published,
        models.PostFieldsInstance.Author,
        models.PostFieldsInstance.CreatedAt,
    },
    DateHierarchy: "created_at",
})

admin.Register(&admin.Config[models.Category]{
    ListDisplay: []admin.Field{
        models.CategoryFieldsInstance.Name,
        models.CategoryFieldsInstance.Slug,
    },
    SearchFields: []admin.Field{
        models.CategoryFieldsInstance.Name,
    },
})
```

## REST API

```go
func RegisterPostViewSet(router *api.Router) {
    viewset := api.NewBaseViewSet(
        func() api.Serializer {
            return NewPostSerializer()
        },
        models.PostObjects.Filter(models.PostFieldsInstance.Published.Equals(true)),
        &models.Post{},
    )
    
    router.Register("posts", viewset)
}
```

## Usage

### Create a Post

```go
post := &models.Post{
    Title:     "My First Post",
    Slug:      "my-first-post",
    Content:   "This is the content...",
    Published: true,
    Author:    author,
}
err := models.PostObjects.Create(ctx, post)
```

### Get Published Posts

```go
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

### Filter by Category

```go
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Categories.Contains(categoryID)).
    All(ctx)
```

## See Also

- [Models Guide](/docs/guides/models) - Learn about models
- [Queries Guide](/docs/guides/queries) - Query examples
- [Admin Guide](/docs/guides/admin) - Admin customization

