---
sidebar_position: 2
description: Complete QuerySet API reference. Type-safe database queries with Filter, OrderBy, Aggregations, and more. All QuerySet methods documented.
keywords:
  - forge queryset api
  - queryset reference
  - database queries
  - type-safe queries
image: /forge-social-card.svg
---

# QuerySet

The `QuerySet` type provides a type-safe way to query your database. All QuerySet methods return a new QuerySet instance, allowing method chaining.

## Methods

### Filtering

#### Filter

Add filter conditions:

```go
qs := User.Objects.Filter(User.Fields.IsActive.Equals(true))
```

Multiple filters are combined with AND:

```go
qs := User.Objects.
    Filter(User.Fields.IsActive.Equals(true)).
    Filter(User.Fields.IsStaff.Equals(true))
```

#### Exclude

Exclude objects matching conditions:

```go
qs := User.Objects.Exclude(User.Fields.IsDeleted.Equals(true))
```

### Ordering

#### OrderBy

Order results:

```go
qs := User.Objects.OrderBy("username")
qs := User.Objects.OrderBy("username", "-date_joined")
```

Use `-` prefix for descending order.

### Limiting

#### Limit

Limit number of results:

```go
qs := User.Objects.Limit(10)
```

#### Offset

Skip results:

```go
qs := User.Objects.Offset(20)
```

### Distinct

Get distinct results:

```go
qs := User.Objects.Distinct()
```

### Field Selection

#### Only

Select only specified fields:

```go
qs := User.Objects.Only("username", "email")
```

#### Defer

Exclude specified fields:

```go
qs := User.Objects.Defer("password", "secret")
```

### Relations

#### SelectRelated

Use JOIN for foreign keys:

```go
qs := Post.Objects.SelectRelated("author")
```

#### PrefetchRelated

Prefetch many relations:

```go
qs := User.Objects.PrefetchRelated("posts", "profile")
```

### Execution

#### All

Get all results:

```go
users, err := User.Objects.All(ctx)
```

#### Get

Get single result by ID:

```go
user, err := User.Objects.Get(ctx, 1)
```

Returns error if not found or multiple results.

#### First

Get first result:

```go
user, err := User.Objects.First(ctx)
```

#### Last

Get last result:

```go
user, err := User.Objects.Last(ctx)
```

#### Count

Count results:

```go
count, err := User.Objects.Count(ctx)
```

#### Exists

Check if any results exist:

```go
exists, err := User.Objects.Exists(ctx)
```

### Aggregations

#### Aggregate

Perform aggregations:

```go
result, err := User.Objects.
    Aggregate(
        aggregates.Count("id"),
        aggregates.Avg("age"),
        aggregates.Max("date_joined"),
    ).
    Get(ctx)
```

### Updates

#### Update

Bulk update:

```go
affected, err := User.Objects.
    Filter(User.Fields.IsActive.Equals(false)).
    Update(ctx, map[string]interface{}{
        "is_active": true,
    })
```

### Deletion

#### Delete

Bulk delete:

```go
affected, err := User.Objects.
    Filter(User.Fields.IsDeleted.Equals(true)).
    Delete(ctx)
```

## Field Expressions

### Equality

```go
User.Fields.Username.Equals("john")
User.Fields.Age.NotEquals(18)
```

### Comparison

```go
User.Fields.Age.Greater(18)
User.Fields.Age.GreaterOrEqual(18)
User.Fields.Age.Less(65)
User.Fields.Age.LessOrEqual(65)
```

### Null Checks

```go
User.Fields.LastLogin.IsNull()
User.Fields.LastLogin.IsNotNull()
```

### Membership

```go
User.Fields.Status.In("active", "pending")
User.Fields.Status.NotIn("deleted", "banned")
```

### String Operations

```go
User.Fields.Username.Contains("john")
User.Fields.Username.StartsWith("admin")
User.Fields.Username.EndsWith(".com")
User.Fields.Username.IContains("JOHN")
```

### Range

```go
User.Fields.Age.Range(18, 65)
User.Fields.CreatedAt.Range(startDate, endDate)
```

### Combining Conditions

```go
User.Fields.IsActive.Equals(true).And(User.Fields.IsStaff.Equals(true))
User.Fields.IsActive.Equals(true).Or(User.Fields.IsStaff.Equals(true))
User.Fields.IsActive.Equals(true).Not()
```

## See Also

- [Manager Reference](/docs/api-reference/manager) - Manager methods
- [Queries Guide](/docs/guides/queries) - Query usage guide

