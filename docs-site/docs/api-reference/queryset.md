---
sidebar_position: 2
description: Comprehensive API Reference for the QuerySet[T] type.
---

# QuerySet API Reference

A `QuerySet[T]` represents a collection of database queries targeting model `T`. QuerySets are immutable and chainable; every transformation method returns a new `QuerySet` instance without modifying the caller.

---

## Filtering & Slicing Methods

### `Filter(expr orm.Expression) orm.QuerySet[T]`
Adds one expression to the SQL `WHERE` clause. Combine multiple expressions with `orm.And`.

```go
qs = qs.Filter(orm.And(ProductExpr.Price.Lte(100.00), ProductExpr.InStock.Eq(true)))
```

### `Exclude(expr orm.Expression) orm.QuerySet[T]`
Negates the expression (`NOT (...)`) and appends it to the `WHERE` clause.

```go
qs = qs.Exclude(ProductExpr.Status.Eq("archived"))
```

### `OrderBy(fields ...string) *QuerySet[T]`
Sets the `ORDER BY` clause. Prefix a field name with `-` for descending order.

```go
qs = qs.OrderBy("-created_at", "title")
```

### `Limit(n int) *QuerySet[T]`
Sets the maximum number of records to return (`LIMIT n`).

### `Offset(n int) *QuerySet[T]`
Sets the number of rows to skip (`OFFSET n`).

### `Distinct(fields ...string) *QuerySet[T]`
Appends SQL `DISTINCT` (or `DISTINCT ON` in PostgreSQL).

---

## Eager Loading Methods

### `SelectRelated(fields ...string) *QuerySet[T]`
Performs SQL `JOIN`s along single-valued foreign key paths, loading parent and child rows in a single query.

```go
qs = qs.SelectRelated("Category", "Supplier")
```

### `PrefetchRelated(fields ...string) *QuerySet[T]`
Executes batch queries for multi-valued relations (`ManyToMany` or reverse foreign keys) and stitches relationships in memory.

```go
qs = qs.PrefetchRelated("Tags", "Reviews")
```

---

## Projection & Field Selection

### `Only(fields ...string) *QuerySet[T]`
Restricts the SQL `SELECT` list to only the specified columns, reducing payload size.

### `Defer(fields ...string) *QuerySet[T]`
Excludes large columns (like long text or binary blobs) from the initial `SELECT` query.

### `Values(fields ...string) *ValuesQuerySet[T]`
Returns query results as a slice of `map[string]any` dictionaries instead of model structs.

### `ValuesList(fields ...string) *ValuesListQuerySet[T]`
Returns query results as flat slices of tuples / interface values.

---

## Aggregations & Annotations

### `Aggregate(funcs ...Aggregate) QuerySet[T]`
Adds aggregates to a queryset. Use `orm.AggregateValues(ctx, qs, funcs...)` to evaluate scalar aggregates (`COUNT`, `SUM`, `AVG`, `MIN`, `MAX`).

```go
stats, err := orm.AggregateValues(ctx, qs, orm.Count("id"), orm.Avg("price"))
```

### `Annotate(anns ...orm.AnnotationExpr) orm.QuerySet[T]`
Adds computed annotation expressions to returned rows. Construct them with `orm.NewAnnotation` or `orm.NewExpressionAnnotation`.

---

## Terminal Execution Methods

Terminal methods execute the SQL statement against the active database connection or transaction.

| Method | Signature | Description |
| :--- | :--- | :--- |
| `All(ctx)` | `All(ctx context.Context) ([]*T, error)` | Evaluates the query and returns all matching records. |
| `Get(ctx)` | `Get(ctx context.Context) (*T, error)` | Returns exactly one row matching the queryset filters; errors for zero or multiple matches. Use `Manager.Get(ctx, id)` for primary-key lookup. |
| `First(ctx)` | `First(ctx context.Context) (*T, error)` | Returns the first matching record, or `nil` if empty. |
| `Last(ctx)` | `Last(ctx context.Context) (*T, error)` | Reverses ordering and returns the first record. |
| `Count(ctx)` | `Count(ctx context.Context) (int64, error)` | Executes `SELECT COUNT(*)` and returns the count. |
| `Exists(ctx)` | `Exists(ctx context.Context) (bool, error)` | Returns `true` if at least one matching row exists (`SELECT 1 ... LIMIT 1`). |

---

## Mutation Methods

### `Update(ctx context.Context, values orm.UpdateMap) (int64, error)`
Executes an SQL `UPDATE` across all rows matching the QuerySet filters. Returns the number of affected rows.

```go
affected, err := qs.Filter(ProductExpr.Stock.Eq(0)).
    Update(ctx, orm.UpdateMap{"in_stock": false})
```

### `Delete(ctx context.Context) (int64, error)`
Executes an SQL `DELETE` for all rows matching the QuerySet filters. Returns the number of deleted rows.

```go
deleted, err := qs.Filter(ProductExpr.DeletedAt.IsNotNull()).
    Delete(ctx)
```

