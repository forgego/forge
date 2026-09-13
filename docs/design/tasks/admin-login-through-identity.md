Your workspace is /home/hamid/Other/projects/foreit-wt/wave0-rest. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Admin login accepts staff/superuser accounts from identity (env pair stays as bootstrap)

Module root: /home/hamid/Other/projects/foreit-wt/wave0-rest/forge (run go commands from there).
Only edit: forge/admin/api/rest/router.go, forge/admin/site.go, and create forge/admin/api/rest/login_authenticator.go, forge/admin/api/rest/login_authenticator_test.go, forge/identity/backends/staff_login.go, forge/identity/backends/staff_login_test.go.

## Principles (mandatory)
- Tests FIRST, run them red, then implement. Smallest correct change; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, sentinel errors, no new package-level mutable state.
- Naming after behaviour, never after tickets.
- forge/admin must NOT import forge/identity (keep admin independent); the adapter lives in identity.

## Problem (verified on master)
forge/admin/api/rest/router.go `handleLogin` (~309-400) only accepts the `FORGE_ADMIN_USERNAME` / `FORGE_ADMIN_PASSWORD` env pair (`adminCredentials()` ~411) and returns 503 when it is unset. Users created with `forge createsuperuser` (identity users with IsStaff/IsSuperuser) can never log in. Django's admin authenticates `is_staff` users through the auth backends.

## Design
1. admin/api/rest/login_authenticator.go:
   ```go
   // LoginAuthenticator verifies admin credentials. It returns the canonical username on success
   // and ErrInvalidLogin when the credentials are wrong or the account may not use the admin.
   type LoginAuthenticator interface {
       AuthenticateAdmin(ctx context.Context, username, password string) (string, error)
   }
   var ErrInvalidLogin = errors.New("invalid admin credentials")
   ```
   plus `func (r *Router) SetLoginAuthenticator(a LoginAuthenticator)`.
2. handleLogin order (keep all existing limiter, payload validation and session-issuing code):
   - If an authenticator is set, call it. Success → issue the session for the returned username. `ErrInvalidLogin` → continue to the env pair. Any other error → 500 with a generic message, no limiter failure.
   - Env pair configured → existing constant-time comparison.
   - Neither an authenticator nor the env pair configured → the existing 503.
   - Every failed attempt records a limiter failure and returns the existing generic invalid-credentials response, so the response does not reveal which path failed.
3. admin/site.go: `func (s *Site) SetLoginAuthenticator(a rest.LoginAuthenticator)` stores it and passes it to the router where `rest.NewRouter(s.registry)` is created (~151). Read how the site builds the router; if the router is created lazily, apply it at creation.
4. identity/backends/staff_login.go: `NewStaffLoginAuthenticator(backend AuthenticationBackend) *StaffLoginAuthenticator` whose `AuthenticateAdmin(ctx, username, password)` calls `backend.Authenticate(ctx, map[string]string{"username": username, "password": password})`. It returns the user's Username only if `IsStaff || IsSuperuser`. Invalid credentials, inactive/locked users, nil user and non-staff users → an error that `errors.Is(err, <sentinel>)`. Because identity must not import admin, define `var ErrAdminLoginDenied = errors.New("invalid admin credentials")` in identity, and make the router treat any authenticator error that is not a context/infra error as invalid login. Simplest: in router, treat `errors.Is(err, ErrInvalidLogin)` OR an error implementing `interface{ InvalidLogin() bool }` as invalid; have identity's error type implement `InvalidLogin() bool`. Any other error goes down the 500 path. Document this in the interface comment.

## Tests
- rest: a fake authenticator: success issues a token (reuse how existing login tests in forge/admin/api/rest/login_security_test.go read the token); invalid → falls back to env pair when set, else 401 with the same body as today's invalid response; infra error → 500; no authenticator and no env → 503 (unchanged).
- identity: a fake backend returning a staff user → username; a non-staff active user → denied error with `InvalidLogin() == true`; backend `ErrInvalidCredentials` → denied; backend `ErrUserInactive` → denied.

## Acceptance (from the module root)
    gofmt -l admin identity                         # empty
    go vet ./admin/... ./identity/...
    staticcheck ./admin/... ./identity/...          # no output
    go test -race ./admin/... ./identity/... -count=1
    go list -deps ./admin/... | grep forge/identity  # no output

Report (max 25 lines): failing output before, files changed, each command result.
