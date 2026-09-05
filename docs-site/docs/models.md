---
sidebar_position: 10
description: Define models, fields, relations, meta, and hooks.
image: /forge-social-card.svg
---

# Models

Models implement the `schema.Schema` interface. You define fields, relations, meta options, and hooks. Forge generates type-safe ORM accessors and field expressions.

## Interface

- `Fields() []schema.Field`
- `Relations() []schema.Relation`
- `Meta() schema.Meta`
- `Hooks() *schema.ModelHooks`

Embed `schema.BaseSchema` for defaults.

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
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}
```

## Field types

- Int64, Int32
- String, Text
- Bool
- Time, Date, DateTime
- Float32, Float64, Decimal
- Email, URL, UUID
- JSON, Bytes
- ForeignKey, OneToOne, ManyToMany

## Field options (selected)

Validation and database options live on `schema.Field`:

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

```go
func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("author", User{}, schema.Required()),
    }
}
```

Supported relation types:

- ForeignKey
- OneToOne
- ManyToMany

Relation options include cascade behavior, constraint config, related names, and through tables.

## Meta options

`schema.Meta` supports:

- TableName, VerboseName, VerboseNamePlural
- AppLabel, DefaultManager, BaseManager
- OrderBy, GetLatestBy
- Indexes, Constraints, UniqueTogether
- Permissions
- DBTablespace, TableComment, TableOptions
- Proxy, Abstract, Managed, SelectOnSave

## Hooks

```go
func (Post) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithBeforeCreate(func(ctx context.Context, instance interface{}) error { return nil }).
        WithAfterCreate(func(ctx context.Context, instance interface{}) error { return nil })
}
```

Hook points:

- BeforeCreate / AfterCreate
- BeforeUpdate / AfterUpdate
- BeforeSave / AfterSave
- BeforeDelete / AfterDelete
- Clean (validation)
- Save / Delete (override)

## Code generation

Run `forge generate` to produce:

- Model accessors
- Field expression helpers
- Manager and QuerySet types

## Next steps

- [ORM](/docs/orm/)
- [Migrations](/docs/migrations/)
