package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverCommand_Definition(t *testing.T) {
	cmd := NewRecoverCommand()
	def := cmd.Definition()

	assert.Equal(t, "recover", def.Use)
	assert.NotEmpty(t, def.Short)
	assert.NotNil(t, def.Flags().Lookup("path"))
	assert.NotNil(t, def.Flags().Lookup("clean"))
	assert.NotNil(t, def.Flags().Lookup("version"))
	assert.NotNil(t, def.Flags().Lookup("verify"))
}

func TestRecoverCommand_Execute_CleanDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Setup SQLite db with clean schema_migrations
	rawDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	_, err = rawDB.Exec(`CREATE TABLE schema_migrations (version uint NOT NULL PRIMARY KEY, dirty boolean NOT NULL);`)
	require.NoError(t, err)
	_, err = rawDB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (1, 0);`)
	require.NoError(t, err)
	rawDB.Close()

	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite")
	cfg.Set("database.name", dbPath)

	cmd := NewRecoverCommand()
	def := cmd.Definition()
	require.NoError(t, def.Flags().Set("path", tmpDir))

	ctx := &core.Context{
		Cmd:    def,
		Config: cfg,
	}

	err = cmd.Execute(ctx, []string{})
	assert.NoError(t, err)
}

func TestRecoverCommand_Execute_DirtyDBAndClean(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "dirty.db")

	rawDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	_, err = rawDB.Exec(`CREATE TABLE schema_migrations (version uint NOT NULL PRIMARY KEY, dirty boolean NOT NULL);`)
	require.NoError(t, err)
	_, err = rawDB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (2, 1);`)
	require.NoError(t, err)
	rawDB.Close()

	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite")
	cfg.Set("database.name", dbPath)

	cmd := NewRecoverCommand()
	def := cmd.Definition()
	require.NoError(t, def.Flags().Set("path", tmpDir))
	require.NoError(t, def.Flags().Set("clean", "true"))
	require.NoError(t, def.Flags().Set("version", "2"))

	ctx := &core.Context{
		Cmd:    def,
		Config: cfg,
	}

	err = cmd.Execute(ctx, []string{})
	assert.NoError(t, err)

	// Verify dirty is now false
	rawDB, err = sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer rawDB.Close()

	var dirty bool
	err = rawDB.QueryRow(`SELECT dirty FROM schema_migrations WHERE version = 2`).Scan(&dirty)
	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestRecoverCommand_Execute_VerifyFlag(t *testing.T) {
	tmpDir := t.TempDir()
	migFile := filepath.Join(tmpDir, "20260907000000_create_test.up.sql")
	require.NoError(t, os.WriteFile(migFile, []byte("CREATE TABLE dummy (id INT);"), 0644))

	dbPath := filepath.Join(tmpDir, "test.db")
	rawDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	rawDB.Close()

	cfg := config.NewConfig()
	cfg.Set("database.driver", "sqlite")
	cfg.Set("database.name", dbPath)

	cmd := NewRecoverCommand()
	def := cmd.Definition()
	require.NoError(t, def.Flags().Set("path", tmpDir))
	require.NoError(t, def.Flags().Set("verify", "true"))

	ctx := &core.Context{
		Cmd:    def,
		Config: cfg,
	}

	err = cmd.Execute(ctx, []string{})
	assert.NoError(t, err)
}
