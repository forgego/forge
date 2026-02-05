---
sidebar_position: 1
description: Build a complete blog application with forge. Models, admin interface, REST API, and authentication - all in one tutorial.
keywords:
  - forge tutorial
  - blog tutorial
  - complete application
  - django go tutorial
  - web app tutorial
image: /forge-social-card.svg
---

# Build a Complete Blog Application

Let's build a fully functional blog with forge. This tutorial covers everything from models to deployment.

## What We'll Build

A blog with:
- Posts, categories, and comments
- User authentication and author profiles
- Admin interface for content management
- Public REST API
- Search and filtering
- Pagination and sorting

## Prerequisites

- Go 1.25+ installed
- PostgreSQL running
- forge CLI installed (`go install github.com/forgego/forge/newforge/cli/cmd@latest`)

## Step 1: Create the Project

```bash
forge new blog
cd blog
```

This creates a new forge project with the basic structure:
```
blog/
├── main.go
├── go.mod
├── config/
│   └── config.yaml
├── models/
│   └── example.go
└── migrations/
```

## Step 2: Configure the Database

Edit `config/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  name: blog_db
  sslmode: disable

server:
  host: localhost
  port: 8000
  
secret_key: "your-secret-key-here"
```

## Step 3: Define the Models

Create `models/user.go`:

```go
package models

import (
    "github.com/forgego/forge/internal/schema"
    "time"
)

type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("username").Required().Unique().MaxLength(150).Build(),
        schema.String("email").Required().Unique().MaxLength(254).Build(),
        schema.String("first_name").MaxLength(30).Build(),
        schema.String("last_name").MaxLength(30).Build(),
        schema.Text("bio").Blank().Build(),
        schema.String("avatar").Blank().Build(),
        schema.Bool("is_active").Default(true).Build(),
        schema.Bool("is_staff").Default(false).Build(),
        schema.Time("date_joined").AutoNowAdd().Build(),
    }
}

func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",
        VerboseName: "User",
        VerboseNamePlural: "Users",
        Ordering: []string{"username"},
    }
}

func (User) Relations() []schema.Relation {
    return []schema.Relation{
        schema.HasMany("posts", "Post", "author_id"),
        schema.HasMany("comments", "Comment", "user_id"),
    }
}
```

Create `models/category.go`:

```go
package models

import (
    "github.com/forgego/forge/internal/schema"
    "time"
)

type Category struct {
    schema.BaseSchema
}

func (Category) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().Unique().MaxLength(100).Build(),
        schema.String("slug").Required().Unique().MaxLength(100).Build(),
        schema.Text("description").Blank().Build(),
        schema.String("color").MaxLength(7).Default("#007bff").Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
    }
}

func (Category) Meta() schema.Meta {
    return schema.Meta{
        TableName: "categories",
        VerboseName: "Category",
        VerboseNamePlural: "Categories",
        Ordering: []string{"name"},
    }
}

func (Category) Relations() []schema.Relation {
    return []schema.Relation{
        schema.HasMany("posts", "Post", "category_id"),
    }
}
```

Create `models/post.go`:

