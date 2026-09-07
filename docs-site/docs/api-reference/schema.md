---
sidebar_position: 1
---

# Schema

Define models and fields using the schema system.

## Key types

- `schema.BaseSchema`
- `schema.Field`
- `schema.Relation`

## Model contract

`BaseSchema` models implement:

- `Fields() []schema.Field`
- `Relations() []schema.Relation`
- `Meta() schema.Meta`
- `Hooks() *schema.ModelHooks`
