package hooks

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestSamplingHook_ConcurrentProcess(t *testing.T) {
	const (
		initial    = 50
		thereafter = 10
		workers    = 100
	)
	hook := NewSamplingHook(initial, thereafter)

	var kept atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, _, ok := hook.Process(zapcore.Entry{Level: zapcore.InfoLevel}, nil)
			if ok {
				kept.Add(1)
			}
		}()
	}

	wg.Wait()

	// With 100 entries: first 50 kept, plus (100-50)/10 = 5 kept => 55 total.
	assert.Equal(t, int64(55), kept.Load())
}

func TestSamplingHook_FirstInitialEntriesKept(t *testing.T) {
	hook := NewSamplingHook(5, 3)

	// First 5 entries must all be kept
	for i := 1; i <= 5; i++ {
		_, _, ok := hook.Process(zapcore.Entry{Level: zapcore.InfoLevel}, nil)
		assert.True(t, ok, "entry %d must be kept", i)
	}

	// 6th entry: (6-5)%3 == 1 != 0 => dropped
	_, _, ok6 := hook.Process(zapcore.Entry{Level: zapcore.InfoLevel}, nil)
	assert.False(t, ok6, "entry 6 must be dropped")

	// 7th entry: (7-5)%3 == 2 != 0 => dropped
	_, _, ok7 := hook.Process(zapcore.Entry{Level: zapcore.InfoLevel}, nil)
	assert.False(t, ok7, "entry 7 must be dropped")

	// 8th entry: (8-5)%3 == 0 => kept
	_, _, ok8 := hook.Process(zapcore.Entry{Level: zapcore.InfoLevel}, nil)
	assert.True(t, ok8, "entry 8 must be kept")
}
