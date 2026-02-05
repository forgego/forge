---
sidebar_position: 1
description: Learn what forge is, how it works, and why you should use it. forge is a Django-like Go framework with type safety and code generation.
keywords:
  - what is forge
  - forge framework
  - go django
  - golang web framework
  - type-safe go
image: /img/forge-social-card.jpg
---

# What is forge?

forge is a web framework for Go that works like Django. If you know Django, you'll feel right at home. If you don't, you'll love how easy it makes building web apps.

## The problem

Building web apps in Go usually sucks because:

- You write the same CRUD code over and over
- SQL queries are strings, so typos only show up at runtime
- Admin interfaces? You build those yourself
- Migrations? Hope you remember to update your schema
- Type safety? Good luck with that

forge fixes all of this. You define your models once, and forge generates everything else. Type-safe queries, auto-generated admin, automatic migrations—it's all there.

## How it works

### 1. Define your models

Write your model like this:

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

That's your schema. No database tables yet, no structs—just the definition.

### 2. Generate code

Run this:

```bash
forge generate
```

Now you have:
- Go structs for your models
- Type-safe field expressions (like `Post.Fields.Title`)
- Managers with Create, Update, Delete methods
- QuerySets for filtering and querying

### 3. Use it

Query your database with full type safety:

```go
posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    OrderBy("-created_at").
    All(ctx)

post := &Post{
    Title:     "My First Post",
    Content:   "This is the content",
    Published: true,
}
err := Post.Objects.Create(ctx, post)
```

No SQL, no string field names, no runtime errors.

### 4. Get an admin interface

One line:

```go
admin.RegisterModel(&models.Post{})
```

Visit `/admin/` and you have a full admin panel. Search, filters, pagination—all there.

## Key ideas

### Schema-first

You don't write structs and then create tables. You write the schema once, and forge generates:
- Database tables (via migrations)
- Go structs (via code generation)
- Type-safe field accessors
- All the CRUD code

### Type safety everywhere

Everything is checked at compile time:
- `Post.Fields.Title` instead of `"title"` (typos won't compile)
- `Post.Fields.Published.Equals(true)` instead of `"published = true"`
- You get `[]*Post` back, not `[]interface{}`

### Sensible defaults

forge follows conventions:
- Model `Post` becomes table `posts`
- Field `Title` becomes column `title`
- Admin interface is auto-generated
- Migrations are auto-generated

You can override everything, but you usually don't need to.

### Extensible

Need something custom? You can:
- Add custom field types
- Write custom validators
- Build custom admin widgets
- Add custom query methods
- Create plugins

## Comparison with other frameworks

### vs. Django

| Feature | Django | forge |
|---------|--------|-------|
| Language | Python | Go |
| Type Safety | Runtime | Compile-time |
| Performance | Good | Excellent |
| Developer Experience | Excellent | Excellent |
| Code Generation | No | Yes |

### vs. GORM

| Feature | GORM | forge |
|---------|------|-------|
| Type Safety | Partial | Full |
| Code Generation | No | Yes |
| Admin Interface | No | Yes |
| Migrations | Manual | Automatic |
| Django-like API | No | Yes |

### vs. Beego

| Feature | Beego | forge |
|---------|-------|-------|
| ORM Type Safety | No | Yes |
| Admin Interface | Basic | Full |
| Code Generation | No | Yes |
| Django-like API | No | Yes |

## When to use forge

Use forge when you want:

- ✅ Django-like developer experience in Go
- ✅ Type-safe database queries
- ✅ Auto-generated admin interface
- ✅ Code generation for less boilerplate
- ✅ Built-in migrations
- ✅ Full-stack framework features

Don't use forge when you need:

- ❌ Maximum performance (use raw SQL)
- ❌ Minimal framework overhead (use net/http directly)
- ❌ Python ecosystem (use Django)

## Next steps

- **[Architecture](/docs/learn/architecture)** - Understand how forge works internally
- **[Getting Started](/docs/getting-started/installation)** - Install and set up forge
- **[Hello World](/docs/getting-started/hello-world)** - Build your first application
