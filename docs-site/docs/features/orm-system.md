---
id: orm-system
title: ORM System
sidebar_label: ORM System
---

# ORM System Feature

## Overview

The ORM System provides a type-safe, Django-like ORM with QuerySet API, Manager CRUD operations, field expressions, and SQL builder.

## Location

`forge/orm/`

## Status

✅ **Complete** - Production ready (with some advanced features structure ready)

## Core Components

### 1. QuerySet Interface (`queryset.go`)

Type-safe QuerySet interface:

```go
type QuerySet[T any] interface {
    // Filtering
    Filter(expr Expression) QuerySet[T]
    Exclude(expr Expression) QuerySet[T]
    
    // Ordering
    OrderBy(fields ...OrderField) QuerySet[T]
    Reverse() QuerySet[T]
    
    // Limiting
    Limit(n int) QuerySet[T]
    Offset(n int) QuerySet[T]
    Distinct(fields ...string) QuerySet[T]
    
    // Field Selection
    Select(fields ...string) QuerySet[T]
    Only(fields ...string) QuerySet[T]
    Defer(fields ...string) QuerySet[T]
    
    // Relations
    SelectRelated(relations ...string) QuerySet[T]
    PrefetchRelated(relations ...string) QuerySet[T]
    
    // Aggregation
    Aggregate(aggs ...Aggregate) QuerySet[T]
    Annotate(anns ...AnnotationExpr) QuerySet[T]
    
    // Values
    Values(fields ...string) ValuesQuerySet[T]
    ValuesList(fields ...string) ValuesListQuerySet[T]
    
    // Execution
    All(ctx context.Context) ([]*T, error)
    Get(ctx context.Context) (*T, error)
    First(ctx context.Context) (*T, error)
    Last(ctx context.Context) (*T, error)
    Count(ctx context.Context) (int64, error)
    Exists(ctx context.Context) (bool, error)
    
    // Updates
    Update(ctx context.Context, updates UpdateMap) (int64, error)
    UpdateBuilder() (*UpdateBuilder[T], error)
    BulkUpdate(ctx context.Context, updates []UpdateMap) error
    
    // Deletes
    Delete(ctx context.Context) (int64, error)
    
    // Set Operations
    Union(other QuerySet[T]) QuerySet[T]
    Intersection(other QuerySet[T]) QuerySet[T]
    Difference(other QuerySet[T]) QuerySet[T]
}
```

### 2. Manager (`manager.go`)

Type-safe Manager with CRUD operations:

```go
type Manager[T any] struct {
    db        interface{}
    table     string
    schema    *ModelSchema
}

// CRUD Operations
func (m *Manager[T]) Create(ctx context.Context, instance *T) error
func (m *Manager[T]) Update(ctx context.Context, instance *T) error
func (m *Manager[T]) Delete(ctx context.Context, instance *T) error
func (m *Manager[T]) Get(ctx context.Context, id interface{}) (*T, error)

// QuerySet Creation
func (m *Manager[T]) Filter(expr Expression) (QuerySet[T], error)
func (m *Manager[T]) All(ctx context.Context) ([]*T, error)
```

### 3. Field Expressions (`field_expr.go`)

Type-safe field accessors:

```go
type FieldExpr[T any, V any] interface {
    Equals(value V) Expression
    NotEquals(value V) Expression
    GreaterThan(value V) Expression
    GreaterThanOrEqual(value V) Expression
    LessThan(value V) Expression
    LessThanOrEqual(value V) Expression
    In(values []V) Expression
    NotIn(values []V) Expression
    IsNull() Expression
    IsNotNull() Expression
    Contains(value string) Expression
    StartsWith(value string) Expression
    EndsWith(value string) Expression
    Like(value string) Expression
    ILike(value string) Expression
}
```

### 4. Expressions (`expression.go`)

Query expression system:

```go
type Expression interface {
    ToSQL(builder *SQLBuilder) (string, []interface{}, error)
    And(other Expression) Expression
    Or(other Expression) Expression
    Not() Expression
}
```

### 5. SQL Builder (`sql_builder.go`)

Type-safe SQL generation:

- Proper identifier escaping
- Parameter binding
- Query optimization
- Dialect support

### 6. Update Builder (`update_builder.go`)

Type-safe update operations:

```go
type UpdateBuilder[T any] struct {
    // Update operations
}

func (b *UpdateBuilder[T]) Set(field string, value interface{}) *UpdateBuilder[T]
func (b *UpdateBuilder[T]) Increment(field string, value interface{}) *UpdateBuilder[T]
func (b *UpdateBuilder[T]) Decrement(field string, value interface{}) *UpdateBuilder[T]
func (b *UpdateBuilder[T]) Execute(ctx context.Context) (int64, error)
```

### 7. Aggregates (`aggregates.go`)

Aggregate functions (structure ready):

