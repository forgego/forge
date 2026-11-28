---
name: forge-api
description: Build REST APIs in Forge using `github.com/forgego/forge/api` ViewSets, serializers, routers, authentication, permissions, throttling, pagination, and `forge generate --api`. Use when creating REST endpoints, serializers, viewsets, or OpenAPI metadata. Do NOT use for the `/admin` UI (use `forge-admin`) or ORM queries outside HTTP handlers (use `forge-orm`).
license: MIT
compatibility: Go 1.26.10+, github.com/forgego/forge v0.2.x
metadata:
  author: forgego
  version: "0.2"
---

# Forge REST API

Forge's REST layer (`api/`) is modelled on Django REST Framework: serializer factories, `ViewSet` / `ViewSetConfig`, `api.Router`, content negotiation, authentication, permissions, throttling, and RFC 7807 problem-detail errors.

## Golden Path

### Option A: Auto-generate from Models

```bash
forge generate --models ./app/catalog --output ./app/catalog --api
```

Requires each model in the package to have a single auto-increment `int64` primary key named `id`. Generated viewsets enable `RejectUnknownRequestFields: true`.

### Option B: Manual Serializer + ViewSet

```go
type ProductSerializer struct {
	*api.BaseSerializer
}

func NewProductSerializer() api.Serializer {
	return &ProductSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *ProductSerializer) New() api.Serializer       { return NewProductSerializer() }
func (s *ProductSerializer) Fields() []string        { return []string{"id", "name", "price", "is_active"} }
func (s *ProductSerializer) ReadOnlyFields() []string { return []string{"id"} }
```

Register on `api.Router` and mount on `*server.Server`:

```go
apiRouter := api.NewRouter("/api/v1")
vs := api.NewBaseViewSet(NewProductSerializer, catalog.ProductObjects, &catalog.Product{})
vs.Authentication = []authentication.Authentication{authentication.NewTokenAuthentication(tokenStore)}
vs.Permissions = []permissions.Permission{permissions.IsAuthenticated{}}
apiRouter.Register("products", vs)

srv.RegisterRoutes(func(r *server.Router) {
	apiRouter.RegisterRoutes(r)
})
```

## Endpoints Registered by `apiRouter.Register("products", vs)`

- `GET /api/v1/products/` (`List`)
- `POST /api/v1/products/` (`Create`)
- `GET /api/v1/products/{id}` (`Retrieve` — no trailing slash)
- `PUT /api/v1/products/{id}` (`Update`)
- `PATCH /api/v1/products/{id}` (`PartialUpdate`)
- `DELETE /api/v1/products/{id}` (`Destroy`)

## Hard Guardrails

- **Always override `New() api.Serializer`** on custom serializer structs and pass `nil` (or initial data map) to `api.NewBaseSerializer(nil)`.
- **`Router.Register` validates at startup**: It checks that the viewset's queryset/manager implements the operations it routes and **panics on mismatch**; set `ReadOnly: true` for read-only resources (`docs-site/docs/api/data-access.md`).
- **Field contract & unknown keys**: Generated viewsets set `RejectUnknownRequestFields: true` (400 on unknown keys). Declare write-only inputs in `WriteOnlyFields()`, read-only fields in `ReadOnlyFields()`, and hidden fields in `ExcludeResponseFields()` (`docs-site/docs/api/field-contract.md`).
- **Mount via `srv.RegisterRoutes(...)`**: It is a method on `*server.Server`, not a package function. `api.Router` does not have a `.Use()` middleware method.
- **No automatic `?search=` on `BaseViewSet`**: `BaseViewSet.List` ignores `?search=`; wire `api/filters.SearchFilter` explicitly when needed.
- **OpenAPI**: `api/docs/openapi.go` serves OpenAPI document metadata/info; route paths and schemas are not auto-populated from ViewSets.

## References

- [REST API, serializers, auth, permissions & errors](references/rest-api.md)
