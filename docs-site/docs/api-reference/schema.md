---
sidebar_position: 1
description: API Reference for the forge/schema package.
---

# Schema API Reference

The `github.com/forgego/forge/schema` package contains the core types and interfaces used to declare models, fields, relations, metadata, and lifecycle hooks.

---

## Core Interfaces

### `schema.Schema`

All Forge models implement the `Schema` interface:

```go
type Schema interface {
    Fields() []Field
    Relations() []Relation
    Meta() Meta
    Hooks() *ModelHooks
}
```

### `schema.BaseSchema`

`BaseSchema` provides no-op default implementations of all `Schema` methods. Embedding `schema.BaseSchema` in your model struct allows you to implement only the methods you need:

```go
type BaseSchema struct{}

func (BaseSchema) Fields() []Field       { return nil }
func (BaseSchema) Relations() []Relation { return nil }
func (BaseSchema) Meta() Meta            { return Meta{} }
func (BaseSchema) Hooks() *ModelHooks    { return nil }
```

---

## Schema Registry

Forge maintains a package-level registry of declared models used by the code generator and migration engine:

```go
// Register a model definition
schema.Register(&Product{})

// Retrieve registered models
models := schema.GetRegisteredModels()
```

---

## `schema.Meta`

Defines table-level configuration, indexes, constraints, and permission definitions:

```go
type Meta struct {
    TableName          string
    VerboseName        string
    VerboseNamePlural  string
    AppLabel           string
    OrderBy            []string
    GetLatestBy        string
    Indexes            []Index
    Constraints        []Constraint
    UniqueTogether     [][]string
    DBTablespace       string
    TableComment       string
}
```

### Common Meta Helpers

```go
// Specify custom table name
func (Product) Meta() schema.Meta {
    return schema.Meta{
        TableName: "inventory_products",
        OrderBy:   []string{"-created_at", "name"},
        Indexes: []schema.Index{
            schema.IndexOn("idx_product_sku", "sku"),
            schema.UniqueIndexOn("idx_product_slug", "slug"),
            schema.GINIndex("idx_product_metadata", "metadata"),
        },
        Constraints: []schema.Constraint{
            schema.Check("chk_product_price", "price >= 0"),
        },
    }
}
```

---

## Related Guides

- **[Fields Reference](/docs/api-reference/fields/)**: Complete list of field types and options.
- **[Relations Reference](/docs/api-reference/relations/)**: Foreign keys, one-to-one, and many-to-many.
- **[Lifecycle Hooks](/docs/api-reference/hooks/)**: Pre and post CRUD hooks.

