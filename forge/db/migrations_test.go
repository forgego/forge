package db

import (
	"context"
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMigrationRunner_NilDB(t *testing.T) {
	runner, err := NewMigrationRunner(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil database, got nil")
	}
	if runner != nil {
		t.Fatalf("expected nil runner for nil database, got %#v", runner)
	}
	if err.Error() != "database connection is nil" {
		t.Fatalf("expected nil-database error, got %q", err.Error())
	}
}

func TestMigrationRunnerForceRejectsVersionAboveMaxInt(t *testing.T) {
	runner := &MigrationRunner{}

	err := runner.Force(context.Background(), uint(math.MaxInt)+1)

	require.EqualError(t, err, "migration version exceeds maximum supported value")
}

func TestFallbackDetailedMigrationStatus_Dirty(t *testing.T) {
	result := fallbackDetailedMigrationStatus(&MigrationStatus{
		Version: 7,
		Dirty:   true,
	})

	if result.Current != "7" {
		t.Fatalf("expected current version 7, got %q", result.Current)
	}
	if result.Status != "DIRTY" {
		t.Fatalf("expected status DIRTY, got %q", result.Status)
	}
	if !result.Dirty {
		t.Fatal("expected dirty=true")
	}
}

func TestFallbackDetailedMigrationStatus_Clean(t *testing.T) {
	result := fallbackDetailedMigrationStatus(&MigrationStatus{
		Version: 12,
		Dirty:   false,
	})

	if result.Current != "12" {
		t.Fatalf("expected current version 12, got %q", result.Current)
	}
	if result.Status != "OK" {
		t.Fatalf("expected status OK, got %q", result.Status)
	}
	if result.Dirty {
		t.Fatal("expected dirty=false")
	}
}

func TestNewMigrationRunner_UsesDatabaseDriver(t *testing.T) {
	t.Setenv("FORGE_DATABASE_DRIVER", "postgres")

	dbPath := filepath.Join(t.TempDir(), "m.db")
	database, err := NewDBWithDriver("sqlite3", dbPath)
	require.NoError(t, err)
	defer database.Close()

	migrationsDir := t.TempDir()
	err = os.WriteFile(filepath.Join(migrationsDir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INTEGER);"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(migrationsDir, "000001_init.down.sql"), []byte("DROP TABLE t;"), 0644)
	require.NoError(t, err)

	runner, err := NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	require.NotNil(t, runner)

	err = runner.Up(context.Background())
	require.NoError(t, err)

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM t").Scan(&count)
	require.NoError(t, err)
}

func TestNewMigrationRunner_EmptyDriver(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	database := &DB{DB: sqlDB, Driver: ""}
	migrationsDir := t.TempDir()
	err = os.WriteFile(filepath.Join(migrationsDir, "000001_init.up.sql"), []byte("CREATE TABLE t(id INTEGER);"), 0644)
	require.NoError(t, err)

	runner, err := NewMigrationRunner(database, migrationsDir)
	require.Error(t, err)
	require.Nil(t, runner)
	require.Contains(t, err.Error(), "database driver is unknown")
}
