package throttling

import (
	"sync/atomic"
	"time"

	internalratelimit "github.com/forgego/forge/internal/ratelimit"
)

// Store represents a backend store for rate limiting.
// Implementations must be safe for concurrent use by multiple goroutines.
type Store interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// StoreFactory builds the store of a throttle that was not given one with
// WithStore or a ...WithStore constructor. name identifies the throttle's
// counters, for example "anon/100/hour"; throttles with the same name share
// counts in a shared store. limit calls are allowed per key per window.
type StoreFactory func(name string, limit int, window time.Duration) Store

var defaultStoreFactory atomic.Pointer[StoreFactory]

// SetDefaultStoreFactory sets the factory for the stores of throttles that
// have no store of their own, including throttles created before the call:
// each such throttle switches to a store from the new factory on its next
// request. nil restores the default, a fixed-window counter in process
// memory. server.NewServer sets a database-backed factory when
// server.stores is database.
func SetDefaultStoreFactory(factory StoreFactory) {
	if factory == nil {
		defaultStoreFactory.Store(nil)
		return
	}
	defaultStoreFactory.Store(&factory)
}

// defaultStore resolves the store of a throttle without an explicit one,
// from the current default factory, and keeps it until the factory changes.
type defaultStore struct {
	name     string
	limit    int
	window   time.Duration
	resolved atomic.Pointer[resolvedStore]
}

type resolvedStore struct {
	factory *StoreFactory
	store   Store
}

func (d *defaultStore) get() Store {
	factory := defaultStoreFactory.Load()
	current := d.resolved.Load()
	if current != nil && current.factory == factory {
		return current.store
	}
	next := &resolvedStore{factory: factory}
	if factory == nil {
		next.store = internalratelimit.NewFixedWindowCounter(d.limit, d.window)
	} else {
		next.store = (*factory)(d.name, d.limit, d.window)
	}
	if d.resolved.CompareAndSwap(current, next) {
		return next.store
	}
	return d.resolved.Load().store
}
