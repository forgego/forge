---
sidebar_position: 3
description: Comprehensive API Reference for the Manager[T] type and CRUD operations.
---

# Manager API Reference

A `Manager[T]` provides the primary database table interface for a given model `T`. It acts as the gateway for single-record CRUD operations and spawns configured `QuerySet[T]` instances for complex queries.

---

## CRUD Operations

### `Create(ctx context.Context, instance *T) (*T, error)`
Inserts a new model instance into the database. Executes registered `BeforeCreate` and `AfterCreate` lifecycle hooks and populates the auto-generated primary key on `instance`.

```go
product, err := ProductManager.Create(ctx, &Product{
    Name:  "Mechanical Keyboard",
    Price: 129.99,
})
```

### `Get(ctx context.Context, pk any) (*T, error)`
Retrieves a single record by primary key. Returns `sql.ErrNoRows` (or `orm.ErrNotFound`) if no record exists.

```go
product, err := ProductManager.Get(ctx, 42)
```

### `Update(ctx context.Context, instance *T) error`
Persists modifications made to an existing model instance. Executes `BeforeUpdate`, `BeforeSave`, and `AfterUpdate` hooks.

```go
product.Price = 119.99
err := ProductManager.Update(ctx, product)
```

### `Delete(ctx context.Context, pk any) error`
Removes a record from the database by primary key. Executes `BeforeDelete` and `AfterDelete` lifecycle hooks and enforces foreign key `CascadeType` policies (e.g. `CascadePROTECT`).

```go
err := ProductManager.Delete(ctx, 42)
```

---

## QuerySet Spawning Methods

A Manager provides convenient shortcut methods that spawn a new `QuerySet[T]`:

- `Manager.All() *QuerySet[T]`: Starts a query selecting all rows.
- `Manager.Filter(conditions ...any) *QuerySet[T]`: Starts a query with initial `WHERE` filters.
- `Manager.Exclude(conditions ...any) *QuerySet[T]`: Starts a query excluding matching rows.
- `Manager.OrderBy(fields ...string) *QuerySet[T]`: Starts a query with ordering.
- `Manager.SelectRelated(fields ...string) *QuerySet[T]`: Starts a query with eager-loaded relations.
- `Manager.QuerySet() *QuerySet[T]`: Spawns a clean, unconstrained `QuerySet[T]`.

---

## Bulk Operations

### `BulkCreate(ctx context.Context, instances []*T, batchSize int) error`
Performs multi-row batch inserts with a configurable chunk size to stay within database query parameter limits.

```go
err := ProductManager.BulkCreate(ctx, batchProducts, 500)
```

---

## Custom Managers

You can extend or subclass a Manager to add domain-specific query methods:

```go
type CustomProductManager struct {
    *orm.Manager[Product]
}

func (m *CustomProductManager) InStockOnly() *orm.QuerySet[Product] {
    return m.Filter(ProductExpr.InStock.Eq(true), ProductExpr.Stock.Gt(0))
}
```

