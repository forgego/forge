package orm

import (
	"fmt"
	"sync"
)

// PathCache provides LRU caching for compiled path tokens
type PathCache struct {
	cache map[string]*PathToken
	mu    sync.RWMutex
	maxSize int
}

// NewPathCache creates a new path cache with the specified maximum size
func NewPathCache(maxSize int) *PathCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default cache size
	}
	return &PathCache{
		cache:   make(map[string]*PathToken),
		maxSize: maxSize,
	}
}

// Get retrieves a compiled path token, compiling and caching if necessary
func (pc *PathCache) Get(path string, schema *ModelSchema) (*PathToken, error) {
	pc.mu.RLock()
	if token, ok := pc.cache[path]; ok {
		pc.mu.RUnlock()
		return token, nil
	}
	pc.mu.RUnlock()

	// Compile path
	token, err := schema.CompilePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to compile path %s: %w", path, err)
	}

	// Cache it
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Check if we need to evict (simple FIFO for now)
	if len(pc.cache) >= pc.maxSize {
		// Remove oldest entry (simplified - in production use proper LRU)
		for k := range pc.cache {
			delete(pc.cache, k)
			break
		}
	}

	pc.cache[path] = token
	return token, nil
}

// Clear clears the cache
func (pc *PathCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache = make(map[string]*PathToken)
}

// Size returns the current cache size
func (pc *PathCache) Size() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.cache)
}

// Global path cache instance
var globalPathCache = NewPathCache(1000)

// GetCachedPathToken gets a path token from the global cache
func GetCachedPathToken(path string, schema *ModelSchema) (*PathToken, error) {
	return globalPathCache.Get(path, schema)
}
