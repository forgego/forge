---
sidebar_position: 1
description: Production REST API framework with ViewSets, ModelSerializers, pagination, and rate throttling.
image: /social-card.png
---

# REST API Framework Overview

Forge includes a batteries-included REST API framework inspired by Django REST Framework (DRF) and designed for Go's type-safety and performance.

---

## Architectural Principles

- **Declarative ViewSets**: Combine CRUD handlers, list querying, search, pagination, and serializer transformations in a single cohesive controller.
- **ModelSerializers**: Automatic bidirectional mapping between Go model structs and JSON payloads with field-level validation and masking.
- **Pagination**: List responses are paginated by page number (`page`, `page_size`).
- **Rate Throttling**: Protect endpoints with token bucket / sliding window throttles for anonymous IPs and authenticated user tokens.
- **OpenAPI skeleton**: `api/docs` serves an OpenAPI 3.0 document with the `info` block only; paths and schemas are not generated yet.

---

## Complete Example

```go
package catalog

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/forgego/forge/api"
    "myapp/models"
)

// 1. Define ModelSerializer
type ProductSerializer struct {
    api.ModelSerializer[models.Product]
}

func (ProductSerializer) Fields() []string {
    return []string{"id", "sku", "name", "price", "price_with_tax", "in_stock"}
}

// 2. Define ViewSet
type ProductViewSet struct {
    api.ModelViewSet[models.Product]
}

func NewProductViewSet() *ProductViewSet {
    return &ProductViewSet{
        ModelViewSet: api.NewModelViewSet[models.Product](models.ProductManager, &ProductSerializer{}),
    }
}

// 3. Register onto Chi Router
func RegisterRoutes(r chi.Router) {
    viewset := NewProductViewSet()
    viewset.Throttling = api.NewUserRateThrottle("100/min")
    
    api.RegisterViewSet(r, "/api/v1/products", viewset)
}
```

The single call to `api.RegisterViewSet` binds the full standard REST lifecycle:
- `GET /api/v1/products/` — List with pagination, filters, and ordering
- `POST /api/v1/products/` — Create new record with input validation
- `GET /api/v1/products/{id}/` — Retrieve single record
- `PUT /api/v1/products/{id}/` — Update record
- `PATCH /api/v1/products/{id}/` — Partial update
- `DELETE /api/v1/products/{id}/` — Delete record

---

## Next Steps

- **[ModelSerializers](/docs/api/serializers/)**: Field whitelisting, nested serialization, and validation hooks.
- **[ViewSets](/docs/api/viewsets/)**: Custom actions and query filtering.
- **[Pagination](/docs/api/pagination/)**: Page-number pagination and page size settings.
- **[OpenAPI](/docs/api/openapi/)**: What the OpenAPI document contains today.

