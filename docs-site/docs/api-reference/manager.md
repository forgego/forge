---
sidebar_position: 3
description: Comprehensive API Reference for the Manager[T] type and CRUD operations.
---

# Manager API Reference

A `Manager[T]` provides the primary database table interface for a given model `T`. It acts as the gateway for single-record CRUD operations and spawns configured `QuerySet[T]` instances for complex queries.

---

## CRUD Operations

### `Create(ctx context.Context, instance *T) error`
Inserts a new model instance into the database. Executes `BeforeCreate`, `BeforeSave`, `Clean`, `AfterCreate`, and `AfterSave` hooks and populates the auto-generated primary key on `instance`.

```go
product := &catalog.Product{
    ProductGenerated: catalog.ProductGenerated{
        Name:  "Mechanical Keyboard",
        Price: 129.99,
    },
}
err := catalog.ProductObjects.Create(ctx, product)
```

### `Get(ctx context.Context, id int64) (*T, error)`
Retrieves a single record by integer primary key. Returns an error if no record exists.

```go
product, err := catalog.ProductObjects.Get(ctx, 42)
```

### `Update(ctx context.Context, instance *T) error`
Persists modifications made to an existing model instance. Executes `BeforeUpdate`, `BeforeSave`, `Clean`, `AfterUpdate`, and `AfterSave` hooks.

```go
product.Price = 119.99
err := catalog.ProductObjects.Update(ctx, product)
```

### `Delete(ctx context.Context, instance *T) error`
Removes the given record from the database. Executes `BeforeDelete` and `AfterDelete` lifecycle hooks.

```go
err := catalog.ProductObjects.Delete(ctx, product)
```

---

## QuerySet Spawning Methods

A Manager provides query shortcuts. `All` evaluates immediately; the others below create querysets:

- `Manager.All(ctx context.Context) ([]*T, error)`: Executes a query selecting all rows.
- `Manager.Filter(expr Expression) (QuerySet[T], error)`: Starts a query with one `WHERE` expression.
- `Manager.Exclude(expr Expression) QuerySet[T]`: Starts a query excluding matching rows.
- `Manager.OrderBy(fields ...any) QuerySet[T]`: Starts a query with ordering.
- `Manager.SelectRelated(relations ...any) QuerySet[T]`: Starts a query with eager-loaded relations.
- `Manager.QuerySet() QuerySet[T]`: Spawns a clean, unconstrained queryset.

---

## Bulk Operations

### `BulkCreate(ctx context.Context, instances []*T) error`
Performs a multi-row insert. The API has no batch-size argument.

```go
err := catalog.ProductObjects.BulkCreate(ctx, batchProducts)
```

---

## Custom Managers

You can extend or subclass a Manager to add domain-specific query methods:

```go
type CustomProductManager struct {
    *orm.Manager[Product]
}

func (m *CustomProductManager) InStockOnly() (orm.QuerySet[Product], error) {
    return m.Filter(orm.And(ProductExpr.InStock.Eq(true), ProductExpr.Stock.Gt(0)))
}
```

