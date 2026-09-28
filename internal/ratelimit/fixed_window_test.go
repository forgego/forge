package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedWindowCounter(t *testing.T) {
	start := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	now := start
	counter := NewFixedWindowCounter(3, time.Minute)
	counter.now = func() time.Time { return now }

	for range 3 {
		allowed, retryAfter := counter.Allow("client")
		assert.True(t, allowed)
		assert.Zero(t, retryAfter)
	}
	allowed, retryAfter := counter.Allow("client")
	assert.False(t, allowed)
	assert.Equal(t, time.Minute, retryAfter)

	now = now.Add(20 * time.Second)
	for range 2 {
		allowed, _ = counter.Allow("client")
		assert.False(t, allowed)
	}
	allowed, retryAfter = counter.Allow("client")
	assert.False(t, allowed)
	assert.Equal(t, 40*time.Second, retryAfter)

	now = start.Add(time.Minute)
	allowed, retryAfter = counter.Allow("client")
	assert.True(t, allowed)
	assert.Zero(t, retryAfter)
}

func TestFixedWindowCounterZeroLimitDenies(t *testing.T) {
	start := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	now := start
	counter := NewFixedWindowCounter(0, time.Minute)
	counter.now = func() time.Time { return now }

	allowed, retryAfter := counter.Allow("client")
	assert.False(t, allowed)
	assert.Equal(t, time.Minute, retryAfter)
	now = now.Add(15 * time.Second)
	allowed, retryAfter = counter.Allow("client")
	assert.False(t, allowed)
	// limit 0 never stores an entry, so every call reports the full window (original behavior).
	assert.Equal(t, time.Minute, retryAfter)
}

func TestFixedWindowCounterSeparateKeys(t *testing.T) {
	counter := NewFixedWindowCounter(1, time.Minute)
	allowed, _ := counter.Allow("first")
	assert.True(t, allowed)
	allowed, _ = counter.Allow("second")
	assert.True(t, allowed)
}

func TestFixedWindowCounterConcurrent(t *testing.T) {
	const limit = 7
	counter := NewFixedWindowCounter(limit, time.Minute)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := counter.Allow("client"); ok {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(limit), allowed.Load())
}
