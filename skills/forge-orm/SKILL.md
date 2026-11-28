---
name: forge-orm
description: Query and persist data in Forge using `github.com/forgego/forge/orm` (`Manager[T]`, `QuerySet[T]`, `orm.Field[T]` expressions, `SelectRelated`, `PrefetchRelated`, `AggregateValues`, `UpdateBuilder`), `db.WithTx` transactions/savepoints, and `filter/` FilterSets. Use when writing database queries, CRUD operations, relations loading, or filter security rules. Do NOT use for defining model schemas/hooks (use `forge-models`) or SQL migrations (use `forge-migrations-schema`).
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15 (SQLite partially tested)
metadata:
  author: forgego
  version: "0.2"
---

# Forge ORM & Filters

Forge generates a typed `<Model>Objects` (`*orm.Manager[T]`) and `<Model>FieldsInstance` for every model. `QuerySet[T]` values are lazy and immutable — each chained method returns a new `QuerySet[T]`.

## Golden Path

```go
// Startup wiring
catalog.ProductObjects.SetDB(database)

// Typed query
p := catalog.ProductFieldsInstance
qs, err := catalog.ProductObjects.Filter(orm.And(
	p.IsActive.Eq(true),
	p.Price.Lte(150),
))
if err != nil {
	return err
}
products, err := qs.OrderBy(p.Price.Desc()).Limit(20).All(ctx)
```

## CRUD & Transactions

```go
// Create / Update / Delete
err := catalog.ProductObjects.Create(ctx, product)
err = catalog.ProductObjects.Update(ctx, product)
err = catalog.ProductObjects.Delete(ctx, product)
err = catalog.ProductObjects.BulkCreate(ctx, []*catalog.Product{p1, p2})

// Transaction + Savepoint
err = database.WithTx(ctx, func(tx *db.Tx) error {
	mgr := catalog.ProductObjects.WithTx(tx)
	if err := mgr.Create(ctx, product); err != nil {
		return err
	}
	sp, err := tx.CreateSavepoint("sp1")
	if err != nil {
		return err
	}
	if err := mgr.Update(ctx, other); err != nil {
		return sp.RollbackToSavepoint()
	}
	return nil
})
```

## Hard Guardrails

- **`Manager.Filter` returns `(QuerySet[T], error)`**: You cannot chain `.Filter(...).OrderBy(...)` directly on `<Model>Objects` without checking `err` first. (On a `QuerySet[T]`, `Filter` and `Exclude` return `QuerySet[T]`.)
- **`int64` PKs only**: `Manager.Create` returns `NotImplemented` for non-integer keys; `Manager.Get(ctx, id)` takes `int64`.
- **`SelectRelated` / `PrefetchRelated`**: No runtime guardrail reports accessing an unloaded relation — always load relations explicitly.
- **Not implemented in QuerySet/Aggregates**: `Union`, `Intersection`, `Difference`, grouped aggregates, `StdDev`, `Variance`, `orm.RegisterAggregate`, and aggregates over many-to-many paths return `NotImplemented` before running SQL. Use raw `*sql.DB` (embedded in `*db.DB`) for window functions, CTEs, or full-text search.
- **Savepoint API**: `tx.CreateSavepoint(name)` returns `(*db.Savepoint, error)`; `RollbackToSavepoint()` is a method on `*db.Savepoint`, not `*db.Tx`.

## References

- [ORM QuerySet, aggregates & transactions](references/orm.md)
- [Shared filter AST & security whitelisting](references/filters.md)
