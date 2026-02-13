---
title: Documentation
slug: /
sidebar_position: 0
description: Forge framework documentation - Django-inspired productivity for Go.
image: /forge-social-card.svg
---

# Forge Documentation

Welcome to Forge, a Go framework that brings Django-style rapid development to the Go ecosystem. Build database-backed web applications with type-safe queries, auto-generated admin interfaces, and REST APIs.

## Getting Started

New to Forge? Start here:

- **[Introduction](/docs/introduction)** - What is Forge and why use it?
- **[Quick Start](/docs/quickstart)** - Get running in 5 minutes
- **[Installation](/docs/installation)** - Detailed installation guide

## Core Guides

Learn the fundamentals:

### Database & Models
- **[Models](/docs/models)** - Define your data structure with schemas
- **[ORM](/docs/orm)** - Query data with type-safe expressions
- **[Migrations](/docs/migrations)** - Manage schema changes

### Build Features
- **[Admin Interface](/docs/admin/overview)** - Auto-generated admin panel
- **[REST APIs](/docs/api/overview)** - Build APIs with serializers
- **[Authentication](/docs/api/authentication)** - User auth and permissions

## Configuration

Set up your application:

- **[Configuration Overview](/docs/config/overview)** - App and server settings
- **[Database Config](/docs/config/database)** - Database connection setup
- **[Logging](/docs/config/logging)** - Logging configuration
- **[Security](/docs/config/security)** - Security settings

## API Reference

Detailed technical documentation:

- **[Schema API](/docs/api-reference/schema)** - Schema definitions
- **[Fields](/docs/api-reference/fields)** - Field types and options
- **[Relations](/docs/api-reference/relations)** - Foreign keys and many-to-many
- **[QuerySet](/docs/api-reference/queryset)** - Query API reference
- **[Manager](/docs/api-reference/manager)** - Manager methods
- **[Hooks](/docs/api-reference/hooks)** - Lifecycle hooks

## Advanced Topics

Go deeper:

- **[Filters](/docs/filters)** - Advanced query filtering
- **[Identity System](/docs/identity)** - User and permission management
- **[Server](/docs/server/overview)** - HTTP server and middleware
- **[Validation](/docs/validation-errors)** - Input validation

## Resources

- **[Features](/docs/features)** - Complete feature list
- **[Changelog](/docs/changelog)** - Version history
- **[Security](/docs/security)** - Security policy
- **[Community](/docs/community)** - Get help and contribute

## What You Get

Forge provides everything you need to build web applications:

- ✅ **Type-Safe ORM** - Query with full compile-time safety
- ✅ **Auto Admin** - Get a complete admin interface automatically
- ✅ **REST APIs** - Serializers, auth, and pagination built-in
- ✅ **Migrations** - Track and apply database changes
- ✅ **Code Generation** - Generate type-safe queries and managers
- ✅ **Security** - CSRF, CORS, rate limiting out of the box
- ✅ **Authentication** - Multiple auth backends included
- ✅ **CLI Tools** - Powerful command-line interface

## Quick Example

Define a model and start querying:

```go
package models

import "github.com/forgego/forge/schema"

type Article struct {
    schema.BaseSchema
}

func (Article) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary()),
        schema.StringField("title", schema.MaxLength(200)),
        schema.TextField("content"),
        schema.TimeField("published_at", schema.AutoNow()),
    }
}
```

Type-safe queries with autocomplete:

```go
// Generated expressions for compile-time safety
articles := models.ArticleManager.
    Filter(models.ArticleExpr.Title.Contains("Go")).
    OrderBy("-published_at").
    Limit(10)
```

## Need Help?

- **Documentation** - You're here! Browse the sidebar
- **GitHub Issues** - [Report bugs or request features](https://github.com/hamidrabedi/foreit/issues)
- **Examples** - [Sample projects](https://github.com/hamidrabedi/foreit/tree/main/examples)

Ready to build? Start with the [Quick Start guide](/docs/quickstart).
