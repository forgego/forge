---
sidebar_position: 2
---

# Tutorial 1: Getting Started

In this tutorial, you'll create your first forge application and learn the basics of the framework.

## What You'll Build

You'll create a simple blog application with:
- A Post model
- An admin interface to manage posts
- A REST API to view posts

## Step 1: Install forge

First, install the forge CLI:

```bash
# Clone the repository
git clone https://github.com/forgego/forge.git
cd forge

# Build the CLI
go build -o forge ./cli/cmd

# Add to PATH (optional)
export PATH=$PATH:$(pwd)
```

## Step 2: Create a New Project

Create a new forge project:

```bash
forge new myblog
cd myblog
```

This creates a project structure:

```
myblog/
├── cmd/
│   └── server/
│       └── main.go
├── app/
│   └── blog/
│       ├── models.go
│       ├── admin.go
│       └── api.go
├── config/
│   └── config.yaml
├── migrations/
└── go.mod
```

## Step 3: Configure Database

Edit `config/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: myblog_db
  sslmode: disable

server:
  host: localhost
  port: "8000"

admin:
  enabled: true
  path: "/admin"
```

Create the database:

```bash
psql -U postgres -c "CREATE DATABASE myblog_db;"
```

## Step 4: Define Your First Model

Edit `app/blog/models.go`:

```go
package blog

import (
    "github.com/forgego/forge/pkg/schema"
)

type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("title").Required().MaxLength(200).Build(),
        schema.Text("content").Required().Build(),
        schema.Bool("published").Default(false).Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
        schema.Time("updated_at").AutoNow().Build(),
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

## Step 5: Generate Code

Generate type-safe code from your models:

```bash
forge generate
```

This creates:
- `app/blog/post.gen.go` - Generated model struct
- `app/blog/post_fields.gen.go` - Type-safe field expressions
- `app/blog/post_manager.gen.go` - Manager with CRUD operations
- `app/blog/post_queryset.gen.go` - QuerySet for filtering

## Step 6: Register Admin

Edit `app/blog/admin.go`:

```go
package blog

import (
    "github.com/forgego/forge/pkg/admin"
)

func init() {
    admin.RegisterModel(&Post{})
}
```

## Step 7: Run Migrations

Create and apply migrations:

```bash
forge makemigrations
forge migrate
```

## Step 8: Start the Server

Start your development server:

```bash
forge runserver
```

Or:

```bash
go run cmd/server/main.go
```

## Step 9: Access Admin Interface

Visit `http://localhost:8000/admin/` in your browser.

You should see:
- A list of registered models
- A link to manage Posts
- The ability to create, edit, and delete posts

## Step 10: Create Your First Post

1. Click on "Posts" in the admin
2. Click "Add Post"
3. Fill in the form:
   - Title: "My First Post"
   - Content: "This is my first blog post!"
   - Published: ✓
4. Click "Save"

Congratulations! You've created your first forge application.

## What's Next?

Continue with these guides and tutorials:
- [Models and relationships](/docs/guides/models/)
- [Queries guide](/docs/guides/queries/)
- [Admin Interface tutorial](/docs/tutorials/admin-interface/)
