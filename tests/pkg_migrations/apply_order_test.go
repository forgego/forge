package migrations

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	migrationcmds "github.com/forgego/forge/cli/commands/migrations"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/tests/testhelpers"
)

// forgeRunStatus executes `forge migrate status` using the same in-process runner pattern as forgeRun.
func forgeRunStatus(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := migrationcmds.NewStatusCommand()
	definition := cmd.Definition()
	var out bytes.Buffer
	definition.SetOut(&out)
	definition.SetErr(&out)
	if len(args) > 0 && args[0] == "migrate" {
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "status" {
		args = args[1:]
	}
	require.NoError(t, definition.ParseFlags(args))
	positional := definition.Flags().Args()
	if definition.Args != nil {
		require.NoError(t, definition.Args(definition, positional))
	}
	definition.SetContext(context.Background())
	ctx := core.NewContextWithConfig(config.NewConfig())
	ctx.Cmd = definition
	err := cmd.Execute(ctx, positional)
	return out.String(), err
}

func TestOutOfOrderMigrationIsReported(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	opts := testhelpers.LocalPostgresOpts(t.Name())
	opts.DBName = fmt.Sprintf("order_%d", time.Now().UnixNano())
	conn, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	if err != nil {
		testhelpers.SkipOrFailNoDB(t, "Postgres unavailable: %v", err)
	}
	defer func() { require.NoError(t, cleanup()) }()

	u, err := url.Parse(dsn)
	require.NoError(t, err)
	password, _ := u.User.Password()
	for key, value := range map[string]string{
		"DRIVER": "postgres", "HOST": u.Hostname(), "PORT": u.Port(), "USER": u.User.Username(),
		"PASSWORD": password, "NAME": strings.TrimPrefix(u.Path, "/"), "SSLMODE": "disable",
	} {
		t.Setenv("FORGE_DATABASE_"+key, value)
	}

	work := t.TempDir()
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	// 1. Create and apply migrations 1 and 3 with forge migrate up.
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_initial.up.sql"), []byte("CREATE TABLE t1 (id SERIAL PRIMARY KEY, val TEXT);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000001_initial.down.sql"), []byte("DROP TABLE t1;"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_evolve.up.sql"), []byte("CREATE TABLE t2 (id SERIAL PRIMARY KEY, ref INT REFERENCES t1(id));"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_evolve.down.sql"), []byte("DROP TABLE t2;"), 0o644))

	out, err := forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "Migrations applied successfully")
	require.Equal(t, migrationState{version: 3, dirty: false}, readMigrationState(t, conn))

	// 2. Add an unapplied migration file between applied versions 1 and 3.
	patchFile := "000002_patch.up.sql"
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, patchFile), []byte("CREATE TABLE t0 (id SERIAL PRIMARY KEY);"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000002_patch.down.sql"), []byte("DROP TABLE t0;"), 0o644))

	// 3. Run forge migrate status: verify the CLI reports the unapplied out-of-order file.
	statusOut, _ := forgeRunStatus(t, "migrate", "status", "--path", migrationsDir)
	require.Contains(t, statusOut, "000002", "status must report the unapplied out-of-order migration")
	require.Contains(t, statusOut, "Out-of-Order Migrations", "status must report out-of-order section")

	// 4. Run forge migrate up: assert that CLI reports non-zero exit naming the file and does not report success.
	upOut, upErr := forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.Error(t, upErr, "forge migrate up must fail on out-of-order migration")
	require.NotContains(t, upOut, "Migrations applied successfully", "must not report success while silently skipping")
	require.Contains(t, upErr.Error(), patchFile, "error must name the unapplied out-of-order file")

	// 5. Verify the DB remains intact at version 3 and t0 was not created.
	require.Equal(t, migrationState{version: 3, dirty: false}, readMigrationState(t, conn))
	var t0Exists bool
	err = conn.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 't0')`).Scan(&t0Exists)
	require.NoError(t, err)
	require.False(t, t0Exists, "out-of-order migration must not have been executed")
}
