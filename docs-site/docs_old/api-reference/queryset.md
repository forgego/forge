---
sidebar_position: 2
---

# QuerySet Reference

Type-safe query builder.

## Common methods

```go
qs := UserObjects.Filter(UserFieldsInstance.IsActive.Equals(true))
qs = qs.OrderBy("-created_at").Limit(10)
users, err := qs.All(ctx)
```

- `Filter`, `Exclude`
- `OrderBy`, `Limit`, `Offset`, `Distinct`
- `Only`, `Defer`
- `All`, `Get`, `First`, `Last`, `Count`, `Exists`

## Next

- [Queries guide](/docs/guides/queries)
