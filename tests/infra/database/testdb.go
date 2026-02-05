package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"           // PostgreSQL
	_ "github.com/mattn/go-sqlite3" // SQLite
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Driver  string
	DSN     string
	Cleanup bool
}

// setupTestDB creates a test database
func setupTestDB(t *testing.T) *sql.DB {
	// Use environment variable or default to SQLite
	driver := os.Getenv("TEST_DB_DRIVER")
	if driver == "" {
		driver = "sqlite3"
	}

	var db *sql.DB
	var err error

	switch driver {
	case "postgres":
		db, err = setupPostgresDB(t)
	case "sqlite3":
		db, err = setupSQLiteDB(t)
	default:
		t.Fatalf("unsupported driver: %s", driver)
	}

	if err != nil {
		t.Fatalf("failed to setup test DB: %v", err)
	}

	return db
}

// setupPostgresDB creates a PostgreSQL test database
func setupPostgresDB(t *testing.T) (*sql.DB, error) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:5432/test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return db, nil
}

// setupSQLiteDB creates an in-memory SQLite test database
func setupSQLiteDB(t *testing.T) (*sql.DB, error) {
	// Use in-memory database for fast tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}

	// For SQLite in-memory, closing is enough
	// For PostgreSQL, we might want to truncate tables
	driver := os.Getenv("TEST_DB_DRIVER")
	if driver == "" || driver == "sqlite3" {
		db.Close()
		return
	}

	// For PostgreSQL, drop all tables in test schema
	ctx := context.Background()
	tables := []string{"books", "authors", "categories"}
	for _, table := range tables {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
	}

	db.Close()
}

// createTestSchema creates test tables
func createTestSchema(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Create authors table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS authors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			bio TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create authors table: %v", err)
	}

	// Create books table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			isbn TEXT UNIQUE,
			author_id INTEGER NOT NULL,
			category_id INTEGER,
			description TEXT,
			pages INTEGER DEFAULT 0,
			available BOOLEAN DEFAULT TRUE,
			price REAL DEFAULT 0.0,
			published_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (author_id) REFERENCES authors(id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create books table: %v", err)
	}

	// Create categories table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create categories table: %v", err)
	}
}

// withTestDB runs a test with a test database
func withTestDB(t *testing.T, fn func(*sql.DB)) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	createTestSchema(t, db)
	fn(db)
}

// withTestDBTx runs a test with a test database transaction (auto-rollback)
func withTestDBTx(t *testing.T, fn func(*sql.DB, *sql.Tx)) {
	withTestDB(t, func(db *sql.DB) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		fn(db, tx)
	})
}
