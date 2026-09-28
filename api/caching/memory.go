package caching

import (
	"sync"
	"time"
)

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	data      map[string]*cacheItem
	mutex     sync.RWMutex
	stop      chan struct{}
	closeOnce sync.Once
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]*cacheItem),
		stop: make(chan struct{}),
	}
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

// Close stops the background cleanup goroutine. Safe to call multiple times.
func (c *MemoryCache) Close() error {
	c.closeOnce.Do(func() {
		close(c.stop)
	})
	return nil
}

// Get gets a value from cache
func (c *MemoryCache) Get(key string) (interface{}, error) {
	c.mutex.RLock()
	item, ok := c.data[key]
	c.mutex.RUnlock()
	if !ok {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		c.mutex.Lock()
		// re-check: another goroutine may have replaced it
		if cur, ok := c.data[key]; ok && time.Now().After(cur.expiresAt) {
			delete(c.data, key)
		}
		c.mutex.Unlock()
		return nil, nil
	}

	return item.value, nil
}

// Set sets a value in cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete deletes a key from cache
func (c *MemoryCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	return nil
}

// Clear clears all cache
func (c *MemoryCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*cacheItem)
	return nil
}

// Exists checks if a key exists
func (c *MemoryCache) Exists(key string) bool {
	c.mutex.RLock()
	item, ok := c.data[key]
	c.mutex.RUnlock()
	if !ok {
		return false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		c.mutex.Lock()
		// re-check: another goroutine may have replaced it
		if cur, ok := c.data[key]; ok && time.Now().After(cur.expiresAt) {
			delete(c.data, key)
		}
		c.mutex.Unlock()
		return false
	}

	return true
}

// cleanup periodically removes expired items
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			now := time.Now()
			for key, item := range c.data {
				if now.After(item.expiresAt) {
					delete(c.data, key)
				}
			}
			c.mutex.Unlock()
		case <-c.stop:
			return
		}
	}
}
