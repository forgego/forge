---
sidebar_position: 1
description: REST API framework with viewsets, serializers, pagination, filtering, and rate throttling.
image: /social-card.png
---

# REST API Framework Overview

Forge includes a REST API framework inspired by Django REST Framework (DRF): a viewset serves the CRUD routes of one model, a serializer shapes its JSON, and a router mounts viewsets on the application's router.

---

## Architectural Principles

- **ViewSets**: `api.BaseViewSet` serves list, create, retrieve, update, partial update and destroy for one model from a manager such as the generated `<Model>Objects`.
- **Serializers**: a factory returning an `api.Serializer`. Optional `Fields()`, `Exclude()`, `ReadonlyFields()` and `WriteOnlyFields()` methods shape responses and requests. Model validation (`validate` tags and `Clean()`) runs on create and update.
- **Pagination**: List responses are paginated by page number (`page`, `page_size`).
- **Filtering and ordering**: `?field=value`, lookups such as `?price__gte=10`, and `?ordering=-price` on list requests.
- **Rate Throttling**: per-user and per-IP rate throttles (`api/throttling`) answer `429` when a client exceeds its rate.
- **OpenAPI skeleton**: `api/docs` serves an OpenAPI 3.0 document with the `info` block only; paths and schemas are not generated yet.

---

## Complete Example

`forge generate` writes the `Product` model's `ProductObjects` manager; `forge add api` writes a registration function like this one.

```go
package catalog

import (
    "github.com/forgego/forge/api"
    "github.com/forgego/forge/api/throttling"
    "github.com/forgego/forge/server"
)

// 1. Define a serializer. Fields limits the keys in every response.
type ProductSerializer struct {
    *api.BaseSerializer
}

func NewProductSerializer() api.Serializer {
    return &ProductSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ProductSerializer) Fields() []string {
    return []string{"id", "sku", "name", "price", "in_stock"}
}

// 2. Build the viewset and register it on the application's router.
func RegisterProductAPI(router *server.Router) {
    viewset := api.NewBaseViewSet(NewProductSerializer, ProductObjects, &Product{})
    viewset.ExcludeResponseFields = api.NonSerializableFields(&Product{})
    viewset.ReadOnlyRequestFields = api.NonEditableFields(&Product{})
    viewset.RejectUnknownRequestFields = true
    viewset.Throttles = []throttling.Throttle{throttling.NewUserRateThrottle("100/min")}

    apiRouter := api.NewRouter("/api/v1")
    apiRouter.Register("products", viewset)
    apiRouter.RegisterRoutes(router)
}
```

`Register` binds the standard REST routes under the router prefix:
- `GET /api/v1/products/` — List with pagination, filters, and ordering
- `POST /api/v1/products/` — Create a record with input validation
- `GET /api/v1/products/{id}` — Retrieve a single record
- `PUT /api/v1/products/{id}` — Full update
- `PATCH /api/v1/products/{id}` — Partial update
- `DELETE /api/v1/products/{id}` — Delete a record

`Register` and `RegisterRoutes` panic when the viewset's data access cannot serve these routes; see [ViewSet Data Access](/docs/api/data-access/).

---

## Next Steps

- **[ModelSerializers](/docs/api/serializers/)**: Field whitelisting and validation.
- **[ViewSets](/docs/api/viewsets/)**: Custom actions and scoping the queryset.
- **[Pagination](/docs/api/pagination/)**: Page-number pagination and page size settings.
- **[OpenAPI](/docs/api/openapi/)**: What the OpenAPI document contains today.
