package ratelimit

import (
	"sync"
	"time"
)

type windowEntry struct {
	count     int
	expiresAt time.Time
}

// FixedWindowCounter tracks request counts within a fixed time window.
type FixedWindowCounter struct {
	mu           sync.Mutex
	limit        int
	window       time.Duration
	entries      map[string]windowEntry
	now          func() time.Time
	lastEviction time.Time
}

// NewFixedWindowCounter creates a new fixed-window counter with the given limit and window.
func NewFixedWindowCounter(limit int, window time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:   limit,
		window:  window,
		entries: make(map[string]windowEntry),
		now:     time.Now,
	}
}

// Allow reports whether an event may happen for key and the retry duration if denied.
func (c *FixedWindowCounter) Allow(key string) (bool, time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	clock := c.now
	if clock == nil {
		clock = time.Now
	}
	current := clock()

	// Evict expired entries at most once per minute.
	if c.lastEviction.IsZero() || current.Before(c.lastEviction) {
		c.lastEviction = current
	} else if current.Sub(c.lastEviction) >= time.Minute {
		for k, entry := range c.entries {
			if !current.Before(entry.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.lastEviction = current
	}

	if c.limit <= 0 {
		if entry, ok := c.entries[key]; ok && current.Before(entry.expiresAt) {
			return false, entry.expiresAt.Sub(current)
		}
		return false, c.window
	}

	entry, ok := c.entries[key]
	if !ok || !current.Before(entry.expiresAt) {
		c.entries[key] = windowEntry{
			count:     1,
			expiresAt: current.Add(c.window),
		}
		return true, 0
	}

	if entry.count >= c.limit {
		retryAfter := entry.expiresAt.Sub(current)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter
	}

	entry.count++
	c.entries[key] = entry
	return true, 0
}
