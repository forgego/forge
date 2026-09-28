package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/lib/pq"
)

// SetupTestDB creates a connection to the test PostgreSQL database
func SetupTestDB(t *testing.T) *sql.DB {
	// Connection details
	user := "postgres"
	password := "123"
	host := "localhost"
	port := 5432
	dbname := "forge_test"

	// Allow overriding via env vars
	if u := os.Getenv("DB_USER"); u != "" {
		user = u
	}
	if p := os.Getenv("DB_PASSWORD"); p != "" {
		password = p
	}
	if h := os.Getenv("DB_HOST"); h != "" {
		host = h
	}
	if n := os.Getenv("DB_NAME"); n != "" {
		dbname = n
	}

	dsn := testDatabaseURL(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname))

	if os.Getenv("FORGE_TEST_DATABASE_URL") == "" {
		// Open connection to 'postgres' db to check/create 'forge_test'
		defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
			host, port, user, password)
		defaultDB, err := sql.Open("postgres", defaultDSN)
		if err == nil {
			defer defaultDB.Close()
			var exists bool
			err = defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
			if err == nil && !exists {
				// Identifier is quoted: DDL names cannot be query parameters.
				// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
				_, err = defaultDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbname))
				if err != nil {
					t.Logf("Failed to create database %s: %v", dbname, err)
				}
			}
		}
	}

	// Open connection
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		skipOrFailNoDB(t, "PostgreSQL not available: %v. Skipping integration DB tests.", err)
	}

	// Clean up database
	CleanupDB(t, sqlDB)

	return sqlDB
}

func testDatabaseURL(defaultURL string) string {
	if u := os.Getenv("FORGE_TEST_DATABASE_URL"); u != "" {
		return u
	}
	return defaultURL
}

func requireDB() bool {
	return os.Getenv("FORGE_REQUIRE_DB") == "1"
}

func dbUnavailableAction() string {
	if requireDB() {
		return "fatal"
	}
	return "skip"
}

func skipOrFailNoDB(t testing.TB, format string, args ...any) {
	t.Helper()
	msg := fmt.Sprintf(format, args...)
	if requireDB() {
		t.Fatalf("%s (FORGE_REQUIRE_DB=1 is set)", msg)
	}
	t.Skipf("%s (set FORGE_REQUIRE_DB=1 to turn into failure)", msg)
}

// CleanupDB truncates all tables in the database
func CleanupDB(t *testing.T, database *sql.DB) {
	rows, err := database.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'schema_migrations'
		AND tablename != 'forge_migration_checksums'
	`)
	if err != nil {
		t.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, table)
	}

	for _, table := range tables {
		// Table names come from information_schema (not user input) and are
		// quoted; TRUNCATE takes no row parameters.
		// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
		_, err := database.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", pq.QuoteIdentifier(table)))
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}
