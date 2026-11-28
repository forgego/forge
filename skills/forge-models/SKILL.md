---
name: forge-models
description: Define and update Forge data models using `github.com/forgego/forge/schema` functional builders, relations, `Meta` constraints/indexes, and `ModelHooks`. Use when creating or modifying model structs, fields, relations, constraints, or lifecycle hooks. Do NOT use for querying models with `QuerySet`/`Manager` (use `forge-orm`) or running SQL migrations (use `forge-migrations-schema`).
license: MIT
compatibility: Go 1.26.10+, github.com/forgego/forge v0.2.x
metadata:
  author: forgego
  version: "0.2"
---

# Forge Models

Models are defined in Go using `github.com/forgego/forge/schema` functional builders. `forge generate` AST-parses the model file and emits `<Model>Generated`, `<Model>Objects`, and `<Model>FieldsInstance` in `gen.go`.

## Golden Path

```go
package blog

import "github.com/forgego/forge/schema"

type Post struct {
	PostGenerated
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.StringField("slug", schema.Required(), schema.Unique(), schema.MaxLength(200)),
		schema.TextField("body", schema.Required()),
		schema.BoolField("published", schema.Default(false)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (Post) Meta() schema.Meta {
	return schema.Meta{
		TableName: "posts",
		OrderBy:   []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_posts_published", Fields: []string{"published"}},
		},
	}
}

func (Post) Relations() []schema.Relation { return nil }
func (Post) Hooks() *schema.ModelHooks  { return nil }
```

## Hard Guardrails

- **Embed `<Model>Generated`**, which embeds `schema.BaseSchema` and gives your struct typed fields and `json`/`db`/`validate` tags.
- **Same-file rule**: `Fields()`, `Meta()`, `Relations()`, and `Hooks()` must live in the same `.go` file as the model struct (`codegen/ast_parser.go` parses each file independently).
- **Functional builders only**: Use `schema.StringField("name", schema.Required(), schema.MaxLength(100))` — never fluent `.Build()` chains (they do not exist).
- **`int64` `id` PK required for `Manager` & `--api`**: `orm.Manager.Create` returns `NotImplemented` for non-integer keys, and `forge generate --api` requires a single auto-increment `int64` primary key named `id`.
- **Hook execution order**: `BeforeCreate`/`BeforeUpdate` -> `BeforeSave` -> `Clean` -> SQL -> `AfterCreate`/`AfterSave` (`orm/hook_order_contract_test.go`). `Clean` takes `func(instance interface{}) error` (no `context.Context`); there is no `Validate` hook on `schema.ModelHooks`.
- **Relations are not auto-added as struct fields**: `forge generate` does not add `post.Author` or `user.Posts` fields to `<Model>Generated`; declare relation struct fields manually if you load them with `SelectRelated` / `PrefetchRelated`.
- **After every model edit**: Run `forge generate --models ./app/<pkg> --output ./app/<pkg>`, then `forge makemigrations <name> --auto --models ./app/<pkg>` and `forge migrate up`.

## References

- [Models, Meta & constraints](references/models.md)
- [Field constructors & SQL types](references/fields.md)
- [Relations & cascade options](references/relations.md)
- [Lifecycle hooks & validation](references/hooks.md)
