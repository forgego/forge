---
sidebar_position: 4
---

# Code Generation

forge uses code generation to create type-safe code from your schema definitions, eliminating boilerplate and ensuring type safety.

## Why code generation?

Code generation solves several problems:

- **Boilerplate reduction** - No need to write repetitive CRUD code
- **Type safety** - Generated code is type-safe at compile time
- **Consistency** - All generated code follows the same patterns
- **Maintainability** - Changes to schemas automatically update generated code

## How it works

### 1. Define your schema

You define your models declaratively:

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("title").Required().MaxLength(200).Build(),
        schema.Text("content").Required().Build(),
    }
}
```

### 2. Run code generation

```bash
forge generate
```

### 3. Generated code

forge generates:

- **Model struct** - Go struct matching your schema
- **Field expressions** - Type-safe field accessors
- **Manager** - CRUD operations
- **QuerySet** - Type-safe query builder

## Generated files

For each model, forge generates several files:

### Model struct (`post.gen.go`)

```go
type Post struct {
    ID        int64
    Title     string
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Field expressions (`post_fields.gen.go`)

```go
type PostFields struct {
    ID        *query.FieldExpr[int64]
    Title     *query.FieldExpr[string]
    Content   *query.FieldExpr[string]
    CreatedAt *query.FieldExpr[time.Time]
}

var PostFields = &PostFields{
    ID:        query.NewFieldExpr[int64]("id", "posts"),
    Title:     query.NewFieldExpr[string]("title", "posts"),
    Content:   query.NewFieldExpr[string]("content", "posts"),
    CreatedAt: query.NewFieldExpr[time.Time]("created_at", "posts"),
}
```

### Manager (`post_manager.gen.go`)

```go
type PostManager struct {
    db *db.DB
}

func (m *PostManager) Create(ctx context.Context, post *Post) error {
    // Generated CRUD implementation
}

func (m *PostManager) Get(ctx context.Context, id int64) (*Post, error) {
    // Generated get implementation
}
```

### QuerySet (`post_queryset.gen.go`)

```go
type PostQuerySet struct {
    *query.BaseQuerySet[Post]
}

func (qs *PostQuerySet) Filter(expr query.QueryExpr) *PostQuerySet {
    // Generated filter implementation
}
```

## Using generated code

Once code is generated, you use it like this:

```go
// Type-safe field access
Post.Fields.Title.Equals("Hello")

// Type-safe queries
posts, err := Post.Objects.
    Filter(Post.Fields.Published.Equals(true)).
    All(ctx)

// Type-safe CRUD
post := &Post{Title: "My Post"}
err := Post.Objects.Create(ctx, post)
```

## Regenerating code

When you change your schema, regenerate code:

```bash
forge generate
```

forge will:
1. Parse your schema definitions
2. Detect changes
3. Regenerate affected files
4. Preserve your custom code

## Custom code preservation

forge preserves your custom code:

- Custom methods on models
- Custom manager methods
- Custom QuerySet methods
- Custom hooks

Only generated sections are overwritten.

## AST parsing

forge uses Go's AST parser to extract schema information:

1. **Parse Go files** - Reads your model definitions
2. **Extract schemas** - Finds schema implementations
3. **Build model graph** - Understands relationships
4. **Generate code** - Creates type-safe code

## Code generation flow

```
Schema Definition (Go)
    ↓
AST Parser
    ↓
Model Definition
    ↓
Code Generator
    ↓
Generated Files
```

## Benefits

### Type safety

All generated code is type-safe:

```go
// ✅ Compile-time checked
Post.Fields.Title.Equals("Hello")

// ❌ Won't compile
Post.Fields.Title.Equals(123)  // Type error
```

### IDE support

Generated code provides full IDE support:

- Autocomplete
- Go to definition
- Refactoring
- Type checking

### Performance

Generated code has no runtime overhead:

- No reflection
- No dynamic dispatch
- Direct method calls
- Optimized by compiler

## Limitations

Code generation has some limitations:

- **Regeneration required** - Must regenerate after schema changes
- **Build step** - Adds a build step to your workflow
- **File management** - Generated files need to be committed

## Best practices

1. **Commit generated files** - Include them in version control
2. **Regenerate often** - After every schema change
3. **Don't edit generated files** - Your changes will be lost
4. **Use custom methods** - Add custom logic in separate files
5. **Review generated code** - Understand what's generated

## Next steps

- **[Extending forge](/docs/learn/extending-forge)** - Extend the framework
- **[Custom fields](/docs/advanced/custom-fields)** - Create custom field types
- **[Plugins](/docs/advanced/plugins)** - Build plugins
