package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
)

// Option configures NewServer.
type Option func(*serverOptions)

type serverOptions struct {
	database *db.DB
}

// WithDatabase gives NewServer the application database. With
// server.stores set to database, sessions and API throttling counters are
// kept in it, so they survive a restart and are shared by every instance.
// With server.stores set to memory it is not used.
func WithDatabase(database *db.DB) Option {
	return func(o *serverOptions) { o.database = database }
}

// sharedStores returns the database stores selected by server.stores, or
// nil for memory. It fails when database is selected without a database or
// before the framework store tables are migrated.
func sharedStores(settings *config.Settings, opts serverOptions) (*stores.Stores, error) {
	kind, err := stores.ParseKind(settings.Server.Stores)
	if err != nil {
		return nil, err
	}
	if kind != stores.KindDatabase {
		return nil, nil
	}
	if opts.database == nil {
		return nil, errors.New("server.stores is database but NewServer has no database; pass server.WithDatabase(database)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stores.CheckMigrated(ctx, opts.database); err != nil {
		return nil, fmt.Errorf("server.stores is database: %w", err)
	}
	return stores.New(opts.database)
}

// useSharedThrottling makes throttles without a store of their own count in
// the database.
func useSharedThrottling(shared *stores.Stores) {
	throttling.SetDefaultStoreFactory(func(name string, limit int, window time.Duration) throttling.Store {
		return shared.RateLimiter("api:"+name, limit, window)
	})
}
