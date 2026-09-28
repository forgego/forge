package migrations

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestBaselineCommands(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	t.Setenv("FORGE_DATABASE_NAME", path)
	raw, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	defer raw.Close()
	_, err = raw.Exec("CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, dirty BOOLEAN); INSERT INTO schema_migrations VALUES (1, false)")
	require.NoError(t, err)
	file := filepath.Join(dir, "000001_test.up.sql")
	require.NoError(t, os.WriteFile(file, []byte("SELECT 1;"), 0600))
	run := func(command interface{ Definition() *cobra.Command }, flags ...string) (string, error) {
		cmd := command.Definition()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(append([]string{"--path", dir}, flags...))
		err := cmd.Execute()
		return out.String(), err
	}
	out, err := run(NewBaselineCommand())
	require.Error(t, err)
	require.Contains(t, out, "1  unverified")
	out, err = run(NewStatusCommand())
	require.Error(t, err)
	require.Contains(t, out, "Checksum baseline:")
	out, err = run(NewBaselineCommand(), "--adopt")
	require.NoError(t, err)
	require.Contains(t, out, "Adopted versions: [1]")
	require.Contains(t, out, "1 verified")
	require.NoError(t, os.WriteFile(file, []byte("SELECT 2;"), 0600))
	out, err = run(NewBaselineCommand(), "--adopt")
	require.Error(t, err)
	require.Contains(t, out, "1  mismatched")
	_, err = run(NewUpCommand())
	require.ErrorContains(t, err, "1 (mismatched)")
	require.ErrorContains(t, err, "forge migrations recover --verify")
	require.NoError(t, os.Remove(file))
	out, err = run(NewStatusCommand())
	require.Error(t, err)
	require.Contains(t, out, "1  missing")
	_, err = raw.Exec("UPDATE schema_migrations SET dirty = true")
	require.NoError(t, err)
	out, err = run(NewRecoverCommand(), "--verify")
	require.Error(t, err)
	require.Contains(t, out, "Dirty migration detected at version 1")
}
