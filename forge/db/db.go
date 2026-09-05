// Package db provides database functionality
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db/dialect"
)

// DB wraps database/sql.DB with additional functionality
type DB struct {
	*sql.DB
	Driver string
	dialect dialect.Dialect
}

// Option is a functional option for DB configuration.
// Use options to customize database connection behavior.
type Option func(*DB)

// WithPoolConfig sets the connection pool configuration.
// This allows customizing MaxOpenConns, MaxIdleConns, ConnMaxLifetime, and ConnMaxIdleTime.
//
// Example:
//
//	cfg := db.PoolConfig{
//	    MaxOpenConns:    50,
//	    MaxIdleConns:    25,
//	    ConnMaxLifetime: 10 * time.Minute,
//	    ConnMaxIdleTime: 5 * time.Minute,
//	}
//	db, err := db.NewDB(dsn, db.WithPoolConfig(cfg))
func WithPoolConfig(config PoolConfig) Option {
	return func(db *DB) {
		config.Apply(db)
	}
}

// WithMaxOpenConns sets the maximum number of open connections.
// This is a convenience option for setting just MaxOpenConns.
func WithMaxOpenConns(n int) Option {
	return func(db *DB) {
		db.DB.SetMaxOpenConns(n)
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
// This is a convenience option for setting just MaxIdleConns.
func WithMaxIdleConns(n int) Option {
	return func(db *DB) {
		db.DB.SetMaxIdleConns(n)
	}
}

// WithConnMaxLifetime sets the maximum lifetime of connections.
// This is a convenience option for setting just ConnMaxLifetime.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *DB) {
		db.DB.SetConnMaxLifetime(d)
	}
}

// WithConnMaxIdleTime sets the maximum idle time of connections.
// This is a convenience option for setting just ConnMaxIdleTime.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *DB) {
		db.DB.SetConnMaxIdleTime(d)
	}
}

// Dialect returns the SQL dialect for this database connection.
// Use this to generate database-agnostic SQL queries.
func (db *DB) Dialect() dialect.Dialect {
	return db.dialect
}

// SetDialect sets the SQL dialect for this database connection.
// This is useful for tests or when creating DB instances directly.
func (db *DB) SetDialect(d dialect.Dialect) {
	db.dialect = d
}

// PoolStats returns current connection pool statistics.
// Use this for monitoring database connection health and performance.
// The returned sql.DBStats contains information about open connections,
// in-use connections, idle connections, wait statistics, and more.
func (db *DB) PoolStats() *sql.DBStats {
	stats := db.DB.Stats()
	return &stats
}

// NewDBFromConfig creates a new database connection from config.
// It reads pool configuration from the config and applies it to the database connection.
// The following config keys are supported:
//   - database.max_open_conns: maximum open connections (default: 25)
//   - database.max_idle_conns: maximum idle connections (default: 10)
//   - database.conn_max_lifetime: connection max lifetime (default: 5m)
//   - database.conn_max_idle_time: connection max idle time (default: 2m)
func NewDBFromConfig(cfg *config.Config) (*DB, error) {
	if cfg == nil {
		return nil, errors.New("database config is nil")
	}

	driver := cfg.GetDriver()
	host := cfg.GetString("database.host", "localhost")
	port := cfg.GetInt("database.port", 5432)
	user := cfg.GetString("database.user", "postgres")
	password := cfg.GetString("database.password", "")
	name := cfg.GetString("database.name", "forge")
	sslmode := cfg.GetString("database.sslmode", "disable")

	var dsn string
	if driver == "postgres" || driver == "postgresql" {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, name, sslmode)
	} else if driver == "sqlite" || driver == "sqlite3" {
		dsn = name // For SQLite, name is the file path
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Get pool config from config and convert to db.PoolConfig
	cfgPool := cfg.GetPoolConfig()
	poolConfig := PoolConfig{
		MaxOpenConns:    cfgPool.MaxOpenConns,
		MaxIdleConns:    cfgPool.MaxIdleConns,
		ConnMaxLifetime: cfgPool.ConnMaxLifetime,
		ConnMaxIdleTime: cfgPool.ConnMaxIdleTime,
	}

	return NewDB(dsn, WithPoolConfig(poolConfig))
}

