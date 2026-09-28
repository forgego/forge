package db

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/forgego/forge/db/migrate/execute"
	"github.com/forgego/forge/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestMigrationFileChecksums(t *testing.T) {
	dir := t.TempDir()
	lf := []byte("SELECT 1;\n")
	crlf := []byte("SELECT 1;\r\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "000007_first.up.sql"), lf, 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "42_second.up.sql"), crlf, 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "42_second.down.sql"), lf, 0600))
	up, down, err := migrationFileChecksums(dir, 7)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(lf)), up)
	require.Nil(t, down)
	other, down, err := migrationFileChecksums(dir, 42)
	require.NoError(t, err)
	require.NotEqual(t, up, other)
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(crlf)), other)
	require.NotNil(t, down)
	require.Equal(t, up, *down)
	_, _, err = migrationFileChecksums(dir, 8)
	require.ErrorContains(t, err, "version 8")
	require.NoError(t, os.Remove(filepath.Join(dir, "000007_first.up.sql")))
	require.NoError(t, os.Symlink(filepath.Join(dir, "absent"), filepath.Join(dir, "000007_first.up.sql")))
	_, _, err = migrationFileChecksums(dir, 7)
	require.ErrorContains(t, err, "version 7")
}

func TestChecksumBaseline_TwoConnections(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Use an isolated schema without altering other tests' migration state.
	schema := fmt.Sprintf("checksum_baseline_%d", time.Now().UnixNano())
	_, err := sqlDB.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = sqlDB.Exec("DROP SCHEMA " + schema + " CASCADE") })
	// SET is session-local, so initialize every connection used by this runner.
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	forConnections := make([]interface{ Close() error }, 0, 2)
	for i := 0; i < 2; i++ {
		conn, err := sqlDB.Conn(context.Background())
		require.NoError(t, err)
		_, err = conn.ExecContext(context.Background(), "SET search_path TO "+schema)
		require.NoError(t, err)
		forConnections = append(forConnections, conn)
	}
	for _, conn := range forConnections {
		require.NoError(t, conn.Close())
	}
	testChecksumBaseline(t, &DB{DB: sqlDB, Driver: "postgres"})
}

func TestChecksumBaseline_SQLiteRecordsEachAppliedVersion(t *testing.T) {
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.DB.Close() })
	testChecksumBaseline(t, database)
}

func testChecksumBaseline(t *testing.T, database *DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dir := t.TempDir()
	expected := map[uint]string{}
	for i := 1; i <= 3; i++ {
		contents := []byte(fmt.Sprintf("CREATE TABLE checksum_test_%d (id INTEGER PRIMARY KEY);\n", i))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%06d_test.up.sql", i)), contents, 0600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%06d_test.down.sql", i)), []byte(fmt.Sprintf("DROP TABLE checksum_test_%d;\n", i)), 0600))
		expected[uint(i)] = fmt.Sprintf("%x", sha256.Sum256(contents))
	}
	runner, err := NewMigrationRunner(database, dir)
	require.NoError(t, err)
	assertRows := func(count int) {
		t.Helper()
		rows, err := database.Query("SELECT version, up_sha256, down_sha256, applied_at FROM forge_migration_checksums ORDER BY version")
		require.NoError(t, err)
		defer rows.Close()
		n := 0
		for rows.Next() {
			var version uint
			var up, down string
			var applied any
			require.NoError(t, rows.Scan(&version, &up, &down, &applied))
			require.Equal(t, expected[version], up)
			require.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("DROP TABLE checksum_test_%d;\n", version)))), down)
			require.NotNil(t, applied)
			n++
		}
		require.NoError(t, rows.Err())
		require.Equal(t, count, n)
	}
	require.NoError(t, runner.Up(ctx))
	assertRows(3)
	// Adoption must share the lock connection even when the pool has only two.
	_, err = database.Exec("DELETE FROM forge_migration_checksums")
	require.NoError(t, err)
	adopted, err := runner.AdoptChecksumBaseline(ctx)
	require.NoError(t, err)
	require.Equal(t, []uint{1, 2, 3}, adopted)
	assertRows(3)

	require.NoError(t, runner.Rollback(ctx))
	assertRows(2)
	require.NoError(t, runner.Up(ctx))
	assertRows(3)
	require.NoError(t, runner.RollbackTo(ctx, 1))
	assertRows(1)
	require.NoError(t, runner.MigrateTo(ctx, 3))
	assertRows(3)
	require.NoError(t, runner.RollbackSteps(ctx, 3))
	assertRows(0)
}