```go
package models

import (
    "github.com/forgego/forge/internal/schema"
    "time"
)

type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("title").Required().MaxLength(200).Build(),
        schema.String("slug").Required().Unique().MaxLength(200).Build(),
        schema.Text("content").Required().Build(),
        schema.Text("excerpt").Blank().Build(),
        schema.String("featured_image").Blank().Build(),
        schema.Bool("published").Default(false).Build(),
        schema.Bool("featured").Default(false).Build(),
        schema.Int("view_count").Default(0).Build(),
        schema.Time("published_at").Blank().Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
        schema.Time("updated_at").AutoNow().Build(),
        schema.Int64("author_id").Required().Build(),
        schema.Int64("category_id").Required().Build(),
    }
}

func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName: "posts",
        VerboseName: "Post",
        VerboseNamePlural: "Posts",
        Ordering: []string{"-created_at"},
        Indexes: []schema.Index{
            {Name: "idx_post_published", Fields: []string{"published"}},
            {Name: "idx_post_author", Fields: []string{"author_id"}},
            {Name: "idx_post_category", Fields: []string{"category_id"}},
        },
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("author_id", "User", "id"),
        schema.ForeignKey("category_id", "Category", "id"),
        schema.HasMany("comments", "Comment", "post_id"),
    }
}

func (Post) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeSave: func(ctx context.Context, model interface{}) error {
            post := model.(*Post)
            
            // Auto-generate slug from title if not set
            if post.Slug == "" {
                post.Slug = generateSlug(post.Title)
            }
            
            // Set published_at when publishing
            if post.Published && post.PublishedAt.IsZero() {
                post.PublishedAt = time.Now()
            }
            
            // Generate excerpt from content if not set
            if post.Excerpt == "" && len(post.Content) > 200 {
                post.Excerpt = post.Content[:200] + "..."
            }
            
            return nil
        },
    }
}

// Helper function for slug generation
func generateSlug(title string) string {
    // Simple slug generation - in production, use a proper library
    slug := strings.ToLower(title)
    slug = strings.ReplaceAll(slug, " ", "-")
    slug = strings.ReplaceAll(slug, "?", "")
    slug = strings.ReplaceAll(slug, "!", "")
    return slug
}
```

Create `models/comment.go`:

```go
package models

import (
    "github.com/forgego/forge/internal/schema"
    "time"
)

type Comment struct {
    schema.BaseSchema
}

func (Comment) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.Text("content").Required().Build(),
        schema.Bool("approved").Default(false).Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
        schema.Time("updated_at").AutoNow().Build(),
        schema.Int64("post_id").Required().Build(),
        schema.Int64("user_id").Required().Build(),
    }
}

func (Comment) Meta() schema.Meta {
    return schema.Meta{
        TableName: "comments",
        VerboseName: "Comment",
        VerboseNamePlural: "Comments",
        Ordering: []string{"-created_at"},
    }
}

func (Comment) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("post_id", "Post", "id"),
        schema.ForeignKey("user_id", "User", "id"),
    }
}
```

## Step 4: Generate Code

```bash
forge generate
```

This creates:
- Model structs with proper types
- Field expressions for type-safe queries
- Managers with CRUD operations
- QuerySets for filtering

## Step 5: Create Migrations

```bash
forge migrate
```

This creates the database tables based on your model definitions.

## Step 6: Set Up the Admin Interface

Create `admin/admin.go`:

```go
package admin

import (
    "github.com/forgego/forge/admin"
    "blog/models"
)

func RegisterModels() {
    // Register User model
    admin.RegisterModel(&models.User{}, admin.Config{
        ListDisplay: []string{"username", "email", "first_name", "last_name", "is_active", "date_joined"},
        ListFilter: []string{"is_active", "is_staff", "date_joined"},
        SearchFields: []string{"username", "email", "first_name", "last_name"},
        ListPerPage: 25,
        Ordering: []string{"-date_joined"},
        Fieldsets: []admin.Fieldset{
            {
                Title: "User Information",
                Fields: []string{"username", "email", "first_name", "last_name"},
            },
            {
                Title: "Profile",
                Fields: []string{"bio", "avatar"},
                Classes: []string{"collapse"},
            },
            {
                Title: "Permissions",
                Fields: []string{"is_active", "is_staff"},
            },
        },
    })

    // Register Category model
    admin.RegisterModel(&models.Category{}, admin.Config{
        ListDisplay: []string{"name", "slug", "post_count", "created_at"},
        ListFilter: []string{"created_at"},
        SearchFields: []string{"name", "description"},
        Ordering: []string{"name"},
        PrepopulatedFields: map[string][]string{
            "slug": {"name"},
        },
    })

    // Register Post model
    admin.RegisterModel(&models.Post{}, admin.Config{
        ListDisplay: []string{"title", "author", "category", "published", "featured", "view_count", "published_at"},
        ListFilter: []string{"published", "featured", "author", "category", "published_at"},
        SearchFields: []string{"title", "content"},
        ListPerPage: 20,
        Ordering: []string{"-created_at"},
        PrepopulatedFields: map[string][]string{
            "slug": {"title"},
        },
        Fieldsets: []admin.Fieldset{
            {
                Title: "Content",
                Fields: []string{"title", "slug", "content", "excerpt"},
            },
            {
                Title: "Metadata",
                Fields: []string{"author", "category", "published", "featured"},
            },
            {
                Title: "Media",
                Fields: []string{"featured_image"},
                Classes: []string{"collapse"},
            },
        },
        Actions: []admin.Action{
            {
                Name: "publish_posts",
                Description: "Publish selected posts",
                Handler: func(queryset admin.QuerySet, form admin.FormData) error {
                    return queryset.Update(map[string]interface{}{
                        "published": true,
                        "published_at": time.Now(),
                    })
                },
            },
            {
                Name: "unpublish_posts",
                Description: "Unpublish selected posts",
                Handler: func(queryset admin.QuerySet, form admin.FormData) error {
                    return queryset.Update(map[string]interface{}{
                        "published": false,
                    })
                },
            },
        },
    })

    // Register Comment model
    admin.RegisterModel(&models.Comment{}, admin.Config{
        ListDisplay: []string{"post", "user", "content_preview", "approved", "created_at"},
        ListFilter: []string{"approved", "created_at"},
        SearchFields: []string{"content"},
        Ordering: []string{"-created_at"},
        Actions: []admin.Action{
            {
                Name: "approve_comments",
                Description: "Approve selected comments",
                Handler: func(queryset admin.QuerySet, form admin.FormData) error {
                    return queryset.Update(map[string]interface{}{
                        "approved": true,
                    })
                },
            },
        },
    })
}
```

## Step 7: Create the REST API

Create `api/serializers.go`:

