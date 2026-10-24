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

func TestMigrationRunner_RollbackSteps(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_rollback.db")
	database, err := NewDBWithDriver("sqlite3", dbPath)
	require.NoError(t, err)
	defer database.Close()

	migrationsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.up.sql"), []byte("CREATE TABLE t1(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.down.sql"), []byte("DROP TABLE t1;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.up.sql"), []byte("CREATE TABLE t2(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.down.sql"), []byte("DROP TABLE t2;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.up.sql"), []byte("CREATE TABLE t3(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.down.sql"), []byte("DROP TABLE t3;"), 0644))

	runner, err := NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Invalid step counts return error
	require.Error(t, runner.RollbackSteps(ctx, 0))
	require.Error(t, runner.RollbackSteps(ctx, -1))

	// Migrate all 3 steps
	require.NoError(t, runner.Migrate(ctx))

	v, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(3), v)

	// Rollback 2 steps
	require.NoError(t, runner.RollbackSteps(ctx, 2))

	v, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(1), v)
}

func TestMigrationRunner_RollbackTo_RejectsUnknownTargetVersion(t *testing.T) {
	// Regression test for: RollbackTo validates target against live directory
	// but executes against runner's source snapshot. If a migration file is added
	// after runner construction, validation passes but execution overshoots.
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test_rollback_target.db"))
	require.NoError(t, err)
	defer database.Close()

	ctx := context.Background()
	migrationsDir := t.TempDir()

	// Create only versions 1 and 3 (gap at version 2)
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.up.sql"), []byte("CREATE TABLE t1(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.down.sql"), []byte("DROP TABLE t1;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.up.sql"), []byte("CREATE TABLE t3(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.down.sql"), []byte("DROP TABLE t3;"), 0644))

	// Build runner with only versions 1 and 3
	runner, err := NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)

	// Apply both migrations (1 and 3)
	require.NoError(t, runner.Migrate(ctx))

	v, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(3), v)

	// NOW add version 2 migration file to the directory (after runner construction)
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.up.sql"), []byte("CREATE TABLE t2(id INTEGER);"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.down.sql"), []byte("DROP TABLE t2;"), 0644))

	// RollbackTo(2) should fail because runner's source doesn't know about version 2
	// Before the fix, this would succeed and overshoot to version 1
	err = runner.RollbackTo(ctx, 2)
	require.Error(t, err)
	require.ErrorContains(t, err, "target version missing from migration source")

	// Version should remain at 3 (no rollback occurred due to error)
	v, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(3), v, "version should not change when rollback target is unknown to runner")
}

func TestMigrationRunner_SkipsVersionAboveMaxInt(t *testing.T) {
	dir := t.TempDir()
	// Invalid SQL so the test fails if the file is treated as pending instead of skipped.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "9223372036854775808_x.up.sql"), []byte("INVALID SQL(((("), 0644))

	runner := &MigrationRunner{migrationsPath: dir}
	ctx := context.Background()

	require.NoError(t, runner.validatePendingMigrations(ctx, 0))
	require.NoError(t, runner.validatePendingMigrationChecksums(ctx, 0))
}

func TestMigrationRunner_OutOfOrderMigrationDetected(t *testing.T) {
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test_ooo.db"))
	require.NoError(t, err)
	defer database.Close()

	ctx := context.Background()
	migrationsDir := t.TempDir()

	// Create and apply migrations 1 and 3 (gap at version 2)
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.up.sql"), []byte("CREATE TABLE t1(id INTEGER);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_first.down.sql"), []byte("DROP TABLE t1;"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.up.sql"), []byte("CREATE TABLE t3(id INTEGER);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_third.down.sql"), []byte("DROP TABLE t3;"), 0o644))

	runner, err := NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)

	require.NoError(t, runner.Migrate(ctx))

	v, dirty, err := runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(3), v)

	// Add an unapplied migration whose version sorts below current version 3
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.up.sql"), []byte("CREATE TABLE t2(id INTEGER);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_second.down.sql"), []byte("DROP TABLE t2;"), 0o644))

	// Recreate runner to pick up newly added file
	runner, err = NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)

	// Migrate must fail and report the out-of-order migration file
	err = runner.Migrate(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrOutOfOrderMigration)
	require.ErrorContains(t, err, "000002_second.up.sql")

	// GetDetailedStatus must report it under OutOfOrder, not Applied
	detailed, err := runner.GetDetailedStatus(ctx)
	require.NoError(t, err)
	require.Contains(t, detailed.OutOfOrder, "[000002] second")
	require.NotContains(t, detailed.Applied, "[000002] second")

	// Database version must remain at 3
	v, dirty, err = runner.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(3), v)
}

