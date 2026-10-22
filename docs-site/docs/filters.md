---
sidebar_position: 22
description: HTTP query parsing and FilterSet configuration.
image: /social-card.png
---

# Filters & FilterSets

`filter.FilterSet[T]` parses allowed HTTP query parameters into ORM expressions. Create it with `filter.NewFilterSet[T]()`, set a base queryset with `WithQueryset`, configure allowed fields and lookups with `WithSecurity`, then call `ApplyQueryParams(ctx, request)`.

```go
fs, err := filter.NewFilterSet[models.Product]()
if err != nil {
    return err
}
security := filter.DefaultSecurityConfig()
security.AllowedFields["product"] = []string{"name", "price", "in_stock", "created_at", "status", "id"}
security.AllowedLookups["name"] = []string{"icontains"}
security.AllowedLookups["price"] = []string{"gte", "lte"}
fs.WithQueryset(baseQS).WithSecurity(security)

filteredQS, err := fs.ApplyQueryParams(r.Context(), r)
if err != nil {
    return err
}
products, err := filteredQS.Limit(20).All(r.Context())
```

For example, request `?name__icontains=pro&price__gte=500&price__lte=1500`. Unknown fields and disallowed lookups return an error. The parser converts numeric comparison values to `float64`, `__in` values to a string slice, and `__isnull` values to booleans.

## Filter objects

The `filter/filters` subpackage provides constructors for filter objects. Register them with `fs.AddFilter(name, object)` when using the filter-object API. `ApplyQueryParams` parses the request through its security configuration; it does not use the `AddFilter` registry.

| Filter | Constructor |
| :--- | :--- |
| Character | `filters.NewCharFilter[T](field)` |
| Number | `filters.NewNumberFilter[T](field)` |
| Boolean | `filters.NewBooleanFilter[T](field)` |
| Choice | `filters.NewChoiceFilter[T](field, choices)` |
| Date | `filters.NewDateFilter[T](field)` |
| In lookup | Use the parser's `field__in=one,two` parameter |

There are no `AddCharFilter`, `AddNumberFilter`, `AddBooleanFilter`, `AddChoiceFilter`, `AddDateFilter`, or `AddInFilter` convenience methods on `FilterSet`.

## Security

`DefaultSecurityConfig` denies field access until `AllowedFields` is configured. `AllowedLookups` limits operators per field. The parser also enforces configured complexity limits. Validate request parameters before executing the returned queryset and return a client error for invalid filters.