- `Count()`
- `Sum(field)`
- `Avg(field)`
- `Min(field)`
- `Max(field)`

### 8. Annotations (`annotations.go`)

Annotation support (structure ready):

- Computed fields
- Expressions
- Subqueries

## Features

### ✅ Complete Features

1. **Type-Safe QuerySet** - Full type safety with generics
2. **Filter/Exclude** - Query filtering
3. **OrderBy/Reverse** - Query ordering
4. **Limit/Offset** - Query limiting
5. **Distinct** - Distinct queries
6. **Select/Only/Defer** - Field selection
7. **All/Get/First/Last** - Query execution
8. **Count/Exists** - Aggregation queries
9. **Update** - Update operations
10. **Delete** - Delete operations
11. **Manager CRUD** - Create, Update, Delete, Get
12. **Field Expressions** - Type-safe field access
13. **SQL Builder** - Type-safe SQL generation
14. **Update Builder** - Type-safe updates
15. **Set Operations** - Union, Intersection, Difference

### 🚧 Structure Ready

1. **SelectRelated** - JOIN queries (structure ready)
2. **PrefetchRelated** - Separate queries for relations (structure ready)
3. **Aggregates** - Aggregate functions (structure ready)
4. **Annotations** - Computed fields (structure ready)
5. **Values/ValuesList** - Dictionary returns (structure ready)

## Usage Examples

### Basic Queries

```go
// Get all active users
users, err := User.Objects.
    Filter(User.Fields.IsActive.Equals(true)).
    All(ctx)

// Get first user
user, err := User.Objects.First(ctx)

// Count users
count, err := User.Objects.Count(ctx)
```

### Filtering

```go
// Filter with multiple conditions
users, err := User.Objects.
    Filter(
        User.Fields.IsActive.Equals(true).
        And(User.Fields.DateJoined.GreaterThan(lastMonth)),
    ).
    All(ctx)

// Exclude
users, err := User.Objects.
    Exclude(User.Fields.IsDeleted.Equals(true)).
    All(ctx)
```

### Ordering

```go
// Order by field
users, err := User.Objects.
    OrderBy(orm.Desc("created_at")).
    All(ctx)

// Reverse order
users, err := User.Objects.
    OrderBy(orm.Asc("username")).
    Reverse().
    All(ctx)
```

### Limiting

```go
// Limit and offset
users, err := User.Objects.
    Limit(10).
    Offset(20).
    All(ctx)
```

### Updates

```go
// Update single instance
user.IsActive = true
err := User.Objects.Update(ctx, user)

// Bulk update
updates := []orm.UpdateMap{
    {"is_active": true},
    {"is_staff": false},
}
err := User.Objects.BulkUpdate(ctx, updates)

// Update with builder
affected, err := User.Objects.
    Filter(User.Fields.IsActive.Equals(false)).
    Update(ctx, orm.UpdateMap{"is_active": true})
```

### Deletes

```go
// Delete single instance
err := User.Objects.Delete(ctx, user)

// Bulk delete
affected, err := User.Objects.
    Filter(User.Fields.IsDeleted.Equals(true)).
    Delete(ctx)
```

### Field Expressions

```go
// Type-safe field access
expr := User.Fields.Username.Equals("admin")
expr := User.Fields.Email.Contains("@example.com")
expr := User.Fields.CreatedAt.GreaterThan(time.Now().AddDate(0, -1, 0))
expr := User.Fields.IsActive.Equals(true).And(
    User.Fields.IsStaff.Equals(false),
)
```

## Integration Points

### Schema System
- Uses schema for field definitions
- Relationship definitions
- Model metadata

### Code Generation
- Generates QuerySet wrappers
- Generates Manager code
- Generates field expressions

### Admin System
- Admin uses QuerySet for data access
- Type-safe field expressions
- Manager integration

### API Framework
- ViewSets use QuerySet
- Serializers work with models
- Filter integration

### Database Layer
- Uses database connection
- Transaction support
- Migration integration

## Extension Points

### Custom QuerySet Methods
- Add custom methods to QuerySet
- Custom query logic
- Custom execution

### Custom Expressions
- Define custom expressions
- Custom SQL generation
- Custom operators

### Custom Aggregates
- Define custom aggregates
- Custom aggregate logic
- Custom SQL generation

## Best Practices

1. **Type Safety** - Always use type-safe field expressions
2. **Context** - Always pass context
3. **Error Handling** - Always check errors
4. **Performance** - Use Select/Only/Defer for optimization
5. **Transactions** - Use transactions for multiple operations
6. **Indexes** - Add indexes for frequently queried fields

## Future Enhancements

- [ ] SelectRelated implementation
- [ ] PrefetchRelated implementation
- [ ] Aggregates implementation
- [ ] Annotations implementation
- [ ] F() expressions
- [ ] Subqueries
- [ ] Window functions
- [ ] Full-text search
- [ ] Raw SQL support