// NewDB creates a new database connection with optional configuration.
// For PostgreSQL, dsn should be a connection string.
// For SQLite, dsn should be a file path.
//
// By default, NewDB applies sensible production pool settings:
//   - MaxOpenConns: 25
//   - MaxIdleConns: 10
//   - ConnMaxLifetime: 5 minutes
//   - ConnMaxIdleTime: 2 minutes
//
// Use functional options to customize these settings:
//
//	db, err := db.NewDB(dsn,
//	    db.WithMaxOpenConns(50),
//	    db.WithMaxIdleConns(25),
//	)
//
// Or use WithPoolConfig for complete control:
//
//	cfg := db.PoolConfig{
//	    MaxOpenConns:    50,
//	    MaxIdleConns:    25,
//	    ConnMaxLifetime: 10 * time.Minute,
//	    ConnMaxIdleTime: 5 * time.Minute,
//	}
//	db, err := db.NewDB(dsn, db.WithPoolConfig(cfg))
func NewDB(dsn string, opts ...Option) (*DB, error) {
	// Try PostgreSQL first
	sqlDB, err := sql.Open("postgres", dsn)
	if err == nil {
		if err := sqlDB.Ping(); err == nil {
			db := &DB{
				DB:      sqlDB,
				Driver:  "postgres",
				dialect: dialect.NewPostgreSQLDialect(),
			}
			// Apply default pool configuration
			defaultConfig := DefaultPoolConfig()
			defaultConfig.Apply(db)
			// Apply custom options (can override defaults)
			for _, opt := range opts {
				opt(db)
			}
			return db, nil
		}
		// Functionally close if ping failed before trying next
		sqlDB.Close()
	}

	// Try SQLite as fallback
	sqlDB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:      sqlDB,
		Driver:  "sqlite3",
		dialect: dialect.NewSQLiteDialect(),
	}
	// Apply default pool configuration
	defaultConfig := DefaultPoolConfig()
	defaultConfig.Apply(db)
	// Apply custom options (can override defaults)
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}

// RebindPlaceholders converts PostgreSQL $N placeholders to SQLite ?N format.
// This is useful when writing database-agnostic code that generates SQL with
// PostgreSQL-style placeholders but needs to run on SQLite.
func (db *DB) RebindPlaceholders(query string) string {
	if db.Driver == "sqlite3" || db.Driver == "sqlite" {
		return rebindPostgresToSQLite(query)
	}
	return query
}

// Rebind modifies the query based on the driver
// specifically mapping Postgres $N placeholders to SQLite ?N
//
// Deprecated: Use RebindPlaceholders() for clarity. Rebind will be removed in v3.0.
// Migration:
//
//	// Old
//	sql := db.Rebind(query)
//	// New
//	sql := db.RebindPlaceholders(query)
func (db *DB) Rebind(query string) string {
	return db.RebindPlaceholders(query)
}

var paramRegex = regexp.MustCompile(`\$([0-9]+)`)

func rebindPostgresToSQLite(query string) string {
	return paramRegex.ReplaceAllString(query, "?$1")
}

// Ping verifies the database connection is still alive.
// It returns an error if the connection is not initialized or cannot be reached.
func (db *DB) Ping(ctx context.Context) error {
	if db.DB == nil {
		return errors.New("database connection not initialized")
	}
	return db.DB.PingContext(ctx)
}

// IsConnected checks if the database connection is valid.
// It has a 5-second timeout for the ping operation.
func (db *DB) IsConnected() bool {
	if db.DB == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Ping(ctx) == nil
}

// NewDBWithValidation creates a new database connection and validates it immediately.
// This is the recommended way to create a DB instance for production use, as it
// ensures the database is reachable before returning.
func NewDBWithValidation(dsn string) (*DB, error) {
	db, err := NewDB(dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
