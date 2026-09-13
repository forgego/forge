TASK B2: atomic throttle counting + stop trusting spoofable forwarding headers (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/cache-concurrency-lifecycle/forge
(git worktree on branch fix/cache-concurrency-lifecycle. Task B1 already landed here:
api/throttling/cache.go MemoryCache now has a mutex. Read it first.)

Files you may modify — ONLY:
  api/throttling/cache.go, api/throttling/anon_rate.go, api/throttling/user_rate.go
  server/ratelimit.go
  NEW package: netutil/clientip.go  +  netutil/clientip_test.go
  tests: api/throttling/throttle_test.go (or new *_test.go in that package),
         NEW server/clientip_test.go
Do NOT edit go.mod / go.sum. Do NOT touch the rate-limit store cleanup logic in
server/ratelimit.go (newRateLimitStore / cleanupExpired / stop) — task B3 owns it.

Write failing tests FIRST.

=== BUG 1 (High): throttle check-then-increment is not atomic ===
api/throttling/anon_rate.go checkRate (and the identical block in user_rate.go checkRate):

    count, err := t.Cache.GetInt(key)
    if count >= limit { return false, t.Cache.GetTTL(key), nil }
    newCount := count + 1
    t.Cache.Set(key, newCount, duration)

Each call is locked individually, but N concurrent requests all read count=K and all write
K+1, so far more than `limit` requests get through.

Fix:
1. In api/throttling (cache.go is fine), declare an OPTIONAL interface — do NOT change the
   existing CacheBackend interface (external backends implement it):

     // atomicCounter is implemented by caches that can check-and-increment atomically.
     type atomicCounter interface {
         IncrementWithinLimit(key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
     }

2. Implement it on *MemoryCache, holding its mutex for the whole operation:
     - entry missing or expired      -> store count 1, expiresAt = now+window; return true, 0, nil
     - stored count >= limit         -> return false, expiresAt-now (never negative), nil
     - otherwise                     -> count++ keeping the SAME expiresAt (fixed window); true, 0, nil
   Treat a non-int stored value as missing.

3. In BOTH checkRate functions, after parseRate succeeds:
     if c, ok := t.Cache.(atomicCounter); ok {
         return c.IncrementWithinLimit(key, limit, duration)
     }
   and keep the existing GetInt/Set code below as the fallback for other backends.

Tests:
  - 200 goroutines call AnonRateThrottle.AllowRequest concurrently for the same client with
    Rate "50/hour" and a MemoryCache. Count allowed==true. Must be EXACTLY 50. Run with -race.
  - Same for UserRateThrottle if its test setup is straightforward (read user_rate.go for how
    the user/scope is derived; skip with a comment if it needs heavy fixtures).
  - Sequential: limit 3 -> allowed, allowed, allowed, denied with retryAfter > 0.
  - Window expiry: Rate with a tiny window if parseRate supports seconds ("3/second" — check
    parseRate); after the window elapses, requests are allowed again.

=== BUG 2 (High, security): client IP taken from client-controlled headers ===
server/ratelimit.go getClientIP returns the FIRST X-Forwarded-For entry, else X-Real-IP, else
RemoteAddr. api/throttling/anon_rate.go getClientIP returns the WHOLE X-Forwarded-For header,
else X-Real-IP, else RemoteAddr (with port). Any client sets `X-Forwarded-For: <random>` per
request and gets a fresh rate-limit bucket every time -> rate limiting is bypassed.

Fix:
1. Create package netutil (forge/netutil/clientip.go):

     // SetTrustedProxies configures which peers may supply X-Forwarded-For / X-Real-IP.
     // Accepts CIDRs ("10.0.0.0/8") or bare IPs ("127.0.0.1"). Empty = trust no one (default).
     func SetTrustedProxies(entries []string) error
     // TrustedProxies returns the current list (safe for concurrent use).
     func TrustedProxies() []netip.Prefix
     // ClientIP returns the best-effort client IP for rate limiting.
     func ClientIP(r *http.Request, trusted []netip.Prefix) string

   Store the list in a sync/atomic.Pointer[[]netip.Prefix] (or RWMutex). SetTrustedProxies
   must be all-or-nothing: on any invalid entry return an error and leave the list unchanged.

   ClientIP algorithm:
     a. peer := host part of r.RemoteAddr via net.SplitHostPort; if that errors, use RemoteAddr
        as-is. Parse with netip.ParseAddr (unmap IPv4-in-IPv6 with .Unmap()).
     b. If peer does not parse, or is NOT contained in any trusted prefix -> return peer string.
        (Headers are ignored entirely for untrusted peers.)
     c. Peer is trusted: split X-Forwarded-For on ",", trim spaces, walk from the RIGHT;
        return the first entry that parses as an IP and is NOT in a trusted prefix.
     d. If none found: if X-Real-IP parses as an IP, return it.
     e. Otherwise return peer.

2. server/ratelimit.go: getClientIP body becomes `return netutil.ClientIP(r, netutil.TrustedProxies())`.
   Delete splitIPs / splitHostPort ONLY if nothing else in package server uses them (grep first).
3. api/throttling/anon_rate.go: getClientIP body becomes the same one-liner.

Tests (netutil/clientip_test.go, table-driven), trusted = ["10.0.0.0/8"] unless stated:
  RemoteAddr "203.0.113.9:5555", XFF "1.2.3.4"                      -> "203.0.113.9"  (untrusted peer, header ignored)
  RemoteAddr "203.0.113.9:5555", X-Real-IP "1.2.3.4"                -> "203.0.113.9"
  RemoteAddr "10.0.0.5:80",      XFF "1.2.3.4, 198.51.100.7"        -> "198.51.100.7" (rightmost untrusted)
  RemoteAddr "10.0.0.5:80",      XFF "1.2.3.4, 10.0.0.9"            -> "1.2.3.4"
  RemoteAddr "10.0.0.5:80",      XFF "garbage, 198.51.100.7"        -> "198.51.100.7"
  RemoteAddr "10.0.0.5:80",      no XFF, X-Real-IP "198.51.100.8"   -> "198.51.100.8"
  RemoteAddr "10.0.0.5:80",      XFF "10.1.1.1"                      -> "10.0.0.5"
  RemoteAddr "[::1]:80", trusted ["::1"], XFF "2001:db8::1"          -> "2001:db8::1"
  RemoteAddr "203.0.113.9" (no port)                                 -> "203.0.113.9"
  trusted empty, RemoteAddr "10.0.0.5:80", XFF "1.2.3.4"             -> "10.0.0.5"
  SetTrustedProxies(["10.0.0.0/8","bad"]) -> error AND TrustedProxies() unchanged.
And server/clientip_test.go: a request through RateLimitByIP(1, time.Minute) — two requests from
the same RemoteAddr with DIFFERENT X-Forwarded-For values: second must get 429. (Reset the
trusted list with netutil.SetTrustedProxies(nil) at test start.)

=== VERIFY ===
  gofmt -l api server netutil          (prints nothing)
  go vet ./...
  go test -race -count=2 ./api/throttling/ ./server/ ./netutil/
  go test ./...        (TestPrefetchRelated_Integration in ./orm is a KNOWN pre-existing failure; ignore it)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Do not change exported signatures or the CacheBackend interface.
- Minimal diffs.
- Final report: files changed, tests added, exact tail of the -race run.
