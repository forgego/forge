---
sidebar_position: 3
description: Complete documentation for the Schema System - the foundation of forge. Learn about field definitions, relationships, metadata, hooks, and more.
keywords:
  - schema system
  - model definitions
  - fields
  - relationships
  - hooks
image: /forge-social-card.svg
---

# Schema System

The Schema System is the foundation of the forge framework. It provides a declarative way to define models with full Django-like field options, relationships, metadata, and lifecycle hooks.

## Location

`forge/schema/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. Schema Interface (`schema.go`)

The core interface that all models must implement:

```go
type Schema interface {
    Fields() []Field
    Relations() []Relation
    Meta() Meta
    Hooks() *ModelHooks
}
```

### 2. Field Definitions (`field.go`)

Complete field type system with all Django field options:

**Field Types:**
- `TypeInt64`, `TypeInt32` - Integer fields
- `TypeString`, `TypeText` - String fields
- `TypeBool` - Boolean fields
- `TypeTime`, `TypeDate`, `TypeDateTime` - Time fields
- `TypeFloat32`, `TypeFloat64`, `TypeDecimal` - Float fields
- `TypeEmail`, `TypeURL` - Specialized string fields
- `TypeUUID` - UUID fields
- `TypeJSON` - JSON fields
- `TypeBytes` - Byte array fields
- `TypeForeignKey`, `TypeManyToMany`, `TypeOneToOne` - Relationship fields

**Field Options:**
- `Required()` - Field is required
- `Blank()` - Field can be blank
- `Default(value)` - Default value
- `Primary()` - Primary key
- `AutoIncrement()` - Auto-incrementing
- `Unique()` - Unique constraint
- `Index()` - Database index
- `MaxLength(n)` - Maximum length
- `MinLength(n)` - Minimum length
- `Choices([]Choice)` - Predefined choices
- `HelpText(text)` - Help text
- `VerboseName(name)` - Human-readable name
- `DBColumn(name)` - Custom column name
- `DBType(type)` - Custom database type
- `DBIndex(bool)` - Database index
- `DBCollation(collation)` - Character collation
- `DBComment(comment)` - Column comment
- `Generated(expr)` - Generated column
- `Validators([]Validator)` - Custom validators
- `ValidationTag(tag)` - Validation tag
- `Editable(bool)` - Editable flag
- `AutoNow()` - Auto-update on save
- `AutoNowAdd()` - Auto-set on creation

### 3. Relationship Definitions (`relation.go`)

Django-style relationship definitions:

**Relation Types:**
- `ForeignKey` - Many-to-one relationship
- `OneToOne` - One-to-one relationship
- `ManyToMany` - Many-to-many relationship
- `OneToMany` - One-to-many relationship (reverse)

**Relation Options:**
- `Target(model)` - Target model
- `Required(bool)` - Required relationship
- `OnDelete(action)` - Cascade action on delete
- `OnUpdate(action)` - Cascade action on update
- `RelatedName(name)` - Reverse relation name
- `Through(model)` - Through model for ManyToMany

### 4. Model Metadata (`meta.go`)

Complete Django Meta options:

**Meta Options:**
- `TableName` - Custom table name
- `OrderBy` - Default ordering
- `VerboseName` - Human-readable name
- `VerboseNamePlural` - Plural name
- `Indexes` - Database indexes
- `UniqueTogether` - Unique constraints
- `Constraints` - Custom constraints
- `Permissions` - Model permissions
- `DefaultPermissions` - Use default permissions

### 5. Lifecycle Hooks (`hooks.go`)

Model lifecycle hooks for custom logic:

**Hooks:**
- `BeforeSave` - Before save (create or update)
- `AfterSave` - After save (create or update)
- `BeforeCreate` - Before create
- `AfterCreate` - After create
- `BeforeUpdate` - Before update
- `AfterUpdate` - After update
- `BeforeDelete` - Before delete
- `AfterDelete` - After delete

**Hook Execution Order:**
- Create: `BeforeSave` → `BeforeCreate` → DB Insert → `AfterCreate` → `AfterSave`
- Update: `BeforeSave` → `BeforeUpdate` → DB Update → `AfterUpdate` → `AfterSave`
- Delete: `BeforeDelete` → DB Delete → `AfterDelete`

### 6. Constraint Builder (`constraint_builder.go`)

Database constraint builders for advanced schema definitions.

## Features

### ✅ Complete Features

1. **All Django Field Types** - Complete field type system
2. **All Django Field Options** - Full field option support
3. **Relationship Definitions** - ForeignKey, OneToOne, ManyToMany
4. **Model Metadata** - Complete Meta options
5. **Lifecycle Hooks** - Before/after save, create, update, delete
6. **Database Constraints** - Indexes, unique constraints, custom constraints
7. **Validation Integration** - Validator support
8. **Code Generation Ready** - Schema designed for code generation

### Field Builder Pattern

Fields use a fluent builder pattern:

```go
fields.String("username").
    Required().
    MaxLength(150).
    Unique().
    HelpText("Required. 150 characters or fewer.").
    Build()
```

## Usage Examples

### Basic Model

```go
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
        fields.Time("updated_at").AutoNow().Build(),
    }
}

func (User) Meta() schema.Meta {
    return schema.Meta{
        TableName: "users",
        VerboseName: "User",
        VerboseNamePlural: "Users",
        OrderBy: []string{"-created_at"},
    }
}
```

### Model with Relationships

```go
type Post struct {
    schema.BaseSchema
}

func (Post) Fields() []schema.Field {
    return []schema.Field{
        fields.Int64("id").Primary().AutoIncrement().Build(),
        fields.String("title").Required().MaxLength(200).Build(),
        fields.Text("content").Required().Build(),
        fields.Bool("published").Default(false).Build(),
        fields.Time("created_at").AutoNowAdd().Build(),
    }
}

func (Post) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("author", "User").Required().OnDelete(relations.CASCADE).Build(),
        relations.ManyToMany("tags", "Tag").Build(),
    }
}
```

### Model with Hooks

```go
func (User) Hooks() *schema.ModelHooks {
    return &schema.ModelHooks{
        BeforeSave: func(ctx context.Context, instance interface{}) error {
            // Custom logic before save
            return nil
        },
        AfterCreate: func(ctx context.Context, instance interface{}) error {
            // Custom logic after create
            return nil
        },
    }
}
```

## Integration Points

### Code Generation
- Schema definitions are parsed by AST parser
- Code generator creates model structs, managers, querysets

### ORM System
- ORM uses schema for query building
- Field expressions generated from schema
- Relationships used for joins

### Admin System
- Admin uses schema for form generation
- Field widgets based on field types
- Relationships used for inlines

### Migration System
- Migrations generated from schema
- Schema changes detected automatically
- Database schema synchronized

## Extension Points

### Custom Fields
- Define custom field types
- Custom field builders
- Custom validation

### Custom Relations
- Custom relationship types
- Custom cascade behaviors

### Custom Hooks
- Additional hook types
- Custom hook execution

## Best Practices

1. **Use BaseSchema** - Embed `schema.BaseSchema` for default implementations
2. **Field Builders** - Use field builders for type safety
3. **Meta Options** - Set appropriate Meta options
4. **Hooks** - Use hooks for business logic, not data access
5. **Validation** - Use validators for data validation
6. **Naming** - Follow Django naming conventions

## Future Enhancements

- [ ] Custom field types
- [ ] Field inheritance
- [ ] Schema versioning
- [ ] Schema migrations
- [ ] Schema validation