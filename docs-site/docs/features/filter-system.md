---
id: filter-system
title: Filter System
sidebar_label: Filter System
---

# Filter System Feature

## Overview

The Filter System provides advanced filtering capabilities with AST support, query parsing, expression conversion, and security validation. It enables complex filtering of QuerySets through query parameters or programmatic API.

## Location

`forge/filter/`

## Status

✅ **Complete** - Production ready

## Core Components

### 1. FilterSet (`filterset.go`)

**Purpose:** Main entry point for filtering QuerySets

**Structure:**
```go
type FilterSet[T any] struct {
    schema    *orm.ModelSchema
    filters   map[string]Filter[T]
    ast       *FilterNode
    security  *SecurityConfig
    optimizer *QueryOptimizer
    queryset  orm.QuerySet[T]
}
```

**Features:**
- Type-safe filtering with generics
- Filter registration and management
- AST-based filtering
- Query parameter parsing
- Security validation
- Query optimization

### 2. Filter Interface (`filter.go`)

**Purpose:** Base interface for all filter types

**Methods:**
- `Parse(value string)` - Parse query parameter
- `Apply(ctx, qs, value)` - Apply filter to QuerySet
- `ToAST(fieldPath, value)` - Convert to AST node
- `ToExpression(fieldPath, value)` - Convert to ORM expression
- `GetWidget()` - Get admin UI widget
- `GetOptions(ctx, qs)` - Get filter options
- `ValidateValue(value)` - Validate filter value

### 3. Filter AST (`ast.go`)

**Purpose:** Abstract Syntax Tree representation of filters

**Structure:**
```go
type FilterNode struct {
    Op       FilterOp      // and, or, not, field
    Field    string        // Field path
    Lookup   string        // Lookup type (eq, gt, contains, etc.)
    Value    interface{}   // Filter value
    Children []*FilterNode // Child nodes for boolean operations
    Metadata *FilterMetadata
}
```

**Operations:**
- `OpAnd` - AND operation
- `OpOr` - OR operation
- `OpNot` - NOT operation
- `OpField` - Field filter

### 4. Query Parser (`parser.go`)

**Purpose:** Parse query parameters into Filter AST

**Features:**
- Query string parsing
- Field path validation
- Lookup type parsing
- Value parsing and validation
- Security validation
- Error handling

**Supported Query Formats:**
```
?field=value                    // Simple equality
?field__lookup=value            // Lookup (gt, lt, contains, etc.)
?field1=value1&field2=value2    // Multiple filters (AND)
?filter=json                    // JSON filter AST
```

### 5. Expression Converter (`expression_converter.go`)

**Purpose:** Convert Filter AST to ORM expressions

**Features:**
- AST to Expression conversion
- Field path resolution
- Lookup type mapping
- Type coercion
- Error handling

### 6. Security (`security.go`)

**Purpose:** Security validation for filters

**Features:**
- Field whitelist/blacklist
- Lookup type restrictions
- Query complexity limits
- Cost estimation
- Role-based filtering

### 7. Query Optimizer (`optimizer.go`)

**Purpose:** Optimize filter queries

**Features:**
- Query plan analysis
- Index usage optimization
- Filter ordering
- Cost estimation

### 8. Filter Types (`filters/`)

**Available Filters:**
- `BooleanFilter` - Boolean field filtering
- `ChoiceFilter` - Choice field filtering
- `DateFilter` - Date field filtering
- `NumberFilter` - Number field filtering
- `TextFilter` - Text field filtering
- `RelatedFilter` - Related model filtering
- `CustomFilter` - Custom filter implementation

### 9. Filter Widgets (`widgets/`)

**Purpose:** UI widgets for filter rendering

**Widgets:**
- Text input
- Select dropdown
- Date picker
- Number range
- Boolean checkbox
- Multi-select

## Features

### ✅ Complete Features

1. **Type-Safe Filtering** - Full type safety with generics
2. **AST-Based Filtering** - Abstract syntax tree for complex filters
3. **Query Parsing** - Parse query parameters into filters
4. **Expression Conversion** - Convert filters to ORM expressions
5. **Security Validation** - Field and lookup validation
6. **Query Optimization** - Optimize filter queries
7. **Multiple Filter Types** - Boolean, Choice, Date, Number, Text, Related
8. **Custom Filters** - Custom filter implementation
9. **Filter Widgets** - UI widgets for filter rendering
10. **Filter Persistence** - Save and load filter configurations
11. **Filter Sharing** - Share filter configurations

## Usage Examples

### Basic FilterSet

