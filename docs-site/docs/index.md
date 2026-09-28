---
title: Documentation
slug: /
sidebar_position: 0
description: Forge framework documentation - Django-inspired productivity for Go.
image: /social-card.png
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

- **[Support contract](/docs/status)** - What is supported, tested and excluded
- **[Deployment](/docs/deployment)** - Run one production instance
- **[Features](/docs/features)** - Feature list
- **[Changelog](/docs/changelog)** - Version history
- **[Security](/docs/security)** - Security policy
- **[Community](/docs/community)** - Get help and contribute

## What You Get

Forge provides everything you need to build web applications:

- **Type-Safe ORM** - Query with full compile-time safety
- **Admin** - Register a model to manage it in the bundled admin
- **REST APIs** - Serializers, auth, and pagination built-in
- **Migrations** - Track and apply database changes
- **Code Generation** - Generate type-safe queries and managers
- **Security** - Sessions, CSRF and production secret checks by default; CORS and rate-limiting middleware to add
- **Authentication** - Multiple auth backends included
- **CLI Tools** - Powerful command-line interface

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
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.MaxLength(200)),
        schema.TextField("content"),
        schema.TimeField("published_at", schema.AutoNow()),
    }
}
```

Type-safe queries with autocomplete:

```go
// ArticleObjects and ArticleFieldsInstance are written by `forge generate`
a := models.ArticleFieldsInstance

qs, err := models.ArticleObjects.Filter(a.Title.Contains("Go"))
if err != nil {
    return err
}
articles, err := qs.OrderBy(a.PublishedAt.Desc()).Limit(10).All(ctx)
```

## Need Help?

- **Documentation** - You're here! Browse the sidebar
- **GitHub Issues** - [Report bugs or request features](https://github.com/forgego/forge/issues)
- **Examples** - [Sample projects](https://github.com/forgego/forge/tree/master/examples)

Ready to build? Start with the [Quick Start guide](/docs/quickstart).
