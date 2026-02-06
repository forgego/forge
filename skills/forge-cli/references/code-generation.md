---
sidebar_position: 1
---

# Code Generation

forge uses AST-based code generation to create type-safe code from your schema definitions.

## How It Works

1. **Parse Models** - AST parser reads your Go model files
2. **Extract Schema** - Extracts field definitions, relations, meta, hooks
3. **Generate Code** - Creates type-safe managers, querysets, and field expressions
4. **Write Files** - Writes generated files to `models/*.gen.go`

## Generated Files

### Model Struct

`models/post.gen.go`:

```go
type Post struct {
    ID          int64
    Title       string
    Content     string
    Published   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    AuthorID    int64
    Author      *User
    Categories  []*Category
}
```

### Field Expressions

`models/post_fields.gen.go`:

```go
type PostFields struct {
    ID        query.FieldExpr[int64]
    Title     query.FieldExpr[string]
    Content   query.FieldExpr[string]
    Published query.FieldExpr[bool]
    CreatedAt query.FieldExpr[time.Time]
    UpdatedAt query.FieldExpr[time.Time]
}

var PostFields = PostFields{
    ID:        query.NewFieldExpr[int64]("id"),
    Title:     query.NewFieldExpr[string]("title"),
    Content:   query.NewFieldExpr[string]("content"),
    Published: query.NewFieldExpr[bool]("published"),
    CreatedAt: query.NewFieldExpr[time.Time]("created_at"),
    UpdatedAt: query.NewFieldExpr[time.Time]("updated_at"),
}
```

### Manager

`models/post_manager.gen.go`:

```go
type PostManagerType struct {
    db *db.DB
}

var Post = PostManagerType{}

func (m *PostManagerType) Create(ctx context.Context, instance *Post) error {
    // Create implementation
}

func (m *PostManagerType) Get(ctx context.Context, id int64) (*Post, error) {
    // Get implementation
}

func (m *PostManagerType) Update(ctx context.Context, instance *Post) error {
    // Update implementation
}

func (m *PostManagerType) Delete(ctx context.Context, instance *Post) error {
    // Delete implementation
}

func (m *PostManagerType) Filter(conditions ...query.QueryExpr) *PostQuerySet {
    // Filter implementation
}
```

### QuerySet

`models/post_queryset.gen.go`:

```go
type PostQuerySet struct {
    *query.BaseQuerySet[Post]
}

func (qs *PostQuerySet) All(ctx context.Context) ([]*Post, error) {
    return qs.BaseQuerySet.All(ctx)
}

func (qs *PostQuerySet) Get(ctx context.Context, id int64) (*Post, error) {
    return qs.BaseQuerySet.Get(ctx, id)
}

// ... other methods
```

## Running Code Generation

```bash
forge generate
```

This will:
1. Scan `models/` directory for model definitions
2. Parse each model file
3. Generate code for each model
4. Write generated files

## Customizing Generation

### Custom Templates

You can customize generation templates, though this is advanced and not recommended for most users.

### Generation Options

```bash
# Generate for specific models
forge generate --models Post,User

# Output to different directory
forge generate --output ./generated
```

## Best Practices

1. **Don't Edit Generated Files** - They will be overwritten
2. **Regenerate After Model Changes** - Always regenerate after modifying models
3. **Commit Generated Files** - Include `*.gen.go` files in version control
4. **Use Type-Safe APIs** - Use generated field expressions and querysets

## Troubleshooting

### Generation Fails

- Check model syntax
- Ensure all imports are correct
- Verify schema interface implementation

### Generated Code Errors

- Regenerate code
- Check for circular dependencies
- Verify model definitions

## See Also

- [Models Guide](/docs/guides/models) - Model definitions
- [Plugins](/docs/advanced/plugins) - Extend code generation

