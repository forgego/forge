Your workspace is /home/hamid/Other/projects/foreit-wt/w0-api. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-5: JSON null must not panic the viewset decoder; pagination links must be valid URLs

Module root: /home/hamid/Other/projects/foreit-wt/w0-api/forge (run go commands from there).
Only edit: forge/api/viewset.go (function setFieldValue only), forge/api/pagination.go (function BuildPaginatedResponse only), and create forge/api/w0_api_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run `go test ./api -run W0 -count=1`, keep the failing output (a panic counts) for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- New helpers < 40 lines, early returns, no new package-level mutable state. Table-driven tests.

## Bug 1 (verified on master): JSON null panics
`populateFromMap` (forge/api/viewset.go ~764) calls `setFieldValue(fieldValue, value)` (~820) for each JSON key. `setFieldValue` does `valueValue := reflect.ValueOf(value)`; for a JSON `null` value is nil, so `valueValue` is the zero `reflect.Value`. The `default:` branch (~860) calls `valueValue.Type()`, which panics for float, pointer, slice, map or other kinds that reach `default`.
Fix: at the top of `setFieldValue` (after the CanSet check), if `!valueValue.IsValid()` (value is nil): when `field.Kind()` is Ptr, Interface, Slice or Map, set `field.Set(reflect.Zero(field.Type()))`; otherwise leave the field unchanged; then return.
Tests (call the unexported function directly, package api):
- `populateFromMap(&s, map[string]interface{}{"price": nil, "note": nil, "tags": nil, "name": nil})` on a struct with `Price float64 json:"price"`, `Note *string json:"note"` (pre-set to non-nil), `Tags []string json:"tags"` (pre-set), `Name string json:"name"` (pre-set "keep"): must not panic (`assert.NotPanics`); Note and Tags become nil; Price and Name keep their previous values.
- A normal value still works: `{"price": 12.5}` sets Price.

## Bug 2 (verified on master): pagination links are not valid URLs
`BuildPaginatedResponse` (forge/api/pagination.go ~98) builds `baseURL := r.URL.Scheme + "://" + r.URL.Host + r.URL.Path`. For server-side requests `r.URL.Scheme` and `r.URL.Host` are empty, so links become `://path?page=2&page_size=10`. It also drops every other query parameter (filters, search, ordering), so following `next` loses the filter.
This function is used by every viewset list response (viewset.go, viewset_enhanced.go) and identity/handlers/user.go.
Fix (DRF `replace_query_param` semantics):
- Extract `func pageURL(r *http.Request, page, pageSize int) string`.
- scheme is "https" if `r.TLS != nil`, else "http"; do NOT trust X-Forwarded-Proto here.
- host is `r.Host`.
- copy `r.URL.Query()`, set "page" and "page_size" (replacing existing values), keep all other parameters.
- return `(&url.URL{Scheme: scheme, Host: host, Path: r.URL.Path, RawQuery: q.Encode()}).String()`.
- Use it for next and previous. Keep the `PaginatedResponse` shape unchanged.
Tests (use `httptest.NewRequest`):
- `GET http://example.com/api/products?page=2&page_size=10&search=shoe&ordering=-price` with totalCount 35: next is `http://example.com/api/products?ordering=-price&page=3&page_size=10&search=shoe` (Encode sorts keys), previous has page=1 and keeps search/ordering.
- a request with `r.TLS` set (assign `&tls.ConnectionState{}`) produces an https link.
- first page: previous is nil; last page: next is nil.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-api/forge)
    gofmt -l api                          # empty
    go vet ./api/...
    go test ./api -run W0 -count=1 -v
    go test -race ./api/... ./identity/handlers -count=1

Report (max 30 lines): failing output before the fix, files changed, each acceptance command with its result.
