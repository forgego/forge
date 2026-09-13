TASK B3: rate-limit store — goroutine leak and arbitrary eviction (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/cache-concurrency-lifecycle/forge
(git worktree on branch fix/cache-concurrency-lifecycle. Tasks B1 and B2 already landed here.
Use absolute paths.)

Files you may modify — ONLY:
  server/ratelimit.go          (store type, newRateLimitStore, getLimiter, cleanupExpired, stop,
                                and the two get*RateLimitStore functions)
  NEW server/ratelimit_store_test.go
Do NOT touch getClientIP or anything in netutil/ or api/ (B2 owns those).
Do NOT edit go.mod / go.sum.

Write failing tests FIRST.

=== BUG 1 (Medium): stop() leaks the cleanup goroutine ===
    func newRateLimitStore(...) { ... cleanup: time.NewTicker(5*time.Minute) ...; go store.cleanupExpired() }
    func (s *rateLimitStore) cleanupExpired() { for range s.cleanup.C { ... } }
    func (s *rateLimitStore) stop() { s.cleanup.Stop() }

time.Ticker.Stop does NOT close the channel, so `for range s.cleanup.C` blocks forever after
stop(). The goroutine is never released.

Fix:
  - add fields `done chan struct{}` and `stopOnce sync.Once`; create `done` in the constructor.
  - cleanupExpired becomes:
        defer s.cleanup.Stop()
        for {
            select {
            case <-s.cleanup.C:
                s.evictIdle(time.Now())
            case <-s.done:
                return
            }
        }
  - stop(): s.stopOnce.Do(func() { close(s.done) })

=== BUG 2 (Low/Medium, security): eviction keeps an arbitrary half of the map ===
cleanupExpired, once len > 1000, copies "the first half" of a Go map (random iteration order)
into a new map. It can drop an ACTIVE abusive client (resetting its limit) while keeping idle ones.

Fix — expire by idle time instead:
  - change the map to `limiters map[string]*limiterEntry` with
        type limiterEntry struct {
            limiter  *rate.Limiter
            lastSeen atomic.Int64 // unix nanos
        }
  - getLimiter updates lastSeen on EVERY call (fast path under RLock is fine because lastSeen
    is atomic) and returns entry.limiter.
  - add field `idleTTL time.Duration` (constructor default 10*time.Minute) and
        func (s *rateLimitStore) evictIdle(now time.Time) {
            s.mu.Lock(); defer s.mu.Unlock()
            cutoff := now.Add(-s.idleTTL).UnixNano()
            for k, e := range s.limiters { if e.lastSeen.Load() < cutoff { delete(s.limiters, k) } }
        }
    Remove the 1000-entry / half-map logic entirely.
  - Keep newRateLimitStore(r, burst)'s signature; set idleTTL inside it.

=== TESTS (server/ratelimit_store_test.go, package server) ===
  1. Stop releases goroutine: before := runtime.NumGoroutine(); s := newRateLimitStore(rate.Limit(1), 1);
     s.stop(); poll up to 1s (10ms sleeps) until NumGoroutine() <= before; fail otherwise.
  2. stop() twice does not panic.
  3. evictIdle removes only idle keys: getLimiter("idle"); getLimiter("active");
     set idle's lastSeen to now-11min via the entry; evictIdle(time.Now());
     "idle" gone, "active" present.
  4. Active abusive client keeps its state across eviction: s with burst 1; getLimiter("x").Allow()
     == true; evictIdle(now) (x is active); getLimiter("x").Allow() == false (limit NOT reset).
  5. Concurrency: 50 goroutines × 500 getLimiter calls on 20 keys while another goroutine calls
     evictIdle in a loop; run under -race.

=== VERIFY ===
  gofmt -l server                 (prints nothing)
  go vet ./server/...
  go test -race -count=3 ./server/
  go test ./...   (TestPrefetchRelated_Integration in ./orm is a KNOWN pre-existing failure)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No exported signature changes. Minimal diff.
- Final report: files changed, tests added, exact tail of the -race run.
