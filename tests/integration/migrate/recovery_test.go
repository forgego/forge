package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/tests/testhelpers"
)

// TestForceRecovery tests recovering from a dirty migration state using Force
func TestForceRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_force_recovery_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "force_recovery_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create two migrations, one good and one that will fail
	upSQL1 := `CREATE TABLE test_table (id BIGSERIAL PRIMARY KEY, name VARCHAR(100));`
	downSQL1 := `DROP TABLE IF EXISTS test_table;`
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_create_test.up.sql"), []byte(upSQL1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_create_test.down.sql"), []byte(downSQL1), 0644))

	// Create a migration that will fail (invalid SQL)
	upSQL2 := `CREATE TABLE test_table2 (id BIGSERIAL PRIMARY KEY); INVALID SQL HERE;`
	downSQL2 := `DROP TABLE IF EXISTS test_table2;`
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_broken_migration.up.sql"), []byte(upSQL2), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_broken_migration.down.sql"), []byte(downSQL2), 0644))

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	// Apply first migration successfully
	err = runner.MigrateTo(ctx, 1)
	require.NoError(t, err)

	version, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(1), version)
	assert.False(t, dirty)

	// Try to apply second migration - it should fail and mark database as dirty
	err = runner.Migrate(ctx)
	require.Error(t, err, "broken migration should fail")

	// Check that database is now dirty
	version, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	assert.True(t, dirty, "database should be in dirty state after failed migration")
	assert.Equal(t, uint(2), version, "version should be at failed migration")

	// Attempting another migration should fail due to dirty state
	err = runner.Migrate(ctx)
	require.Error(t, err, "should not allow migration while dirty")
	require.Contains(t, err.Error(), "dirty", "error should mention dirty state")

	// Use Force to mark version as clean and roll back to version 1
	err = runner.Force(ctx, 1)
	require.NoError(t, err, "Force should recover from dirty state")

	// Verify database is no longer dirty
	version, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	assert.False(t, dirty, "database should no longer be dirty")
	assert.Equal(t, uint(1), version, "version should be forced to 1")

	// Verify we can now apply migrations again (after fixing the broken migration)
	// Replace the broken migration with a working one
	upSQL2Fixed := `CREATE TABLE test_table2 (id BIGSERIAL PRIMARY KEY, data TEXT);`
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_broken_migration.up.sql"), []byte(upSQL2Fixed), 0644))

	// Should now be able to migrate
	err = runner.Migrate(ctx)
	require.NoError(t, err, "should be able to migrate after force recovery")

	version, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(2), version)
}

// TestStatus tests migration status reporting
func TestMigrationStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_status_%d", time.Now().UnixNano()),
	}
	postgresDB, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()
	defer postgresDB.Close()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "status_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create three migrations
	migrations := []struct {
		version string
		up      string
		down    string
	}{
		{
			version: "000001",
			up:      `CREATE TABLE table1 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table1;`,
		},
		{
			version: "000002",
			up:      `CREATE TABLE table2 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table2;`,
		},
		{
			version: "000003",
			up:      `CREATE TABLE table3 (id BIGSERIAL PRIMARY KEY);`,
			down:    `DROP TABLE IF EXISTS table3;`,
		},
	}

	for _, mig := range migrations {
		upPath := filepath.Join(migrationsDir, mig.version+"_migration.up.sql")
		downPath := filepath.Join(migrationsDir, mig.version+"_migration.down.sql")
		require.NoError(t, os.WriteFile(upPath, []byte(mig.up), 0644))
		require.NoError(t, os.WriteFile(downPath, []byte(mig.down), 0644))
	}

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	// Initially, no migrations applied
	status, err := runner.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(0), status.Version, "no migrations applied yet")
	assert.False(t, status.Dirty)

	// Apply first two migrations
	err = runner.MigrateTo(ctx, 2)
	require.NoError(t, err)

	status, err = runner.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(2), status.Version, "should be at version 2")
	assert.False(t, status.Dirty)

	// Get detailed status
	detailedStatus, err := runner.GetDetailedStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(2), detailedStatus.Current)
	assert.False(t, detailedStatus.Dirty)
	assert.Len(t, detailedStatus.Applied, 2, "two migrations should be applied")
	assert.Len(t, detailedStatus.Pending, 1, "one migration should be pending")

	// Apply remaining migration
	err = runner.Migrate(ctx)
	require.NoError(t, err)

	status, err = runner.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(3), status.Version)
	assert.False(t, status.Dirty)

	detailedStatus, err = runner.GetDetailedStatus(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(3), detailedStatus.Current)
	assert.Len(t, detailedStatus.Applied, 3, "all migrations should be applied")
	assert.Len(t, detailedStatus.Pending, 0, "no pending migrations")
}

// TestVersionMethod tests the Version method
func TestVersionMethod(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := testhelpers.PostgresOpts{
		UseDirect: true,
		Host:      "localhost",
		Port:      "5432",
		User:      "postgres",
		Password:  "123",
		DBName:    fmt.Sprintf("test_version_%d", time.Now().UnixNano()),
	}
	_, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	require.NoError(t, err)
	defer cleanup()

	tempDir, cleanupTemp := testhelpers.TempDirInTests(t, "version_")
	defer cleanupTemp()
	migrationsDir := filepath.Join(tempDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	// Create a simple migration
	upSQL := `CREATE TABLE version_test (id BIGSERIAL PRIMARY KEY);`
	downSQL := `DROP TABLE IF EXISTS version_test;`
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_test.up.sql"), []byte(upSQL), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_test.down.sql"), []byte(downSQL), 0644))

	database, err := db.NewDB(dsn)
	require.NoError(t, err)
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)
	defer runner.Close()

	// Check initial version (should be 0 or no version)
	version, dirty, err := runner.Version(ctx)
	if err != nil {
		// No version is okay initially
		require.Contains(t, err.Error(), "no migration")
	} else {
		assert.Equal(t, uint(0), version)
		assert.False(t, dirty)
	}

	// Apply migration
	err = runner.Migrate(ctx)
	require.NoError(t, err)

	// Check version after migration
	version, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(1), version)
	assert.False(t, dirty)

	// Rollback
	err = runner.Rollback(ctx)
	require.NoError(t, err)

	// Check version after rollback
	version, dirty, err = runner.Version(ctx)
	if err != nil {
		// After rolling back all migrations, we might get an error
		require.Contains(t, err.Error(), "no migration")
	} else {
		assert.Equal(t, uint(0), version)
		assert.False(t, dirty)
	}
}
