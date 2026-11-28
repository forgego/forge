# REST API Reference

## Serializers (`api/serializer.go`, `api/serializer_fields.go`)

Always provide a factory (`func() api.Serializer`) and override `New()` so each request gets an isolated serializer instance:

```go
type UserSerializer struct {
	*api.BaseSerializer
}

func NewUserSerializer() api.Serializer {
	return &UserSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *UserSerializer) New() api.Serializer          { return NewUserSerializer() }
func (s *UserSerializer) Fields() []string           { return []string{"id", "username", "email", "is_active"} }
func (s *UserSerializer) ReadOnlyFields() []string   { return []string{"id", "date_joined"} }
func (s *UserSerializer) WriteOnlyFields() []string  { return []string{"password"} }
```

## ViewSets & `ViewSetConfig` (`api/viewset.go`, `api/viewset_config.go`)

```go
vs := &api.ViewSetConfig{
	Model:                      &catalog.Product{},
	Queryset:                   catalog.ProductObjects,
	Serializer:                 NewProductSerializer(),
	PerPage:                    20,
	RejectUnknownRequestFields: true,
	Authentication:             []authentication.Authentication{authentication.NewJWTAuthentication(jwtCfg)},
	Permissions:                []permissions.Permission{permissions.IsAuthenticated{}},
	Throttles:                  []throttling.Throttle{throttling.NewUserRateThrottle(100, time.Minute)},
}
apiRouter.Register("products", vs)
```

For read-only endpoints (or managers without write methods), set `ReadOnly: true` so `Router.Register` only requires read methods.

## Filtering, Ordering & Pagination

- **Pagination**: `?page=2&page_size=50`
- **Ordering**: `?ordering=-price,name`
- **Field filters** (`api/viewset.go` `buildFilterExpr`): `?is_active=true`, `?name__contains=desk`, `?price__gte=10`, `?price__lte=100`, `?name__startswith=a`, `?name__endswith=z`, `?id__in=1,2,3`, `?deleted_at__isnull=true` (`iexact` and `range` are not handled by `BaseViewSet.buildFilterExpr`).

## Error Format (`api/errors/`)

Validation and request-contract errors return `400 Bad Request` with problem-detail JSON:

```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "detail": "Validation failed",
  "errors": {
    "price": ["price must be positive"]
  }
}
```
