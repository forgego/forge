Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# One API error response format: exceptions are written through api/errors (RFC 7807)

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit forge/api/exceptions, forge/api/errors, forge/api/api.go, forge/api/middleware_integration.go, forge/api/viewset.go (only its handleException), forge/api/serializer_test.go, and their tests.

## Principles (mandatory)
- Tests FIRST. gofmt; functions < 40 lines; no new package-level mutable state; naming after behaviour.
- Keep the DRF-style exception constructors (`exceptions.NewValidationError`, `NewPermissionDenied`, `NewNotFound`, `NewThrottled`, ...) and their types: callers keep using them.

## Facts (verified on master)
- Two response writers for the same errors: `exceptions.HandleExceptionHTTP` (forge/api/exceptions/handler.go:61) writes `{error, code, message, details}`; `api/errors.Handler.HandleError` (forge/api/errors/handler.go:93) writes RFC 7807 problems and already maps every exception type through `ErrorMapper.MapError` (mapper.go:27, cases for `*exceptions.ValidationError`, `*AuthenticationFailed`, ...). forge/server/errors.go uses api/errors, so the rest of the framework already speaks problem+json.
- `exceptions.SetExceptionHandler` stores `globalHandler`, but `HandleExceptionHTTP` never reads it (it builds a new default handler when passed nil), and every caller passes nil. So `api.SetExceptionHandler` (api.go:161) has no effect.
- api/errors imports api/exceptions, so the writer must live in api/errors.

## Change
1. In api/errors add `func WriteError(w http.ResponseWriter, r *http.Request, err error)` that uses a default `Handler` (reuse `NewHandler(DefaultHandlerConfig())` without a package-level mutable variable: construct per call, or a package-level value that is never reassigned) and calls `HandleError`. Make sure `*exceptions.Throttled` still sets `Retry-After` (check mapper/handler; add it if missing).
2. Replace every `exceptions.HandleExceptionHTTP(...)` call (forge/api/middleware_integration.go:23,25,48 and `BaseViewSet.handleException` in forge/api/viewset.go if it exists after the viewset merge) with `apierrors.WriteError`.
3. Delete from api/exceptions: `HandleExceptionHTTP`, `ExceptionHandler`, `DefaultExceptionHandler`, `NewDefaultExceptionHandler`, `globalHandler`, `SetExceptionHandler`, `GetExceptionHandler`, plus `ErrorResponse`/`ToResponse` if nothing else uses them after the change (grep the workspace, examples, cli templates, docs-site). Delete `api.SetExceptionHandler`. Port tests: exception_test.go cases for status codes become tests of `WriteError` in api/errors (status code, `application/problem+json`, Retry-After for Throttled, 500 with no internal message for unknown errors).
4. forge/api/serializer_test.go `TestBaseSerializer_Validate_Invalid` is skipped as a "known bug", but `BaseSerializer.Validate` is documented "to be overridden": a base serializer without fields accepts any data (DRF behaves the same). Rewrite the test to use a small serializer that embeds `*BaseSerializer`, overrides `Validate` and calls `AddError`, and assert `IsValid() == false` and the error map; remove the skip.

## Acceptance (from the module root)
    gofmt -l api                                      # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./api/... ./server/...
    staticcheck ./api/... ./server/...                # no output
    go test -race ./api/... ./server/... -count=1

Report (max 25 lines): removed identifiers with grep evidence, callers switched, response format change, each command result.