```go
package api

import (
    "github.com/forgego/forge/api/serializers"
    "blog/models"
    "time"
)

type UserSerializer struct {
    serializers.BaseSerializer
}

func (s *UserSerializer) Fields() []serializers.Field {
    return []serializers.Field{
        serializers.Int64Field("id").ReadOnly().Build(),
        serializers.StringField("username").Build(),
        serializers.StringField("email").Build(),
        serializers.StringField("first_name").Build(),
        serializers.StringField("last_name").Build(),
        serializers.TextField("bio").Build(),
        serializers.StringField("avatar").Build(),
        serializers.BooleanField("is_active").Build(),
        serializers.DateTimeField("date_joined").ReadOnly().Build(),
        serializers.MethodField("post_count", s.getPostCount).Build(),
    }
}

func (s *UserSerializer) getPostCount(obj interface{}) (interface{}, error) {
    user := obj.(*models.User)
    count, _ := models.Post.Objects.Filter(
        models.Post.Fields.AuthorID.Equals(user.ID),
    ).Count(context.Background())
    return count, nil
}

type CategorySerializer struct {
    serializers.BaseSerializer
}

func (s *CategorySerializer) Fields() []serializers.Field {
    return []serializers.Field{
        serializers.Int64Field("id").ReadOnly().Build(),
        serializers.StringField("name").Build(),
        serializers.StringField("slug").Build(),
        serializers.TextField("description").Build(),
        serializers.StringField("color").Build(),
        serializers.DateTimeField("created_at").ReadOnly().Build(),
        serializers.MethodField("post_count", s.getPostCount).Build(),
    }
}

func (s *CategorySerializer) getPostCount(obj interface{}) (interface{}, error) {
    category := obj.(*models.Category)
    count, _ := models.Post.Objects.Filter(
        models.Post.Fields.CategoryID.Equals(category.ID),
    ).Count(context.Background())
    return count, nil
}

type PostSerializer struct {
    serializers.BaseSerializer
}

func (s *PostSerializer) Fields() []serializers.Field {
    return []serializers.Field{
        serializers.Int64Field("id").ReadOnly().Build(),
        serializers.StringField("title").Build(),
        serializers.StringField("slug").Build(),
        serializers.TextField("content").Build(),
        serializers.TextField("excerpt").Build(),
        serializers.StringField("featured_image").Build(),
        serializers.BooleanField("published").Build(),
        serializers.BooleanField("featured").Build(),
        serializers.IntegerField("view_count").ReadOnly().Build(),
        serializers.DateTimeField("published_at").Build(),
        serializers.DateTimeField("created_at").ReadOnly().Build(),
        serializers.DateTimeField("updated_at").ReadOnly().Build(),
        serializers.ForeignKeyField("author", &UserSerializer{}).Build(),
        serializers.ForeignKeyField("category", &CategorySerializer{}).Build(),
        serializers.MethodField("comment_count", s.getCommentCount).Build(),
    }
}

func (s *PostSerializer) getCommentCount(obj interface{}) (interface{}, error) {
    post := obj.(*models.Post)
    count, _ := models.Comment.Objects.Filter(
        models.Comment.Fields.PostID.Equals(post.ID),
        models.Comment.Fields.Approved.Equals(true),
    ).Count(context.Background())
    return count, nil
}

type CommentSerializer struct {
    serializers.BaseSerializer
}

func (s *CommentSerializer) Fields() []serializers.Field {
    return []serializers.Field{
        serializers.Int64Field("id").ReadOnly().Build(),
        serializers.TextField("content").Build(),
        serializers.BooleanField("approved").Build(),
        serializers.DateTimeField("created_at").ReadOnly().Build(),
        serializers.ForeignKeyField("post", &PostSerializer{}).Build(),
        serializers.ForeignKeyField("user", &UserSerializer{}).Build(),
    }
}
```

Create `api/views.go`:

```go
package api

import (
    "github.com/forgego/forge/api"
    "github.com/go-chi/chi/v5"
    "blog/models"
    "net/http"
    "context"
)

func PostViewSet() *api.BaseViewSet {
    return api.NewBaseViewSet(
        &PostSerializer{},
        models.Post.Objects.All(),
        &models.Post{},
    )
}

func CategoryViewSet() *api.BaseViewSet {
    return api.NewBaseViewSet(
        &CategorySerializer{},
        models.Category.Objects.All(),
        &models.Category{},
    )
}

func CommentViewSet() *api.BaseViewSet {
    return api.NewBaseViewSet(
        &CommentSerializer{},
        models.Comment.Objects.All(),
        &models.Comment{},
    )
}

func PublicPostViewSet() *api.BaseViewSet {
    // Only show published posts
    queryset := models.Post.Objects.Filter(
        models.Post.Fields.Published.Equals(true),
    )
    
    return api.NewBaseViewSet(
        &PostSerializer{},
        queryset,
        &models.Post{},
    )
}

// Custom actions
func (v *PostViewSet) PublishPost(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    post, err := models.Post.Objects.Get(context.Background(),
        models.Post.Fields.ID.Equals(parseID(id)))
    if err != nil {
        v.RenderError(w, err, http.StatusNotFound)
        return
    }
    
    post.Published = true
    post.PublishedAt = time.Now()
    models.Post.Objects.Save(post)
    
    v.RenderResponse(w, post)
}

func (v *PostViewSet) LikePost(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    // Increment view count
    err := models.Post.Objects.Filter(
        models.Post.Fields.ID.Equals(parseID(id)),
    ).Update(map[string]interface{}{
        "view_count": models.Post.Fields.ViewCount + 1,
    })
    
    if err != nil {
        v.RenderError(w, err, http.StatusInternalServerError)
        return
    }
    
    // Return updated post
    post, _ := models.Post.Objects.Get(context.Background(),
        models.Post.Fields.ID.Equals(parseID(id)))
    v.RenderResponse(w, post)
}
```

