---
sidebar_position: 11
description: Type-safe querying with QuerySet and Manager APIs.
image: /forge-social-card.svg
---

# ORM

The ORM provides type-safe queries through `QuerySet` and `Manager` APIs, with support for filtering, relations, aggregation, and updates.

## QuerySet API (core)

- Filter, Exclude
- OrderBy, Reverse
- Limit, Offset, Distinct
- Select, Only, Defer
- SelectRelated, PrefetchRelated
- Aggregate, Annotate
- Values, ValuesList
- All, Get, First, Last, Count, Exists
- Update, UpdateBuilder, BulkUpdate
- Delete
- Union, Intersection, Difference

## Filtering

```go
posts, err := PostObjects.
    Filter(PostFieldsInstance.Published.Equals(true)).
    OrderBy(orm.Desc("created_at")).
    Limit(10).
    All(ctx)
```

## Relations

```go
posts, err := PostObjects.
    SelectRelated("author").
    PrefetchRelated("tags").
    All(ctx)
```

## Aggregation

```go
stats, err := PostObjects.
    Aggregate(orm.Count("id"), orm.Max("created_at")).
    All(ctx)
```

## Updates

```go
updated, err := PostObjects.
    Filter(PostFieldsInstance.Published.Equals(false)).
    Update(ctx, orm.UpdateMap{"published": true})
```

## Values

```go
rows, err := PostObjects.
    ValuesList(PostFieldsInstance.Id, PostFieldsInstance.Title).
    All(ctx)
```

## Expressions

Field expression helpers are generated during codegen. Use them to build type-safe filters instead of string field names.

## Next steps

- [Filters](/docs/filters/)
- [API Reference](/docs/api-reference/)
