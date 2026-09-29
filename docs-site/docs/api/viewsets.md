---
sidebar_position: 3
description: CRUD endpoints with BaseViewSet, custom action routes, and scoping the records a viewset serves.
image: /social-card.png
---

# ViewSets & Endpoint Handlers

An `api.BaseViewSet` serves the standard CRUD operations of one model, with pagination, filtering, ordering, authentication, permission and throttle checks. Build one with `api.NewBaseViewSet(serializerFactory, manager, &Model{})` and register it on an `api.Router`.

---

## Standard CRUD Operations

After `apiRouter := api.NewRouter("/api/v1")`, `apiRouter.Register("products", viewset)` and `apiRouter.RegisterRoutes(router)`, these handlers are bound:

| Method | Endpoint | Action Method | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/products/` | `List` | Paginated list with filtering and ordering. |
| `POST` | `/api/v1/products/` | `Create` | Create a new record with validation. |
| `GET` | `/api/v1/products/{id}` | `Retrieve` | Fetch a single record by ID. |
| `PUT` | `/api/v1/products/{id}` | `Update` | Full record update. |
| `PATCH` | `/api/v1/products/{id}` | `PartialUpdate` | Partial record field update. |
| `DELETE` | `/api/v1/products/{id}` | `Destroy` | Delete record. |

`List` accepts `page` and `page_size`, `ordering` (for example `?ordering=-price`), and a filter per serialized model field: `?is_featured=true`, or a lookup such as `?price__gte=10`, `?name__icontains=desk`, `?id__in=1,2,3` or `?deleted_at__isnull=true`.

Set `ReadOnly` to serve only `List` and `Retrieve`; see [ViewSet Data Access](/docs/api/data-access/).

---

## Adding Custom Action Routes

Like `@action` in Django REST Framework, `Router.Action` adds an endpoint next to a resource's routes. `Detail: true` puts it under `/{id}/`; `Methods` defaults to `GET` and `URLPath` to the action name. Register actions on the same `api.Router` before calling `RegisterRoutes`:

```go
import (
    "net/http"
    "strconv"

    "github.com/forgego/forge/api"
    "github.com/forgego/forge/orm"
    "github.com/forgego/forge/server"
    "github.com/go-chi/chi/v5"
)

func RegisterProductActions(apiRouter *api.Router) {
    // Collection endpoint: GET /api/v1/products/featured
    apiRouter.Action("products", "featured", api.ActionConfig{}, func(w http.ResponseWriter, r *http.Request) {
        featured, err := ProductObjects.QuerySet().
            Filter(orm.F("is_featured").Eq(true)).
            Limit(6).
            All(r.Context())
        if err != nil {
            _ = server.SendError(w, http.StatusInternalServerError, "could not load products")
            return
        }
        _ = server.SendJSON(w, http.StatusOK, api.SerializeMany(featured))
    })

    // Detail endpoint: POST /api/v1/products/{id}/apply-discount
    apiRouter.Action("products", "apply_discount", api.ActionConfig{
        Methods: []string{http.MethodPost},
        Detail:  true,
        URLPath: "apply-discount",
    }, func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
        if err != nil {
            _ = server.SendError(w, http.StatusBadRequest, "invalid id")
            return
        }
        var body struct {
            Percent float64 `json:"percent"`
        }
        if err := server.GetJSON(r, &body); err != nil || body.Percent <= 0 || body.Percent >= 100 {
            _ = server.SendError(w, http.StatusBadRequest, "percent must be between 0 and 100")
            return
        }
        product, err := ProductObjects.Get(r.Context(), id)
        if err != nil {
            _ = server.SendError(w, http.StatusNotFound, "not found")
            return
        }
        product.Price = product.Price * (1 - body.Percent/100)
        if err := ProductObjects.Update(r.Context(), product); err != nil {
            _ = server.SendError(w, http.StatusInternalServerError, "could not save product")
            return
        }
        _ = server.SendJSON(w, http.StatusOK, api.SerializeModel(product))
    })
}
```

An action handler is a plain `http.HandlerFunc`: the viewset's authentication, permission and throttle checks and its serializer do not run for it. `api.SerializeModel` and `api.SerializeMany` return every serializable model field.

---

## Scoping the Records a ViewSet Serves

`BaseViewSet` has no per-request queryset hook. It calls the methods of its `Queryset` by name (see [ViewSet Data Access](/docs/api/data-access/)), so to hide records from every route, pass a type that embeds the manager and narrows the operations the routes use. `List` paginates through `QuerySet()`; `Retrieve`, `Update`, `PartialUpdate` and `Destroy` load the record through `Get`:

```go
// ActiveProducts hides archived products from every route of a viewset.
type ActiveProducts struct {
    *orm.Manager[Product]
}

func (m ActiveProducts) QuerySet() orm.QuerySet[Product] {
    return m.Manager.QuerySet().Filter(orm.F("archived").Eq(false))
}

func (m ActiveProducts) Get(ctx context.Context, id int64) (*Product, error) {
    return m.QuerySet().Filter(orm.F("id").Eq(id)).Get(ctx)
}

viewset := api.NewBaseViewSet(NewProductSerializer, ActiveProducts{ProductObjects}, &Product{})
```

An archived product is missing from the list, and `GET /api/v1/products/{id}` answers `404` for it. Scoping that depends on the request, such as the signed-in user's records, needs a custom viewset: embed `*api.BaseViewSet` and override the handlers that must see the request.

---

## Next Steps

- **[Pagination](/docs/api/pagination/)**: Page-number pagination and page size settings.
- **[Throttling](/docs/api/throttling/)**: Prevent API abuse with rate limits.
- **[OpenAPI Documentation](/docs/api/openapi/)**: What the OpenAPI document contains today.
