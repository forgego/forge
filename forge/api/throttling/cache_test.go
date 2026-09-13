package throttling

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
					_, _ = cache.GetInt(key)
				case 2:
					_ = cache.GetTTL(key)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryCache_Expiry(t *testing.T) {
	cache := NewMemoryCache()
	err := cache.Set("expiry-key", 123, 10*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond)

	val, err := cache.GetInt("expiry-key")
	require.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, time.Duration(0), cache.GetTTL("expiry-key"))
}
