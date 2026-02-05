package throttling

import "time"

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

// MemoryCache is an in-memory cache backend
type MemoryCache struct {
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
	c.data[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// GetTTL gets the remaining TTL for a key
func (c *MemoryCache) GetTTL(key string) time.Duration {
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
	delete(c.data, key)
	return nil
}
