---
sidebar_position: 11
description: Type-safe querying with QuerySet and Manager APIs.
image: /forge-social-card.svg
---

# ORM

The ORM provides type-safe queries through `QuerySet` and `Manager` APIs, with support for filtering, relations, aggregation, and updates.

## QuerySet API

Core capabilities:

- Filtering: `Filter`, `Exclude`
- Ordering: `OrderBy`, `Reverse`
- Limits: `Limit`, `Offset`, `Distinct`
- Field selection: `Select`, `Only`, `Defer`
- Relations: `SelectRelated`, `PrefetchRelated`
- Aggregates: `Aggregate`, `Annotate`
- Values: `Values`, `ValuesList`
- Execution: `All`, `Get`, `First`, `Last`, `Count`, `Exists`
- Updates: `Update`, `UpdateBuilder`, `BulkUpdate`
- Deletes: `Delete`
- Set operations: `Union`, `Intersection`, `Difference`

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

## Aggregation and annotation

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
