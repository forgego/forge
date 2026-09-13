package throttling

import (
	"sync"
	"time"
)

// CacheBackend is the interface for throttle cache backends
type CacheBackend interface {
	// GetInt gets an integer value from cache
	GetInt(key string) (int, error)
	// Set sets a value in cache with TTL
	Set(key string, value interface{}, ttl time.Duration) error
	// GetTTL gets the remaining TTL for a key
	GetTTL(key string) time.Duration
	// Delete deletes a key from cache
	Delete(key string) error
}

// atomicCounter is implemented by caches that can check-and-increment atomically.
type atomicCounter interface {
	IncrementWithinLimit(key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// MemoryCache is an in-memory cache backend
type MemoryCache struct {
	mu   sync.Mutex
	data map[string]*cacheEntry
}

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]*cacheEntry),
	}
}

// GetInt gets an integer value from cache
func (c *MemoryCache) GetInt(key string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.data[key]
	if !ok {
		return 0, nil
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return 0, nil
	}

	if val, ok := entry.value.(int); ok {
		return val, nil
	}

	return 0, nil
}

// Set sets a value in cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// GetTTL gets the remaining TTL for a key
func (c *MemoryCache) GetTTL(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.data[key]
	if !ok {
		return 0
	}

	if time.Now().After(entry.expiresAt) {
		delete(c.data, key)
		return 0
	}

	return time.Until(entry.expiresAt)
}

// Delete deletes a key from cache
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// IncrementWithinLimit checks and increments the counter atomically within limit
func (c *MemoryCache) IncrementWithinLimit(key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if limit <= 0 {
		if entry, ok := c.data[key]; ok && now.Before(entry.expiresAt) {
			return false, entry.expiresAt.Sub(now), nil
		}
		return false, window, nil
	}

	entry, ok := c.data[key]
	if !ok || now.After(entry.expiresAt) {
		c.data[key] = &cacheEntry{
			value:     1,
			expiresAt: now.Add(window),
		}
		return true, 0, nil
	}

	count, ok := entry.value.(int)
	if !ok {
		c.data[key] = &cacheEntry{
			value:     1,
			expiresAt: now.Add(window),
		}
		return true, 0, nil
	}

	if count >= limit {
		retryAfter := entry.expiresAt.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, retryAfter, nil
	}

	entry.value = count + 1
	return true, 0, nil
}
