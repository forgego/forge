---
sidebar_position: 10
description: Define models, fields, relations, meta, and hooks.
image: /forge-social-card.svg
---

# Models

Models are defined by implementing the `schema.Schema` interface. You describe fields, relations, metadata, and hooks, and forge generates type-safe ORM accessors.

## Schema interface

A model implements these methods:

- `Fields() []schema.Field`
- `Relations() []schema.Relation`
- `Meta() schema.Meta`
- `Hooks() *schema.ModelHooks`

You can embed `schema.BaseSchema` to get empty defaults.

## Basic model

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
    }
}
```

## Field types

Built-in field types include:

- Int64, Int32
- String, Text
- Bool
- Time, Date, DateTime
- Float32, Float64, Decimal
- Email, URL, UUID
- JSON, Bytes
- ForeignKey, OneToOne, ManyToMany

## Field options

Common field options supported by `schema.Field`:

- Required, Unique, Blank
- MinLength, MaxLength
- MinValue, MaxValue
- MaxDigits, DecimalPlaces
- Choices, Validators, ValidationTag
- ErrorMessages
- DBColumn, DBType, DBCollation, DBComment, DBTablespace
- DBIndex, DBDefault
- AutoIncrement, AutoNow, AutoNowAdd
- VerboseName, HelpText, Editable, Serialize

## Relations

Use relations to connect models:

- ForeignKey
- OneToOne
- ManyToMany

Relations support cascade behaviors and constraint configuration.

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("author", User{}, schema.Required()),
    }
}
```

## Meta options

`schema.Meta` supports table naming and advanced database options, including:

- TableName, VerboseName, VerboseNamePlural
- AppLabel, DefaultManager, BaseManager
- OrderBy, GetLatestBy
- Indexes, Constraints, UniqueTogether
- Permissions
- DBTablespace, TableComment, TableOptions
- Proxy, Abstract, Managed, SelectOnSave

## Hooks

Use hooks to run logic around create, update, save, and delete.

```go
func (Post) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithBeforeCreate(func(ctx context.Context, instance interface{}) error { return nil }).
        WithAfterCreate(func(ctx context.Context, instance interface{}) error { return nil })
}
```

Supported hook points:

- BeforeCreate / AfterCreate
- BeforeUpdate / AfterUpdate
- BeforeSave / AfterSave
- BeforeDelete / AfterDelete
- Clean (validation)
- Save / Delete (override)

## Validation

Validation can be expressed through field options and validators. Forge also includes typed validators in the validation package for custom checks.

## Code generation

Run `forge generate` to produce:

- Model accessors
- Field expression helpers
- Manager and QuerySet types

## Next steps

- [ORM](/docs/orm/)
- [Migrations](/docs/migrations/)
