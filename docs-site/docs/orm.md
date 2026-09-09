---
sidebar_position: 11
description: High-performance type-safe querying, QuerySets, Q-expressions, relations, and aggregations.
image: /forge-social-card.svg
---

# Type-Safe ORM & QuerySets

Forge provides a compiled, type-safe ORM inspired by Django's QuerySet semantics and optimized for Go's concurrency and static type system. 

Key advantages:
- **Compile-Time Safety**: Field names, types, and relational traversal paths are verified by the Go compiler.
- **Zero Stringly-Typed Bugs**: No runtime typos like `db.Where("stauts = ?", 1)`.
- **Zero Reflection in Hot Execution**: High-speed SQL generation with query parameter binding.
- **Django-Grade Query Fluency**: Method chaining, lazy execution, complex boolean queries (`orm.Q`), and JOIN/prefetch optimization.

---

## The QuerySet Philosophy

QuerySets are **lazy**. Constructing, filtering, or slicing a QuerySet does not touch the database until an execution terminal method (`All()`, `Get()`, `First()`, `Count()`, `Exists()`) is called.

Every filter or ordering method returns an immutable, cloned `QuerySet`, making it completely thread-safe to branch queries across requests:

```go
// Base active products queryset
activeProducts := ProductManager.Filter(ProductExpr.InStock.Eq(true))

// Branch 1: Cheap items
budgetItems, err := activeProducts.
    Filter(ProductExpr.Price.Lte(25.00)).
    OrderBy("price").
    Limit(10).
    All(ctx)

// Branch 2: High-end electronics
premiumElectronics, err := activeProducts.
    Filter(ProductExpr.Category.Slug.Eq("electronics")).
    Filter(ProductExpr.Price.Gte(500.00)).
    All(ctx)
```

---

## Type-Safe Expressions & Lookups

Forge generates expression helpers for each model during `forge generate`. For example, `ProductExpr` provides typed methods for all fields:

| Lookup Method | SQL Operator | Description |
| :--- | :--- | :--- |
| `Field.Eq(val)` | `=` | Exact equality |
| `Field.Ne(val)` | `!=` | Inequality |
| `Field.Gt(val)` | `>` | Greater than |
| `Field.Gte(val)` | `>=` | Greater than or equal |
| `Field.Lt(val)` | `<` | Less than |
| `Field.Lte(val)` | `<=` | Less than or equal |
| `Field.In(vals...)` | `IN (...)` | Contained in slice/args |
| `Field.NotIn(vals...)`| `NOT IN (...)` | Not contained in slice |
| `Field.Contains(str)` | `LIKE '%str%'` | Substring match |
| `Field.IContains(str)`| `ILIKE '%str%'` | Case-insensitive substring match |
| `Field.StartsWith(s)` | `LIKE 's%'` | Prefix match |
| `Field.EndsWith(s)` | `LIKE '%s'` | Suffix match |
| `Field.IsNull(bool)` | `IS NULL` / `IS NOT NULL` | Nullability check |

```go
// Example of expressive lookups
products, err := ProductManager.
    Filter(ProductExpr.Title.IContains("mechanical keyboard")).
    Filter(ProductExpr.Price.Lte(150.00)).
    Filter(ProductExpr.DeletedAt.IsNull(true)).
    All(ctx)
```

---

## Complex Boolean Logic with `orm.Q` / `QueryExpr`

Compose complex boolean expressions using `And()`, `Or()`, and negation:

```go
// WHERE (price < 50 OR in_stock = true) AND NOT (status = 'discontinued')
condition := orm.And(
    orm.Or(
        ProductExpr.Price.Lt(50.00),
        ProductExpr.InStock.Eq(true),
    ),
    ProductExpr.Status.Ne("discontinued"),
)

items, err := ProductManager.Filter(condition).All(ctx)
```

You can also use SQL-like explicit `orm.Where(field, operator, value)` when building dynamic search queries from user input:

```go
qs := ProductManager.QuerySet()
if minPrice > 0 {
    qs = qs.Filter(orm.Where("price", orm.OpGreaterOrEqual, minPrice))
}
if search != "" {
    qs = qs.Filter(orm.Where("name", orm.OpIContains, search))
}
results, err := qs.All(ctx)
```

