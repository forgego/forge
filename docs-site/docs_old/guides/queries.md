---
sidebar_position: 2
description: Query data with the QuerySet API.
keywords:
  - forge queries
  - queryset
  - query building
image: /forge-social-card.svg
---

# Queries Guide

Use the QuerySet API to read and filter data.

## Basic queries

```go
// Get all
users, err := UserObjects.All(ctx)

// Get by ID
user, err := UserObjects.Get(ctx, 1)
```

## Filtering

```go
users, err := UserObjects.
    Filter(UserFieldsInstance.IsActive.Equals(true)).
    All(ctx)
```

## Ordering

```go
users, err := UserObjects.
    OrderBy("username", "-created_at").
    All(ctx)
```

## Limiting

```go
users, err := UserObjects.
    Limit(10).
    Offset(20).
    All(ctx)
```

## Distinct

```go
users, err := UserObjects.Distinct().All(ctx)
```

## Field selection

```go
users, err := UserObjects.Only("username", "email").All(ctx)
users, err := UserObjects.Defer("password").All(ctx)
```

## Aggregation

```go
count, err := UserObjects.Count(ctx)
exists, err := UserObjects.Exists(ctx)
```

## Next steps

- [Models guide](/docs/guides/models)
- [Admin guide](/docs/guides/admin)
- [API reference](/docs/api-reference/queryset)