func checksumRegressionRunner(t *testing.T) (*DB, *MigrationRunner) {
	t.Helper()
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.DB.Close() })
	dir := t.TempDir()
	for i := 1; i <= 3; i++ {
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d_create.up.sql", i)), []byte("SELECT 1;"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d_create_rollback.down.sql", i)), []byte("SELECT 2;"), 0600))
	}
	runner, err := NewMigrationRunner(database, dir)
	require.NoError(t, err)
	require.NoError(t, runner.Up(context.Background()))
	return database, runner
}

func TestChecksumBaseline_DifferentDownStem(t *testing.T) {
	database, _ := checksumRegressionRunner(t)
	var down string
	require.NoError(t, database.QueryRow("SELECT down_sha256 FROM forge_migration_checksums WHERE version = 1").Scan(&down))
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("SELECT 2;"))), down)
}

func TestChecksumBaseline_ReconcileLeftover(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		t.Run(fmt.Sprintf("rollback=%t", rollback), func(t *testing.T) {
			database, runner := checksumRegressionRunner(t)
			_, err := database.Exec("INSERT INTO forge_migration_checksums (version, up_sha256) VALUES (4, 'leftover')")
			require.NoError(t, err)
			if rollback {
				require.NoError(t, runner.Rollback(context.Background()))
			} else {
				require.NoError(t, runner.Up(context.Background()))
			}
			version, _, err := runner.Version(context.Background())
			require.NoError(t, err)
			expected := uint(3)
			if rollback {
				expected = 2
			}
			require.Equal(t, expected, version)
			var count int
			require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums WHERE version > $1", expected).Scan(&count))
			require.Zero(t, count)
		})
	}
}

func TestForce_ChecksumCleanup(t *testing.T) {
	database, runner := checksumRegressionRunner(t)
	require.NoError(t, runner.Force(context.Background(), 1))
	var versions []uint
	rows, err := database.Query("SELECT version FROM forge_migration_checksums ORDER BY version")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var version uint
		require.NoError(t, rows.Scan(&version))
		versions = append(versions, version)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []uint{1}, versions)
}

func TestChecksumFileIndex_DifferentStems(t *testing.T) {
	dir := t.TempDir()
	names := []string{"1_create.up.sql", "1_create_rollback.down.sql", "000002_second.up.sql", "2_remove.down.sql", "42_last.up.sql", "ignore.txt"}
	for _, name := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0600))
	}
	index, err := indexMigrationFiles(dir)
	require.NoError(t, err)
	require.Equal(t, migrationFileIndex{
		1:  {up: filepath.Join(dir, names[0]), down: filepath.Join(dir, names[1])},
		2:  {up: filepath.Join(dir, names[2]), down: filepath.Join(dir, names[3])},
		42: {up: filepath.Join(dir, names[4])},
	}, index)
	// The index retains paths without requiring another directory scan.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "3_later.up.sql"), []byte("SELECT 1;"), 0600))
	_, _, err = index.checksums(3)
	require.ErrorContains(t, err, "version 3 not found")
}

func TestChecksumBaseline_OneConnectionRejected(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqlDB.SetMaxOpenConns(1)
	_, err := NewMigrationRunner(&DB{DB: sqlDB, Driver: "postgres"}, t.TempDir())
	require.ErrorContains(t, err, "at least 2 connections")
}

func TestChecksumBaseline_RollbackErrorReconciles(t *testing.T) {
	database, runner := checksumRegressionRunner(t)
	// A dirty version makes Steps fail before executing SQL.
	_, err := database.Exec("UPDATE schema_migrations SET version = 1, dirty = 1")
	require.NoError(t, err)
	require.Error(t, runner.rollbackChecksumStep(context.Background(), database))
	// When dirty, reconcileMigrationChecksums skips deleting rows; all 3 versions should remain.
	var count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums").Scan(&count))
	require.Equal(t, 3, count)
}

func TestChecksumBaseline_CleanRollbackErrorReconciles(t *testing.T) {
	database, runner := checksumRegressionRunner(t)
	// Database is clean at version 1, but checksums have versions 1, 2, 3.
	_, err := database.Exec("UPDATE schema_migrations SET version = 1, dirty = 0")
	require.NoError(t, err)
	// Making version 1's down migration unreadable makes Steps(-1) fail without making state dirty.
	downFile := filepath.Join(runner.migrationsPath, "1_create_rollback.down.sql")
	require.NoError(t, os.Chmod(downFile, 0000))
	t.Cleanup(func() { _ = os.Chmod(downFile, 0600) })
	err = runner.rollbackChecksumStep(context.Background(), database)
	require.Error(t, err)
	// Reconcile must still run and remove rows above the clean version (1).
	var count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums WHERE version > 1").Scan(&count))
	require.Zero(t, count)
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums").Scan(&count))
	require.Equal(t, 1, count)
}

