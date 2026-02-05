---
id: code-generation
title: Code Generation
sidebar_label: Code Generation
---

# Code Generation Feature

## Overview

The Code Generation system automatically generates type-safe Go code from schema definitions. It uses AST parsing to extract model definitions and generates model structs, managers, querysets, and field expressions.

## Location

`forge/codegen/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. AST Parser (`ast_parser.go`)

**Purpose:** Parse Go source code to extract schema definitions

**Features:**
- Go AST parsing
- Schema method extraction
- Field definition parsing
- Relation definition parsing
- Meta option parsing
- Hook extraction

**Parsed Information:**
- Model struct definitions
- Field definitions with all options
- Relationship definitions
- Meta options
- Lifecycle hooks

### 2. Code Generator (`generator.go`)

**Purpose:** Generate type-safe Go code from schema definitions

**Generated Code:**
- Model structs with proper types
- FieldExpr definitions for type-safe field access
- Manager with CRUD operations
- QuerySet wrappers with type safety
- Field accessors and setters

**Generation Process:**
1. Parse schema definitions
2. Analyze model structure
3. Generate model structs
4. Generate field expressions
5. Generate manager code
6. Generate queryset code
7. Write generated files

### 3. Code Writer (`writer.go`)

**Purpose:** Write generated code to files

**Features:**
- File writing with formatting
- Import management
- Code formatting (goimports)
- Error handling
- Backup and restore

### 4. Templates (`templates.go`, `templates/`)

**Purpose:** Code generation templates

**Templates:**
- Model struct template
- FieldExpr template
- Manager template
- QuerySet template
- Field accessor template

**Template Features:**
- Go template syntax
- Type-safe generation
- Customizable output
- Import management

## Features

### ✅ Complete Features

1. **AST Parsing** - Parse Go source code to extract schemas
2. **Model Generation** - Generate model structs
3. **FieldExpr Generation** - Generate type-safe field expressions
4. **Manager Generation** - Generate managers with CRUD operations
5. **QuerySet Generation** - Generate QuerySet wrappers
6. **Type Safety** - Full type safety in generated code
7. **Import Management** - Automatic import management
8. **Code Formatting** - Automatic code formatting
9. **Error Handling** - Comprehensive error handling
10. **Template System** - Flexible template system

## Usage Examples

### Basic Schema Definition

```go
// models/user.go
package models

import (
    "github.com/forgego/forge/schema"
    "github.com/forgego/forge/schema/fields"
)

type User struct {
    schema.BaseSchema
}

func (User) Fields() []schema.Field {
    return []schema.Field{
        fields.Int64("id").Primary().AutoIncrement().Build(),
        fields.String("username").Required().Unique().MaxLength(150).Build(),
        fields.String("email").Required().Unique().MaxLength(254).Build(),
        fields.Bool("is_active").Default(true).Build(),
        fields.Time("created_at").AutoNowAdd().Build(),
    }
}

func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",
        VerboseName: "User",
        VerboseNamePlural: "Users",
    }
}
```

### Generate Code

```go
import (
    "github.com/forgego/forge/codegen"
)

// Parse models
parser := codegen.NewASTParser()
defs, err := parser.ParseDirectory("./models")
if err != nil {
    log.Fatal(err)
}

// Generate code
generator := codegen.NewGenerator()
err = generator.Generate(defs, "./generated")
if err != nil {
    log.Fatal(err)
}
```

### Generated Code Structure

**Generated Model:**
```go
// generated/models/user_gen.go
package models

type User struct {
    ID        int64     `db:"id"`
    Username  string    `db:"username"`
    Email     string    `db:"email"`
    IsActive  bool      `db:"is_active"`
    CreatedAt time.Time `db:"created_at"`
}
```

**Generated FieldExpr:**
```go
// generated/models/user_fields.go
package models