```go
import "github.com/forgego/forge/filter"

// Create FilterSet
fs, err := filter.NewFilterSet[User]()
if err != nil {
    log.Fatal(err)
}

// Add filters
fs.AddFilter("is_active", filter.NewBooleanFilter[User](
    filter.BoolField("is_active", getter, setter),
))

fs.AddFilter("email", filter.NewTextFilter[User](
    filter.StringField("email", getter, setter),
))

// Apply to QuerySet
baseQS := User.Objects
filteredQS, err := fs.WithQueryset(baseQS).
    ApplyQueryParams(ctx, r)
if err != nil {
    log.Fatal(err)
}

users, err := filteredQS.All(ctx)
```

### Query Parameter Filtering

```go
// Query: ?is_active=true&email__contains=@example.com
filteredQS, err := fs.WithQueryset(User.Objects).
    ApplyQueryParams(ctx, r)

// Equivalent to:
// User.Objects.
//     Filter(User.Fields.IsActive.Equals(true)).
//     Filter(User.Fields.Email.Contains("@example.com"))
```

### Programmatic Filtering

```go
// Build filter AST
ast := filter.NewAndNode(
    filter.NewFieldNode("is_active", "eq", true),
    filter.NewFieldNode("email", "contains", "@example.com"),
)

// Apply AST
filteredQS, err := fs.WithQueryset(User.Objects).
    SetAST(ast).
    ApplyAST(ctx, ast)
```

### Custom Filter

```go
// Register custom filter
err := filter.RegisterCustom("recent_users", &filter.CustomFilterHandler{
    ID:      "recent_users",
    Name:    "Recent Users",
    Handler: func(value interface{}) (orm.Expression, error) {
        days, ok := value.(int)
        if !ok {
            return nil, fmt.Errorf("invalid value type")
        }
        cutoff := time.Now().AddDate(0, 0, -days)
        return User.Fields.CreatedAt.GreaterThan(cutoff), nil
    },
})

// Use custom filter
fs.AddFilter("recent", filter.NewCustomFilter[User](
    "created_at",
    "recent_users",
))
```

### Security Configuration

```go
// Configure security
security := filter.NewSecurityConfig()
security.AllowField("is_active")
security.AllowField("email")
security.BlockField("password")
security.AllowLookup("eq", "contains")
security.BlockLookup("raw_sql")
security.SetMaxComplexity(10)

// Apply security
fs.WithSecurity(security)
```

### Filter Persistence

```go
// Save filter
filterConfig := fs.GetAST()
data, err := json.Marshal(filterConfig)
if err != nil {
    log.Fatal(err)
}
// Save to database or cache

// Load filter
var ast filter.FilterNode
err = json.Unmarshal(data, &ast)
if err != nil {
    log.Fatal(err)
}

fs.SetAST(&ast)
```

## Integration Points

### ORM System
- Filters convert to ORM expressions
- QuerySet integration
- Type-safe field access

### Admin System
- Filter widgets in admin UI
- Filter sidebar integration
- Filter persistence

### API Framework
- Query parameter filtering
- Filter serialization
- Filter documentation

### HTTP Server
- Query parameter parsing
- Request context
- Response formatting

## Filter Lookups

**Supported Lookups:**
- `eq` - Equals
- `ne` - Not equals
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `in` - In list
- `nin` - Not in list
- `contains` - Contains substring
- `icontains` - Case-insensitive contains
- `startswith` - Starts with
- `endswith` - Ends with
- `isnull` - Is null
- `isnotnull` - Is not null
- `range` - Range (for numbers/dates)

## Security Features

### Field Whitelist/Blacklist
- Allow only specific fields
- Block sensitive fields
- Field path validation

### Lookup Restrictions
- Allow only safe lookups
- Block dangerous operations
- Cost-based restrictions

### Query Complexity
- Maximum complexity limits
- Cost estimation
- Query timeout

### Role-Based Filtering
- Role-based field access
- Permission-based filtering
- Custom security rules

## Best Practices

1. **Always Validate** - Validate all filter inputs
2. **Use Security Config** - Configure security restrictions
3. **Whitelist Fields** - Only allow safe fields
4. **Limit Complexity** - Set complexity limits
5. **Type Safety** - Use type-safe filters
6. **Error Handling** - Handle filter errors properly
7. **Performance** - Use query optimization
8. **Documentation** - Document filter usage
9. **Testing** - Test filter combinations
10. **Security** - Regular security audits

## Future Enhancements

- [ ] Full-text search filters
- [ ] Geographic filters
- [ ] Advanced date range filters
- [ ] Filter templates
- [ ] Filter analytics
- [ ] Filter recommendations
- [ ] Visual filter builder
