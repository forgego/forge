---
sidebar_position: 1
description: Learn about forge, a Django-like Go framework that brings Django's developer experience to Go with full type safety and modern Go features.
keywords:
  - go framework
  - golang framework
  - django go
  - type-safe orm
  - go web framework
  - forge framework
image: /forge-social-card.svg
---

# Introduction

**forge** is a Go framework that feels like Django. If you've ever used Django and wished you could have that same developer experience in Go, this is it.

## What is forge?

forge is basically Django for Go. You get the same productivity and ease of use, but with Go's speed and type safety.

No more writing the same CRUD code over and over. Define your models once, and forge generates all the type-safe code you need. No more string-based SQL queries that break at runtime. You get compile-time type checking for everything. And forget about building admin panels—forge creates one automatically from your models.

Here's a simple example—a blog post model:

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

Then query it like this:

```go
posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

That's it. No SQL, no string field names, no runtime surprises.

## Why forge?

Ever used Django and thought "this is great, but I wish it was in Go"? That's exactly why forge exists.

You get Django's developer experience—the admin interface, the ORM, the sensible defaults—but with Go's performance and type safety. No more runtime errors from typos in field names. No more wondering if your query is correct until it runs.

### What you get

- **Django's ease of use** - Same declarative models, same admin interface, same "it just works" feeling
- **Go's speed** - Compiled, fast, efficient. Perfect for production workloads
- **Type safety** - Your IDE knows what fields exist. Refactoring is safe. No more `map[string]interface{}` everywhere
- **Modern Go** - Uses generics, follows Go conventions, plays nice with the ecosystem

## What forge gives you

### Type-Safe ORM

No more `"SELECT * FROM users WHERE is_active = true"`. Write queries that your compiler checks:

```go
users, err := User.Objects.
    Filter(User.Fields.IsActive.Equals(true)).
    OrderBy("-date_joined").
    Limit(10).
    All(ctx)
```

If you typo a field name, your code won't compile. That's the point.

### Auto-Generated Admin

Remember Django's admin? Same thing here. Register your model and you're done:

```go
admin.RegisterModel(&models.Post{})
```

Visit `/admin/` and there's your full CRUD interface. Search, filters, pagination—all there.

### Code Generation

Run `forge generate` and it creates all the boilerplate:
- Model structs with the right types
- Field expressions for type-safe queries
- Managers with CRUD methods
- QuerySets for filtering

You write the schema once, forge handles the rest.

### REST API Framework

Building an API? forge has you covered. It's like Django REST Framework but for Go:

```go
viewset := api.NewBaseViewSet(
    NewPostSerializer(),
    Post.Objects.Filter(),
    &Post{},
)
router.Register("posts", viewset)
```

Now you have full CRUD endpoints at `/api/posts/`. Pagination, filtering, authentication—all built in.

### User System

Users, authentication, sessions, permissions. It's all there:

```go
user, err := users.Authenticate(ctx, email, password)
if users.HasPermission(user, "posts.add_post") {
    // user can add posts
}
```

## Current status

**MVP Complete - Production Ready! 🎉**

All core features are implemented and working. You can build real applications with forge today.

### ✅ Fully Implemented Features

**Core Systems:**
- ✅ **Schema System** - Complete field types (Int64, String, Bool, Time, Date, DateTime, Float32/64, Decimal, Text, Email, URL, UUID, JSON, Bytes) with all field options
- ✅ **Code Generation** - AST-based generation of models, FieldExpr, managers, QuerySets with proper type safety
- ✅ **Type-Safe ORM** - Complete QuerySet API (Filter, Exclude, OrderBy, Limit, Offset, Distinct, All, Get, First, Last, Count, Exists)
- ✅ **Manager CRUD** - Create, Update, Delete with lifecycle hooks and proper error handling

**Admin Interface:**
- ✅ **Type-Safe Admin** - Generic-based Admin[T] with complete Config[T] system
- ✅ **List Views** - Pagination, filtering, searching, sorting with field expressions
- ✅ **Form Views** - Create/update forms with validation and widgets
- ✅ **Filter System** - Boolean, Choice, Date, Number, Text, Related filters with widgets
- ✅ **Widget System** - Text, Textarea, Select, Checkbox, Radio, Date, Time, DateTime, File, Image, RichText, Autocomplete, RawID
- ✅ **Actions** - Bulk operations with custom action support
- ✅ **Export** - CSV and JSON export functionality
- ✅ **Inlines** - Related model editing (nested forms)
- ✅ **Fieldsets** - Form field grouping and organization

**REST API Framework:**
- ✅ **BaseViewSet** - Complete CRUD operations with type safety
- ✅ **Serializer System** - Field serializers with validation
- ✅ **Authentication** - Token, JWT, Basic, Session, API Key backends
- ✅ **Permissions** - AllowAny, IsAuthenticated, IsAdminUser, IsOwnerOrReadOnly, and more
- ✅ **Throttling** - AnonRateThrottle, UserRateThrottle, ScopedRateThrottle
- ✅ **Renderers** - JSON, XML, YAML, HTML, CSV response formats
- ✅ **Parsers** - JSON, XML, Form, MultiPart request parsers
- ✅ **Filters** - Field filtering and search functionality
- ✅ **Pagination** - PageNumber and LimitOffset pagination
- ✅ **Versioning** - API versioning support
- ✅ **OpenAPI Docs** - Auto-generated API documentation

**Advanced Systems:**
- ✅ **Filter System** - AST-based filtering with security validation, query optimization, and persistence
- ✅ **Identity System** - Complete user management with authentication, sessions, permissions, groups, tokens
- ✅ **Database Layer** - Connection pooling, transactions with savepoints, migration integration
- ✅ **Migration System** - AST-based schema detection, change detection, SQL generation, rollback support
- ✅ **HTTP Server** - Chi router with complete middleware stack (Request ID, Real IP, Recoverer, Logger, Session, CSRF, Authentication)
- ✅ **Security** - CSRF protection, XSS prevention, SQL injection prevention, input sanitization
- ✅ **Logging** - Structured logging with zap, multiple outputs, contextual logging
- ✅ **Configuration** - Viper-based YAML/JSON/env var configuration
- ✅ **Validation** - go-playground/validator integration
- ✅ **CLI Tools** - forge new, generate, migrate, runserver commands

**Technology Stack:**
- Go 1.25+ with generics
- PostgreSQL with database/sql
- Chi v5 router
- golang-migrate for migrations
- zap for logging
- viper for configuration
- And 15+ other carefully selected libraries

## Getting started

Ready to try it? Here's where to start:

1. **[Install forge](/docs/getting-started/installation)** - Get it running on your machine
2. **[Hello World](/docs/getting-started/hello-world)** - Build your first app (takes 5 minutes)
3. **[Learn the basics](/docs/learn/what-is-forge)** - How forge works under the hood
4. **[Check out examples](/docs/examples/blog)** - See real code in action

Want to dive deeper?

- **[Architecture](/docs/learn/architecture)** - How forge is built
- **[API Reference](/docs/api-reference/schema)** - Every function, every method
- **[Guides](/docs/guides/models)** - Step-by-step tutorials
