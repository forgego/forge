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
- forge CLI installed (`go install github.com/forgego/forge/cli/cmd@latest`)

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

security:
  secret_key: "your-secret-key-here"
  csrf_secret_key: "your-csrf-secret-here"
  session_secret: "your-session-secret-here"
```

## Step 3: Define the Models

Create `models/user.go`:

```go
package models

import (
    "github.com/forgego/forge/schema"
    "time"
)

type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("username", schema.Required(), schema.Unique(), schema.MaxLength(150)),
        schema.StringField("email", schema.Required(), schema.Unique(), schema.MaxLength(254)),
        schema.StringField("first_name", schema.MaxLength(30)),
        schema.StringField("last_name", schema.MaxLength(30)),
        schema.TextField("bio", schema.Blank()),
        schema.StringField("avatar", schema.Blank()),
        schema.BoolField("is_active", schema.Default(true)),
        schema.BoolField("is_staff", schema.Default(false)),
        schema.TimeField("date_joined", schema.AutoNowAdd()),
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
    return []schema.Relation{}
}
```

Create `models/category.go`:

```go
package models

import (
    "github.com/forgego/forge/schema"
    "time"
)

type Category struct {
    schema.BaseSchema
}

func (Category) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("name", schema.Required(), schema.Unique(), schema.MaxLength(100)),
        schema.StringField("slug", schema.Required(), schema.Unique(), schema.MaxLength(100)),
        schema.TextField("description", schema.Blank()),
        schema.StringField("color", schema.MaxLength(7), schema.Default("#007bff")),
        schema.TimeField("created_at", schema.AutoNowAdd()),
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
    return []schema.Relation{}
}
```

Create `models/post.go`:

```go
package models

import (
    "github.com/forgego/forge/schema"
    "time"
)

type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.Required(), schema.MaxLength(200)),
        schema.StringField("slug", schema.Required(), schema.Unique(), schema.MaxLength(200)),
        schema.TextField("content", schema.Required()),
        schema.TextField("excerpt", schema.Blank()),
        schema.StringField("featured_image", schema.Blank()),
        schema.BoolField("published", schema.Default(false)),
        schema.BoolField("featured", schema.Default(false)),
        schema.Int64Field("view_count", schema.Default(0)),
        schema.TimeField("published_at", schema.Blank()),
        schema.TimeField("created_at", schema.AutoNowAdd()),
        schema.TimeField("updated_at", schema.AutoNow()),
        schema.Int64Field("author_id", schema.Required()),
        schema.Int64Field("category_id", schema.Required()),
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
        schema.ForeignKeyField("author_id", "User",
            schema.RelatedName("posts"),
            schema.OnDelete(schema.CascadeCASCADE),
        ),
        schema.ForeignKeyField("category_id", "Category",
            schema.RelatedName("posts"),
            schema.OnDelete(schema.CascadeCASCADE),
        ),
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
    "github.com/forgego/forge/schema"
    "time"
)

type Comment struct {
    schema.BaseSchema
}

