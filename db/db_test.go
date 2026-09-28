package db

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewDBFromConfig_NilConfig(t *testing.T) {
	database, err := NewDBFromConfig(nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
	if database != nil {
		t.Fatalf("expected nil database for nil config, got %#v", database)
	}
	if !strings.Contains(err.Error(), "config is nil") {
		t.Fatalf("expected nil-config error message, got %q", err.Error())
	}
}

func TestPing_NilDB(t *testing.T) {
	var database DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := database.Ping(ctx)
	if err == nil {
		t.Fatal("expected error for nil database, got nil")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("expected not initialized error message, got %q", err.Error())
	}
}

func TestIsConnected_NilDB(t *testing.T) {
	var database DB
	if database.IsConnected() {
		t.Fatal("expected IsConnected to return false for nil database")
	}
}

func TestNewDBWithValidation_InvalidDSN(t *testing.T) {
	// Test with an invalid DSN that will fail to connect
	database, err := NewDBWithValidation("invalid:dsn@tcp(localhost:3306)/nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
	if database != nil {
		t.Fatalf("expected nil database for invalid DSN, got %#v", database)
	}
	// The error could be either "failed to connect" or "failed to ping database"
	// depending on which driver is attempted first
	errStr := err.Error()
	if !strings.Contains(errStr, "failed to connect") && !strings.Contains(errStr, "failed to ping") {
		t.Fatalf("expected connection-related error message, got %q", errStr)
	}
}

func TestNewDBWithDriver_SQLite(t *testing.T) {
	database, err := NewDBWithDriver("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("expected NewDBWithDriver to succeed, got error: %v", err)
	}
	defer database.Close()

	if database.Driver != "sqlite3" {
		t.Fatalf("expected Driver to be 'sqlite3', got %q", database.Driver)
	}
}

func TestNewDBWithDriver_UnsupportedDriver(t *testing.T) {
	database, err := NewDBWithDriver("mysql", "x")
	if err == nil {
		database.Close()
		t.Fatal("expected error for unsupported driver 'mysql', got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported driver error, got: %v", err)
	}
}

func TestNewDBWithDriver_PostgresFailureNoFallback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	dsn := "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1"
	database, err := NewDBWithDriver("postgres", dsn)
	if err == nil {
		database.Close()
		t.Fatal("expected error for unreachable postgres connection, got nil")
	}

	// Verify no SQLite file named like the DSN was created in the current dir
	if _, statErr := os.Stat(dsn); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file named %q to exist, but statErr was %v", dsn, statErr)
	}
}

func TestNewDB_PostgresFailureNoFallback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	dsn := "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable connect_timeout=1"
	database, err := NewDB(dsn)
	if err == nil {
		database.Close()
		t.Fatal("expected error for unreachable postgres DSN, got nil")
	}

	// Verify no SQLite file named like the DSN was created in the current dir
	if _, statErr := os.Stat(dsn); !os.IsNotExist(statErr) {
		t.Fatalf("expected no file named %q to exist, but statErr was %v", dsn, statErr)
	}
}

func TestDB_RebindPlaceholders_SkipsLiterals(t *testing.T) {
	d := &DB{Driver: "sqlite"}

	input := "SELECT * FROM notes WHERE title ILIKE $1 AND note = 'costs $1 ILIKE' AND author = 'O''Reilly $2'"
	expected := "SELECT * FROM notes WHERE title LIKE ?1 AND note = 'costs $1 ILIKE' AND author = 'O''Reilly $2'"

	got := d.RebindPlaceholders(input)
	if got != expected {
		t.Fatalf("got %q, want %q", got, expected)
	}
}

func TestRebindPlaceholders_SQLiteBackslashBeforeQuote(t *testing.T) {
	database := &DB{Driver: "sqlite3"}
	query := `SELECT * FROM t WHERE note = 'C:\' AND title ILIKE $1`
	want := `SELECT * FROM t WHERE note = 'C:\' AND title LIKE ?1`
	if got := database.RebindPlaceholders(query); got != want {
		t.Fatalf("RebindPlaceholders() = %q, want %q", got, want)
	}
}
