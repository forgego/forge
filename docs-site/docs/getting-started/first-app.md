---
sidebar_position: 5
---

# Building Your First Application

This guide walks you through building a complete blog application with forge. You'll learn how to create models, build APIs, customize the admin interface, and more.

## Project Overview

We'll build a blog application with:

- **Post model** - Blog posts with title, content, and publishing status
- **Author model** - Blog authors
- **Category model** - Post categories
- **Admin interface** - Auto-generated admin for managing posts
- **REST API** - API endpoints for frontend consumption

## Step 1: Project Setup

Create a new project:

```bash
forge new myblog
cd myblog
```

Configure your database in `config/config.yaml`:

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: myblog_db
  sslmode: disable
```

## Step 2: Define Models

Create `models/author.go`:

```go
package models

import (
    "github.com/forgego/forge/schema"
)

type Author struct {
    schema.BaseSchema
}

func (Author) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("name", schema.Required(), schema.MaxLength(100)),
        schema.StringField("email", schema.Required(), schema.Unique(), schema.MaxLength(255)),
        schema.TextField("bio"),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}

func (Author) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "authors",
        VerboseName:      "Author",
        VerboseNamePlural: "Authors",
    }
}

func (Author) Relations() []schema.Relation {
    return []schema.Relation{}
}

func (Author) Hooks() *schema.ModelHooks {
    return nil
}
```

Create `models/category.go`:

```go
package models

import (
    "github.com/forgego/forge/schema"
)

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

Create `models/post.go`:

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

func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "posts",
        VerboseName:      "Post",
        VerboseNamePlural: "Posts",
        OrderBy:          []string{"-created_at"},
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("author", "Author",
            schema.OnDelete(schema.CascadeCASCADE),
            schema.RelatedName("posts"),
        ),
        schema.ManyToManyField("categories", "Category",
            schema.RelatedName("posts"),
        ),
    }
}

func (Post) Hooks() *schema.ModelHooks {
    return nil
}
```

## Step 3: Generate Code

Generate type-safe code:

```bash
forge generate
```

## Step 4: Set Up Main Application

Update `main.go`:

```go
package main

import (
    "fmt"
    stdlog "log"
    "net/http"

    "github.com/forgego/forge/admin"
    "github.com/forgego/forge/config"
    "github.com/forgego/forge/db"
    forgelog "github.com/forgego/forge/log"
    "github.com/forgego/forge/server"
    "myblog/models"
)

func main() {
    cfg := config.NewConfig()
    settings := config.LoadSettings(cfg)

    logger, err := forgelog.NewLogger(settings.App.Debug)
    if err != nil {
        stdlog.Fatal(err)
    }
    defer logger.Sync()

    database, err := db.NewDBFromConfig(cfg)
    if err != nil {
        stdlog.Fatal(err)
    }
    defer database.Close()

    // Wire ORM managers
    models.AuthorObjects.SetDB(database)
    models.CategoryObjects.SetDB(database)
    models.PostObjects.SetDB(database)

    adminSite := admin.DefaultSite
    uiConfig := adminSite.GetUIConfig()
    uiConfig.Prefix = settings.Admin.Path
    adminSite.WithUIConfig(uiConfig)
    adminSite.SetDB(database)

    if _, err := admin.Register(&admin.Config[models.Author]{}); err != nil {
        stdlog.Fatal(err)
    }
    if _, err := admin.Register(&admin.Config[models.Category]{}); err != nil {
        stdlog.Fatal(err)
    }
    if _, err := admin.Register(&admin.Config[models.Post]{}); err != nil {
        stdlog.Fatal(err)
    }

    srv, err := server.NewServer(cfg, settings, logger)
    if err != nil {
        stdlog.Fatal(err)
    }

    srv.RegisterRoutes(func(router *server.Router) {
        router.Get("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Welcome to MyBlog!")
        })

        if settings.Admin.Enabled {
            router.Mount(settings.Admin.Path, adminSite.Handler())
        }
    })

    fmt.Printf("Starting server on %s:%s\n", settings.Server.Host, settings.Server.Port)
    if err := srv.Start(); err != nil {
        stdlog.Fatal(err)
    }
}
```

## Step 5: Run Migrations

```bash
forge makemigrations
forge migrate up
```

## Step 6: Start the Server

```bash
forge runserver
```

Visit `http://localhost:8000/admin/` and create some authors, categories, and posts.

## Step 7: Add Views

Create `views/posts.go`:

```go
package views

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"
    "myblog/models"
    "github.com/go-chi/chi/v5"
)

func ListPosts(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    posts, err := models.PostObjects.
        Filter(models.PostFieldsInstance.Published.Equals(true)).
        OrderBy("-created_at").
        All(ctx)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // Extract ID from URL (you'll need to implement URL parameter extraction)
    // For example, using chi router: chi.URLParam(r, "id")
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
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

Register routes in `main.go`:

```go
server.RegisterRoutes(func(router *httplib.Router) {
    router.Get("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Welcome to MyBlog!")
    })
    
    router.Get("/api/posts", views.ListPosts)
    router.Get("/api/posts/\\{id\\}", views.GetPost)
    
    if settings.Admin.Enabled {
        admin.RegisterAdminRoutes(router, settings.Admin.Path)
    }
})
```

## Step 8: Use the ORM

Here are some common patterns:

```go
// Get all published posts
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)

// Get posts by author
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Author.Equals(authorID)).
    All(ctx)

// Get posts in category
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Categories.Contains(categoryID)).
    All(ctx)

// Search posts
posts, err := models.PostObjects.
    Filter(models.PostFieldsInstance.Title.Contains("django")).
    All(ctx)

// Create a post
post := &models.Post{
    Title:     "My Post",
    Content:   "Content here",
    Slug:      "my-post",
    Author:    author,
    Published: true,
}
err := models.PostObjects.Create(ctx, post)
```

## Next Steps

You now have a working blog application! Next, you can:

- [Learn about Models](/docs/guides/models) - Deep dive into model definitions
- [Explore Queries](/docs/guides/queries) - Advanced querying techniques
- [Customize Admin](/docs/guides/admin) - Customize the admin interface
- [Build REST APIs](/docs/guides/rest-api) - Use the REST API system
- [Check Examples](/docs/examples/blog) - See more complete examples