func (Comment) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.TextField("content", schema.Required()),
        schema.BoolField("approved", schema.Default(false)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
        schema.TimeField("updated_at", schema.AutoNow()),
        schema.Int64Field("post_id", schema.Required()),
        schema.Int64Field("user_id", schema.Required()),
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
        schema.ForeignKeyField("post_id", "Post",
            schema.RelatedName("comments"),
            schema.OnDelete(schema.CascadeCASCADE),
        ),
        schema.ForeignKeyField("user_id", "User",
            schema.RelatedName("comments"),
            schema.OnDelete(schema.CascadeCASCADE),
        ),
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
forge makemigrations
forge migrate up
```

This creates the database tables based on your model definitions.

## Step 6: Set Up the Admin Interface

Create `admin/admin.go`:

```go
package admin

import (
    "context"
    "time"

    "github.com/forgego/forge/admin"
    "blog/models"
)

func RegisterModels() {
    // Register User model
    admin.Register(&admin.Config[models.User]{
        ListDisplay: []admin.Field{
            models.UserFieldsInstance.Username,
            models.UserFieldsInstance.Email,
            models.UserFieldsInstance.FirstName,
            models.UserFieldsInstance.LastName,
            models.UserFieldsInstance.IsActive,
            models.UserFieldsInstance.DateJoined,
        },
        ListFilter: []admin.Field{
            models.UserFieldsInstance.IsActive,
            models.UserFieldsInstance.IsStaff,
            models.UserFieldsInstance.DateJoined,
        },
        SearchFields: []admin.Field{
            models.UserFieldsInstance.Username,
            models.UserFieldsInstance.Email,
            models.UserFieldsInstance.FirstName,
            models.UserFieldsInstance.LastName,
        },
        ListPerPage: 25,
        Ordering: []admin.Field{
            admin.Computed("-date_joined"),
        },
        Fieldsets: []admin.Fieldset[models.User]{
            admin.NewFieldset[models.User]("User Information", "username", "email", "first_name", "last_name"),
            admin.NewFieldset[models.User]("Profile", "bio", "avatar").WithCollapsed(true),
            admin.NewFieldset[models.User]("Permissions", "is_active", "is_staff"),
        },
    })

    // Register Category model
    admin.Register(&admin.Config[models.Category]{
        ListDisplay: []admin.Field{
            models.CategoryFieldsInstance.Name,
            models.CategoryFieldsInstance.Slug,
            admin.Computed("post_count"),
            models.CategoryFieldsInstance.CreatedAt,
        },
        ListFilter: []admin.Field{
            models.CategoryFieldsInstance.CreatedAt,
        },
        SearchFields: []admin.Field{
            models.CategoryFieldsInstance.Name,
            models.CategoryFieldsInstance.Description,
        },
        Ordering: []admin.Field{
            models.CategoryFieldsInstance.Name,
        },
        PrepopulatedFields: map[string][]string{
            "slug": {"name"},
        },
    })

    // Register Post model
    admin.Register(&admin.Config[models.Post]{
        ListDisplay: []admin.Field{
            models.PostFieldsInstance.Title,
            models.PostFieldsInstance.AuthorId,
            models.PostFieldsInstance.CategoryId,
            models.PostFieldsInstance.Published,
            models.PostFieldsInstance.Featured,
            models.PostFieldsInstance.ViewCount,
            models.PostFieldsInstance.PublishedAt,
        },
        ListFilter: []admin.Field{
            models.PostFieldsInstance.Published,
            models.PostFieldsInstance.Featured,
            models.PostFieldsInstance.AuthorId,
            models.PostFieldsInstance.CategoryId,
            models.PostFieldsInstance.PublishedAt,
        },
        SearchFields: []admin.Field{
            models.PostFieldsInstance.Title,
            models.PostFieldsInstance.Content,
        },
        ListPerPage: 20,
        Ordering: []admin.Field{
            admin.Computed("-created_at"),
        },
        PrepopulatedFields: map[string][]string{
            "slug": {"title"},
        },
        Fieldsets: []admin.Fieldset[models.Post]{
            admin.NewFieldset[models.Post]("Content", "title", "slug", "content", "excerpt"),
            admin.NewFieldset[models.Post]("Metadata", "author_id", "category_id", "published", "featured"),
            admin.NewFieldset[models.Post]("Media", "featured_image").WithCollapsed(true),
        },
        Actions: []admin.Action[models.Post]{
            {
                Name:        "publish_posts",
                Label:       "Publish selected posts",
                Description: "Publish selected posts",
                Handler: func(ctx context.Context, instances []*models.Post) error {
                    for _, post := range instances {
                        post.Published = true
                        post.PublishedAt = time.Now()
                        if err := models.PostObjects.Update(ctx, post); err != nil {
                            return err
                        }
                    }
                    return nil
                },
            },
            {
                Name:        "unpublish_posts",
                Label:       "Unpublish selected posts",
                Description: "Unpublish selected posts",
                Handler: func(ctx context.Context, instances []*models.Post) error {
                    for _, post := range instances {
                        post.Published = false
                        if err := models.PostObjects.Update(ctx, post); err != nil {
                            return err
                        }
                    }
                    return nil
                },
            },
        },
    })

    // Register Comment model
    admin.Register(&admin.Config[models.Comment]{
        ListDisplay: []admin.Field{
            models.CommentFieldsInstance.PostId,
            models.CommentFieldsInstance.UserId,
            admin.Computed("content_preview"),
            models.CommentFieldsInstance.Approved,
            models.CommentFieldsInstance.CreatedAt,
        },
        ListFilter: []admin.Field{
            models.CommentFieldsInstance.Approved,
            models.CommentFieldsInstance.CreatedAt,
        },
        SearchFields: []admin.Field{
            models.CommentFieldsInstance.Content,
        },
        Ordering: []admin.Field{
            admin.Computed("-created_at"),
        },
        Actions: []admin.Action[models.Comment]{
            {
                Name:        "approve_comments",
                Label:       "Approve selected comments",
                Description: "Approve selected comments",
                Handler: func(ctx context.Context, instances []*models.Comment) error {
                    for _, comment := range instances {
                        comment.Approved = true
                        if err := models.CommentObjects.Update(ctx, comment); err != nil {
                            return err
                        }
                    }
                    return nil
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
    count, _ := models.PostObjects.Filter(
        models.PostFieldsInstance.AuthorID.Equals(user.ID),
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
    count, _ := models.PostObjects.Filter(
        models.PostFieldsInstance.CategoryID.Equals(category.ID),
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
    count, _ := models.CommentObjects.Filter(
        models.CommentFieldsInstance.PostID.Equals(post.ID),
        models.CommentFieldsInstance.Approved.Equals(true),
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
        models.PostObjects.All(),
        &models.Post{},
    )
}

func CategoryViewSet() *api.BaseViewSet {
    return api.NewBaseViewSet(
        &CategorySerializer{},
        models.CategoryObjects.All(),
        &models.Category{},
    )
}

func CommentViewSet() *api.BaseViewSet {
    return api.NewBaseViewSet(
        &CommentSerializer{},
        models.CommentObjects.All(),
        &models.Comment{},
    )
}

func PublicPostViewSet() *api.BaseViewSet {
    // Only show published posts
    queryset := models.PostObjects.Filter(
        models.PostFieldsInstance.Published.Equals(true),
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
    post, err := models.PostObjects.Get(context.Background(),
        models.PostFieldsInstance.ID.Equals(parseID(id)))
    if err != nil {
        v.RenderError(w, err, http.StatusNotFound)
        return
    }
    
    post.Published = true
    post.PublishedAt = time.Now()
    models.PostObjects.Save(post)
    
    v.RenderResponse(w, post)
}

func (v *PostViewSet) LikePost(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    // Increment view count
    err := models.PostObjects.Filter(
        models.PostFieldsInstance.ID.Equals(parseID(id)),
    ).Update(map[string]interface{}{
        "view_count": models.PostFieldsInstance.ViewCount + 1,
    })
    
    if err != nil {
        v.RenderError(w, err, http.StatusInternalServerError)
        return
    }
    
    // Return updated post
    post, _ := models.PostObjects.Get(context.Background(),
        models.PostFieldsInstance.ID.Equals(parseID(id)))
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
    models.UserObjects.Create(admin)
    
    author := &models.User{
        Username:  "author",
        Email:     "author@blog.com",
        FirstName: "John",
        LastName:  "Doe",
        Bio:       "Tech writer and blogger",
        IsActive:  true,
    }
    models.UserObjects.Create(author)
    
    // Create categories
    tech := &models.Category{
        Name:        "Technology",
        Slug:        "technology",
        Description: "Posts about technology and programming",
        Color:       "#007bff",
    }
    models.CategoryObjects.Create(tech)
    
    lifestyle := &models.Category{
        Name:        "Lifestyle",
        Slug:        "lifestyle",
        Description: "Posts about lifestyle and personal development",
        Color:       "#28a745",
    }
    models.CategoryObjects.Create(lifestyle)
    
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
        models.PostObjects.Create(post)
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

forge gives you Django-like productivity with Go's performance and type safety.
If you're building CRUD-heavy apps, it helps you move faster without losing
compile-time checks.
