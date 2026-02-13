// Package db provides database functionality
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/forgego/forge/config"
)

// DB wraps database/sql.DB with additional functionality
type DB struct {
	*sql.DB
	Driver string
}

// NewDBFromConfig creates a new database connection from config
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

	return NewDB(dsn)
}

// NewDB creates a new database connection
// For PostgreSQL, dsn should be a connection string
// For SQLite, dsn should be a file path
func NewDB(dsn string) (*DB, error) {
	// Try PostgreSQL first
	sqlDB, err := sql.Open("postgres", dsn)
	if err == nil {
		if err := sqlDB.Ping(); err == nil {
			return &DB{DB: sqlDB, Driver: "postgres"}, nil
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

	return &DB{DB: sqlDB, Driver: "sqlite3"}, nil
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
