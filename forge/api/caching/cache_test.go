package caching

import (
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
