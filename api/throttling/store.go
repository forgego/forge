package throttling

import (
	"context"
	stdlog "log"
	"net/http"
	"sync/atomic"
	"time"

	internalratelimit "github.com/forgego/forge/internal/ratelimit"
)

// Store represents a backend store for rate limiting.
// Implementations must be safe for concurrent use by multiple goroutines.
type Store interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// ContextStore is a Store whose check does I/O, such as a database-backed
// counter. Throttles call AllowContext with the request's context, so a slow
// backend stops when the client goes away. An error allows the request and
// is logged: an unreachable store must not refuse traffic.
type ContextStore interface {
	Store
	AllowContext(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error)
}

// allow checks key against store, through AllowContext with r's context when
// store supports it.
func allow(store Store, r *http.Request, key string) (bool, time.Duration) {
	cs, ok := store.(ContextStore)
	if !ok {
		return store.Allow(key)
	}
	allowed, retryAfter, err := cs.AllowContext(r.Context(), key)
	if err != nil {
		stdlog.Printf("forge/throttling: store error, allowing request: %v", err)
		return true, 0
	}
	return allowed, retryAfter
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
