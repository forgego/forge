---
sidebar_position: 2
description: Get a running Forge app in 5 minutes.
image: /forge-social-card.svg
---

# Quick Start

Get a working Forge application running in 5 minutes. This guide will walk you through installation, project creation, and building your first model.

## Prerequisites

- Go 1.21 or higher ([download](https://go.dev/dl/))
- PostgreSQL 12+ ([download](https://www.postgresql.org/download/))
- Basic Go knowledge

## Step 1: Install Forge CLI

Install the `forge` command globally:

```bash
go install github.com/forgego/forge/cli/cmd@latest
```

Verify installation:

```bash
forge --version
```

:::tip
Make sure `$GOPATH/bin` or `$GOBIN` is in your PATH. If `forge` command isn't found, add this to your shell profile:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```
:::

## Step 2: Create a New Project

Create your first Forge project:

```bash
forge new blog
cd blog
```

This creates:
```
blog/
├── config/
│   └── config.yaml       # Database and app configuration
├── models/
│   └── example.go        # Sample model
├── migrations/           # Auto-generated migrations
├── main.go              # Entry point
└── go.mod
```

## Step 3: Configure Database

Edit `config/config.yaml` with your PostgreSQL connection:

```yaml
database:
  host: localhost
  port: 5432
  name: blog_db
  user: postgres
  password: your_password
  
server:
  host: 0.0.0.0
  port: 8000
  debug: true
```

Create the database:

```bash
createdb blog_db
```

## Step 4: Define Your First Model

Create `models/article.go`:

```go
package models

import "github.com/forgego/forge/schema"

type Article struct {
    schema.BaseSchema
}

func (Article) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", 
            schema.Primary(), 
            schema.AutoIncrement(),
        ),
        schema.StringField("title", 
            schema.Required(),
            schema.MaxLength(200),
        ),
        schema.TextField("content", 
            schema.Required(),
        ),
        schema.StringField("author",
            schema.MaxLength(100),
        ),
        schema.TimeField("created_at", 
            schema.AutoNowAdd(),
        ),
        schema.TimeField("updated_at", 
            schema.AutoNow(),
        ),
        schema.BoolField("published",
            schema.Default(false),
        ),
    }
}

func (Article) Meta() schema.Meta {
    return schema.Meta{
        TableName:   "articles",
        VerboseName: "Article",
        Ordering:    []string{"-created_at"},
    }
}

func (Article) Relations() []schema.Relation {
    return nil
}

func (Article) Hooks() *schema.ModelHooks {
    return nil
}
```

## Step 5: Generate Code

Generate type-safe managers and field expressions:

```bash
forge generate
```

This creates:
- `models/article_gen.go` - Generated manager and expressions
- Type-safe query builders
- Admin registration

## Step 6: Create and Run Migrations

Generate migration files from your models:

```bash
forge makemigrations init --auto
```

Apply migrations to the database:

```bash
forge migrate
```

Check migration status:

```bash
forge migrate status
```

## Step 7: Create a Superuser

Create an admin user to access the admin panel:

```bash
forge createsuperuser
```

Follow the prompts to enter email and password.

## Step 8: Start the Development Server

Run your application:

```bash
forge runserver
```

You should see:

```
Starting Forge server...
Server running at http://localhost:8000
Admin panel: http://localhost:8000/admin/
```

## Step 9: Explore the Admin Interface

Open your browser to:
- **Admin Panel**: http://localhost:8000/admin/
- **API Root**: http://localhost:8000/api/

Log in with the superuser credentials you created.

In the admin panel, you can:
- View and manage articles
- Add, edit, and delete records
- Search and filter data
- Export data to CSV

## Step 10: Query Data in Code

Edit `main.go` to query your models:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "blog/models"
    "github.com/forgego/forge/orm"
    "github.com/forgego/forge/server"
)

func main() {
    // Initialize database
    db, err := orm.InitDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    ctx := context.Background()
    
    // Create an article
    article := &models.Article{}
    err = models.ArticleManager.Create(ctx, article, map[string]interface{}{
        "title":     "Getting Started with Forge",
        "content":   "Forge makes Go web development fast...",
        "author":    "Jane Doe",
        "published": true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Type-safe query
    articles, err := models.ArticleManager.
        Filter(models.ArticleExpr.Published.Equals(true)).
        OrderBy("-created_at").
        Limit(10).
        All(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, a := range articles {
        fmt.Printf("Article: %s by %s\n", a.Title, a.Author)
    }
    
    // Start server
    srv := server.New()
    srv.Start()
}
```

## What's Next?

Now that you have a working app, learn more about Forge:

### Core Concepts
- [Models & Fields](/docs/models) - Define your data structure
- [ORM & Queries](/docs/orm) - Query data type-safely
- [Migrations](/docs/migrations) - Manage schema changes

### Build Features
- [Admin Interface](/docs/admin/overview) - Customize the admin
- [REST APIs](/docs/api/overview) - Build APIs with serializers
- [Authentication](/docs/api/authentication) - Add user auth

### Advanced Topics
- [Relations](/docs/api-reference/relations) - Foreign keys and many-to-many
- [Validation](/docs/validation-errors) - Input validation
- [Filters](/docs/filters) - Advanced query filters

## Common Commands

Here's a quick reference of Forge CLI commands:

```bash
# Project management
forge new <project>         # Create new project
forge generate              # Generate code from models
forge runserver             # Start development server

# Database migrations
forge makemigrations        # Create migration files
forge migrate               # Apply migrations
forge migrate status        # Check migration status
forge migrate rollback      # Rollback last migration

# User management
forge createsuperuser       # Create admin user

# Development
forge version               # Show version
forge --help                # Show all commands
```

## Troubleshooting

### Command not found: forge

Make sure `$GOPATH/bin` is in your PATH:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Database connection failed

Check your `config/config.yaml` and ensure PostgreSQL is running:

```bash
psql -U postgres -c "SELECT version();"
```

### Migration failed

Reset and try again:

```bash
forge migrate rollback
forge makemigrations init --auto
forge migrate
```

## Getting Help

- [Documentation](/docs/) - Full documentation
- [GitHub Issues](https://github.com/hamidrabedi/foreit/issues) - Report bugs
- [Examples](https://github.com/hamidrabedi/foreit/tree/main/examples) - Sample projects

Ready to dive deeper? Continue with the [Models guide](/docs/models).
