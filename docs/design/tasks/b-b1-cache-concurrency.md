TASK B1: make three in-memory caches concurrency-safe and stoppable (Go backend).

Working directory: /home/hamid/Other/projects/foreit-wt/cache-concurrency-lifecycle/forge
(git worktree on branch fix/cache-concurrency-lifecycle. Work ONLY there.)

Files you may modify — ONLY:
  api/throttling/cache.go      + api/throttling/throttle_test.go (or a new cache_test.go there)
  api/caching/memory.go        + api/caching/cache_test.go
  filter/cache.go              + NEW filter/cache_test.go
Do NOT touch server/*, api/throttling/anon_rate.go, user_rate.go — a later task owns them.
Do NOT edit go.mod / go.sum.

Write the failing tests FIRST (they will fail or crash under -race), then fix.

=== BUG 1 (High): api/throttling/cache.go has no lock at all ===
    type MemoryCache struct { data map[string]*cacheEntry }
GetInt reads AND deletes, Set writes, GetTTL reads — all unsynchronised. Concurrent HTTP
requests hit this map -> data race / "fatal error: concurrent map read and map write".

Fix: add `mu sync.Mutex` to MemoryCache and hold it for the whole body of EVERY method
on the type (read the file; lock all of them, including ones not named here).

=== BUG 2 (High): two caches delete map entries while holding only RLock ===
api/caching/memory.go  Get():
    c.mutex.RLock(); defer c.mutex.RUnlock()
    ... if time.Now().After(item.expiresAt) { delete(c.data, key) ... }   // write under READ lock
filter/cache.go  GetParsedTree() (and very likely its sibling getters for compiledSQL and
metadata — read the whole file):
    c.mu.RLock(); defer c.mu.RUnlock()
    ... delete(c.parsedTrees, key)                                          // same bug

Several readers hold RLock simultaneously, so concurrent deletes race.

Fix pattern for every such getter:
    c.mu.RLock()
    entry, ok := c.m[key]
    c.mu.RUnlock()
    if !ok { return zero, false }
    if time.Now().After(entry.ExpiresAt) {
        c.mu.Lock()
        // re-check: another goroutine may have replaced it
        if cur, ok := c.m[key]; ok && time.Now().After(cur.ExpiresAt) {
            delete(c.m, key)
        }
        c.mu.Unlock()
        return zero, false
    }
    return entry.value, true
Never call delete on a map while holding only RLock anywhere in these files.

=== BUG 3 (Medium): cleanup goroutines can never be stopped ===
Both NewMemoryCache (api/caching) and NewFilterCache (filter) do `go cache.cleanup()`,
an infinite ticker loop. Every cache constructed leaks a goroutine forever (tests, per-request
caches, hot reload).

Fix, in BOTH types:
  - add fields `stop chan struct{}` and `closeOnce sync.Once`; create the chan in the constructor.
  - add exported method
        // Close stops the background cleanup goroutine. Safe to call multiple times.
        func (c *T) Close() error { c.closeOnce.Do(func(){ close(c.stop) }); return nil }
  - cleanup loop becomes:
        ticker := time.NewTicker(<existing interval>)
        defer ticker.Stop()
        for { select { case <-ticker.C: <existing body>; case <-c.stop: return } }
    keep the existing interval value and the existing cleanup body (which must use the
    WRITE lock when deleting — check it).
  - Add Close ONLY as a method on the concrete struct. If an interface (e.g. a Cache
    interface) exists in that package, do NOT add Close to the interface.

=== TESTS (required) ===
For each of the three caches:
  1. Concurrency: 50 goroutines × 200 iterations doing mixed Set / Get(/GetInt) / GetTTL
     on a small key set (e.g. 5 keys) with a tiny TTL (1ms) so expiry-deletes happen.
     Use sync.WaitGroup. This test must pass under `go test -race`.
  2. Expiry: Set with 10ms TTL, sleep 30ms, Get reports missing.
For api/caching and filter caches additionally:
  3. Close stops the goroutine: record runtime.NumGoroutine() before constructing, construct,
     Close(), then poll up to 1s (loop with 10ms sleeps) until NumGoroutine() <= before.
     Fail if it never returns.
  4. Close() called twice does not panic.

=== VERIFY (from working directory) ===
  gofmt -l api filter                     (prints nothing)
  go vet ./api/... ./filter/...
  go test -race -count=3 ./api/caching/ ./api/throttling/ ./filter/
  go test ./...

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Do not change any existing exported signature; only ADD Close().
- Minimal diffs, no unrelated refactors.
- Final report: files changed, tests added, exact tail of the -race command.
