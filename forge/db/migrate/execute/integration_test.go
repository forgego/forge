package execute

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/internal/testutils"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Integration_Workflow(t *testing.T) {
	db := testutils.SetupTestDB(t)

    // Explicit cleanup
    _, _ = db.Exec("TRUNCATE TABLE schema_migrations")
    _, _ = db.Exec("DROP TABLE IF EXISTS users_test")
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	// Create temp migrations dir
	tmpDir := t.TempDir()
	
	// Migration 1: Create Users
	m1Up := `CREATE TABLE users_test (id SERIAL PRIMARY KEY, name TEXT);`
	m1Down := `DROP TABLE users_test;`
	createMigration(t, tmpDir, "000001_create_users", m1Up, m1Down)

	// Migration 2: Add Email
	m2Up := `ALTER TABLE users_test ADD COLUMN email TEXT;`
	m2Down := `ALTER TABLE users_test DROP COLUMN email;`
	createMigration(t, tmpDir, "000002_add_email", m2Up, m2Down)

	// Initialize Executor
	executor, err := NewExecutor(driver, tmpDir)
	require.NoError(t, err)
	defer executor.Close()

	ctx := context.Background()

	// 1. Migrate Up (All)
	err = executor.Migrate(ctx)
	require.NoError(t, err)

	// Verify schema
	assertTableExists(t, db, "users_test")
	assertColumnExists(t, db, "users_test", "email")

	// 2. Rollback (One step)
	err = executor.Rollback(ctx)
	require.NoError(t, err)

	// Verify schema (email should be gone)
	assertTableExists(t, db, "users_test")
	assertColumnNotExists(t, db, "users_test", "email")

	// 3. Rollback (Last step)
	err = executor.Rollback(ctx)
	require.NoError(t, err)

	// Verify schema (users should be gone)
	assertTableNotExists(t, db, "users_test")

	// 4. Migrate Up again
	err = executor.Migrate(ctx)
	require.NoError(t, err)
	assertTableExists(t, db, "users_test")
	assertColumnExists(t, db, "users_test", "email")
}

func createMigration(t *testing.T, dir, name, up, down string) {
	err := os.WriteFile(filepath.Join(dir, name+".up.sql"), []byte(up), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, name+".down.sql"), []byte(down), 0644)
	require.NoError(t, err)
}

func assertTableExists(t *testing.T, db *sql.DB, table string) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = $1)", table).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "Table %s should exist", table)
}

func assertTableNotExists(t *testing.T, db *sql.DB, table string) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = $1)", table).Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "Table %s should not exist", table)
}

func assertColumnExists(t *testing.T, db *sql.DB, table, column string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			AND column_name = $2
		)`, table, column).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "Column %s.%s should exist", table, column)
}

func assertColumnNotExists(t *testing.T, db *sql.DB, table, column string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.columns 
			WHERE table_schema = 'public' 
			AND table_name = $1 
			AND column_name = $2
		)`, table, column).Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "Column %s.%s should not exist", table, column)
}
