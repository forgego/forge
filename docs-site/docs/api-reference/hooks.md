---
sidebar_position: 6
description: API Reference for model lifecycle hooks, validation, and execution order.
---

# Lifecycle Hooks API Reference

The `github.com/forgego/forge/schema` package provides Django-style lifecycle hooks for intercepting, modifying, validating, or reacting to ORM persistence events.

---

## The `ModelHooks` Struct

Models register hooks by returning `*schema.ModelHooks` from their `Hooks()` method:

```go
func (Product) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithBeforeCreate(beforeCreateHook).
        WithAfterCreate(afterCreateHook).
        WithBeforeSave(beforeSaveHook).
        WithClean(cleanValidationHook)
}
```

---

## Complete List of Hook Points

| Hook Method | Signature | Trigger Event |
| :--- | :--- | :--- |
| `WithBeforeCreate` | `func(ctx context.Context, instance any) error` | Immediately before a new record is inserted into the database. |
| `WithAfterCreate` | `func(ctx context.Context, instance any) error` | Immediately after a new record is successfully committed to the database. |
| `WithBeforeUpdate` | `func(ctx context.Context, instance any) error` | Immediately before an existing record is updated in the database. |
| `WithAfterUpdate` | `func(ctx context.Context, instance any) error` | Immediately after an existing record is updated in the database. |
| `WithBeforeSave` | `func(ctx context.Context, instance any) error` | Before either create or update execution. |
| `WithAfterSave` | `func(ctx context.Context, instance any) error` | After either create or update execution succeeds. |
| `WithBeforeDelete` | `func(ctx context.Context, instance any) error` | Before a record is deleted from the database. |
| `WithAfterDelete` | `func(ctx context.Context, instance any) error` | After a record has been deleted from the database. |
| `WithClean` | `func(instance any) error` | Invoked during `Validate()` or pre-save validation passes. |

---

## Execution Lifecycle Order

### On Record Creation (`Create` / `Save` on new instance)
1. `Clean(instance)` — Validate cross-field business constraints.
2. `BeforeSave(ctx, instance)` — Common pre-save normalization (e.g. slug generation).
3. `BeforeCreate(ctx, instance)` — Creation-specific tasks (e.g. assigning UUIDs or hashing passwords).
4. **Database Transaction / SQL INSERT**
5. `AfterCreate(ctx, instance)` — Fire event, publish audit log.
6. `AfterSave(ctx, instance)` — Common post-save notifications.

### On Record Update (`Update` / `Save` on existing instance)
1. `Clean(instance)` — Validate updated fields.
2. `BeforeSave(ctx, instance)` — Re-calculate computed attributes.
3. `BeforeUpdate(ctx, instance)` — Check permission or log old vs new diffs.
4. **Database Transaction / SQL UPDATE**
5. `AfterUpdate(ctx, instance)` — Clear caches.
6. `AfterSave(ctx, instance)` — Common post-save triggers.

---

## Practical Example

```go
func (Order) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithClean(func(instance interface{}) error {
            order := instance.(*Order)
            if order.ShippingAmount < 0 {
                return errors.New("shipping amount cannot be negative")
            }
            return nil
        }).
        WithBeforeSave(func(ctx context.Context, instance interface{}) error {
            order := instance.(*Order)
            // Compute total amount from items, tax, and shipping
            order.TotalAmount = order.Subtotal + order.TaxAmount + order.ShippingAmount
            return nil
        }).
        WithAfterCreate(func(ctx context.Context, instance interface{}) error {
            order := instance.(*Order)
            // Fire asynchronous webhook or email event
            go NotifyOrderPlaced(order.ID)
            return nil
        })
}
```

