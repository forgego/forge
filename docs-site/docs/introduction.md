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

**forge** is a Django-inspired Go framework that keeps Go's speed and type
safety while giving you a batteries-included workflow for web apps.

## What is forge?

forge is a full-stack Go framework for building database-backed apps. You
define schemas, and forge gives you:

- Type-safe ORM and query building
- Code generation for models, managers, and field expressions
- An auto-generated admin UI for CRUD workflows
- A REST API layer with serializers, authentication, and throttling
- Migrations and CLI tooling for everyday development

Here's a simple blog post model:

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("title", schema.Required(), schema.MaxLength(200)),
        schema.TextField("content", schema.Required()),
        schema.BoolField("published", schema.Default(false)),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}
```

Then query it like this:

```go
posts, err := PostObjects.
    Filter(PostFieldsInstance.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

That's it. No stringly-typed field names and far fewer runtime surprises.

The generated code gives you `PostObjects` for manager methods and
`PostFieldsInstance` for typed field expressions.

## Why forge?

If you're building internal tools, admin-heavy apps, or REST APIs in Go,
forge helps you move fast without giving up Go's strengths.

### What you get

- **Django-like workflow** without the dynamic typing surprises
- **Go-first performance** with compile-time checks and IDE-friendly types
- **Opinionated, but flexible**: override the pieces you need without forks

## Where forge fits well

- Admin and ops dashboards
- API-first services with a small number of models
- CRUD-heavy internal tools that need strong typing and speed

## Current status

forge has solid coverage for core ORM, admin, and API workflows. For a
details-first checklist, see the implementation status page.

- [Implementation status](/docs/status/implementation)

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
