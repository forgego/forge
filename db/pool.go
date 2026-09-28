package db

import "time"

// PoolConfig holds database connection pool configuration.
// These settings control how database connections are managed, which is essential
// for production deployments to prevent connection exhaustion and ensure optimal
// performance.
type PoolConfig struct {
	// MaxOpenConns is the maximum number of open connections to the database.
	// A value of 0 means unlimited connections. Default: 25
	// Set this based on your database server's max_connections setting and
	// the number of application instances.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of connections in the idle connection pool.
	// A value of 0 means no idle connections are retained. Default: 10
	// Setting this too low can cause connection churn; too high wastes resources.
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	// A value of 0 means connections are never closed due to age. Default: 5 minutes
	// Set this lower than your database server's wait_timeout to prevent
	// "connection closed" errors.
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// A value of 0 means connections are never closed due to idle time. Default: 2 minutes
	// This helps release unused connections back to the database server.
	ConnMaxIdleTime time.Duration
}

// DefaultPoolConfig returns a PoolConfig with sensible production defaults.
// These defaults are suitable for most web applications:
//   - MaxOpenConns: 25 (prevents connection exhaustion)
//   - MaxIdleConns: 10 (reduces connection churn)
//   - ConnMaxLifetime: 5 minutes (prevents stale connections)
//   - ConnMaxIdleTime: 2 minutes (releases unused connections)
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}
}

// Apply applies the pool configuration to the given sql.DB.
// This is an internal method used by NewDB and WithPoolConfig.
func (c PoolConfig) Apply(db *DB) {
	if c.MaxOpenConns > 0 {
		db.DB.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.MaxIdleConns > 0 {
		db.DB.SetMaxIdleConns(c.MaxIdleConns)
	}
	if c.ConnMaxLifetime > 0 {
		db.DB.SetConnMaxLifetime(c.ConnMaxLifetime)
	}
	if c.ConnMaxIdleTime > 0 {
		db.DB.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	}
}
