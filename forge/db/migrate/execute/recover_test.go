package execute

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMigrationsTable(t *testing.T, db interface{ Exec(string, ...any) (any, error) }) {
    // Check if table exists
    // In testutils we truncate tables.
    // We can just create it.
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
    if err != nil {
        // Retry with just Exec if interface mismatch (sql.DB Exec returns Result, error)
        // This helper signature is tricky.
    }
}

func TestRecovery_GetDirtyMigrationInfo(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()

    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

	t.Run("No migrations", func(t *testing.T) {
        // Clear table first
        _, err := db.Exec("TRUNCATE TABLE schema_migrations")
        require.NoError(t, err)
        
		info, err := recovery.GetDirtyMigrationInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint(0), info.Version)
		assert.False(t, info.Dirty)
	})

	t.Run("Clean migration", func(t *testing.T) {
        _, err := db.Exec("TRUNCATE TABLE schema_migrations")
        require.NoError(t, err)

		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, false)")
		require.NoError(t, err)

		info, err := recovery.GetDirtyMigrationInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.Version)
		assert.False(t, info.Dirty)
	})

	t.Run("Dirty migration", func(t *testing.T) {
        _, err := db.Exec("TRUNCATE TABLE schema_migrations")
        require.NoError(t, err)
        
        // Insert clean first? No, update implies existence.
		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, true)")
		require.NoError(t, err)

		info, err := recovery.GetDirtyMigrationInfo(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.Version)
		assert.True(t, info.Dirty)
	})
}

func TestRecovery_RecoverDirtyState(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()
    
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

	t.Run("No dirty migrations", func(t *testing.T) {
        _, err := db.Exec("TRUNCATE TABLE schema_migrations")
        require.NoError(t, err)

		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, false)")
		require.NoError(t, err)

		dirtyMig, err := recovery.RecoverDirtyState(ctx, "")
		require.NoError(t, err)
		assert.Nil(t, dirtyMig)
	})

	t.Run("Dirty migration found", func(t *testing.T) {
        _, err := db.Exec("TRUNCATE TABLE schema_migrations")
        require.NoError(t, err)

		_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (2, true)")
		require.NoError(t, err)

		dirtyMig, err := recovery.RecoverDirtyState(ctx, "")
		require.NoError(t, err)
		require.NotNil(t, dirtyMig)
		assert.Equal(t, uint(2), dirtyMig.Version)
	})
}

func TestRecovery_MarkMigrationClean(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()
    
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

    // Clean first
    _, err = db.Exec("TRUNCATE TABLE schema_migrations")
    require.NoError(t, err)

	// Insert dirty migration
	_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, true)")
	require.NoError(t, err)

	// Mark as clean
	err = recovery.MarkMigrationClean(ctx, 1)
	require.NoError(t, err)

	// Verify
	var dirty bool
	err = db.QueryRow("SELECT dirty FROM schema_migrations WHERE version = 1").Scan(&dirty)
	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestRecovery_GetAppliedMigrations(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()
    
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

	// Insert test migrations
	migrations := []struct {
		version uint
		dirty   bool
	}{
		{1, false},
		{2, false},
		{3, true},
		{4, false},
	}

    // Clean first
    _, err = db.Exec("TRUNCATE TABLE schema_migrations")
    require.NoError(t, err)

	for _, m := range migrations {
		_, err := db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES ($1, $2)", m.version, m.dirty)
		require.NoError(t, err)
	}

	// Get all migrations
	applied, err := recovery.GetAppliedMigrations(ctx)
	require.NoError(t, err)
	assert.Len(t, applied, 4)

	// Verify order (should be ascending)
	for i := 0; i < len(applied)-1; i++ {
		assert.Less(t, applied[i].Version, applied[i+1].Version)
	}

	// Verify dirty flag
	assert.False(t, applied[0].Dirty)
	assert.True(t, applied[2].Dirty) // Version 3
}


