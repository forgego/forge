package migrations

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"fmt"
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

func TestRecoverCommand_Execute_DirtyDB_CleanAndVerify(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "dirty.db")

	migUp := filepath.Join(tmpDir, "000001_init.up.sql")
	migDown := filepath.Join(tmpDir, "000001_init.down.sql")
	upContent := []byte("CREATE TABLE dummy (id INT);")
	downContent := []byte("DROP TABLE dummy;")
	require.NoError(t, os.WriteFile(migUp, upContent, 0644))
	require.NoError(t, os.WriteFile(migDown, downContent, 0644))

	upHash := fmt.Sprintf("%x", sha256.Sum256(upContent))
	downHash := fmt.Sprintf("%x", sha256.Sum256(downContent))

	rawDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	_, err = rawDB.Exec(`CREATE TABLE schema_migrations (version uint NOT NULL PRIMARY KEY, dirty boolean NOT NULL);`)
	require.NoError(t, err)
	_, err = rawDB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (1, 1);`)
	require.NoError(t, err)
	_, err = rawDB.Exec(`CREATE TABLE forge_migration_checksums (version INTEGER PRIMARY KEY, up_sha256 TEXT NOT NULL, down_sha256 TEXT, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`)
	require.NoError(t, err)
	_, err = rawDB.Exec(`INSERT INTO forge_migration_checksums (version, up_sha256, down_sha256) VALUES (1, $1, $2);`, upHash, downHash)
	require.NoError(t, err)
	rawDB.Close()

	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", dbPath)

	cmd := NewRecoverCommand().Definition()
	require.NoError(t, cmd.Flags().Set("path", tmpDir))
	require.NoError(t, cmd.Flags().Set("clean", "true"))
	require.NoError(t, cmd.Flags().Set("verify", "true"))
	require.NoError(t, cmd.Flags().Set("version", "1"))

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err = cmd.Execute()
	require.NoError(t, err)

	outputText := out.String()
	require.Contains(t, outputText, "1 verified")
	require.Contains(t, outputText, "has been marked as clean")

	rawDB, err = sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer rawDB.Close()
	var dirty bool
	err = rawDB.QueryRow(`SELECT dirty FROM schema_migrations WHERE version = 1`).Scan(&dirty)
	require.NoError(t, err)
	require.False(t, dirty)
}
