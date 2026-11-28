# Code Generation (`forge generate`)

Forge uses Go AST parsing (`codegen/ast_parser.go`) to generate deterministic, type-safe code from your schema definitions without runtime reflection.

## How It Works

1. **Parse Models**: Reads Go source files in `--models`. Each file is parsed independently, so `Fields()`, `Meta()`, `Relations()`, and `Hooks()` **must live in the same `.go` file** as the model struct (otherwise `Meta()` is ignored and the default table name is used).
2. **Extract Schema**: Evaluates the supported subset of Go schema expressions. Use `--strict` to fail with file and line numbers on any unsupported expression.
3. **Emit `gen.go`**: Writes `<Model>Generated`, `Validate()`, `<Model>Objects` (`*orm.Manager[Model]`), `<Model>Fields`, and `<Model>FieldsInstance` to `gen.go` in `--output`.
4. **Emit `api_gen.go` (with `--api`)**: Generates REST serializers, ViewSets, and route registration helpers.

## Generated Artifacts (`gen.go`)

```go
type PostGenerated struct {
	schema.BaseSchema
	ID        int64     `json:"id" db:"id" validate:""`
	Title     string    `json:"title" db:"title" validate:"required,max=200"`
	Published bool      `json:"published" db:"published" validate:""`
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:""`
}

var PostObjects = orm.MustNewManager[Post]("posts")

type PostFields struct {
	ID        orm.Field[int64]
	Title     orm.Field[string]
	Published orm.Field[bool]
	CreatedAt orm.Field[time.Time]
}

var PostFieldsInstance = PostFields{ ... }
```

Bind `<Model>Objects.SetDB(database)` at startup before serving requests.

## Flags & Requirements

- `--models`: Source directory to scan (defaults to `./app` if it exists, otherwise `./models`).
- `--output`: Output directory for `gen.go` / `api_gen.go` (defaults to `./models`; always pass `--output` when `--models` is an app directory).
- `--api`: Also generate `api_gen.go`. Requires every model in the package to have a single auto-increment `int64` primary key named `id`, plus a writable `int64` ID (declared `ID`/`Id`, embedded `<Model>Generated`, or `GetID`/`SetID` on the model itself).
- `--strict`: Fail on any schema expression the AST parser cannot statically evaluate.

**Caution:** Both `gen.go` and `api_gen.go` are staged before replacement with best-effort rollback. Never hand-edit `*.gen.go` / `gen.go` / `api_gen.go`; re-run `forge generate` after every model change or framework upgrade.
