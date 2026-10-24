package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
)

// Option configures NewServer.
type Option func(*serverOptions)

type serverOptions struct {
	database *db.DB
}

// WithDatabase gives NewServer the application database. Sessions, API
// throttling counters, and other framework stores are kept in it, so they
// survive a restart and are shared by every instance.
func WithDatabase(database *db.DB) Option {
	return func(o *serverOptions) { o.database = database }
}

// sharedStores returns the framework database stores. It fails when NewServer has
// no database configured or before the framework store tables are migrated.
func sharedStores(opts serverOptions) (*stores.Stores, error) {
	if opts.database == nil || opts.database.DB == nil {
		return nil, errors.New("no database configured: pass server.WithDatabase(database) and run `forge migrate up`")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stores.CheckMigrated(ctx, opts.database); err != nil {
		return nil, fmt.Errorf("framework store tables not migrated (run `forge migrate up`): %w", err)
	}
	return stores.New(opts.database)
}

// useSharedThrottling sets the process-wide default store factory for
// throttles without a store of their own to use the database stores.
func useSharedThrottling(shared *stores.Stores) {
	if shared == nil {
		throttling.SetDefaultStoreFactory(nil)
		return
	}
	throttling.SetDefaultStoreFactory(func(name string, limit int, window time.Duration) throttling.Store {
		return shared.RateLimiter("api:"+name, limit, window)
	})
}
