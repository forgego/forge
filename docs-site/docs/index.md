---
title: Documentation
slug: /
sidebar_position: 0
description: Forge framework documentation - Django-inspired productivity for Go.
image: /social-card.png
---

# Forge Documentation

Welcome to Forge, a Go framework that brings Django-style rapid development to the Go ecosystem. Build database-backed web applications with type-safe queries, auto-generated admin interfaces, and REST APIs. The [status](./status.md) page is the source of truth for capability tiers and verification status.

## Getting Started

New to Forge? Start here:

- **[Introduction](/docs/introduction)** - What is Forge and why use it?
- **[Quick Start](/docs/quickstart)** - Get running in 5 minutes
- **[Installation](/docs/installation)** - Detailed installation guide

## Core Guides

Learn the fundamentals:

### Database & Models
- **[Models](/docs/models)** - Define your data structure with schemas
- **[Code Generation](/docs/codegen)** - AST parsing, `gen.go`, and `api_gen.go`
- **[ORM](/docs/orm)** - Query data with type-safe expressions
- **[Migrations](/docs/migrations)** - Manage schema changes
- **[Testing](/docs/testing)** - Unit, API, and PostgreSQL integration tests

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

- **[Support contract](./status.md)** - What is supported, tested and excluded
- **[CLI Reference](/docs/cli)** - All `forge` commands and flags
- **[AI Agent Skills & llms.txt](/docs/ai-skills)** - Skills for Claude Code, Codex, Cursor, OpenCode, and Gemini CLI
- **[Deployment](/docs/deployment)** - Run one or more production instances
- **[Upgrading](/docs/upgrading)** - Upgrade checklist and regeneration steps
- **[Features](/docs/features)** - Feature list
- **[Changelog](/docs/changelog)** - Version history
- **[Security](/docs/security)** - Security policy
- **[Community](/docs/community)** - Get help and contribute

## What You Get

Forge provides everything you need to build web applications:

- **Type-Safe ORM** - Query with full compile-time safety
- **Admin** ([partially tested](./status.md)) - Register a model to manage it in the bundled admin
- **REST APIs** ([partially tested](./status.md)) - Serializers, auth, and pagination built-in
- **Migrations** - Track and apply database changes
- **Code Generation** ([partially tested](./status.md)) - Generate type-safe queries and managers
- **Security** - Sessions, CSRF and production secret checks by default; CORS and rate-limiting middleware to add
- **Authentication** - Multiple auth backends included
- **CLI Tools** - Powerful command-line interface

## Quick Example

Define a model and start querying:

```go
package models

import "github.com/forgego/forge/schema"

type Article struct {
    ArticleGenerated
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
