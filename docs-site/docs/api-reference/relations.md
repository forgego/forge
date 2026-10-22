---
sidebar_position: 5
description: API Reference for model relationships, through models, and cascade options.
---

# Relations API Reference

The `github.com/forgego/forge/schema` package provides expressive constructors for declaring relational links between models.

---

## Relation Types

Forge supports three relational cardinality patterns:

### 1. `schema.ForeignKeyField` (Many-to-One)

Establishes a foreign key column on the model pointing to a target model's primary key:

```go
schema.ForeignKeyField("author_id", "User", schema.OnDelete(schema.CascadeCASCADE),
    schema.RelatedName("posts"), schema.RelatedQueryName("post"))
```

### 2. `schema.OneToOneField` (One-to-One)

Declares a one-to-one relation in model metadata. Migrations currently create its foreign key but do not add a `UNIQUE` constraint automatically. Mark the corresponding field `schema.Unique()` or declare a unique constraint in `Meta()` to enforce 1:1 cardinality in the database:

```go
schema.OneToOneField("profile_id", "UserProfile", schema.OnDelete(schema.CascadeCASCADE),
    schema.RelatedName("user"))
```

### 3. `schema.ManyToManyField` (Many-to-Many)

Establishes an M:N association across an intermediary through table:

```go
schema.ManyToManyField("tags", "Tag", schema.Through("post_tags"),
    schema.RelatedName("posts"))
```

Create `post_tags` in a manual migration; migration generation does not create
the through table from this declaration.

---

## Cascade Deletion Behaviors

The `CascadeType` parameter defines what happens to dependent rows when a referenced record is deleted:

| Cascade Type | Behavior | Best Use Case |
| :--- | :--- | :--- |
| `schema.CascadeCASCADE` | Recursively deletes dependent rows. | Child line items, sub-components, order items. |
| `schema.CascadePROTECT` | Blocks deletion and raises an integrity error if dependents exist. | Orders referencing a customer, invoices referencing a company. |
| `schema.CascadeSET_NULL` | Sets the foreign key column to `NULL`. Requires nullable field. | Optional associations like assigning a support ticket to an agent. |
| `schema.CascadeSET_DEFAULT` | Sets the foreign key column to its SQL default value. | Reassigning orphaned content to a default system account. |
| `schema.CascadeDO_NOTHING` | Takes no database action. | Advanced databases with custom database-level trigger scripts. |

---

## Relation Options

- `RelatedName(name string)`: Stores the reverse relation name in schema metadata; it does not generate a Go accessor such as `user.Posts`.
- `RelatedQueryName(name string)`: Sets the query lookup name for reverse filters.
- `Through(tableName string)`: Specifies the junction table for `ManyToMany` relationships.
- `DBConstraint(bool)`: Stores the requested constraint setting in relation metadata. Migration generation currently does not honor `false`; it still emits the foreign key constraint. Do not use this option to disable database enforcement.

