Your workspace is /home/hamid/Other/projects/foreit-wt/wave1. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

ROLE: You are the delegated coder. The rule in CLAUDE.md/AGENTS.md that forbids the primary agent from editing code applies to Claude, not to you: edit the files yourself. Do NOT call any other agent. Use US spelling in comments.

# API throttles keep their original fixed-window semantics

Module root: /home/hamid/Other/projects/foreit-wt/wave1/forge (run go commands from there).
Only edit forge/internal/ratelimit (new file fixed_window.go + test) and forge/api/throttling (anon_rate.go, user_rate.go, tests). Another agent edits forge/api/helpers.go, forge/filter, forge/identity and forge/db at the same time: do not touch those; ignore build errors there.
A read-only copy of master is at /home/hamid/Other/projects/foreit-wt/verify-master/forge (read forge/api/throttling/cache.go and user_rate.go there; never edit it).

## Problem (verified)
Before the cleanup, `UserRateThrottle`/`AnonRateThrottle` counted with a fixed window (`MemoryCache.IncrementWithinLimit` in master's forge/api/throttling/cache.go): the first request in a key's window sets `expiresAt = now + window`; requests are allowed while `count < limit`; once `count >= limit` they are denied with `retryAfter = expiresAt - now`; after `expiresAt` the counter restarts at 1; `limit <= 0` always denies with `retryAfter` = remaining window (or the full window when no entry exists).
The cleanup made the default store `internal/ratelimit.KeyedLimiter` (token bucket, burst = limit, refill = window/limit). That allows up to about 2x the limit within the first window and reports different retry times: a behaviour change for "N per period" throttles.

## Change
1. forge/internal/ratelimit/fixed_window.go: `type FixedWindowCounter struct` with a mutex, `map[string]windowEntry{count int; expiresAt time.Time}`, an injectable clock (`now func() time.Time`, default `time.Now`), and idle eviction of expired entries at most once per minute (same approach as KeyedLimiter's eviction; no background goroutine required). `func NewFixedWindowCounter(limit int, window time.Duration) *FixedWindowCounter` and `func (c *FixedWindowCounter) Allow(key string) (bool, time.Duration)` implementing exactly the semantics above.
2. forge/api/throttling: the default store for `NewUserRateThrottle(rate)` and `NewAnonRateThrottle(rate)` becomes `ratelimit.NewFixedWindowCounter(limit, window)` from the parsed rate. `...WithStore` / `WithStore` keep accepting any `Store`. Leave server middleware (forge/server/ratelimit.go) on KeyedLimiter.
3. Tests (table-driven, injected clock, no sleeps):
   - "3/minute": calls 1-3 allowed; call 4 denied with retryAfter == remaining window; 6 calls inside one window allow exactly 3.
   - after the window passes, the next call is allowed and the count restarts.
   - limit 0 denies.
   - separate keys are independent; concurrent Allow from 50 goroutines on one key allows exactly `limit` (run under -race).
   - throttle level: a UserRateThrottle at "2/minute" allows 2 then returns ThrottledError with WaitDuration > 0.

## Acceptance (from the module root)
    gofmt -l internal api/throttling                    # empty
    go vet ./internal/... ./api/throttling/...
    staticcheck ./internal/... ./api/throttling/...     # no output
    go test -race ./internal/... ./api/throttling/... -count=1
    ~/go/bin/golangci-lint run --timeout=5m --config=../.golangci.yml ./internal/... ./api/throttling/...   # 0 issues

Report (max 15 lines): files changed, tests, each command result.
