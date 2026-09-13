Your workspace is /home/hamid/Other/projects/foreit-wt/w0-perms. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-6: owner/admin permission checks must match real users; an unknown token must be rejected

Module root: /home/hamid/Other/projects/foreit-wt/w0-perms/forge (run go commands from there).
Only edit: forge/api/permissions/is_owner.go, forge/api/permissions/is_admin.go, forge/api/permissions/helpers.go, forge/api/authentication/token.go, and create forge/api/permissions/w0_permissions_test.go and forge/api/authentication/w0_token_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run them, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- Helpers < 40 lines, early returns, no new package-level mutable state. Table-driven tests.
- Do not create import cycles: before importing another forge package, check with `go list -deps` or grep that it does not import the package you are editing.

## Bug 1 (verified on master): IsOwnerOrReadOnly always denies
- `NewIsOwnerOrReadOnly("")` sets `OwnerField: "user_id"` and `UserIDField: "id"` (forge/api/permissions/is_owner.go ~22-30).
- `getOwnerID(obj, "user_id")` and `getUserID(user, "id")` resolve names through `getField` (helpers.go ~34-48), which uses `reflect.Value.FieldByName`. That only matches Go field names such as `UserID` and `ID`, so with the defaults the owner lookup always fails and writes are always denied.
- The fallback list in `getOwnerID` (`user_id, owner_id, UserID, OwnerID`) is never reached, because the constructor never passes "".

Fix:
- In helpers.go, make `getField` resolve a name in order:
  1. exact Go field name;
  2. a field whose `db` tag or `json` tag name (the part before the first comma) equals the name;
  3. case-insensitive Go field name.
  Keep it a small function; extract a `fieldByTag` helper.
- In `NewIsOwnerOrReadOnly`, leave `OwnerField` empty when the argument is "", so the fallback list is used. Keep `UserIDField: "id"`: the user side already tries a `GetID()` method first, and `getField` now finds `ID` through its `json:"id"` tag.

Tests (w0_permissions_test.go, package permissions):
- An object struct `{ UserID int64 \`json:"user_id" db:"user_id"\` }` with owner 7, and a user struct `{ ID int64 \`json:"id"\` }` with ID 7.
- A POST with that user on the request context is allowed by `NewIsOwnerOrReadOnly("")`; user ID 8 is denied; GET is allowed without a user.
- Put the user on the request the way the package's existing tests or `authentication.SetUserOnRequest` do; read forge/api/authentication/auth.go to see how.

## Bug 2 (verified on master): IsAdminUser never matches framework users
- `isAdmin` (forge/api/permissions/is_admin.go ~44-62) only checks an `IsAdmin` method or field.
- The framework user `identity/models.User` has `IsStaff` and `IsSuperuser` fields and no `IsAdmin`, so staff and superusers are always denied.

Fix (DRF `IsAdminUser` is `is_staff`):
- Keep the `IsAdmin` method/field check first.
- Then treat the user as admin if a bool method or field `IsStaff` is true, or `IsSuperuser` is true.
- Extract a `boolMember(obj, name) (value, found bool)` helper so the logic is not repeated.

Tests: users with `IsStaff: true`, `IsSuperuser: true`, `IsAdmin: true` pass; a user with all of them false fails; an unauthenticated request fails. Use small local structs; do NOT import identity/models.

## Bug 3 (verified on master): an unknown token falls through to anonymous
- `TokenAuthentication.Authenticate` (forge/api/authentication/token.go ~24-56) returns `nil, nil` when `TokenLookup` returns no user for a token the client supplied.
- `AuthenticateRequest` (auth.go ~36-52) treats `nil, nil` as "not applicable" and tries the next class, so a request with a wrong `Authorization: Token xyz` is handled as anonymous. DRF raises AuthenticationFailed here.

Fix:
- When the header has the Token scheme with a non-empty token, a lookup function is configured, and the lookup returns `nil, nil`: return an error.
- Use the same error type the other authentication classes in forge/api/authentication return for invalid credentials. Read jwt.go, basic.go and apikey.go and match them exactly. If none exists, define `var ErrInvalidToken = errors.New("invalid token")` in token.go.
- Keep `nil, nil` for: no header, a different scheme, an empty token, or no lookup function configured.

Tests (w0_token_test.go): an unknown token returns a non-nil error, and `AuthenticateRequest` with `[TokenAuthentication]` returns that error; a valid token returns the user; no header returns `nil, nil`; "Bearer x" returns `nil, nil`.

## Acceptance (from /home/hamid/Other/projects/foreit-wt/w0-perms/forge)
    gofmt -l api                                   # empty
    go vet ./api/...
    go test ./api/permissions ./api/authentication -run W0 -count=1 -v
    go test -race ./api/... -count=1

Report (max 30 lines): failing output before the fix, files changed, the error type used for Bug 3 and why, each acceptance command with its result.