func TestChecksumBaseline_DirtyRollbackPreservesChecksums(t *testing.T) {
	// Regression test for: when a down migration fails, golang-migrate marks
	// the target version as dirty but the actual schema is still at the higher version.
	// We must preserve checksums for all versions including the one that failed to roll back.
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.DB.Close() })

	ctx := context.Background()
	dir := t.TempDir()

	// Create 3 migrations with valid up/down
	for i := 1; i <= 3; i++ {
		upSQL := fmt.Sprintf("CREATE TABLE t%d (id INTEGER);\n", i)
		downSQL := fmt.Sprintf("DROP TABLE t%d;\n", i)
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%06d_mig.up.sql", i)), []byte(upSQL), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, fmt.Sprintf("%06d_mig.down.sql", i)), []byte(downSQL), 0600))
	}

	runner, err := NewMigrationRunner(database, dir)
	require.NoError(t, err)

	// Apply all 3 migrations
	require.NoError(t, runner.Up(ctx))

	// Verify all 3 checksums exist
	var count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums").Scan(&count))
	require.Equal(t, 3, count)

	// Make version 3's down migration invalid
	require.NoError(t, os.WriteFile(filepath.Join(dir, "000003_mig.down.sql"), []byte("INVALID SQL THAT WILL FAIL;\n"), 0600))

	// Attempt rollback - this should fail because down migration is invalid
	err = runner.Rollback(ctx)
	require.Error(t, err)

	// Version should now be 2 with dirty=true (golang-migrate protocol)
	version, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(2), version)
	require.True(t, dirty)

	// IMPORTANT: Checksums for versions 1, 2, 3 should ALL still exist
	// because the actual schema is still at version 3 (down migration failed)
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums").Scan(&count))
	require.Equal(t, 3, count, "all checksums should be preserved when rollback fails and state is dirty")

	// Verify each version's checksum is present
	for v := uint(1); v <= 3; v++ {
		var upHash string
		require.NoError(t, database.QueryRow("SELECT up_sha256 FROM forge_migration_checksums WHERE version = ?", v).Scan(&upHash))
		require.NotEmpty(t, upHash, "checksum for version %d should exist", v)
	}
}

func TestChecksumBaseline_EmptyRollbackReconciles(t *testing.T) {
	database, runner := checksumRegressionRunner(t)
	require.NoError(t, runner.RollbackSteps(context.Background(), 3))
	_, err := database.Exec("INSERT INTO forge_migration_checksums (version, up_sha256) VALUES (3, 'leftover')")
	require.NoError(t, err)
	require.ErrorContains(t, runner.Rollback(context.Background()), "no migrations")
	var count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM forge_migration_checksums").Scan(&count))
	require.Zero(t, count)
}

func TestChecksumBaseline_VerifyDifferentDownStem(t *testing.T) {
	database, runner := checksumRegressionRunner(t)
	ctx := context.Background()
	recovery := execute.NewRecovery(database.DB)
	report, err := recovery.VerifyAgainstBaseline(ctx, runner.migrationsPath)
	require.NoError(t, err)
	require.True(t, report.AllVerified())
	_, err = database.Exec("DELETE FROM forge_migration_checksums")
	require.NoError(t, err)
	adopted, err := runner.AdoptChecksumBaseline(ctx)
	require.NoError(t, err)
	require.Equal(t, []uint{1, 2, 3}, adopted)
	report, err = recovery.VerifyAgainstBaseline(ctx, runner.migrationsPath)
	require.NoError(t, err)
	require.True(t, report.AllVerified())
	require.NoError(t, os.WriteFile(filepath.Join(runner.migrationsPath, "1_create_rollback.down.sql"), []byte("SELECT 3;"), 0600))
	report, err = recovery.VerifyAgainstBaseline(ctx, runner.migrationsPath)
	require.NoError(t, err)
	require.Equal(t, execute.BaselineMismatched, report.Entries[0].Status)
}

