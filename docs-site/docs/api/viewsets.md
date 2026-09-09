---
sidebar_position: 3
description: Type-safe CRUD endpoints, custom action routes, permissions, and query filtering with ViewSets.
image: /forge-social-card.svg
---

# ViewSets & Endpoint Handlers

A `ModelViewSet[T]` combines standard CRUD lifecycle operations into a single cohesive, type-safe controller with built-in pagination, rate limiting, and permission verification.

---

## Standard CRUD Operations

When registered with `api.RegisterViewSet(r, "/api/v1/products", viewset)`, the following REST handlers are automatically bound:

| Method | Endpoint | Action Method | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/products/` | `List` | Paginated list with filtering and search. |
| `POST` | `/api/v1/products/` | `Create` | Create a new record with validation. |
| `GET` | `/api/v1/products/{id}/` | `Retrieve` | Fetch single record by ID. |
| `PUT` | `/api/v1/products/{id}/` | `Update` | Full record update. |
| `PATCH` | `/api/v1/products/{id}/` | `PartialUpdate` | Partial record field update. |
| `DELETE` | `/api/v1/products/{id}/` | `Destroy` | Delete record. |

---

## Adding Custom Action Routes

Just like `@action` in Django REST Framework, you can attach custom sub-endpoints to a ViewSet:

```go
type ProductViewSet struct {
    api.ModelViewSet[models.Product]
}

func (v *ProductViewSet) CustomActions() []api.ActionRoute {
    return []api.ActionRoute{
        // Custom Collection Endpoint: GET /api/v1/products/featured/
        api.NewAction("featured", http.MethodGet, false, func(w http.ResponseWriter, r *http.Request) {
            featured, err := models.ProductManager.
                Filter(models.ProductExpr.IsFeatured.Eq(true)).
                Limit(6).
                All(r.Context())
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            json.NewEncoder(w).Encode(featured)
        }),

        // Custom Detail Endpoint: POST /api/v1/products/{id}/apply-discount/
        api.NewAction("apply-discount", http.MethodPost, true, func(w http.ResponseWriter, r *http.Request) {
            id := chi.URLParam(r, "id")
            // Parse discount percentage and update product
            w.WriteHeader(http.StatusOK)
        }),
    }
}
```

---

## Customizing the Base QuerySet

Override `GetQuerySet(r *http.Request)` to restrict records based on multi-tenancy, soft deletion, or user ownership:

```go
func (v *ProductViewSet) GetQuerySet(r *http.Request) *orm.QuerySet[models.Product] {
    // Hide soft-deleted products from public catalog endpoints
    return models.ProductManager.Filter(models.ProductExpr.DeletedAt.IsNull(true))
}
```

---

## Next Steps

- **[Pagination](/docs/api/pagination/)**: Configure pagination strategies.
- **[Throttling](/docs/api/throttling/)**: Prevent API abuse with rate limits.
- **[OpenAPI Documentation](/docs/api/openapi/)**: Interactive API docs.

