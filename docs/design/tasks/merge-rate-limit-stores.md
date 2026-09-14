Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# One rate-limit store: API throttles use the server's x/time/rate keyed store; login limiter evicts stale entries

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit forge/server/ratelimit.go (+ a new forge/server/keyed_limiter.go if you split it), forge/api/throttling/*, forge/api/api.go and forge/api/helpers.go where they construct throttles, forge/admin/api/rest/login_limiter.go, and their tests.

## Principles (mandatory)
- Tests FIRST. gofmt; functions < 40 lines; no new package-level mutable state; naming after behaviour.
- DRF semantics stay: `"100/hour"` rate strings, `ThrottledError.WaitDuration`, anon throttles key by client IP, user throttles key by user id and fall back to IP for anonymous users.

## Facts (verified on master)
- forge/server/ratelimit.go has a keyed `rateLimitStore` on `golang.org/x/time/rate` with idle eviction (`newRateLimitStore`, `getLimiter`, `evictIdle`, `cleanupExpired`, `stop`).
- forge/api/throttling implements a second algorithm: fixed windows stored in its own `MemoryCache` (cache.go, 146 lines, `CacheBackend` interface), used by `UserRateThrottle` / `AnonRateThrottle` (`checkRate`).
- forge/admin/api/rest/login_limiter.go counts failed logins with lockout; its cleanup only runs above 10k entries and holds the lock while iterating, so rotating usernames grows memory (B30). Its lockout semantics are different from request rate limiting: keep them.

## Change
1. Export a small keyed limiter from package server: `type KeyedLimiter struct{...}`, `func NewKeyedLimiter(requests int, window time.Duration) *KeyedLimiter`, `func (l *KeyedLimiter) Reserve(key string) (allowed bool, retryAfter time.Duration)`, `func (l *KeyedLimiter) Close()`. Implement it on the existing `rateLimitStore` (limit = `rate.Every(window/requests)`, burst = `requests`); make `RateLimitByIP`/`ByUser`/`General` use it so there is one implementation. `retryAfter` comes from `Reservation.Delay()` (cancel the reservation when not allowed).
2. api/throttling: `UserRateThrottle` and `AnonRateThrottle` hold a `*server.KeyedLimiter` created from the parsed rate. Change constructors to `NewUserRateThrottle(rate string)` / `NewAnonRateThrottle(rate string)` (drop the cache parameter), return `ThrottledError{WaitDuration: retryAfter}` on deny. Delete cache.go (`CacheBackend`, `MemoryCache`) and its tests after grepping the workspace (examples, cli templates, docs-site) for users; update every caller, including forge/api/api.go and helpers.go. Check for an import cycle (`go list -deps ./server | grep forge/api`): if server imports api, put `KeyedLimiter` in a new package forge/internal/ratelimit instead and use it from both.
3. login_limiter.go: evict entries whose lockout and failure window have expired on a time-based schedule (at most once per minute, like the server store) instead of only above 10k entries; keep the public behaviour and existing tests passing. Add a test with an injected clock (`getNow` already exists) showing that 20k distinct usernames with expired windows are evicted.
4. Tests: throttling — 3 requests allowed at "3/minute", the 4th denied with WaitDuration > 0; anonymous vs user keys are separate. server — KeyedLimiter allows burst then denies with retryAfter > 0; different keys are independent.

## Acceptance (from the module root)
    gofmt -l server api admin                         # empty
    go build ./... && (cd ../examples/ecommerce && go build ./...)
    go vet ./server/... ./api/... ./admin/...
    staticcheck ./server/... ./api/... ./admin/...     # no output
    go test -race ./server/... ./api/... ./admin/... -count=1

Report (max 25 lines): files deleted, constructor changes and callers updated, each command result.
