---
sidebar_position: 1
description: Learn about forge, a Django-inspired Go framework that brings type safety and batteries-included workflows to Go.
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

forge is a Django-inspired Go framework for database-backed applications. It combines Go performance with a batteries-included workflow for schema, ORM, admin, REST API, and migrations.

## What you get

- **Type-safe ORM** with generated managers and field expressions
- **Schema system** for declarative models and validation
- **Admin UI** for CRUD workflows and exports
- **REST API layer** with serializers, auth, and permissions
- **Migrations and CLI tooling** for daily development

## Example model

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

Query it like this:

```go
posts, err := PostObjects.
    Filter(PostFieldsInstance.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)
```

## When forge is a good fit

- You want Django-like productivity in Go.
- You need type-safe access to data, queries, and API contracts.
- You are building admin-heavy or API-driven products.

## Next steps

1. [Quick Start](/docs/getting-started/quickstart/) - get a running app fast
2. [What is forge](/docs/learn/what-is-forge/) - concept-level overview
3. [Models guide](/docs/guides/models/) - schema and ORM workflow
