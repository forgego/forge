---
sidebar_position: 2
description: Comprehensive API Reference for the QuerySet[T] type.
---

# QuerySet API Reference

A `QuerySet[T]` represents a collection of database queries targeting model `T`. QuerySets are immutable and chainable; every transformation method returns a new `QuerySet` instance without modifying the caller.

---

## Filtering & Slicing Methods

### `Filter(conditions ...any) *QuerySet[T]`
Adds conditions to the SQL `WHERE` clause combined via `AND`. Accepts generated field expressions, `orm.Q` / `QueryExpr` boolean trees, or `orm.Where()`.

```go
qs = qs.Filter(ProductExpr.Price.Lte(100.00), ProductExpr.InStock.Eq(true))
```

### `Exclude(conditions ...any) *QuerySet[T]`
Negates the provided conditions (`NOT (...)`) and appends them to the `WHERE` clause.

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

### `Aggregate(ctx context.Context, funcs ...Aggregate) (map[string]any, error)`
Executes an immediate aggregate query (`COUNT`, `SUM`, `AVG`, `MIN`, `MAX`) and returns a map of evaluated scalar values.

```go
stats, err := qs.Aggregate(ctx, orm.Count("id"), orm.Avg("price"))
```

### `Annotate(funcs ...Aggregate) *QuerySet[T]`
Annotates each returned row with a calculated aggregate value (e.g. counting total reviews per product).

---

## Terminal Execution Methods

Terminal methods execute the SQL statement against the active database connection or transaction.

| Method | Signature | Description |
| :--- | :--- | :--- |
| `All(ctx)` | `All(ctx context.Context) ([]*T, error)` | Evaluates the query and returns all matching records. |
| `Get(ctx, pk)` | `Get(ctx context.Context, pk any) (*T, error)` | Returns exactly one matching record. Returns `ErrNotFound` if missing, or `ErrMultipleObjects` if more than 1 matches. |
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

