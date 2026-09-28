package caching

import "time"

// CacheBackend is the interface for cache backends
type CacheBackend interface {
	// Get gets a value from cache
	Get(key string) (interface{}, error)
	// Set sets a value in cache with TTL
	Set(key string, value interface{}, ttl time.Duration) error
	// Delete deletes a key from cache
	Delete(key string) error
	// Clear clears all cache
	Clear() error
	// Exists checks if a key exists
	Exists(key string) bool
}

// CacheKeyGenerator generates cache keys
type CacheKeyGenerator interface {
	// GenerateKey generates a cache key
	GenerateKey(parts ...string) string
}

// DefaultCacheKeyGenerator is the default cache key generator
type DefaultCacheKeyGenerator struct {
	Prefix string
}

// NewDefaultCacheKeyGenerator creates a new cache key generator
func NewDefaultCacheKeyGenerator(prefix string) *DefaultCacheKeyGenerator {
	return &DefaultCacheKeyGenerator{
		Prefix: prefix,
	}
}

// GenerateKey generates a cache key from parts
func (g *DefaultCacheKeyGenerator) GenerateKey(parts ...string) string {
	key := g.Prefix
	for _, part := range parts {
		if key != "" {
			key += ":"
		}
		key += part
	}
	return key
}
