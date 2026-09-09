---
sidebar_position: 22
description: Advanced filtering with FilterSets, URL query parameter binding, search filters, and AST parsing.
image: /forge-social-card.svg
---

# Filters & FilterSets

Forge provides a powerful, declarative filtering layer inspired by `django-filter`. FilterSets bridge HTTP query parameters directly into type-safe ORM `QuerySet` constraints with input validation, type coercion, and security whitelisting.

---

## Defining a FilterSet

A `FilterSet` declares which fields can be filtered, their matching operators, and default behaviors:

```go
package catalog

import (
    "github.com/forgego/forge/filter"
    "myapp/models"
)

type ProductFilterSet struct {
    filter.BaseFilterSet[models.Product]
}

func NewProductFilterSet() *ProductFilterSet {
    fs := &ProductFilterSet{}
    
    // Exact match on category slug
    fs.AddCharFilter("category", "category.slug", filter.OpExact)
    
    // Case-insensitive substring search on name
    fs.AddCharFilter("q", "name", filter.OpIContains)
    
    // Price range filters
    fs.AddNumberFilter("min_price", "price", filter.OpGte)
    fs.AddNumberFilter("max_price", "price", filter.OpLte)
    
    // Boolean availability
    fs.AddBooleanFilter("in_stock", "in_stock")
    
    // Choice filter for status
    fs.AddChoiceFilter("status", "status", []string{"active", "sale", "featured"})

    return fs
}
```

---

## Applying FilterSets to HTTP Requests

In your Chi handler or ViewSet, pass the `http.Request` URL query parameters into the FilterSet:

```go
func ListProductsHandler(w http.ResponseWriter, r *http.Request) {
    fs := NewProductFilterSet()
    
    // Start with base QuerySet (e.g. only non-deleted items)
    baseQS := models.ProductManager.Filter(models.ProductExpr.DeletedAt.IsNull(true))
    
    // FilterSet parses r.URL.Query(), validates types, and builds SQL WHERE tree
    filteredQS, err := fs.Filter(r.Context(), baseQS, r.URL.Query())
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Execute query with pagination
    products, err := filteredQS.Limit(20).All(r.Context())
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(products)
}
```

Now, frontend clients can query your API naturally:

```http
GET /api/v1/products?category=laptops&min_price=500&max_price=1500&in_stock=true&q=pro
```

---

## Built-in Filter Types

| Filter Type | Constructors | URL Example | Generated SQL |
| :--- | :--- | :--- | :--- |
| **CharFilter** | `fs.AddCharFilter(param, field, op)` | `?q=forge` | `name ILIKE '%forge%'` |
| **NumberFilter** | `fs.AddNumberFilter(param, field, op)` | `?min_price=100` | `price >= 100.00` |
| **BooleanFilter** | `fs.AddBooleanFilter(param, field)` | `?in_stock=true` | `in_stock = true` |
| **DateFilter** | `fs.AddDateFilter(param, field, op)` | `?created_after=2026-01-01` | `created_at >= '2026-01-01'` |
| **ChoiceFilter** | `fs.AddChoiceFilter(param, field, choices)` | `?status=active` | `status = 'active'` |
| **InFilter** | `fs.AddInFilter(param, field)` | `?id=1,2,3,4` | `id IN (1, 2, 3, 4)` |

---

## Security & Query Whitelisting

Direct client-to-SQL parameter binding can introduce vulnerabilities or Denial of Service via unbounded Cartesian joins. Forge FilterSets prevent this by:
1. **Strict Whitelisting**: Only fields explicitly added to the FilterSet are permitted. Unknown query parameters are ignored or rejected.
2. **Type Coercion & Bounds**: Parameters are strongly parsed (e.g., parsing a non-numeric string into a NumberFilter returns a 400 Validation Error).
3. **Max Limits**: Prevents unbounded limit sizes and guards against expensive operations.

---

## Next Steps

- **[REST API ViewSets](/docs/api/viewsets/)**: Use FilterSets seamlessly inside REST API ViewSets.
- **[QuerySet API Reference](/docs/api-reference/queryset/)**: Complete list of terminal execution and lookup methods.

