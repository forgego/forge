TASK G1c: remove the string "user" context key (staticcheck SA1029) in identity.

Workspace: /home/hamid/Other/projects/foreit-wt/staticcheck (Go module in ./forge). Always use absolute paths.

=== PROBLEM ===
`staticcheck ./...` (from /home/hamid/Other/projects/foreit-wt/staticcheck/forge, binary /home/hamid/go/bin/staticcheck)
reports SA1029 only here:
  identity/middleware/auth.go:62, :79         context.WithValue(ctx, "user", user)
  identity/middleware/auth_test.go:65,94,120,148   context.WithValue(req.Context(), "user", &models.User{...})
Readers of the string key:
  identity/middleware/auth.go:137, :186, :207  ctx.Value("user").(*models.User)
  identity/handlers/auth.go:184                ctx.Value("user").(*models.User)   (inside GetUserFromContext)

The middleware ALREADY stores the same user under the typed key via `core.WithUser(ctx, user)`
(api/core/context.go: `WithUser`, `UserFromContext(ctx) (interface{}, bool)`, typed `UserKey`). The
string key is a redundant legacy copy.

=== CHANGE ===
1. identity/middleware/auth.go: in RequireAuth and OptionalAuth delete the `context.WithValue(ctx, "user", user)`
   line and the redundant `context.WithValue(ctx, core.UserKey, user)` line; keep `ctx = core.WithUser(ctx, user)`.
   Replace every `ctx.Value("user").(*models.User)` read with a lookup through `core.UserFromContext(ctx)` followed
   by a `.(*models.User)` type assertion (a small unexported helper in the file is fine to avoid repetition).
2. identity/handlers/auth.go GetUserFromContext: read via `core.UserFromContext` the same way. Keep its signature.
3. identity/middleware/auth_test.go: build request contexts with `core.WithUser(req.Context(), &models.User{...})`.
4. Search the WHOLE repo (/home/hamid/Other/projects/foreit-wt/staticcheck, including examples/) for any other
   `Value("user")` or `WithValue(..., "user", ...)` usage and migrate it the same way. Report each one.

=== VERIFY (all from .../staticcheck/forge) ===
  /home/hamid/go/bin/staticcheck ./...        -> no output at all
  go vet ./... && go build ./...
  go test ./identity/... ./api/...
  cd ../examples/ecommerce && go build ./...
  gofmt -l on edited files -> empty

=== HARD CONSTRAINTS ===
- NO git commands. Only touch the files above (plus any extra string-key users you find in step 4).
- No behaviour change beyond dropping the duplicate string key.
- Final report: files changed with one line each, and the verification output.
