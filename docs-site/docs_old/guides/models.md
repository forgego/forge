---
sidebar_position: 1
description: Define models, fields, relations, and metadata in forge.
keywords:
  - forge models
  - forge schema
  - forge relations
image: /forge-social-card.svg
---

# Models Guide

Define data models with schema definitions. This is the core of a forge app.

## 1. Create a model

```go
type Post struct {
    schema.BaseSchema
}
```

## 2. Define fields

```go
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

## 3. Add relations (optional)

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("author", User{}, schema.Required()),
    }
}
```

## 4. Add metadata (optional)

```go
func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName: "posts",
        Ordering: []string{"-created_at"},
    }
}
```

## 5. Generate and migrate

```bash
forge generate
forge makemigrations posts --auto
forge migrate
```

## Next steps

- [Queries guide](/docs/guides/queries)
- [Admin guide](/docs/guides/admin)
- [REST API guide](/docs/guides/rest-api)
