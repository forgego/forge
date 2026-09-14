Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

ROLE: You are the delegated coder: edit the files yourself; do not call other agents. Use US spelling in comments.

# Restore three useful API capabilities that the Waves 1-2 cleanup removed, on the merged code

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit forge/api (router/viewset files, api/errors, api/throttling), forge/internal/ratelimit, forge/server/ratelimit.go, and tests. Keep everything that the cleanup already merged (one BaseViewSet, one RFC 7807 writer, one keyed limiter): add, don't revert.

## Principles (mandatory)
- Tests FIRST for each capability. gofmt; functions < 40 lines; errors wrapped with %w; no new package-level mutable state except where stated; naming after behaviour.

## 1. Custom viewset actions on api.Router (Django REST Framework `@action`)
The removed `EnhancedRouter` let users add extra endpoints per resource: list actions `/{resource}/{name}` and detail actions `/{resource}/{id}/{name}`, with any HTTP method. `api.Router` today only has `Get`/`Post`/`Handle` custom routes (forge/api/viewset.go ~527-580) with raw paths.
Add to `api.Router`:
```go
// ActionConfig describes an extra endpoint on a registered resource.
type ActionConfig struct {
    Methods []string // e.g. http.MethodPost; defaults to GET
    Detail  bool     // true: /{resource}/{id}/{path}; false: /{resource}/{path}
    URLPath string   // defaults to the action name
}
func (r *Router) Action(resource, name string, cfg ActionConfig, handler http.HandlerFunc)
```
`RegisterRoutes` registers actions under the same prefix, after the CRUD routes, so `/{resource}/{id}/publish` does not clash with `/{resource}/{id}`. Registering an action for a resource that is not registered returns no error at registration, but `RegisterRoutes` must still mount it under the prefix. Unknown methods in `Methods` panic at registration with a clear message (programmer error, like chi does).
Also restore OPTIONS metadata: when a resource is registered, `OPTIONS /{resource}/` calls the existing `docs.OptionsHandler` (forge/api/docs/openapi.go:31) with the viewset. Read its signature first.
Tests (router_actions_test.go): a list action with POST and a detail action with GET are reachable through a chi router built by `RegisterRoutes` and receive the `id` URL param; `/resource/{id}` still hits Retrieve; OPTIONS returns 200 with metadata JSON.

## 2. Pluggable rate-limit store for API throttles
The cleanup replaced throttling's `CacheBackend` parameter with an in-process `internal/ratelimit.KeyedLimiter`. That removed the only extension point for sharing limits across server instances (for example a Redis-backed store).
- In api/throttling define `type Store interface { Allow(key string) (allowed bool, retryAfter time.Duration) }` (a user-facing interface; document that implementations must be safe for concurrent use).
- `NewUserRateThrottle(rate string)` / `NewAnonRateThrottle(rate string)` keep working and use the in-process keyed limiter. Add `WithStore(store Store)` as a method or a constructor variant (`NewUserRateThrottleWithStore(rate string, store Store)`), whichever fits the existing code better; the store receives the full key the throttle already builds (scope + user/IP).
- Test: a fake store that denies after N calls makes the throttle return `ThrottledError` with its retryAfter; the default constructor still allows burst then denies.

## 3. Custom exception handling hook on the RFC 7807 writer
The cleanup deleted `api.SetExceptionHandler`, which had never worked (the stored handler was never read). DRF's `EXCEPTION_HANDLER` is a real feature: map application errors to responses.
- In api/errors add a `Handler` option (check `HandlerConfig` in forge/api/errors/handler.go) `CustomMapper func(err error, r *http.Request) *Problem`: if it returns non-nil, that problem is written; if nil, the default mapping runs.
- `WriteError(w, r, err)` uses the default handler; add `func NewWriter(cfg *HandlerConfig) func(http.ResponseWriter, *http.Request, error)` or equivalent so users can install a custom mapper without package-level mutable state, and let `BaseViewSet` accept an optional `ErrorWriter func(http.ResponseWriter, *http.Request, error)` field (and `ViewSetConfig` pass-through) that `handleException` uses when set.
- Tests: a custom mapper turns a sentinel application error into 409 problem+json; nil from the mapper falls back to default 500; a BaseViewSet with `ErrorWriter` uses it.

## Acceptance (from the module root)
    gofmt -l api internal server                      # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./api/... ./internal/... ./server/...
    staticcheck ./api/... ./internal/... ./server/... # no output
    go test -race ./api/... ./internal/... ./server/... -count=1
    ~/go/bin/golangci-lint run --timeout=5m --config=../.golangci.yml ./api/... ./internal/... ./server/...   # 0 issues

Report (max 25 lines): public API added, tests added, each command result.