---

## Eager Loading: Solving the N+1 Problem

Forge provides two eager-loading mechanisms optimized for relational databases:

### 1. `SelectRelated` (SQL JOIN)
Use `SelectRelated` for single-valued relationships (`ForeignKey` and `OneToOne`). It performs an SQL `INNER JOIN` or `LEFT JOIN`, populating the parent and related structs in a single round-trip:

```go
// Fetches order and joins customer + shipping address in 1 SQL query
order, err := OrderManager.
    SelectRelated("Customer", "ShippingAddress").
    Get(ctx, orderID)

fmt.Println(order.Customer.Email) // Instant access, no extra DB query
```

### 2. `PrefetchRelated` (Batched Queries)
Use `PrefetchRelated` for multi-valued relationships (`ManyToMany` or reverse foreign keys). Forge executes a secondary batched query using `WHERE in_id IN (...)` and stitches the records in Go memory:

```go
// Fetches 20 products, then fetches all associated tags and reviews in 2 batch queries
products, err := ProductManager.
    Filter(ProductExpr.InStock.Eq(true)).
    PrefetchRelated("Tags", "Reviews").
    Limit(20).
    All(ctx)
```

---

## Aggregations & Grouping

Execute database-level aggregates with `Aggregate()`:

```go
stats, err := ProductManager.
    Filter(ProductExpr.InStock.Eq(true)).
    Aggregate(ctx,
        orm.Count("id"),
        orm.Avg("price"),
        orm.Max("price"),
        orm.Min("price"),
        orm.Sum("inventory_count"),
    )

fmt.Printf("Total: %v, Avg Price: %v\n", stats["count"], stats["avg"])
```

Supported aggregate functions:
- `orm.Count(field)`
- `orm.Sum(field)`
- `orm.Avg(field)`
- `orm.Min(field)`
- `orm.Max(field)`
- `orm.StdDev(field)`
- `orm.Variance(field)`

---

## Mutations & Bulk Operations

### Single Record Operations

```go
// Create
product, err := ProductManager.Create(ctx, &Product{
    Name:  "Wireless Mouse",
    Price: 29.99,
})

// Update
product.Price = 24.99
err = ProductManager.Update(ctx, product)

// Delete
err = ProductManager.Delete(ctx, product.ID)
```

### High-Performance Bulk Operations

Avoid thousands of individual round-trips with bulk operations:

```go
// Bulk Create (Single multi-row INSERT statement)
newItems := []*Product{
    {Name: "Item A", Price: 10.0},
    {Name: "Item B", Price: 20.0},
}
err := ProductManager.BulkCreate(ctx, newItems, 500) // batch size 500

// QuerySet-level Update (Single SQL UPDATE statement)
rowsAffected, err := ProductManager.
    Filter(ProductExpr.Status.Eq("pending")).
    Update(ctx, orm.UpdateMap{
        "status": "archived",
        "archived_at": time.Now(),
    })

// QuerySet-level Delete (Single SQL DELETE statement)
deletedCount, err := ProductManager.
    Filter(ProductExpr.DeletedAt.IsNotNull()).
    Delete(ctx)
```

---

## Transactions

Execute multiple mutations within an ACID transaction:

```go
err := orm.WithTransaction(ctx, db, func(txCtx context.Context) error {
    order, err := OrderManager.Create(txCtx, newOrder)
    if err != nil {
        return err
    }

    _, err = InventoryManager.
        Filter(InventoryExpr.ProductID.Eq(order.ProductID)).
        Update(txCtx, orm.UpdateMap{
            "stock": orm.F("stock") - order.Quantity,
        })
    return err
})
```

---

## Next Steps

- **[FilterSets & Search](/docs/filters/)**: Advanced parameter filtering and search widgets.
- **[QuerySet API Reference](/docs/api-reference/queryset/)**: Full method reference.
- **[Admin Integration](/docs/admin/overview/)**: Wiring ORM queries into the React Admin console.