## Step 8: Set Up Routing

Create `routes.go`:

```go
package main

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "blog/admin"
    "blog/api"
)

func setupRoutes() *chi.Mux {
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    
    // Admin routes
    r.Route("/admin", func(r chi.Router) {
        admin.SetupAdminRoutes(r)
    })
    
    // API routes
    r.Route("/api", func(r chi.Router) {
        // Public endpoints
        r.Route("/public", func(r chi.Router) {
            r.Get("/posts", api.PublicPostViewSet().List)
            r.Get("/posts/{id}", api.PublicPostViewSet().Retrieve)
            r.Get("/categories", api.CategoryViewSet().List)
            r.Get("/categories/{id}", api.CategoryViewSet().Retrieve)
        })
        
        // Admin endpoints (require authentication)
        r.Route("/admin", func(r chi.Router) {
            r.Use(authMiddleware)
            
            // Posts
            r.Get("/posts", api.PostViewSet().List)
            r.Post("/posts", api.PostViewSet().Create)
            r.Get("/posts/{id}", api.PostViewSet().Retrieve)
            r.Put("/posts/{id}", api.PostViewSet().Update)
            r.Delete("/posts/{id}", api.PostViewSet().Destroy)
            r.Post("/posts/{id}/publish", api.PostViewSet().PublishPost)
            r.Post("/posts/{id}/like", api.PostViewSet().LikePost)
            
            // Categories
            r.Get("/categories", api.CategoryViewSet().List)
            r.Post("/categories", api.CategoryViewSet().Create)
            r.Get("/categories/{id}", api.CategoryViewSet().Retrieve)
            r.Put("/categories/{id}", api.CategoryViewSet().Update)
            r.Delete("/categories/{id}", api.CategoryViewSet().Destroy)
            
            // Comments
            r.Get("/comments", api.CommentViewSet().List)
            r.Post("/comments", api.CommentViewSet().Create)
            r.Get("/comments/{id}", api.CommentViewSet().Retrieve)
            r.Put("/comments/{id}", api.CommentViewSet().Update)
            r.Delete("/comments/{id}", api.CommentViewSet().Destroy)
        })
    })
    
    // Static files
    r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    
    return r
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check authentication here
        // For now, just pass through
        next.ServeHTTP(w, r)
    })
}
```

## Step 9: Update Main Application

Update `main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "github.com/forgego/forge/server"
    "github.com/forgego/forge/config"
    "blog/admin"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db, err := server.NewDatabase(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Register admin models
    admin.RegisterModels()
    
    // Setup routes
    router := setupRoutes()
    
    // Create server
    srv := &server.Server{
        Config: cfg,
        Router: router,
        DB:     db,
    }
    
    // Start server
    log.Printf("Starting server on %s", cfg.Server.Address())
    if err := srv.Start(); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

## Step 10: Test the Application

```bash
# Start the server
forge runserver

# The application is now running at:
# http://localhost:8000/admin/ - Admin interface
# http://localhost:8000/api/public/posts - Public API
# http://localhost:8000/api/admin/posts - Admin API
```

## Step 11: Add Some Sample Data

Create a script `scripts/seed.go`:

```go
package main

import (
    "context"
    "log"
    "time"
    "blog/models"
)