func TestChecksumBaseline_RunnerUpBlockedOnMismatch(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := NewDBWithDriver("sqlite3", dbPath)
	require.NoError(t, err)
	defer database.Close()

	file1Up := filepath.Join(dir, "000001_first.up.sql")
	file1Down := filepath.Join(dir, "000001_first.down.sql")
	file2Up := filepath.Join(dir, "000002_second.up.sql")
	file2Down := filepath.Join(dir, "000002_second.down.sql")

	require.NoError(t, os.WriteFile(file1Up, []byte("CREATE TABLE t1 (id INT);\n"), 0600))
	require.NoError(t, os.WriteFile(file1Down, []byte("DROP TABLE t1;\n"), 0600))
	require.NoError(t, os.WriteFile(file2Up, []byte("CREATE TABLE t2 (id INT);\n"), 0600))
	require.NoError(t, os.WriteFile(file2Down, []byte("DROP TABLE t2;\n"), 0600))

	runner, err := NewMigrationRunner(database, dir)
	require.NoError(t, err)
	defer runner.Close()

	// Apply migration 1 only
	require.NoError(t, runner.MigrateTo(ctx, 1))
	v, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), v)
	require.False(t, dirty)

	// Edit migration 1's applied file
	require.NoError(t, os.WriteFile(file1Up, []byte("CREATE TABLE t1_modified (id INT);\n"), 0600))

	// runner.Up must fail with ErrChecksumBaselineBlocked and apply nothing new
	err = runner.Up(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumBaselineBlocked)

	// Verify nothing new was applied: version is still 1, table t2 does not exist
	v, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), v)
	require.False(t, dirty)

	var t2Count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='t2'").Scan(&t2Count))
	require.Zero(t, t2Count)

	// Rollback must not be blocked (recovery must stay possible)
	require.NoError(t, runner.Rollback(ctx))
	v, _, _ = runner.Version(ctx)
	require.Equal(t, uint(0), v)
}

func TestChecksumBaseline_RunnerUpBlockedOnMismatch_Postgres(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Use an isolated schema without altering other tests' migration state.
	schema := fmt.Sprintf("checksum_baseline_blocked_%d", time.Now().UnixNano())
	_, err := sqlDB.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = sqlDB.Exec("DROP SCHEMA " + schema + " CASCADE") })
	// SET is session-local, so initialize every connection used by this runner.
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	forConnections := make([]interface{ Close() error }, 0, 2)
	for i := 0; i < 2; i++ {
		conn, err := sqlDB.Conn(context.Background())
		require.NoError(t, err)
		_, err = conn.ExecContext(context.Background(), "SET search_path TO "+schema)
		require.NoError(t, err)
		forConnections = append(forConnections, conn)
	}
	for _, conn := range forConnections {
		require.NoError(t, conn.Close())
	}
	database := &DB{DB: sqlDB, Driver: "postgres"}

	ctx := context.Background()
	dir := t.TempDir()

	file1Up := filepath.Join(dir, "000001_first.up.sql")
	file1Down := filepath.Join(dir, "000001_first.down.sql")
	file2Up := filepath.Join(dir, "000002_second.up.sql")
	file2Down := filepath.Join(dir, "000002_second.down.sql")

	require.NoError(t, os.WriteFile(file1Up, []byte("CREATE TABLE t1 (id INT);\n"), 0600))
	require.NoError(t, os.WriteFile(file1Down, []byte("DROP TABLE t1;\n"), 0600))
	require.NoError(t, os.WriteFile(file2Up, []byte("CREATE TABLE t2 (id INT);\n"), 0600))
	require.NoError(t, os.WriteFile(file2Down, []byte("DROP TABLE t2;\n"), 0600))

	runner, err := NewMigrationRunner(database, dir)
	require.NoError(t, err)
	defer runner.Close()

	// Apply migration 1 only
	require.NoError(t, runner.MigrateTo(ctx, 1))
	v, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), v)
	require.False(t, dirty)

	// Edit migration 1's applied file
	require.NoError(t, os.WriteFile(file1Up, []byte("CREATE TABLE t1_modified (id INT);\n"), 0600))

	// runner.Up must fail with ErrChecksumBaselineBlocked and apply nothing new
	err = runner.Up(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumBaselineBlocked)
	require.True(t, errors.Is(err, ErrChecksumBaselineBlocked))

	// Verify nothing new was applied: version is still 1, table t2 does not exist
	v, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(1), v)
	require.False(t, dirty)

	var t2Count int
	require.NoError(t, sqlDB.QueryRow("SELECT count(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name = 't2'", schema).Scan(&t2Count))
	require.Zero(t, t2Count)

	// Rollback must not be blocked (recovery must stay possible)
	require.NoError(t, runner.Rollback(ctx))
	v, _, _ = runner.Version(ctx)
	require.Equal(t, uint(0), v)
}
