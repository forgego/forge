TASK D2: admin login and auth hygiene (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/server-security-edge-cases/forge
(git worktree on branch fix/server-security-edge-cases; D1 already landed here.)

Files you may modify — ONLY:
  admin/api/rest/router.go            (handleLogin, adminCredentials)
  the file that defines the type of `r.sessions` in admin/api/rest (find it: grep "func (.*) Issue(" admin/api/rest)
  log/middleware.go
  NEW admin/api/rest/login_limiter.go
  tests: existing *_test.go in admin/api/rest and log, or NEW ones in those packages
Do NOT edit go.mod / go.sum. Do NOT touch server/ or identity/ (already done).

Write failing tests FIRST.

=== BUG 1 (High): API keys and tokens written to access logs ===
log/middleware.go logs `zap.String("query", sanitizeLogString(r.URL.RawQuery))`.
api/authentication/apikey.go accepts the key from the `api_key` query parameter, so every
authenticated request writes the live key into the logs.

Fix: add an unexported func `redactQuery(raw string) string` in log/middleware.go:
  - parse with url.ParseQuery; on parse error return "[unparseable]"
  - for any key whose lowercased name is one of: api_key, apikey, key, token, access_token,
    refresh_token, password, secret, signature, session, session_key
    replace EVERY value with "REDACTED"
  - re-encode with url.Values.Encode()  (key order becomes sorted — acceptable)
  - empty input -> ""
Log `sanitizeLogString(redactQuery(r.URL.RawQuery))` instead.
Tests: "api_key=abc&page=2" -> contains "api_key=REDACTED" and "page=2", not "abc";
"Token=x" (case-insensitive) redacted; "" -> ""; "%zz" -> "[unparseable]".

=== BUG 2 (Medium): passwords are trimmed before comparison ===
router.go handleLogin: `payload.Password = strings.TrimSpace(payload.Password)`, and
adminCredentials trims FORGE_ADMIN_PASSWORD. A password with leading/trailing spaces can never be
used exactly as set, and "pw " is accepted for "pw".
Fix:
  - Do NOT trim payload.Password. Keep trimming the username.
  - The empty check becomes: username == "" || password == "".
  - In adminCredentials, trim the password env var ONLY of trailing "\r\n" (strings.TrimRight(v, "\r\n"))
    so values loaded from files with a newline still work. Keep TrimSpace on the username.
Tests (set env with t.Setenv): configured "p w " -> login with "p w " succeeds, "p w" fails (401).

=== BUG 3 (High): no brute-force protection on admin login ===
handleLogin has no limit on failed attempts.
Fix: create admin/api/rest/login_limiter.go:

    type loginLimiter struct {
        mu       sync.Mutex
        failures map[string]*failureWindow // key -> window
        max      int           // default 5
        window   time.Duration // default 15 * time.Minute
        now      func() time.Time
    }
    type failureWindow struct { count int; resetAt time.Time }

    func newLoginLimiter() *loginLimiter
    // blocked reports whether key is locked out and for how long.
    func (l *loginLimiter) blocked(key string) (bool, time.Duration)
    func (l *loginLimiter) fail(key string)
    func (l *loginLimiter) success(key string)   // deletes the key

  Windows are fixed: the first failure sets resetAt = now+window; after resetAt the entry resets.
  In fail(), also delete expired entries opportunistically when len(failures) > 10000.

  Wire into Router: add a field (e.g. `loginLimiter *loginLimiter`) initialised wherever the Router
  and r.sessions are constructed (find the constructor). In handleLogin, AFTER decoding the payload:
    ip  := host part of req.RemoteAddr (net.SplitHostPort; fall back to RemoteAddr)
    keys := "ip:"+ip and "user:"+strings.ToLower(username)
    if either key is blocked -> respond 429, code "too_many_attempts", message
       "Too many failed login attempts. Try again later.", and set header
       Retry-After to the ceil of the longer remaining duration in seconds.
    on invalid credentials -> fail() both keys, then the existing 401.
    on success -> success() both keys.
  Do NOT read X-Forwarded-For here.
Tests: with limiter now() stubbed or window small, 5 wrong passwords -> 6th request (even with the
CORRECT password) returns 429 with Retry-After > 0; after the window passes, correct password -> 200;
a different username from the same IP is also blocked (ip key); success clears counters.

=== BUG 4 (Medium): expired/revoked admin sessions accumulate forever ===
Read the type behind r.sessions (Issue / Revoke / validation). If it stores tokens in a map with
expiry and never deletes expired ones, add an unexported purgeExpiredLocked(now) that deletes expired
(and revoked, if revocation is tracked separately) entries, and call it from Issue while the lock is
held, at most once per minute (track lastPurge). If the store is not an in-memory map (e.g. DB-backed
or signed stateless tokens), make NO change for this bug and say so in the report.
Test (only if you changed it): issue a token with a 1ms TTL, sleep 5ms, issue another -> internal map
size is 1.

=== VERIFY ===
  gofmt -l admin/api/rest log      (your files must not appear)
  go vet ./admin/... ./log/...
  go test -race ./admin/api/rest/ ./log/
  go test ./...   (TestPrefetchRelated_Integration in ./orm is known-flaky; report, don't fix)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes. Minimal diff.
- Final report: files changed, tests added, what you did for BUG 4, exact tail of the -race run.
