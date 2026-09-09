---
sidebar_position: 5
description: API Reference for model relationships, through models, and cascade options.
---

# Relations API Reference

The `github.com/forgego/forge/schema` package provides expressive constructors for declaring relational links between models.

---

## Relation Types

Forge supports three relational cardinality patterns:

### 1. `schema.ForeignKey` (Many-to-One)

Establishes a foreign key column on the model pointing to a target model's primary key:

```go
schema.ForeignKey("author_id", "User", schema.CascadeCASCADE).
    RelatedName("posts").
    RelatedQueryName("post")
```

### 2. `schema.OneToOne` (One-to-One)

Establishes a foreign key with an automatic `UNIQUE` database constraint, guaranteeing a strict 1:1 relationship:

```go
schema.OneToOne("profile_id", "UserProfile", schema.CascadeCASCADE).
    RelatedName("user")
```

### 3. `schema.ManyToMany` (Many-to-Many)

Establishes an M:N association across an intermediary through table:

```go
schema.ManyToMany("tags", "Tag").
    Through("post_tags").
    RelatedName("posts")
```

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

- `RelatedName(name string)`: Sets the reverse accessor name on the target model (e.g. `user.Posts`).
- `RelatedQueryName(name string)`: Sets the query lookup name for reverse filters.
- `Through(tableName string)`: Specifies the junction table for `ManyToMany` relationships.
- `DBConstraint(bool)`: If `false`, omits the foreign key constraint from migrations while maintaining ORM JOIN capabilities.

