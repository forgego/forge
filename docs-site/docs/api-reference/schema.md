---
sidebar_position: 1
description: API Reference for the schema package.
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

For models you pass to `forge generate`, embed `<Model>Generated` (for example,
`ArticleGenerated`) when you need the generated Go fields and CRUD methods.
The generated type includes `BaseSchema`.

---

## Model Discovery

The code generator and migration generator parse model declarations from the configured models directory. They do not use a runtime model registry: `schema.Register` and `schema.GetRegisteredModels` are not package APIs. Declare models as Go structs implementing `schema.Schema` and pass their source directory to generation commands.

The registry in the `schema` package is for custom **field types**. Use `schema.RegisterFieldType` to add a field factory and `schema.ListRegisteredFieldTypes` to list those field type names.

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
            {Name: "chk_product_price", Type: schema.ConstraintTypeCheck, Condition: "price >= 0"},
        },
    }
}
```

`forge makemigrations` reads the literal `Constraint` above. The `schema.Check`
helper builds equivalent runtime metadata but is not yet supported by source
generation. `GINIndex` is PostgreSQL-specific; use B-tree indexes on SQLite.

---

## Related Guides

- **[Fields Reference](/docs/api-reference/fields/)**: Complete list of field types and options.
- **[Relations Reference](/docs/api-reference/relations/)**: Foreign keys, one-to-one, and many-to-many.
- **[Lifecycle Hooks](/docs/api-reference/hooks/)**: Pre and post CRUD hooks.

