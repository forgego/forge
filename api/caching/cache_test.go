package caching

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("key1", "value1", time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)
}

func TestMemoryCache_GetInt(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("key1", 42, time.Minute)
	require.NoError(t, err)

	value, err := cache.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("key1", "value1", 100*time.Millisecond)
	require.NoError(t, err)

	// Should exist immediately
	value, err := cache.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	value, err = cache.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", "value1", time.Minute)
	cache.Delete("key1")

	value, err := cache.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()

	assert.False(t, cache.Exists("key1"))

	cache.Set("key1", "value1", time.Minute)
	assert.True(t, cache.Exists("key1"))

	cache.Delete("key1")
	assert.False(t, cache.Exists("key1"))
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("key1", "value1", time.Minute)
	cache.Set("key2", "value2", time.Minute)

	err := cache.Clear()
	require.NoError(t, err)

	assert.False(t, cache.Exists("key1"))
	assert.False(t, cache.Exists("key2"))
}

func TestDefaultCacheKeyGenerator(t *testing.T) {
	generator := NewDefaultCacheKeyGenerator("api")

	key := generator.GenerateKey("users", "123")
	assert.Equal(t, "api:users:123", key)
}

func TestDefaultCacheKeyGenerator_NoPrefix(t *testing.T) {
	generator := NewDefaultCacheKeyGenerator("")

	key := generator.GenerateKey("users", "123")
	assert.Equal(t, "users:123", key)
}

func TestDefaultCacheKeyGenerator_MultipleParts(t *testing.T) {
	generator := NewDefaultCacheKeyGenerator("cache")

	key := generator.GenerateKey("users", "123", "orders", "456")
	assert.Equal(t, "cache:users:123:orders:456", key)
}

func TestMemoryCache_Concurrency(t *testing.T) {
	cache := NewMemoryCache()
	var wg sync.WaitGroup
	numGoroutines := 50
	iterations := 200

	keys := []string{"key-0", "key-1", "key-2", "key-3", "key-4"}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if j%10 == 0 {
					time.Sleep(2 * time.Millisecond)
				}
				key := keys[(gID+j)%len(keys)]
				switch (gID + j) % 3 {
				case 0:
					_ = cache.Set(key, j, 1*time.Millisecond)
				case 1:
					_, _ = cache.Get(key)
				case 2:
					_ = cache.Exists(key)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryCache_Expiry(t *testing.T) {
	cache := NewMemoryCache()
	err := cache.Set("expiry-key", "value", 10*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond)

	val, err := cache.Get("expiry-key")
	require.NoError(t, err)
	assert.Nil(t, val)
	assert.False(t, cache.Exists("expiry-key"))
}

func TestMemoryCache_Close(t *testing.T) {
	before := runtime.NumGoroutine()
	cache := NewMemoryCache()

	err := cache.Close()
	require.NoError(t, err)

	deadline := time.Now().Add(1 * time.Second)
	stopped := false
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before {
			stopped = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, stopped, "cleanup goroutine did not stop within 1s")
}

func TestMemoryCache_CloseTwice(t *testing.T) {
	cache := NewMemoryCache()
	err := cache.Close()
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		err = cache.Close()
		assert.NoError(t, err)
	})
}
