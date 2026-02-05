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
    "github.com/forgego/forge/pkg/schema"
)

type Author struct {
    schema.BaseSchema
}

func (Author) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().MaxLength(100).Build(),
        schema.String("email").Unique().Required().MaxLength(255).Build(),
        schema.Text("bio").Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
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
    "github.com/forgego/forge/pkg/schema"
)

type Category struct {
    schema.BaseSchema
}

func (Category) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().Unique().MaxLength(100).Build(),
        schema.String("slug").Required().Unique().MaxLength(100).Build(),
        schema.Text("description").Build(),
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
    "github.com/forgego/forge/pkg/schema"
    "github.com/forgego/forge/pkg/schema/relations"
)

type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("title").Required().MaxLength(200).Build(),
        schema.String("slug").Unique().MaxLength(200).Build(),
        schema.Text("content").Required().Build(),
        schema.Text("excerpt").MaxLength(500).Build(),
        schema.Bool("published").Default(false).Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
        schema.Time("updated_at").AutoNow().Build(),
        schema.Time("published_at").Build(),
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
        relations.ForeignKey("author", "Author").
            Required().
            OnDelete(schema.Cascade).
            RelatedName("posts"),
        relations.ManyToMany("categories", "Category").
            RelatedName("posts"),
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
    "log"
    "net/http"

    "github.com/forgego/forge/pkg/admin"
    "github.com/forgego/forge/pkg/config"
    "github.com/forgego/forge/pkg/db"
    "github.com/forgego/forge/pkg/logging"
    httplib "github.com/forgego/forge/pkg/http"
    "github.com/forgego/forge/pkg/registry"
    "myblog/models"
)

func main() {
    cfg := config.NewConfig()
    settings := config.LoadSettings(cfg)

    logger, err := logging.NewLogger(cfg.IsDevelopment())
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    database, err := db.NewDBFromConfig(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()

    // Register all models
    registry.RegisterModel(&models.Author{})
    registry.RegisterModel(&models.Category{})
    registry.RegisterModel(&models.Post{})

    // Register admin models
    admin.RegisterModel(&models.Author{})
    admin.RegisterModel(&models.Category{})
    admin.RegisterModel(&models.Post{})

    server, err := httplib.NewServer(cfg, settings, logger)
    if err != nil {
        log.Fatal(err)
    }

    server.RegisterRoutes(func(router *httplib.Router) {
        router.Get("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Welcome to MyBlog!")
        })

        if settings.Admin.Enabled {
            admin.RegisterAdminRoutes(router, settings.Admin.Path)
        }
    })

    fmt.Printf("Starting server on %s:%s\n", settings.Server.Host, settings.Server.Port)
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Step 5: Run Migrations

```bash
forge makemigrations
forge migrate
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
    
    posts, err := models.Post.Objects.
        Filter(models.Post.Fields.Published.Equals(true)).
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
    
    post, err := models.Post.Objects.
        Filter(models.Post.Fields.ID.Equals(id)).
        Filter(models.Post.Fields.Published.Equals(true)).
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
posts, err := models.Post.Objects.
    Filter(models.Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)

// Get posts by author
posts, err := models.Post.Objects.
    Filter(models.Post.Fields.Author.Equals(authorID)).
    All(ctx)

// Get posts in category
posts, err := models.Post.Objects.
    Filter(models.Post.Fields.Categories.Contains(categoryID)).
    All(ctx)

// Search posts
posts, err := models.Post.Objects.
    Filter(models.Post.Fields.Title.Contains("django")).
    All(ctx)

// Create a post
post := &models.Post{
    Title:     "My Post",
    Content:   "Content here",
    Slug:      "my-post",
    Author:    author,
    Published: true,
}
err := models.Post.Objects.Create(ctx, post)
```

## Next Steps

You now have a working blog application! Next, you can:

- [Learn about Models](/docs/guides/models) - Deep dive into model definitions
- [Explore Queries](/docs/guides/queries) - Advanced querying techniques
- [Customize Admin](/docs/guides/admin) - Customize the admin interface
- [Build REST APIs](/docs/guides/rest-api) - Use the REST API system
- [Check Examples](/docs/examples/blog) - See more complete examples

