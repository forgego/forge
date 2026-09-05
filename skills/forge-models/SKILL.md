---
name: forge-models
description: Define and update Forge models using the schema system. Use when creating or modifying model structs, fields, relations, Meta options, or hooks, or when guidance is needed for model codegen inputs.
---

# Forge Models

## Overview
Define data models using `github.com/forgego/forge/schema`. This skill focuses on model structure that drives code generation: fields, relations, metadata, and hooks.

## When to Use
- The task mentions model structs, fields, relations, Meta options, or hooks.
- You need to change model schemas or codegen inputs.
- The user asks about renames, constraints, or lifecycle hooks.

## Quick Start
1. Create or edit a model struct embedding `schema.BaseSchema`.
2. Implement `Fields()`, `Meta()`, `Relations()`, and `Hooks()`.
3. Run `forge generate`, then `forge makemigrations` and `forge migrate`.

## Common Tasks
- Add or change fields and options (required, default, unique, max length).
- Add relations (ForeignKey, OneToOne, ManyToMany).
- Configure `Meta()` (table name, ordering, indexes, unique constraints).
- Add lifecycle hooks in `Hooks()`.

## Gotchas
- Renaming fields or relations requires migrations and can affect existing data.
- Always run `forge generate` after model changes.
- Model changes drive admin and API behavior, so keep naming consistent.

## References
- [Models guide](references/models.md)
- [Schema API](references/schema.md)
- [Fields API](references/fields.md)
- [Relations API](references/relations.md)
- [Hooks API](references/hooks.md)
