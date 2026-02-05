---
sidebar_position: 3
description: Complete quick start guide with concepts, fast startup, and core logic. Learn forge fundamentals and build your first application.
keywords:
  - forge quickstart
  - forge concepts
  - forge logic
  - forge tutorial
  - get started with forge
image: /img/forge-social-card.jpg
---

# Quick Start Guide

Get up and running with forge in minutes. This guide covers core concepts, fast startup, and the fundamental logic behind the framework.

## Core Concepts

### What is forge?

forge is a Django-like Go framework that brings Django's developer experience to Go with full type safety. You define models declaratively, and forge generates all the type-safe code you need.

### Key Concepts

**1. Schema Definition**
Define your models using Go structs that implement the `Schema` interface:

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("title").Required().MaxLength(200).Build(),
    }
}
```

**2. Code Generation**
Run `forge generate` to automatically create:
- Type-safe model structs
- Field expressions for queries
- Manager with CRUD operations
- QuerySet for filtering

**3. Type-Safe Queries**
Query your data with compile-time type checking:

```go
posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

**4. Auto-Generated Admin**
Register your model and get a full admin interface:

```go
admin.RegisterModel(&models.Post{})
```

## Fast Startup (5 Minutes)

### Step 1: Install forge

```bash
# Build from source (recommended)
git clone https://github.com/forgego/forge.git
cd forge/newforge
go build -o forge ./cli/cmd

# Or install via go install
go install github.com/forgego/forge/newforge/cli/cmd@latest
```

### Step 2: Create Project

```bash
forge new myapp
cd myapp
```

### Step 3: Configure Database

Edit `config/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  name: myapp_db
  sslmode: disable

server:
  host: localhost
  port: 8000
  
secret_key: "your-secret-key-here"
```

Create the database:

```bash
psql -U postgres -c "CREATE DATABASE myapp_db;"
```

### Step 4: Define Model

Edit `models/post.go`:

```go
package models

import "github.com/forgego/forge/schema"

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
    }
}

func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName: "posts",
        VerboseName: "Post",
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{}
}

func (Post) Hooks() *schema.ModelHooks {
    return nil
}
```

### Step 5: Generate Code

```bash
forge generate
```

This creates all the type-safe code you need.

### Step 6: Register & Run

Update `main.go`:

```go
package main

import (
    "log"
    "github.com/forgego/forge/admin"
    "github.com/forgego/forge/server"
    "github.com/forgego/forge/config"
    "myapp/models"
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
    
    // Register models for admin
    admin.RegisterModel(&models.Post{})
    
    // Setup routes and start server
    srv := &server.Server{
        Config: cfg,
        DB:     db,
    }
    
    log.Printf("Starting server on %s", cfg.Server.Address())
    if err := srv.Start(); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

### Step 7: Migrate & Run

```bash
forge generate
forge migrate
forge runserver
```

Visit `http://localhost:8000/admin/` - you have a working admin interface!

## Core Logic

### How forge Works

**1. Schema → Code Generation Flow**

```
Schema Definition (Go)
  ↓
AST Parser extracts definitions
  ↓
Code Generator creates:
  - Model structs
  - FieldExpr for type-safe access
  - Manager with CRUD
  - QuerySet for queries
  ↓
Type-safe Go code ready to use
```

**2. Query Execution Flow**

```
QuerySet.Filter(...)
  ↓
QueryExpr built from FieldExpr
  ↓
SQL Builder generates SQL
  ↓
Parameter binding (SQL injection safe)
  ↓
Database execution
  ↓
Results scanned into model instances
```

**3. Request Lifecycle**

```
HTTP Request
  ↓
Chi Router
  ↓
Middleware Stack (logging, CSRF, auth)
  ↓
Handler (Admin/API/Custom)
  ↓
QuerySet/Manager
  ↓
Database
  ↓
Response (JSON/HTML)
```

### Type Safety

forge uses Go generics to ensure type safety:

```go
// Type-safe field access
Post.Fields.Title  // Compiler knows this is a string field

// Type-safe queries
Post.Objects.Filter(
    Post.Fields.Published.Equals(true)  // Compiler validates
)

// Type-safe results
posts, err := Post.Objects.All(ctx)  // []*Post, not []interface{}
```

### Code Generation Benefits

1. **No Reflection at Runtime** - All field access is direct
2. **IDE Autocomplete** - Full IntelliSense support
3. **Compile-Time Errors** - Catch mistakes before deployment
4. **Performance** - Generated code is optimized

### Admin Auto-Generation

When you register a model:

```go
admin.RegisterModel(&models.Post{})
```

forge automatically:
- Generates list view with pagination
- Creates create/edit forms
- Adds search and filters
- Provides delete functionality
- Handles all HTTP routing

### Extension Points

Everything in forge is extensible:

- **Custom Admin Config** - Override list display, filters, etc.
- **Model Hooks** - BeforeSave, AfterCreate, etc.
- **Custom QuerySet Methods** - Add domain-specific queries
- **Middleware** - Custom request/response handling
- **Plugins** - Extend framework functionality

## Next Steps

Now that you understand the basics:

1. **[Learn Models](/docs/guides/models)** - Deep dive into model definitions
2. **[Explore Queries](/docs/guides/queries)** - Master the QuerySet API
3. **[Build APIs](/docs/guides/rest-api)** - Create REST endpoints
4. **[Customize Admin](/docs/guides/admin)** - Tailor the admin interface
5. **[See Examples](/docs/examples/blog)** - Real-world applications

## Common Patterns

### Pattern 1: Filter Published Posts

```go
publishedPosts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

### Pattern 2: Create with Validation

```go
post := &Post{
    Title: "My Post",
    Content: "Content here",
    Published: false,
}

err := Post.Objects.Create(ctx, post)
// Hooks run automatically (BeforeSave, BeforeCreate, etc.)
```

### Pattern 3: Update Specific Fields

```go
post.Title = "Updated Title"
err := Post.Objects.Update(ctx, post)
```

### Pattern 4: Complex Queries

```go
posts, err := Post.Objects.
    Filter(
        Post.Fields.Published.Equals(true).
            And(Post.Fields.CreatedAt.GreaterThan(someDate)),
    ).
    Exclude(Post.Fields.Deleted.Equals(true)).
    OrderBy("-created_at").
    Limit(10).
    All(ctx)
```

## Troubleshooting

**Code generation errors?**
- Make sure models embed `schema.BaseSchema`
- Check that `Fields()`, `Meta()`, `Relations()`, `Hooks()` are defined

**Database connection issues?**
- Verify PostgreSQL is running
- Check credentials in `config.yaml`
- Ensure database exists

**Admin not showing?**
- Register models with `admin.RegisterModel()`
- Enable admin in config: `admin.enabled: true`
- Check admin path matches config

## Summary

forge gives you:
- ✅ **Type Safety** - Compile-time checking
- ✅ **Code Generation** - No boilerplate
- ✅ **Auto Admin** - Full CRUD interface
- ✅ **Django Experience** - Familiar patterns
- ✅ **Go Performance** - Fast and efficient

Ready to build? Start with the [Installation Guide](/docs/getting-started/installation) or explore the [Full Guides](/docs/guides/models).