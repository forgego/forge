---
sidebar_position: 1
description: Forge brings Django's productivity to Go with type safety and performance.
image: /social-card.png
---

# Introduction

Forge is a Go framework that brings Django-style rapid development to the Go ecosystem. Build database-backed applications with type-safe queries, auto-generated admin interfaces, and REST APIs—all without sacrificing Go's performance or type safety.

## What is Forge?

Forge provides the building blocks for web applications:

- **Type-Safe ORM** - Query your database with full compile-time safety. No string-based queries or runtime surprises.
- **Built-in Admin** - Register a model to manage it in the bundled React admin.
- **REST API Framework** - Build APIs with serializers, authentication, and pagination built-in.
- **Schema Migrations** - Track and apply database changes automatically, like Django or Rails.
- **Code Generation** - Define schemas once, generate type-safe queries and managers automatically.

## Why Forge?

### Familiar Patterns, Go Performance

If you've used Django, Rails, or Laravel, Forge will feel natural. Same patterns, same productivity, but with Go's speed and type safety.

```go
// Define your schema
type Article struct {
    schema.BaseSchema
}

func (Article) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.MaxLength(200)),
        schema.TextField("content"),
        schema.TimeField("published_at", schema.AutoNow()),
    }
}
```

### Type Safety Without the Boilerplate

Get IDE autocomplete, compile-time validation, and refactoring support for all your queries:

```go
// ArticleObjects and ArticleFieldsInstance are written by `forge generate`
a := models.ArticleFieldsInstance

qs, err := models.ArticleObjects.Filter(a.Title.Contains("Go"))
if err != nil {
    return err
}
articles, err := qs.OrderBy(a.PublishedAt.Desc()).Limit(10).All(ctx)
```

### Batteries Included

Everything you need is already there:
- Authentication and authorization
- Input validation
- Security middleware (CSRF and sessions by default; CORS and rate limiting to add)
- Logging and error handling
- Health checks and metrics

## When to Use Forge

**Forge is great for:**
- CRUD applications and admin panels
- REST APIs with database backends
- Prototypes that need to scale
- Teams who want Django's DX in Go

**Consider alternatives if:**
- You need GraphQL (Forge focuses on REST)
- You're building a microservice without database needs
- You prefer minimal frameworks or manual control

## Core Concepts

### Models Define Everything

In Forge, you start by defining models. A model describes your data structure:

```go
type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("email", schema.Unique(), schema.MaxLength(255)),
        schema.StringField("name", schema.MaxLength(150)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}
```

### Generate Type-Safe Code

Run `forge generate` and get:
- Manager with CRUD methods
- Type-safe field expressions for queries
- With `--api`, REST serializers, viewsets and routes

### Migrations Handle Schema Changes

When you change models, run:
```bash
forge makemigrations
forge migrate up
```

Forge detects changes and generates SQL migrations automatically.

### Build APIs Fast

`forge generate --api` writes a serializer, a viewset and route registration
for each model. Mount them on the server's router:

```go
srv.RegisterRoutes(func(router *server.Router) {
    models.RegisterAPIRoutes(router) // serves /api/v1/articles/
})
```

## Architecture

Forge follows a layered architecture:

1. **Schema Layer** - Define your data structure
2. **ORM Layer** - Query and manipulate data
3. **Admin Layer** - Auto-generated UI for data management
4. **API Layer** - REST endpoints with serializers
5. **Server Layer** - HTTP server with middleware

Each layer can be used independently or together.

## Next Steps

Ready to start? Follow the quickstart guide:

- [Quick Start](/docs/quickstart) - Get running in 5 minutes
- [Installation](/docs/installation) - Detailed setup guide
- [Models Guide](/docs/models) - Learn about schemas and fields

Or explore specific features:

- [ORM & Queries](/docs/orm)
- [Admin Interface](/docs/admin/overview)
- [REST APIs](/docs/api/overview)
