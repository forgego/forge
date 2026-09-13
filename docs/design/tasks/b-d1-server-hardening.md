TASK D1: three server/auth security edge cases (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/server-security-edge-cases/forge
(That is a git worktree on branch fix/server-security-edge-cases. Work ONLY there.)

Files you may modify — ONLY these:
  server/server.go
  server/response.go
  identity/middleware/auth.go
  and their *_test.go files in the same packages (create new _test.go files if needed)

Do NOT touch server/ratelimit.go or api/throttling/* — another agent owns them.

Write the failing test FIRST for each bug, run it and see it fail, then fix.

=== BUG 1 (Medium, security): CSRF exemption uses raw prefix matching ===
server/server.go, func isCSRFExemptPath (around line 221). It ends with:

    if strings.HasPrefix(path, prefix) {
        return true
    }

Exempting "/hook" therefore also exempts "/hook-attacker" and "/hookx/steal".

Fix: normalise the prefix by trimming ONE trailing "/" (unless the prefix is exactly
"/", which must still exempt everything — keep that existing branch). Then match only
when `path == prefix` OR `strings.HasPrefix(path, prefix+"/")`.

Required test cases (table-driven), prefixes -> path -> expected:
  ["/hook"]   "/hook"            true
  ["/hook"]   "/hook/"           true
  ["/hook"]   "/hook/github"     true
  ["/hook"]   "/hook-attacker"   false
  ["/hook"]   "/hookx"           false
  ["/hook/"]  "/hook"            true
  ["/hook/"]  "/hook/github"     true
  ["hook"]    "/hook/a"          true    (missing leading slash is still added)
  ["/"]       "/anything"        true
  []          "/hook"            false
  ["  "]      "/hook"            false

=== BUG 2 (Medium, crash): Redirect panics on a request-less Response ===
server/response.go:

    func NewResponse(w http.ResponseWriter) *Response  // request: nil
    func (r *Response) Redirect(url string, code int) {
        http.Redirect(r, r.Request(), url, code)
    }

http.Redirect dereferences the request for relative URLs, so Redirect on a Response
built with NewResponse panics.

Fix: if r.request is nil, pass a minimal non-nil request instead:
    req := r.request
    if req == nil {
        req = &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/"}, Header: http.Header{}}
    }
    http.Redirect(r, req, target, code)
(rename the `url` parameter to `target` so it does not shadow net/url.)
Behaviour with a real request must be unchanged.

Required tests, using httptest.NewRecorder():
  - NewResponse(rec).Redirect("/login", 302)  -> no panic, status 302, Location "/login"
  - NewResponse(rec).Redirect("https://example.com/x", 301) -> 301, Location exact
  - NewResponseWithRequest(rec, httptest.NewRequest("GET","/a/b",nil)).Redirect("c", 302)
      -> Location "/a/c"   (proves request-relative resolution still works)

=== BUG 3 (High, security): session auth accepts inactive and locked users ===
identity/middleware/auth.go, session branch (around line 116):

    session, err := m.sessionRepo.GetByKey(ctx, sessionKey)
    if err == nil && !session.IsExpired() {
        user, err := m.userRepo.GetByID(ctx, session.UserID)
        if err == nil && user != nil {
            return user, nil
        }
    }

The user model (identity/models/user.go) has `IsActive bool` and `IsLocked bool`.
A deactivated or locked account keeps full access through any existing session.

Fix: only return the user when `user.IsActive && !user.IsLocked`. Otherwise fall
through to the existing `return nil, nil` (anonymous). Do not change any other branch
of the function. Also guard: if m.sessionRepo or m.userRepo is nil, skip the session
branch instead of panicking.

Required tests (use or extend the fakes/mocks already in that package's tests — read
them first; if none exist, write tiny in-test fakes implementing only the methods used):
  - active, unlocked user with valid session  -> user returned
  - IsActive=false                            -> nil user, nil error
  - IsLocked=true                             -> nil user, nil error
  - expired session                           -> nil user (unchanged behaviour)
  - no session key                            -> nil user (unchanged behaviour)

=== VERIFY (all must pass, run from the working directory above) ===
  gofmt -l server identity/middleware        (must print nothing)
  go vet ./server/... ./identity/...
  go test -race ./server/... ./identity/...
  go test ./...

=== HARD CONSTRAINTS ===
- Do NOT run any git command. The coordinator commits.
- Do NOT change public function signatures.
- Do NOT edit go.mod / go.sum.
- Minimal diffs. Do not reformat or refactor unrelated code.
- Report at the end: files changed, each test name added, and the exact output of
  the go test -race command.