type UserFields struct {
    ID        *orm.FieldExpr[User, int64]
    Username  *orm.FieldExpr[User, string]
    Email     *orm.FieldExpr[User, string]
    IsActive  *orm.FieldExpr[User, bool]
    CreatedAt *orm.FieldExpr[User, time.Time]
}

var UserFields = &UserFields{
    ID:        orm.NewFieldExpr[User, int64]("id"),
    Username: orm.NewFieldExpr[User, string]("username"),
    Email:     orm.NewFieldExpr[User, string]("email"),
    IsActive:  orm.NewFieldExpr[User, bool]("is_active"),
    CreatedAt: orm.NewFieldExpr[User, time.Time]("created_at"),
}
```

**Generated Manager:**
```go
// generated/models/user_manager.go
package models

type UserManager struct {
    *orm.Manager[User]
}

var UserObjects = &UserManager{
    Manager: orm.NewManager[User](db, "users", schema),
}
```

**Generated QuerySet:**
```go
// generated/models/user_queryset.go
package models

type UserQuerySet struct {
    *orm.BaseQuerySet[User]
}

func (m *UserManager) Filter(expr orm.Expression) *UserQuerySet {
    return &UserQuerySet{
        BaseQuerySet: m.Manager.Filter(expr),
    }
}
```

### Using Generated Code

```go
// Type-safe queries
users, err := User.Objects.
    Filter(User.Fields.IsActive.Equals(true)).
    Filter(User.Fields.Email.Contains("@example.com")).
    OrderBy(orm.Desc("created_at")).
    Limit(10).
    All(ctx)

// Type-safe field access
user := &User{
    Username: "john",
    Email:    "john@example.com",
    IsActive:  true,
}

// Create
err = User.Objects.Create(ctx, user)

// Update
user.IsActive = false
err = User.Objects.Update(ctx, user)

// Delete
err = User.Objects.Delete(ctx, user)
```

## Integration Points

### Schema System
- Generates code from schema definitions
- Preserves all schema options
- Maintains type safety

### ORM System
- Generates ORM-compatible code
- Type-safe QuerySet generation
- Manager integration

### Admin System
- Generated code works with admin
- Field expressions for admin
- Type-safe admin configuration

### API Framework
- Generated models for serialization
- Type-safe API endpoints
- Serializer generation

## Generation Process

### 1. Parsing Phase
- Parse Go source files
- Extract schema definitions
- Build model definitions

### 2. Analysis Phase
- Analyze model structure
- Resolve relationships
- Validate schema definitions

### 3. Generation Phase
- Generate model structs
- Generate field expressions
- Generate manager code
- Generate queryset code

### 4. Writing Phase
- Format generated code
- Manage imports
- Write to files

## Template System

Templates use Go's `text/template` package:

**Model Template:**
```go
type {{.Name}} struct {
{{range .Fields}}
    {{.Name}} {{.Type}} `db:"{{.DBColumn}}"`
{{end}}
}
```

**FieldExpr Template:**
```go
type {{.Name}}Fields struct {
{{range .Fields}}
    {{.Name}} *orm.FieldExpr[{{.Name}}, {{.Type}}]
{{end}}
}
```

## Best Practices

1. **Schema First** - Define schemas before generating
2. **Version Control** - Commit generated code
3. **Regenerate** - Regenerate after schema changes
4. **Review Generated Code** - Review generated code
5. **Type Safety** - Use generated type-safe APIs
6. **Error Handling** - Handle generation errors
7. **Testing** - Test generated code
8. **Documentation** - Document schema definitions
9. **Incremental Generation** - Generate incrementally
10. **Backup** - Backup before regeneration

## CLI Integration

```bash
# Generate code from models
forge generate

# Generate specific model
forge generate --model User

# Generate with custom output
forge generate --output ./generated
```

## Future Enhancements

- [ ] Incremental generation
- [ ] Watch mode for auto-generation
- [ ] Custom template support
- [ ] Plugin system for generators
- [ ] Code validation
- [ ] Migration generation integration
- [ ] API generation from models
- [ ] Admin generation from models
