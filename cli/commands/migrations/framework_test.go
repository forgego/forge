package migrations

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runMigrateCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func sqliteTableExists(t *testing.T, path, table string) bool {
	t.Helper()
	raw, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	defer raw.Close()
	var count int
	require.NoError(t, raw.QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count))
	return count == 1
}

func TestMigrateUp_AppliesFrameworkStoreTablesWhenStoresIsDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	migrations := filepath.Join(dir, "migrations")
	require.NoError(t, os.MkdirAll(migrations, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(migrations, "000001_init.up.sql"), []byte("CREATE TABLE widgets (id INTEGER PRIMARY KEY);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrations, "000001_init.down.sql"), []byte("DROP TABLE widgets;"), 0o644))
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", dbPath)
	t.Setenv("FORGE_SERVER_STORES", "database")

	out, err := runMigrateCommand(t, NewUpCommand().Definition(), "--path", migrations)
	require.NoError(t, err, out)
	assert.Contains(t, out, "Framework store tables migrated to version 1")
	for _, table := range []string{"widgets", "forge_sessions", "forge_rate_limits", "forge_admin_tokens", "forge_admin_saved_views", "forge_admin_log", "forge_framework_migrations"} {
		assert.True(t, sqliteTableExists(t, dbPath, table), table)
	}

	out, err = runMigrateCommand(t, NewUpCommand().Definition(), "--path", migrations)
	require.NoError(t, err, out)
	assert.NotContains(t, out, "Framework store tables migrated", "nothing left to apply")

	out, err = runMigrateCommand(t, NewStatusCommand().Definition(), "--path", migrations)
	require.NoError(t, err, out)
	assert.Contains(t, out, "Framework store tables (server.stores: database): version 1 of 1, up to date")
	assert.Contains(t, out, "Current Version: 1", "the application's version is tracked separately")
}

func TestMigrateUp_MemoryStoresLeaveFrameworkTablesOut(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "000001_init.up.sql"), []byte("CREATE TABLE widgets (id INTEGER PRIMARY KEY);"), 0o644))
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", dbPath)
	t.Setenv("FORGE_SERVER_STORES", "")

	out, err := runMigrateCommand(t, NewUpCommand().Definition(), "--path", dir)
	require.NoError(t, err, out)
	assert.True(t, sqliteTableExists(t, dbPath, "widgets"))
	assert.False(t, sqliteTableExists(t, dbPath, "forge_sessions"), "server.stores defaults to memory")
	assert.False(t, sqliteTableExists(t, dbPath, "forge_framework_migrations"))
}

func TestMigrateUp_FrameworkTablesWithoutApplicationMigrations(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "app.db")
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", dbPath)
	t.Setenv("FORGE_SERVER_STORES", "database")

	out, err := runMigrateCommand(t, NewUpCommand().Definition(), "--path", dir)
	require.NoError(t, err, out)
	assert.Contains(t, out, "No application migrations in")
	assert.True(t, sqliteTableExists(t, dbPath, "forge_sessions"))

	out, err = runMigrateCommand(t, NewStatusCommand().Definition(), "--path", dir)
	require.NoError(t, err, out)
	assert.Contains(t, out, "Framework store tables (server.stores: database): version 1 of 1, up to date",
		"the framework version shows even without application migrations")
}

func TestMigrateUp_RejectsUnknownStoresSetting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", filepath.Join(dir, "app.db"))
	t.Setenv("FORGE_SERVER_STORES", "redis")

	_, err := runMigrateCommand(t, NewUpCommand().Definition(), "--path", dir)
	require.ErrorContains(t, err, `invalid server.stores "redis"`)

	_, err = runMigrateCommand(t, NewUpCommand().Definition(), "--path", dir, "--dry-run")
	require.ErrorContains(t, err, `invalid server.stores "redis"`, "a dry run rejects what the real run rejects")
}
