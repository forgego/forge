Your workspace is /home/hamid/Other/projects/foreit-wt/wave0-rest. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# API: multipart boundary, ordering/search filters, content negotiation without renderers

Module root: /home/hamid/Other/projects/foreit-wt/wave0-rest/forge (run go commands from there).
Only edit files under forge/api/parsers, forge/api/filters, and forge/api/content_negotiation.go (plus new _test.go files in those places). Another agent is editing other packages at the same time: if `go build ./...` fails in a package you did not touch, ignore it.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; gofmt. Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming after behaviour, never after tickets. No "W0", "B16" etc. in names.

## Bugs (verified on master)
1. forge/api/parsers/multipart.go:24 `multipart.NewReader(r, "")`: the boundary is empty, so every multipart body fails to parse. Fix: take the boundary from the request `Content-Type` with `mime.ParseMediaType`; return a clear error if missing. Read the parser interface first to see how the content type reaches the parser; if it does not, parse via `r.ParseMultipartForm`-equivalent logic on the data available and explain in the report.
   Test: build a real multipart body with `mime/multipart.Writer` (a text field and a file) and assert both are parsed.
2. forge/api/filters/ordering.go ~70-73 calls variadic `OrderBy(...)` via `reflect.Call` with a single `[]string` value, so the ORM gets one slice argument. Fix: pass each field as its own argument (`CallSlice` or one reflect.Value per field, matching the real `OrderBy` signature in forge/orm).
   Test: a fake queryset type with a variadic `OrderBy(fields ...string)` method records its arguments; `?ordering=-price,name` must call it with `"-price","name"`.
3. forge/api/filters/search.go ~30-52 looks up a `Search` method that `orm.QuerySet` does not have, so `?search=` is silently ignored. Fix: build `orm.Or(orm.F(field).IContains(q)...)` over `SearchFields` and apply it with the queryset's `Filter` method (check the exact names in forge/orm, and how forge/admin/core/admin.go does its search, and mirror it). Empty query or no fields → queryset unchanged.
   Test: with a fake queryset recording the `Filter` argument, `?search=abc` over two fields calls Filter once with a non-nil expression; empty search does not call it.
4. forge/api/content_negotiation.go lines 30, 47, 58, 74 index `Renderers[0]` / `Parsers[0]` without a length check, so an empty configuration panics. Fix: return nil / an error the callers already handle (read the callers) instead of panicking; do not change behaviour when lists are non-empty.
   Test: empty Renderers/Parsers do not panic.

## Acceptance (from the module root)
    gofmt -l api                                   # empty
    go vet ./api/...
    staticcheck ./api/...                          # no output
    go test -race ./api/... -count=1

Report (max 25 lines): failing output before each fix, files changed, each acceptance command with its result.