func TestRecovery_ForceCleanState(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()
    
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

    // Clean first
    _, err = db.Exec("TRUNCATE TABLE schema_migrations")
    require.NoError(t, err)

	// Insert dirty migration
	_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, true)")
	require.NoError(t, err)

	// Force clean
	err = recovery.ForceCleanState(ctx, 1)
	require.NoError(t, err)

	// Verify
	var dirty bool
	err = db.QueryRow("SELECT dirty FROM schema_migrations WHERE version = 1").Scan(&dirty)
	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestRecovery_ValidateMigrationIntegrity(t *testing.T) {
	recovery := NewRecovery(nil) // No DB needed for file operations

	// Create temp directory with test migrations
	tmpDir := t.TempDir()

	testMigrations := []struct {
		filename string
		content  string
	}{
		{"20240101000001_create_users.up.sql", "CREATE TABLE users (id INT);"},
		{"20240101000002_add_index.up.sql", "CREATE INDEX idx_users ON users(id);"},
	}

	for _, tm := range testMigrations {
		path := filepath.Join(tmpDir, tm.filename)
		err := os.WriteFile(path, []byte(tm.content), 0644)
		require.NoError(t, err)
	}

	// Validate integrity
	checksums, err := recovery.ValidateMigrationIntegrity(tmpDir)
	require.NoError(t, err)
	assert.Len(t, checksums, 2)

	// Verify checksums are consistent
	checksums2, err := recovery.ValidateMigrationIntegrity(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, checksums, checksums2)
}

func TestRecovery_CompareChecksums(t *testing.T) {
	recovery := NewRecovery(nil)

	// Create temp directory with test migrations
	tmpDir := t.TempDir()
	migrationFile := filepath.Join(tmpDir, "20240101000001_test.up.sql")
	
	// Initial content
	err := os.WriteFile(migrationFile, []byte("CREATE TABLE test;"), 0644)
	require.NoError(t, err)

	// Get initial checksums
	initialChecksums, err := recovery.ValidateMigrationIntegrity(tmpDir)
	require.NoError(t, err)

	t.Run("No modifications", func(t *testing.T) {
		modified, err := recovery.CompareChecksums(tmpDir, initialChecksums)
		require.NoError(t, err)
		assert.Empty(t, modified)
	})

	t.Run("Modified migration", func(t *testing.T) {
		// Modify the file
		err := os.WriteFile(migrationFile, []byte("CREATE TABLE test_modified;"), 0644)
		require.NoError(t, err)

		modified, err := recovery.CompareChecksums(tmpDir, initialChecksums)
		require.NoError(t, err)
		assert.Contains(t, modified, uint(20240101000001))
	})
}

func TestRecovery_RollbackPartialMigration(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer db.Close()
    
    _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`)
	require.NoError(t, err)

	recovery := NewRecovery(db)
	ctx := context.Background()

    // Clean first
    _, err = db.Exec("TRUNCATE TABLE schema_migrations")
    require.NoError(t, err)
    _, _ = db.Exec("DROP TABLE IF EXISTS test_table")

	// Create a test table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)

	// Insert migration record
	_, err = db.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (1, true)")
	require.NoError(t, err)

	// Rollback with down SQL
	downSQL := "DROP TABLE test_table;"
	err = recovery.RollbackPartialMigration(ctx, 1, downSQL)
	require.NoError(t, err)

	// Verify table is dropped
	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'test_table')").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists)

	// Verify migration record is removed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRecovery_SplitSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected int
	}{
		{
			name:     "Single statement",
			sql:      "CREATE TABLE users (id INT);",
			expected: 1,
		},
		{
			name:     "Multiple statements",
			sql:      "CREATE TABLE users (id INT); CREATE INDEX idx ON users(id);",
			expected: 2,
		},
		{
			name:     "Statements with whitespace",
			sql:      "  CREATE TABLE t1;  \n\n  CREATE TABLE t2;  ",
			expected: 2,
		},
		{
			name:     "Empty statements",
			sql:      ";;;",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements := splitSQL(tt.sql)
			assert.Len(t, statements, tt.expected)
		})
	}
}

func TestRecovery_GetRecoverySteps(t *testing.T) {
	recovery := NewRecovery(nil)

	t.Run("Generic error", func(t *testing.T) {
		steps := recovery.GetRecoverySteps("001", "some error")
		assert.NotEmpty(t, steps)
		assert.Contains(t, steps[0], "error message")
	})

	t.Run("Constraint error", func(t *testing.T) {
		steps := recovery.GetRecoverySteps("001", "constraint violation")
		found := false
		for _, step := range steps {
			if containsIgnoreCase(step, "foreign key") || containsIgnoreCase(step, "constraint") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should mention constraint handling")
	})

	t.Run("Column error", func(t *testing.T) {
		steps := recovery.GetRecoverySteps("001", "column already exists")
		found := false
		for _, step := range steps {
			if containsIgnoreCase(step, "column") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should mention column handling")
	})
}

func containsIgnoreCase(s, substr string) bool {
	return contains(toLower(s), toLower(substr))
}

func toLower(s string) string {
	// Simple toLower implementation
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr) >= 0
}

func findSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

