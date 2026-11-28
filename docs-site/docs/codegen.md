---
title: Code Generation
description: How forge generate parses Go ASTs to emit typed models, managers, fields, and REST APIs.
image: /social-card.png
---

# Code Generation

`forge generate` parses your model `.go` files using Go's AST (`codegen/ast_parser.go`) and writes deterministic, type-safe Go code without runtime reflection.

## Running `forge generate`

```bash
# Default: scans ./app (if present) or ./models, writes to ./models
forge generate

# Recommended for `forge new` apps: pass both --models and --output
forge generate --models ./app/blog --output ./app/blog

# Generate REST ViewSets, serializers, and routes (api_gen.go) too
forge generate --models ./app/blog --output ./app/blog --api

# Fail immediately on unsupported schema expressions
forge generate --models ./app/blog --output ./app/blog --strict
```

## What Gets Generated

### `gen.go`
- **`<Model>Generated`**: Struct embedding `schema.BaseSchema` with all schema fields and `json`, `db`, and `validate` tags. Embed this in your model struct.
- **`Validate()`**: Generated validation method on your model.
- **`<Model>Objects`**: Package-level `*orm.Manager[Model]` bound to `Meta().TableName` (or snake_case plural of the model name). Bind its database at startup with `<Model>Objects.SetDB(database)`.
- **`<Model>Fields` & `<Model>FieldsInstance`**: Typed `orm.Field[T]` accessors for compile-time-checked queries.

### `api_gen.go` (with `--api`)
- Generated REST serializers, `ViewSet` wiring (with `RejectUnknownRequestFields: true`), and route registration helpers.

## Rules & Caveats

- **Same-file requirement**: The AST parser inspects each `.go` file independently. `Fields()`, `Meta()`, `Relations()`, and `Hooks()` must live in the same file as the model struct.
- **`--api` primary key requirement**: `forge generate --api` refuses any model whose primary key is not a single auto-increment `int64` named `id`, plus a writable `int64` ID (declared `ID`/`Id`, embedded `<Model>Generated`, or `GetID`/`SetID` on the model itself).
- **Never hand-edit generated files**: Changes to `gen.go` or `api_gen.go` are overwritten on the next `forge generate` run. Always regenerate after model edits or Forge upgrades.
- **`forge auth` + `--api`**: Running `forge generate --api` on an `app/auth` package that already defines user serializers can duplicate symbols; inspect `api_gen.go` if customizing `app/auth`.
