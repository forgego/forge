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
    "github.com/forgego/forge/schema"
)

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
schema.Int64Field("id", schema.Primary(), schema.AutoIncrement())
schema.Int32Field("age")
schema.Float64Field("price")
schema.DecimalField("amount", schema.MaxDigits(10), schema.DecimalPlaces(2))
```

### String Fields

```go
schema.StringField("username", schema.MaxLength(150), schema.Required())
schema.TextField("description")
schema.EmailField("email", schema.Unique(), schema.Required())
schema.URLField("website")
schema.StringField("slug", schema.MaxLength(200))
```

### Boolean Fields

```go
schema.BoolField("is_active", schema.Default(true))
schema.BoolField("is_staff", schema.Default(false))
```

### Date and Time Fields

```go
schema.TimeField("created_at", schema.AutoNowAdd())
schema.TimeField("updated_at", schema.AutoNow())
schema.DateField("birth_date")
schema.DateTimeField("last_login")
```

### Special Fields

```go
schema.UUIDField("id", schema.Primary())
schema.JSONField("metadata")
schema.BytesField("avatar")
```

## Field Options

All field types support chainable options:

```go
schema.StringField(
    "username",
    schema.Required(),
    schema.Unique(),
    schema.Primary(),
    schema.DBIndex(),
    schema.DBColumn("user_name"),
    schema.Default("guest"),
    schema.MaxLength(150),
    schema.MinLength(3),
    schema.HelpText("Username"),
    schema.VerboseName("Username"),
    schema.ChoicesFromPairsOpts("admin", "Admin", "user", "User"),
    schema.Blank(),
    schema.Editable(false),
)
```

## Relations

### ForeignKey (Many-to-One)

```go
import "github.com/forgego/forge/schema"

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("author", "User",
            schema.OnDelete(schema.CascadeCASCADE),
            schema.RelatedName("posts"),
        ),
    }
}
```

### OneToOne

```go
func (User) Relations() []schema.Relation {
    return []schema.Relation{
        schema.OneToOneField("profile", "UserProfile",
            schema.OnDelete(schema.CascadeCASCADE),
        ),
    }
}
```

### ManyToMany

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ManyToManyField("tags", "Tag",
            schema.Through("post_tags"),
            schema.RelatedName("posts"),
        ),
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
    "github.com/forgego/forge/schema"
)

type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField(
            "username",
            schema.Unique(),
            schema.Required(),
            schema.MaxLength(150),
            schema.HelpText("Required. 150 characters or fewer."),
        ),
        schema.StringField(
            "email",
            schema.Unique(),
            schema.Required(),
            schema.MaxLength(255),
            schema.HelpText("Required. Must be a valid email address."),
        ),
        schema.StringField(
            "password",
            schema.Required(),
            schema.MaxLength(128),
            schema.HelpText("Required. Password for authentication."),
        ),
        schema.BoolField(
            "is_active",
            schema.Default(true),
            schema.HelpText("Designates whether this user should be treated as active."),
        ),
        schema.BoolField(
            "is_staff",
            schema.Default(false),
            schema.HelpText("Designates whether this user can access the admin site."),
        ),
        schema.TimeField(
            "date_joined",
            schema.AutoNowAdd(),
            schema.HelpText("Date when user was created."),
        ),
        schema.TimeField(
            "last_login",
            schema.Optional(),
            schema.HelpText("Last login timestamp."),
        ),
    }
}

func (User) Relations() []schema.Relation {
    return []schema.Relation{
        schema.OneToOneField("profile", "UserProfile",
            schema.OnDelete(schema.CascadeCASCADE),
        ),
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

