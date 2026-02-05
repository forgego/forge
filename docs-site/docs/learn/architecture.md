---
sidebar_position: 2
description: Understand forge's architecture, design principles, and how it works. Learn about code generation, layers, components, and extension points.
keywords:
  - forge architecture
  - forge design
  - code generation
  - forge internals
  - framework architecture
image: /forge-social-card.svg
---

# Architecture

forge is built to feel like Django but work like Go. Here's how it all fits together.

## The Big Picture

You write model definitions in Go, forge generates type-safe code, and you get a complete web framework. No reflection magic, no string-based queries that break at runtime.

```
You define models → forge generates code → you build your app
     ↓                    ↓                    ↓
  schema.go → models.go, fields.go → admin, API, views
```

## How It Works

### 1. You Define Models

```go
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
```

### 2. forge Generates Code

When you run `forge generate`, forge:

- Parses your Go code with AST
- Generates type-safe structs
- Creates field expressions for queries
- Builds managers and QuerySets

### 3. You Use the Generated Code

```go
// Type-safe queries - no string field names
posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)

// Admin interface - just register your model
admin.RegisterModel(&Post{})

// REST API - one line of code
router.Handle("/api/posts", api.NewBaseViewSet(serializer, Post.Objects, &Post{}))
```

## The Layers

### Your Code (Top Layer)
This is where you spend most of your time:
- Model definitions in `models/`
- Your business logic
- Custom views and handlers
- Configuration

### Generated Code (Middle Layer)
forge creates this for you:
- Model structs with proper types
- Field expressions for type-safe queries
- Managers with CRUD operations
- QuerySets for filtering and ordering

### Framework Core (Bottom Layer)
The forge framework handles:
- Database connections and transactions
- HTTP routing and middleware
- Security (CSRF, XSS, SQL injection)
- Admin interface rendering
- API serialization

## Key Components

### Schema System
**Location:** `forge/schema/`  
**What it does:** Defines your models declaratively

```go
func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.String("title").Required().Build(),
        // More fields...
    }
}
```

### Code Generator
**Location:** `forge/codegen/`  
**What it does:** Turns your schema definitions into Go code

- AST parser reads your Go files
- Template system generates code
- Go formats the output
- Imports are handled automatically

### ORM System
**Location:** `forge/orm/`  
**What it does:** Type-safe database operations

```go
// This is compile-time checked
Post.Fields.Title.Contains("golang")

// Not this (runtime error waiting to happen)
db.Query("SELECT * FROM posts WHERE title LIKE ?", "%golang%")
```

### Admin System
**Location:** `forge/admin/`  
**What it does:** Auto-generates Django-style admin interface

- List views with pagination
- Create/edit forms
- Search and filtering
- Bulk actions

### API Framework
**Location:** `forge/api/`  
**What it does:** Django REST Framework equivalent

- Serializers for data validation
- ViewSets for CRUD operations
- Authentication and permissions
- OpenAPI documentation

## Data Flow

### Request Flow
```
HTTP Request → Chi Router → Middleware → Handler → QuerySet → Database → Response
```

1. **HTTP Request** comes in
2. **Chi Router** matches the URL
3. **Middleware** runs (auth, CSRF, logging, etc.)
4. **Handler** processes the request
5. **QuerySet** builds a type-safe query
6. **Database** executes with parameter binding
7. **Response** goes back through middleware

### Code Generation Flow
```
Schema Definition → AST Parser → Template → Generated Code
```

1. You write schema definitions
2. AST parser reads the Go code
3. Template system generates Go code
4. Generated code is written to files

## Design Decisions

### Why Generics?
Before Go 1.18, frameworks used `interface{}` everywhere. That meant:
- No type safety
- Runtime errors instead of compile-time errors
- Terrible IDE support

forge uses generics everywhere:
```go
QuerySet[Post]     // Type-safe
Manager[Post]      // Type-safe  
Admin[Post]        // Type-safe
FieldExpr[string]  // Type-safe
```

### Why Code Generation?
Some frameworks use reflection at runtime. forge generates code because:

- **Performance** - No reflection overhead
- **Type Safety** - Generated code is just Go code
- **IDE Support** - Your IDE understands generated code
- **Debugging** - You can step through generated code

### Why Chi?
Chi is a lightweight router that:
- Works with Go's `net/http`
- Has great middleware support
- Is fast and battle-tested
- Doesn't try to do too much

### Why PostgreSQL?
PostgreSQL gives you:
- Strong typing
- Great JSON support
- Full-text search
- Advanced features (arrays, hstore, etc.)

## Technology Stack

### Core Dependencies
- **Go 1.25+** - Language with generics
- **database/sql** - Standard database interface
- **chi/v5** - HTTP router
- **lib/pq** - PostgreSQL driver

### Key Libraries
- **golang-migrate** - Database migrations
- **zap** - Structured logging
- **viper** - Configuration
- **gorilla/csrf** - CSRF protection
- **alexedwards/scs** - Session management
- **go-playground/validator** - Validation

## Performance Considerations

### What's Fast?
- **Database queries** - Uses connection pooling
- **Code generation** - One-time cost
- **HTTP routing** - Chi is very fast
- **JSON serialization** - Uses standard library

### What's Not Free?
- **AST parsing** - Only runs during generation
- **Template rendering** - Only in admin/API
- **Reflection** - Minimized usage

### Optimization Tips
1. **Use select_related** for foreign keys
2. **Cache expensive queries**
3. **Use pagination** for large lists
4. **Profile your queries** with EXPLAIN

## Security Architecture

### Defense in Depth
1. **Input Validation** - At multiple layers
2. **SQL Injection Prevention** - Parameter binding everywhere
3. **CSRF Protection** - Built into forms
4. **XSS Prevention** - Output encoding
5. **Authentication** - Multiple secure options

### Security by Default
- CSRF protection is enabled by default
- All database queries use parameter binding
- Form input is validated and sanitized
- Sessions are secure by default

## Extensibility

### Plugin Points
- **Custom Fields** - Create your own field types
- **Custom Widgets** - Build form widgets
- **Custom Auth** - Authentication backends
- **Custom Middleware** - Request/response processing
- **Custom Serializers** - API serialization

### Hook System
- **Model Hooks** - Before/After save, create, update, delete
- **Request Hooks** - Before/After request processing
- **Admin Hooks** - Custom admin behavior

## Testing Architecture

### Test Support
- **Test Database** - Automatic test database setup
- **Mock Support** - Easy mocking of database operations
- **Test Helpers** - Utilities for testing forge apps
- **Integration Tests** - Database testing support

### Test Patterns
```go
func TestPostModel(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer db.Close()
    
    // Create test data
    post := &Post{Title: "Test"}
    err := Post.Objects.Create(post)
    assert.NoError(t, err)
    
    // Test queries
    found, err := Post.Objects.Get(Post.Fields.ID.Equals(post.ID))
    assert.NoError(t, err)
    assert.Equal(t, "Test", found.Title)
}
```

## What's Next?

This architecture gives you:
- **Django's productivity** - Admin interface, ORM, sensible defaults
- **Go's performance** - Compiled, fast, efficient
- **Type safety** - Compile-time checking everywhere
- **Extensibility** - Everything can be customized

Want to dive deeper?
- [Features Overview](/docs/features/overview) - All features in detail
- [Schema System](/docs/features/schema-system) - Model definitions
- [Admin System](/docs/features/admin-system) - Auto-generated admin
- [API Framework](/docs/features/api-framework) - REST API development