func main() {
    // Create users
    admin := &models.User{
        Username:  "admin",
        Email:     "admin@blog.com",
        FirstName: "Admin",
        LastName:  "User",
        IsActive:  true,
        IsStaff:   true,
    }
    models.User.Objects.Create(admin)
    
    author := &models.User{
        Username:  "author",
        Email:     "author@blog.com",
        FirstName: "John",
        LastName:  "Doe",
        Bio:       "Tech writer and blogger",
        IsActive:  true,
    }
    models.User.Objects.Create(author)
    
    // Create categories
    tech := &models.Category{
        Name:        "Technology",
        Slug:        "technology",
        Description: "Posts about technology and programming",
        Color:       "#007bff",
    }
    models.Category.Objects.Create(tech)
    
    lifestyle := &models.Category{
        Name:        "Lifestyle",
        Slug:        "lifestyle",
        Description: "Posts about lifestyle and personal development",
        Color:       "#28a745",
    }
    models.Category.Objects.Create(lifestyle)
    
    // Create posts
    posts := []*models.Post{
        {
            Title:       "Getting Started with Go",
            Slug:        "getting-started-with-go",
            Content:     "Go is a powerful programming language...",
            Excerpt:     "Learn the basics of Go programming",
            Published:   true,
            Featured:    true,
            PublishedAt: time.Now(),
            AuthorID:    author.ID,
            CategoryID:  tech.ID,
        },
        {
            Title:       "Web Development Trends 2024",
            Slug:        "web-development-trends-2024",
            Content:     "The web development landscape is evolving...",
            Excerpt:     "What's new in web development this year",
            Published:   true,
            PublishedAt: time.Now(),
            AuthorID:    author.ID,
            CategoryID:  tech.ID,
        },
        {
            Title:       "Productivity Tips",
            Slug:        "productivity-tips",
            Content:     "Being productive is about working smarter...",
            Excerpt:     "Tips to boost your productivity",
            Published:   true,
            PublishedAt: time.Now(),
            AuthorID:    author.ID,
            CategoryID:  lifestyle.ID,
        },
    }
    
    for _, post := range posts {
        models.Post.Objects.Create(post)
    }
    
    log.Println("Sample data created successfully!")
}
```

Run the seed script:
```bash
go run scripts/seed.go
```

## Step 12: Test the API

Test the public API:

```bash
# Get all published posts
curl http://localhost:8000/api/public/posts

# Get a specific post
curl http://localhost:8000/api/public/posts/1

# Get categories
curl http://localhost:8000/api/public/categories
```

Test the admin API (you'll need to add authentication first):

```bash
# Create a new post
curl -X POST http://localhost:8000/api/admin/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Post",
    "content": "This is a new post...",
    "author_id": 1,
    "category_id": 1
  }'
```

## What We Built

You now have a complete blog application with:

### ✅ Models & Database
- User, Category, Post, Comment models
- Proper relationships and constraints
- Auto-generated migrations

### ✅ Admin Interface
- Full CRUD for all models
- Search, filtering, and pagination
- Custom actions (publish/unpublish posts)
- Fieldsets and custom display

### ✅ REST API
- Public API for blog content
- Admin API for content management
- Custom serializers with method fields
- Type-safe viewsets

### ✅ Features
- Auto-generated slugs
- View counting
- Comment approval workflow
- Featured posts
- Search functionality

## Next Steps

1. **Add Authentication** - Implement JWT or session-based auth
2. **Add Frontend** - Build a React/Vue frontend
3. **Add Testing** - Write unit and integration tests
4. **Add Deployment** - Dockerize and deploy to production
5. **Add Features** - Tags, RSS feeds, email notifications

## Deployment

### Docker Deployment

Create `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN forge generate
RUN go build -o blog .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/blog .
COPY --from=builder /app/config ./config
COPY --from=builder /app/static ./static

EXPOSE 8000
CMD ["./blog"]
```

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: blog_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  web:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
    environment:
      DATABASE_HOST: db
      DATABASE_PASSWORD: password

volumes:
  postgres_data:
```

Deploy with:
```bash
docker-compose up -d
```

## Summary

In this tutorial, you built a complete blog application with forge. You learned:

- How to define models with relationships
- How to set up the admin interface
- How to create REST APIs
- How to add custom business logic
- How to deploy the application

forge gives you Django's productivity with Go's performance and type safety. Everything is type-safe, fast, and production-ready.
