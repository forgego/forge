package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
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

	// Open connection to 'postgres' db to check/create 'forge_test'
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)
	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err == nil {
		defer defaultDB.Close()
		var exists bool
		err = defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
		if err == nil && !exists {
			_, err = defaultDB.Exec("CREATE DATABASE " + dbname)
			if err != nil {
				t.Logf("Failed to create database %s: %v", dbname, err)
			}
		}
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open connection
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v. Make sure PostgreSQL is running and database '%s' exists.", err, dbname)
	}

	// Clean up database
	CleanupDB(t, sqlDB)

	return sqlDB
}

// CleanupDB truncates all tables in the database
func CleanupDB(t *testing.T, database *sql.DB) {
	rows, err := database.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'schema_migrations'
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
		_, err := database.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}
}
