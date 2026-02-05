# Type-Safe ORM API

The ORM now supports both string-based (dynamic) and fully type-safe (generic) APIs using the same method names. This provides compile-time type safety while maintaining backward compatibility.

## Features

- **Unified API**: Same method names work with both strings and `FieldExpression[T]`
- **Type Safety**: Compile-time checking when using `FieldExpression[T]`
- **Backward Compatible**: All existing string-based code continues to work
- **Mixed Usage**: Can mix strings and `FieldExpression` in the same call
- **Fluent API**: `.Asc()` and `.Desc()` methods for ordering

## Usage Examples

### Type-Safe Approach

```go
// Get field accessor
fa, _ := manager.GetFieldAccessor()

// Create type-safe field expressions
priceField := orm.FieldFor[Book, float64](fa, "price")
titleField := orm.FieldFor[Book, string](fa, "title")

// Use in queries
qs.Select(priceField, titleField)  // Type-safe!
qs.OrderBy(priceField.Desc(), titleField.Asc())  // Fluent API!
qs.Values(priceField, titleField)  // Type-safe!
```

### Dynamic Approach (Still Works)

```go
// String-based - still works as before
qs.Select("price", "title")
qs.OrderBy(Asc("price"), Desc("title"))
qs.Values("price", "title")
```

### Mixed Approach

```go
// Mix both in the same call!
qs.Select(priceField, "description")  // Type-safe + string
qs.OrderBy(priceField.Desc(), Asc("title"))  // Mixed ordering
```

### Generated Fields (After Code Generation)

```go
// After code generation, you get:
// var BookFields = BookFields{
//     Price: orm.NewField[float64]("price", "books"),
//     Title: orm.NewField[string]("title", "books"),
// }

// Use generated fields directly
qs.Select(BookFields.Price, BookFields.Title)
qs.OrderBy(BookFields.Price.Desc(), BookFields.Title.Asc())
```

## Available Methods

All these methods now accept both `string` and `FieldExpression[T]`:

- `Select(fields ...any) QuerySet[T]`
- `Only(fields ...any) QuerySet[T]`
- `Defer(fields ...any) QuerySet[T]`
- `Distinct(fields ...any) QuerySet[T]`
- `Values(fields ...any) ValuesQuerySet[T]`
- `ValuesList(fields ...any) ValuesListQuerySet[T]`
- `SelectRelated(relations ...any) QuerySet[T]`
- `PrefetchRelated(relations ...any) QuerySet[T]`
- `OrderBy(fields ...any) QuerySet[T]`

## Ordering with .Asc() and .Desc()

```go
// Type-safe ordering
field := orm.FieldFor[Book, float64](fa, "price")
qs.OrderBy(field.Asc())   // Ascending
qs.OrderBy(field.Desc())  // Descending

// Or with generated fields
qs.OrderBy(BookFields.Price.Asc(), BookFields.Title.Desc())
```

## Type-Safe Relations

```go
// Create relation expression
authorRel := orm.NewRelationExpression("author")

// Use in queries
qs.SelectRelated(authorRel)
qs.PrefetchRelated(authorRel)

// Or mix with strings
qs.SelectRelated(authorRel, "publisher")  // Mixed!
```

## Migration Guide

1. **Existing code continues to work** - no changes needed
2. **Gradually adopt type-safe APIs** where beneficial
3. **Use code generation** to get ready-to-use field expressions
4. **Mix both approaches** as needed in the same codebase

## Benefits

- **Compile-time Safety**: Catch field name typos at compile time
- **IDE Support**: Autocomplete for field names
- **Refactoring Safety**: Renaming fields updates all references
- **Flexibility**: Still support dynamic queries when needed
- **No Breaking Changes**: Existing code works as-is
