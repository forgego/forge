---
sidebar_position: 1
description: Complete API reference for the Schema interface. Learn how to define models, fields, relations, meta options, and hooks in forge.
keywords:
  - forge schema api
  - schema interface
  - model definition
  - forge api reference
image: /img/forge-social-card.jpg
---

# Schema

The `Schema` interface is the foundation of all models in forge. It defines how your data is structured and how it behaves.

## Interface

All models must implement the `Schema` interface:

```go
type Schema interface {
    Fields() []Field
    Relations() []Relation
    Meta() Meta
    Hooks() *ModelHooks
}
```

## BaseSchema

Embed `BaseSchema` in your models:

```go
type User struct {
    schema.BaseSchema
}
```

## Fields() Method

Returns all field definitions for the model:

```go
func (User) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("username").Required().MaxLength(150).Build(),
    }
}
```

## Relations() Method

Returns relationship definitions:

```go
func (User) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("profile", "UserProfile").
            OnDelete(schema.Cascade),
    }
}
```

## Meta() Method

Returns model metadata:

```go
func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName:        "users",
        VerboseName:      "User",
        VerboseNamePlural: "Users",
        OrderBy:          []string{"-date_joined"},
    }
}
```

## Hooks() Method

Returns lifecycle hooks:

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeCreate: func(ctx context.Context, instance interface{}) error {
            // Hook logic
            return nil
        },
    }
}
```

## Meta Options

### TableName

Custom table name:

```go
TableName: "custom_users"
```

### VerboseName

Human-readable singular name:

```go
VerboseName: "User"
```

### VerboseNamePlural

Human-readable plural name:

```go
VerboseNamePlural: "Users"
```

### OrderBy

Default ordering:

```go
OrderBy: []string{"-date_joined", "username"}
```

Use `-` prefix for descending order.

### Indexes

Custom indexes:

```go
Indexes: []schema.Index{
    {
        Name: "idx_user_email",
        Fields: []string{"email"},
        Unique: false,
    },
}
```

### UniqueTogether

Composite unique constraints:

```go
UniqueTogether: [][]string{
    {"username", "email"},
}
```

### Constraints

Check constraints:

```go
Constraints: []schema.UniqueConstraint{
    {
        Name: "check_age",
        Check: "age >= 0",
    },
}
```

## ModelHooks

Lifecycle hooks:

```go
type ModelHooks struct {
    BeforeCreate func(ctx context.Context, instance interface{}) error
    AfterCreate  func(ctx context.Context, instance interface{}) error
    BeforeUpdate func(ctx context.Context, instance interface{}) error
    AfterUpdate  func(ctx context.Context, instance interface{}) error
    BeforeSave   func(ctx context.Context, instance interface{}) error
    AfterSave    func(ctx context.Context, instance interface{}) error
    BeforeDelete func(ctx context.Context, instance interface{}) error
    AfterDelete  func(ctx context.Context, instance interface{}) error
    Clean        func(ctx context.Context, instance interface{}) error
    Validate     func(ctx context.Context, instance interface{}) error
}
```

## See Also

- [Fields Reference](/docs/api-reference/fields) - Complete field type reference
- [Relations Reference](/docs/api-reference/relations) - Relationship types
- [Hooks Reference](/docs/api-reference/hooks) - Lifecycle hooks

