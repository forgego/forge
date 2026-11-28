# ORM, QuerySets, Aggregates & Transactions

## Field Expressions (`orm/field_expr.go`, `orm/expression.go`)

Generated `<Model>FieldsInstance` exposes typed `orm.Field[T]` values with operators:
- Comparison: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`
- String: `Contains`, `StartsWith`, `EndsWith` (and case-insensitive `IContains`, `IStartsWith`, `IEndsWith`)
- Set / Null: `In`, `NotIn`, `IsNull`, `IsNotNull`
- Combinators: `orm.And(exprs...)`, `orm.Or(exprs...)`, `orm.Not(expr)`

## Eager Loading

- `qs.SelectRelated("author")` — JOIN-based eager loading for ForeignKey / OneToOne relations.
- `qs.PrefetchRelated("tags")` — separate-query prefetching for ManyToMany / reverse relations.

## Ungrouped Aggregates (`orm/aggregates.go`)

Supported functions: `orm.Count`, `orm.Sum`, `orm.Avg`, `orm.Min`, `orm.Max` via `orm.AggregateValues(ctx, qs, ...)`:

```go
stats, err := orm.AggregateValues(ctx, qs,
	orm.Count("id").As("total"),
	orm.Avg("price").As("avg_price"),
)
```

Grouped aggregates, `StdDev`, `Variance`, and aggregates across many-to-many paths return `NotImplemented`.

## Raw SQL Escape Hatch

`*db.DB` embeds `*sql.DB`, so drop down to `database/sql` directly for window functions, recursive CTEs, or full-text search:

```go
rows, err := database.QueryContext(ctx, "SELECT id, name FROM products WHERE price > $1", 100)
```
