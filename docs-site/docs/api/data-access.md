---
sidebar_position: 3
description: The manager and queryset operations a BaseViewSet calls, read-only resources, and the configuration check that runs at startup.
image: /social-card.png
---

# ViewSet Data Access

`api.BaseViewSet` does not require its `Queryset` to implement a fixed Go interface. It finds the methods it needs by name when a request arrives. To keep a mistake there from becoming a `500` on the first request, the router checks every viewset's data access when you register it and again when you mount its routes. A viewset that cannot serve its routes stops the application at startup, with a message that names the resource and the missing or mismatched operation.

---

## Operations by action

`Queryset` holds a manager (anything with the write methods below, such as `orm.Manager[T]`) or, for a read-only resource, any value with the read methods. `Model` must be a non-nil pointer to the model struct, for example `&Product{}`.

| Action | Route | Operation on `Queryset` |
| :--- | :--- | :--- |
| `list` | `GET /resource/` | `Count(context.Context) (int64, error)` and `All(context.Context) ([]T, error)` |
| `retrieve` | `GET /resource/{id}` | `Get(context.Context, int64) (*T, error)` |
| `create` | `POST /resource/` | `Create(context.Context, *T) error` |
| `update`, `partial_update` | `PUT`/`PATCH /resource/{id}` | `Get`, then `Update(context.Context, *T) error` |
| `destroy` | `DELETE /resource/{id}` | `Get`, then `Delete(context.Context, *T) error` |

The parameter types can be wider than shown. `*T` can be `interface{}`, and the context parameter can be any interface that `context.Context` satisfies. `All` can return any slice. The object `Get` returns is what `Update` and `Delete` receive, so their parameter must accept it. On a writable resource, `Get` must return a pointer or an interface, because the viewset populates that object in place.

### List pipeline

`list` calls these optional methods when they exist. If a method is present, its signature must match:

| Method | Signature | Used for |
| :--- | :--- | :--- |
| `Filter` | `Filter(orm.Expression) Q` (the parameter may be any interface that `orm.Expression` satisfies) | `?field=value` filters |
| `OrderBy` | `OrderBy(...string) Q` (`...any` also works) | `?ordering=` and the default order |
| `Offset`, `Limit` | `Offset(int) Q`, `Limit(int) Q` | pagination |

If the `Queryset` lacks `Offset` or `Limit` but has a `QuerySet()` constructor that takes no arguments, `list` runs against that constructor's result. That is how `orm.Manager[T]` paginates in SQL: `Count`, `All`, `Filter`, `OrderBy`, `Offset` and `Limit` are checked on `orm.QuerySet[T]`.

`orm.Manager[T]`, the `<Model>Objects` managers that `forge generate` emits, and the generated `<Model>ViewSet` all pass the check.

---

## Read-only resources

A resource that should only be listed and retrieved sets `ReadOnly`:

```go
vs := api.NewBaseViewSet(NewReportSerializer, ReportObjects, &Report{})
vs.ReadOnly = true
router.Register("reports", vs)
```

A read-only viewset needs only `Count`, `All` and `Get`. `POST`, `PUT`, `PATCH` and `DELETE` respond `405 Method Not Allowed` with `Allow: GET`. Those responses come after authentication, permission and throttle checks, and they never call the queryset. `ViewSetConfig` has the same `ReadOnly` field.

A writable viewset whose queryset has no `Create`, `Update` or `Delete` fails the check, and the message suggests `ReadOnly`.

---

## When the check runs

- `Router.Register` checks the viewset and panics if the check fails.
- `Router.RegisterRoutes` checks every registered viewset again before it mounts any route, so a field changed after `Register` is also caught.

Both calls panic because the problem is a programming error that no request can recover from. This matches `Router.Action`, which panics on an unknown HTTP method, and `http.ServeMux.Handle`, which panics on an invalid pattern. To get the problem as an error instead, for example in a test, call `CheckConfiguration()` on the viewset.

The check reports every problem at once:

```text
api: resource "reports": Queryset *reports.Store: missing Create(context.Context, *Model) error, needed by create; Queryset *reports.Store: missing Update(context.Context, *Model) error, needed by update and partial_update; Queryset *reports.Store: missing Delete(context.Context, *Model) error, needed by destroy; set ReadOnly to serve only list and retrieve
```

Messages contain the resource name, Go type names and operation names. They never contain configured values, so a connection string held by a queryset does not reach logs or crash reports.

Beyond the operations, the check rejects a nil or typed-nil viewset, a nil serializer factory or one that returns nil or panics, a nil `Model`, a typed-nil `Model` such as `(*Product)(nil)`, a `Model` that is not a pointer to a struct, and a nil or typed-nil `Queryset`.

---

## Custom viewsets

- A type that implements `api.ViewSet` without embedding `*api.BaseViewSet` is checked only for nil. Its handlers own their data access.
- A type that embeds `*api.BaseViewSet` inherits `CheckConfiguration`, so the embedded viewset is checked. If the type serves some actions from its own data, define `CheckConfiguration() error` on it to state its real requirements, or return `nil`.
- `api.ViewSetConfig` checks that `Serializer` is set, then checks the viewset it would build. The check does not build or cache that viewset, so fields you set after `Register` still apply.
