package server

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// 1. Stop releases goroutine: before := runtime.NumGoroutine(); s := newRateLimitStore(rate.Limit(1), 1);
// s.stop(); poll up to 1s (10ms sleeps) until NumGoroutine() <= before; fail otherwise.
func TestRateLimitStore_StopReleasesGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()
	s := newRateLimitStore(rate.Limit(1), 1)
	s.stop()

	deadline := time.Now().Add(1 * time.Second)
	released := false
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			released = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !released {
		t.Fatalf("goroutine leak: NumGoroutine did not decrease to <= %d (got %d)", before, runtime.NumGoroutine())
	}
}

// 2. stop() twice does not panic.
func TestRateLimitStore_StopTwiceDoesNotPanic(t *testing.T) {
	s := newRateLimitStore(rate.Limit(1), 1)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("stop() panicked on second call: %v", r)
		}
	}()
	s.stop()
	s.stop()
}

// 3. evictIdle removes only idle keys: getLimiter("idle"); getLimiter("active");
// set idle's lastSeen to now-11min via the entry; evictIdle(time.Now());
// "idle" gone, "active" present.
func TestRateLimitStore_EvictIdle(t *testing.T) {
	s := newRateLimitStore(rate.Limit(1), 1)
	defer s.stop()

	s.getLimiter("idle")
	s.getLimiter("active")

	s.mu.Lock()
	idleEntry := s.limiters["idle"]
	if idleEntry == nil {
		s.mu.Unlock()
		t.Fatal("expected 'idle' entry to exist")
	}
	idleEntry.lastSeen.Store(time.Now().Add(-11 * time.Minute).UnixNano())
	s.mu.Unlock()

	s.evictIdle(time.Now())

	s.mu.RLock()
	_, idleExists := s.limiters["idle"]
	_, activeExists := s.limiters["active"]
	s.mu.RUnlock()

	if idleExists {
		t.Error("expected 'idle' key to be evicted")
	}
	if !activeExists {
		t.Error("expected 'active' key to remain")
	}
}

// 4. Active abusive client keeps its state across eviction: s with burst 1; getLimiter("x").Allow()
// == true; evictIdle(now) (x is active); getLimiter("x").Allow() == false (limit NOT reset).
func TestRateLimitStore_ActiveAbusiveClientRetainsLimit(t *testing.T) {
	s := newRateLimitStore(rate.Limit(1), 1)
	defer s.stop()

	limiter := s.getLimiter("x")
	if !limiter.Allow() {
		t.Fatal("expected first request for 'x' to be allowed")
	}

	s.evictIdle(time.Now())

	if s.getLimiter("x").Allow() {
		t.Fatal("expected limit NOT to be reset across eviction for active client")
	}
}

// 5. Concurrency: 50 goroutines × 500 getLimiter calls on 20 keys while another goroutine calls
// evictIdle in a loop; run under -race.
func TestRateLimitStore_Concurrency(t *testing.T) {
	s := newRateLimitStore(rate.Limit(10), 10)
	defer s.stop()

	const (
		numGoroutines = 50
		numCalls      = 500
		numKeys       = 20
	)

	var wg sync.WaitGroup
	stopEvict := make(chan struct{})

	// Goroutine calling evictIdle in a loop
	go func() {
		for {
			select {
			case <-stopEvict:
				return
			default:
				s.evictIdle(time.Now())
				runtime.Gosched()
			}
		}
	}()

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numCalls; j++ {
				key := fmt.Sprintf("key-%d", (id+j)%numKeys)
				lim := s.getLimiter(key)
				if lim == nil {
					t.Errorf("getLimiter returned nil for %s", key)
				}
			}
		}(i)
	}

	wg.Wait()
	close(stopEvict)
}
