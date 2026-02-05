---
sidebar_position: 1
description: Learn how to define models in forge. Declarative model definitions with type-safe fields, relations, and hooks. Complete guide to forge ORM models.
keywords:
  - forge models
  - forge orm
  - go models
  - type-safe models
  - forge schema
  - database models
image: /forge-social-card.svg
---

# Models

Models are how you define your data in forge. Instead of writing database schemas and Go structs separately, you write the model once and forge handles the rest.

## Why models?

Models solve real problems:

- **One source of truth** - Define your data structure once, not in three places
- **Type safety** - Everything is checked at compile time
- **Less code** - No more writing the same CRUD code for every model
- **Auto migrations** - Database schema updates automatically
- **Free admin** - Admin interface appears automatically

## Defining a model

Here's a simple example:

```go
package models

import (
    "github.com/forgego/forge/pkg/schema"
)

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

func (Post) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "posts",
        VerboseName:      "Post",
        VerboseNamePlural: "Posts",
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{}
}

func (Post) Hooks() *schema.ModelHooks {
    return nil
}
```

## Field types

forge supports many field types. Here are the most common:

### Numeric Fields

```go
schema.Int64("id").Primary().AutoIncrement().Build()
schema.Int32("age").Build()
schema.Float64("price").Build()
schema.Decimal("amount").MaxDigits(10).DecimalPlaces(2).Build()
```

### String Fields

```go
schema.String("username").MaxLength(150).Required().Build()
schema.Text("description").Build()
schema.Email("email").Unique().Required().Build()
schema.URL("website").Build()
schema.Slug("slug").MaxLength(200).Build()
```

### Boolean Fields

```go
schema.Bool("is_active").Default(true).Build()
schema.Bool("is_staff").Default(false).Build()
```

### Date and Time Fields

```go
schema.Time("created_at").AutoNowAdd().Build()
schema.Time("updated_at").AutoNow().Build()
schema.Date("birth_date").Build()
schema.DateTime("last_login").Build()
```

### Special Fields

```go
schema.UUID("id").Primary().Build()
schema.JSON("metadata").Build()
schema.Bytes("avatar").Build()
```

## Field Options

All field types support chainable options:

```go
schema.String("username")
    .Required()
    .Unique()
    .Primary()
    .Index()
    .DBColumn("user_name")
    .Default("guest")
    .MaxLength(150)
    .MinLength(3)
    .HelpText("Username")
    .VerboseName("Username")
    .Choices(...)
    .Null()
    .Blank()
    .Editable(false)
    .AutoNow()
    .AutoNowAdd()
    .Build()
```

## Relations

### ForeignKey (Many-to-One)

```go
import "github.com/forgego/forge/pkg/schema/relations"

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("author", "User").
            Required().
            OnDelete(schema.Cascade).
            RelatedName("posts"),
    }
}
```

### OneToOne

```go
func (User) Relations() []schema.Relation {
    return []schema.Relation{
        relations.OneToOne("profile", "UserProfile").
            OnDelete(schema.Cascade),
    }
}
```

### ManyToMany

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ManyToMany("tags", "Tag").
            Through("post_tags").
            RelatedName("posts"),
    }
}
```

## Meta Options

The `Meta()` method returns metadata about your model:

```go
func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",              // Custom table name
        OrderBy: []string{"-date_joined"}, // Default ordering
        VerboseName: "User",              // Singular name
        VerboseNamePlural: "Users",       // Plural name
        Indexes: []schema.Index{          // Custom indexes
            {
                Name: "idx_user_email",
                Fields: []string{"email"},
                Unique: false,
            },
        },
        UniqueTogether: [][]string{       // Unique constraints
            {"username", "email"},
        },
    }
}
```

## Model Hooks

Hooks allow you to run code at specific points in a model's lifecycle:

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            return nil
        },
        AfterCreate: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        BeforeUpdate: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        AfterUpdate: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        BeforeSave: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        AfterSave: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        BeforeDelete: func(ctx context.Context, instance interface{}) error {
            return nil
        },
        AfterDelete: func(ctx context.Context, instance interface{}) error {
            return nil
        },
    }
}
```

## Complete Example

Here's a complete model with all features:

```go
package models

import (
    "context"
    "github.com/forgego/forge/pkg/schema"
    "github.com/forgego/forge/pkg/schema/relations"
)

type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("username").
            Unique().
            Required().
            MaxLength(150).
            HelpText("Required. 150 characters or fewer.").
            Build(),
        schema.String("email").
            Unique().
            Required().
            MaxLength(255).
            HelpText("Required. Must be a valid email address.").
            Build(),
        schema.String("password").
            Required().
            MaxLength(128).
            HelpText("Required. Password for authentication.").
            Build(),
        schema.Bool("is_active").
            Default(true).
            HelpText("Designates whether this user should be treated as active.").
            Build(),
        schema.Bool("is_staff").
            Default(false).
            HelpText("Designates whether this user can access the admin site.").
            Build(),
        schema.Time("date_joined").
            AutoNowAdd().
            HelpText("Date when user was created.").
            Build(),
        schema.Time("last_login").
            Null().
            HelpText("Last login timestamp.").
            Build(),
    }
}

func (User) Relations() []schema.Relation {
    return []schema.Relation{
        relations.OneToMany("posts", "Post").
            RelatedName("author"),
        relations.OneToOne("profile", "UserProfile").
            OnDelete(schema.Cascade),
    }
}

func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",
        OrderBy: []string{"-date_joined"},
        VerboseName: "User",
        VerboseNamePlural: "Users",
        Indexes: []schema.Index{
            {
                Name: "idx_user_email",
                Fields: []string{"email"},
            },
        },
    }
}

func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            user := instance.(*User)
            if user.Password != "" {
                hashed, err := auth.HashPassword(user.Password)
                if err != nil {
                    return err
                }
                user.Password = hashed
            }
            return nil
        },
    }
}
```

## Best Practices

1. **Use descriptive field names** - Make your code self-documenting
2. **Set appropriate constraints** - Use `Required()`, `Unique()`, `MaxLength()` where needed
3. **Use indexes wisely** - Index frequently queried fields
4. **Keep hooks simple** - Hooks should be fast and not block
5. **Use relations appropriately** - Choose the right relation type for your use case

## Next Steps

- [Learn about Queries](/docs/guides/queries) - Query your models
- [Explore Relations](/docs/api-reference/relations) - Deep dive into relationships
- [Check Field Reference](/docs/api-reference/fields) - Complete field type reference

