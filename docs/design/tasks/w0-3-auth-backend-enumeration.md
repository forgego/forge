Your workspace is /home/hamid/Other/projects/foreit-wt/w0-auth. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# W0-3: password backend must not reveal whether an account exists or its status

Module root: /home/hamid/Other/projects/foreit-wt/w0-auth/forge (run go commands from there).
Only edit: forge/identity/backends/password.go, and create forge/identity/backends/password_enumeration_test.go.

## Principles (mandatory)
- Reproduce first: write the tests FIRST, run `go test ./identity/backends -run Enumeration -count=1`, keep the failing output for your report, then fix.
- Smallest correct change; no unrelated edits; gofmt.
- New functions < 40 lines, early returns, errors wrapped with %w.
- The only allowed package-level value is an immutable precomputed dummy hash (see below); no other new globals.
- Tests must NOT need Postgres: use an in-memory fake implementing `repository.UserRepository` (find the interface in forge/identity/repository; embed the interface in the fake struct so only the methods you need are implemented, e.g. `GetByEmail`, `GetByUsername`).

## Bug (verified on master), `passwordBackend.Authenticate` (forge/identity/backends/password.go ~41-90)
1. When `GetByEmail`/`GetByUsername` fails (unknown user), it returns `ErrInvalidCredentials` immediately, without any bcrypt work, while a known user costs a full bcrypt comparison. The response time reveals whether the account exists.
2. It returns `ErrUserInactive` / `ErrUserLocked` BEFORE checking the password, so anyone can learn an account's status without knowing the password.
3. Every repository error (including a database outage) is disguised as `ErrInvalidCredentials`.

## Required behaviour (Django ModelBackend semantics)
- Unknown user: run exactly one bcrypt comparison against a dummy hash, then return `ErrInvalidCredentials`.
  - Dummy hash: a package-level `var dummyPasswordHash = mustHash("forge-dummy-password")` produced ONCE with `utils.HashPassword` (same `DefaultCost` as real hashes).
  - `mustHash` panics on error (it runs at init with a constant input).
- "Unknown user" means the repository returned `sql.ErrNoRows` or an error that `errors.Is` matches to it. Check what forge/identity/repository/user.go returns on not-found (~lines 148, 204, 260) and match that exactly. Any other repository error: return `fmt.Errorf("authenticate: lookup user: %w", err)`.
- Known user with a wrong password: `ErrInvalidCredentials`, regardless of `IsActive`/`IsLocked`.
- Correct password and inactive user: `ErrUserInactive`. Correct password and locked user: `ErrUserLocked`. Correct password and active, unlocked user: return the user.
- Keep the existing `nil, nil` "not applicable" returns when username/email or password keys are missing.
- Extract small helpers if `Authenticate` grows past ~40 lines (e.g. `lookupUser`, `checkStatus`).

## Tests (password_enumeration_test.go, table-driven, fake repository)
| case | expect |
|---|---|
| unknown username | ErrInvalidCredentials, and the bcrypt path ran (see below) |
| unknown email | ErrInvalidCredentials |
| inactive user + wrong password | ErrInvalidCredentials (NOT ErrUserInactive) |
| locked user + wrong password | ErrInvalidCredentials (NOT ErrUserLocked) |
| inactive user + correct password | ErrUserInactive |
| locked user + correct password | ErrUserLocked |
| active user + correct password | user returned |
| repository returns a non-not-found error | error wraps it (errors.Is) and is NOT ErrInvalidCredentials |

Proving bcrypt ran for unknown users: do not use wall-clock timing assertions (flaky). Instead, route the comparison through an unexported package-level function variable only if you cannot otherwise observe it (e.g. `var comparePassword = utils.CheckPassword`), and have the test swap it to count calls. Restore it with t.Cleanup. That variable is the one exception to the no-globals rule.

## Acceptance (run from /home/hamid/Other/projects/foreit-wt/w0-auth/forge)
    gofmt -l identity                              # empty
    go vet ./identity/...
    go test ./identity/backends -run Enumeration -count=1 -v
    go test -race ./identity/backends ./identity/service -count=1

Report (max 30 lines): failing output before the fix, files changed, each acceptance command with its result.