func TestOutOfOrderVersions(t *testing.T) {
	tests := []struct {
		name           string
		recorded       map[uint]bool
		currentVersion uint
		fileVersions   []uint
		want           []uint
	}{
		{
			name:           "recorded{3,4}, current 4, files 1..4 -> nil",
			recorded:       map[uint]bool{3: true, 4: true},
			currentVersion: 4,
			fileVersions:   []uint{1, 2, 3, 4},
			want:           nil,
		},
		{
			name:           "recorded{1,2,4}, current 4, files 1..4 -> [3]",
			recorded:       map[uint]bool{1: true, 2: true, 4: true},
			currentVersion: 4,
			fileVersions:   []uint{1, 2, 3, 4},
			want:           []uint{3},
		},
		{
			name:           "recorded{} -> nil",
			recorded:       map[uint]bool{},
			currentVersion: 4,
			fileVersions:   []uint{1, 2, 3, 4},
			want:           nil,
		},
		{
			name:           "nil recorded -> nil",
			recorded:       nil,
			currentVersion: 4,
			fileVersions:   []uint{1, 2, 3, 4},
			want:           nil,
		},
		{
			name:           "currentVersion 0 -> nil",
			recorded:       map[uint]bool{1: true, 2: true},
			currentVersion: 0,
			fileVersions:   []uint{1, 2},
			want:           nil,
		},
		{
			name:           "multiple missing versions sorted ascending",
			recorded:       map[uint]bool{1: true, 5: true},
			currentVersion: 5,
			fileVersions:   []uint{4, 2, 3, 1, 5},
			want:           []uint{2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := outOfOrderVersions(tt.recorded, tt.currentVersion, tt.fileVersions)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLoadRecordedVersions_TableNotExist(t *testing.T) {
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test_not_exist.db"))
	require.NoError(t, err)
	defer database.Close()

	ctx := context.Background()
	recorded, err := loadRecordedVersions(ctx, database, 5)
	require.NoError(t, err)
	require.NotNil(t, recorded)
	require.Empty(t, recorded)
}

func TestLoadRecordedVersions_QueryErrorWrapped(t *testing.T) {
	database, err := NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "test_err.db"))
	require.NoError(t, err)
	require.NoError(t, database.Close())

	ctx := context.Background()
	_, err = loadRecordedVersions(ctx, database, 5)
	require.Error(t, err)
	require.ErrorContains(t, err, "query migration checksum versions")
}

func TestMigrationRunner_PreChecksumDatabaseNotFlagged(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_pre_checksum.db")
	database, err := NewDBWithDriver("sqlite3", dbPath)
	require.NoError(t, err)
	defer database.Close()

	ctx := context.Background()
	migrationsDir := t.TempDir()

	// Create migration files 1, 2, 3
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_m1.up.sql"), []byte("CREATE TABLE m1(id INT);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_m1.down.sql"), []byte("DROP TABLE m1;"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_m2.up.sql"), []byte("CREATE TABLE m2(id INT);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_m2.down.sql"), []byte("DROP TABLE m2;"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_m3.up.sql"), []byte("CREATE TABLE m3(id INT);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_m3.down.sql"), []byte("DROP TABLE m3;"), 0o644))

	// Simulate pre-checksum migration: apply migrations 1 and 2 manually with schema_migrations only
	_, err = database.ExecContext(ctx, "CREATE TABLE schema_migrations (version uint64, dirty bool);")
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "INSERT INTO schema_migrations (version, dirty) VALUES (2, 0);")
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "CREATE TABLE m1(id INT);")
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "CREATE TABLE m2(id INT);")
	require.NoError(t, err)

	// Create forge_migration_checksums table with only version 2 recorded (minRecorded = 2, so version 1 is legacy pre-checksum)
	_, err = database.ExecContext(ctx, sqliteMigrationChecksumsDDL)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, "INSERT INTO forge_migration_checksums (version, up_sha256) VALUES (2, 'hash2');")
	require.NoError(t, err)

	runner, err := NewMigrationRunner(database, migrationsDir)
	require.NoError(t, err)

	// Detailed status should not flag version 1 as out-of-order
	detailed, err := runner.GetDetailedStatus(ctx)
	require.NoError(t, err)
	require.Empty(t, detailed.OutOfOrder)
	require.Contains(t, detailed.Applied, "[000001] m1")
	require.Contains(t, detailed.Applied, "[000002] m2")
	require.Contains(t, detailed.Pending, "[000003] m3")
}
