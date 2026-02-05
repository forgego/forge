---
sidebar_position: 3
---

# Your First API

Build a REST API with forge in 10 minutes.

## What you'll build

A simple blog API with:
- A Post model
- CRUD endpoints (Create, Read, Update, Delete)
- JSON responses

## Step 1: Create a model

Create `app/blog/models.go`:

```go
package blog

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
        schema.TextField("content", schema.Required()),
        schema.BoolField("published", schema.Default(false)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
        schema.TimeField("updated_at", schema.AutoNow()),
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
    return []schema.Relation{}
}

func (Post) Hooks() *schema.ModelHooks {
    return nil
}
```

## Step 2: Generate code

```bash
forge generate
```

This creates:
- `app/blog/post.gen.go` - Generated model struct
- `app/blog/post_fields.gen.go` - Type-safe field expressions
- `app/blog/post_manager.gen.go` - Manager with CRUD operations
- `app/blog/post_queryset.gen.go` - QuerySet for filtering

## Step 3: Run migrations

```bash
forge makemigrations
forge migrate up
```

## Step 4: Create API handlers

Create `app/blog/api.go`:

```go
package blog

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
)

func RegisterPostRoutes(router chi.Router) {
    router.Route("/api/posts", func(r chi.Router) {
        r.Get("/", listPosts)
        r.Post("/", createPost)
        r.Get("/{id}", getPost)
        r.Put("/{id}", updatePost)
        r.Delete("/{id}", deletePost)
    })
}

func listPosts(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    posts, err := PostObjects.
        Filter(PostFieldsInstance.Published.Equals(true)).
        OrderBy("-created_at").
        All(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func createPost(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var post Post
    if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := PostObjects.Create(ctx, &post); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(post)
}

func getPost(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    post, err := PostObjects.Get(ctx, id)
    if err != nil {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

func updatePost(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    post, err := PostObjects.Get(ctx, id)
    if err != nil {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := PostObjects.Update(ctx, post); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    post, err := PostObjects.Get(ctx, id)
    if err != nil {
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    if err := PostObjects.Delete(ctx, post); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

## Step 5: Register routes

Update `cmd/server/main.go`:

```go
import (
    "your-module/app/blog"
)

func main() {
    // ... your setup code ...

    server.RegisterRoutes(func(router *server.Router) {
        blog.RegisterPostRoutes(router)
    })

    // ... start server ...
}
```

## Step 6: Test your API

Start the server:

```bash
forge runserver
```

### Create a post

```bash
curl -X POST http://localhost:8000/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "This is my first blog post!",
    "published": true
  }'
```

### List posts

```bash
curl http://localhost:8000/api/posts
```

### Get a post

```bash
curl http://localhost:8000/api/posts/1
```

### Update a post

```bash
curl -X PUT http://localhost:8000/api/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "content": "Updated content",
    "published": true
  }'
```

### Delete a post

```bash
curl -X DELETE http://localhost:8000/api/posts/1
```

## What you learned

- How to define models
- How to generate code
- How to run migrations
- How to create API endpoints
- How to use the ORM

## Next steps

- **[Use ViewSets](/docs/guides/rest-api)** - Use forge's ViewSet system for cleaner APIs
- **[Add authentication](/docs/guides/authentication)** - Secure your API
- **[Add validation](/docs/guides/validation)** - Validate request data
- **[Explore examples](/docs/examples/blog)** - See complete examples
