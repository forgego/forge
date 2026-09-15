package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cli "github.com/forgego/forge/cli/commands/migrations"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/forgego/forge/tests/testhelpers"
	gm "github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

func TestChecksumBaseline(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy_%t", legacy), func(t *testing.T) {
			ctx := context.Background()
			raw, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, testhelpers.LocalPostgresOpts(t.Name()))
			if err != nil {
				if os.Getenv("FORGE_REQUIRE_DB") == "1" {
					t.Fatalf("Postgres unavailable: %v", err)
				}
				t.Skipf("Postgres unavailable: %v", err)
			}
			defer func() { require.NoError(t, cleanup()) }()
			defer raw.Close()
			dir := t.TempDir()
			file := func(v int, direction string) string {
				return filepath.Join(dir, fmt.Sprintf("%06d_test.%s.sql", v, direction))
			}
			for v := 1; v <= 3; v++ {
				require.NoError(t, os.WriteFile(file(v, "up"), []byte(fmt.Sprintf("CREATE TABLE baseline_%d (id INT);", v)), 0600))
				require.NoError(t, os.WriteFile(file(v, "down"), []byte(fmt.Sprintf("DROP TABLE baseline_%d;", v)), 0600))
			}
			if legacy {
				m, err := gm.New("file://"+dir, dsn)
				require.NoError(t, err)
				require.NoError(t, m.Up())
				sourceErr, dbErr := m.Close()
				require.NoError(t, sourceErr)
				require.NoError(t, dbErr)
			} else {
				runner, err := db.NewMigrationRunner(&db.DB{DB: raw, Driver: "postgres"}, dir)
				require.NoError(t, err)
				require.NoError(t, runner.Up(ctx))
				// The existing Close method reports the driver closing its DB pool.
				_ = runner.Close()
			}
			fresh, err := sql.Open("postgres", dsn)
			require.NoError(t, err)
			defer fresh.Close()
			runner, err := db.NewMigrationRunner(&db.DB{DB: fresh, Driver: "postgres"}, dir)
			require.NoError(t, err)
			defer runner.Close()
			recovery := execute.NewRecovery(fresh)
			report, err := recovery.VerifyAgainstBaseline(ctx, dir)
			require.NoError(t, err)
			require.Len(t, report.Entries, 3)
			if legacy {
				require.False(t, report.AllVerified())
				require.False(t, report.HasBlocking())
				require.Equal(t, 3, report.Counts()[execute.BaselineUnverified])
				var exists bool
				require.NoError(t, fresh.QueryRow("SELECT to_regclass('forge_migration_checksums') IS NOT NULL").Scan(&exists))
				require.False(t, exists, "verification must not create the baseline table")
				adopted, err := runner.AdoptChecksumBaseline(ctx)
				require.NoError(t, err)
				require.Equal(t, []uint{1, 2, 3}, adopted)
				report, err = recovery.VerifyAgainstBaseline(ctx, dir)
				require.NoError(t, err)
			}
			require.True(t, report.AllVerified())
			require.Equal(t, 3, report.Counts()[execute.BaselineVerified])
			require.NoError(t, os.WriteFile(file(2, "up"), []byte("-- edited bytes\nCREATE TABLE baseline_2 (id INT);"), 0600))
			report, err = recovery.VerifyAgainstBaseline(ctx, dir)
			require.NoError(t, err)
			require.True(t, report.HasBlocking())
			require.Equal(t, execute.BaselineMismatched, report.Entries[1].Status)
			require.Contains(t, report.Entries[1].Detail, "up")
			adopted, err := runner.AdoptChecksumBaseline(ctx)
			require.NoError(t, err)
			require.Empty(t, adopted)
			report, err = recovery.VerifyAgainstBaseline(ctx, dir)
			require.NoError(t, err)
			require.True(t, report.HasBlocking(), "adoption must not overwrite mismatches")
			u, err := url.Parse(dsn)
			require.NoError(t, err)
			password, _ := u.User.Password()
			for key, value := range map[string]string{"DRIVER": "postgres", "HOST": u.Hostname(), "PORT": u.Port(), "USER": u.User.Username(), "PASSWORD": password, "NAME": strings.TrimPrefix(u.Path, "/"), "SSLMODE": "disable"} {
				t.Setenv("FORGE_DATABASE_"+key, value)
			}
			cmd := cli.NewRecoverCommand().Definition()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{"--verify", "--path", dir})
			require.Error(t, cmd.Execute())
			require.Contains(t, out.String(), "2  mismatched")
			require.Contains(t, out.String(), "state is clean")
			require.NoError(t, os.Remove(file(1, "up")))
			require.NoError(t, os.Remove(file(1, "down")))
			report, err = recovery.VerifyAgainstBaseline(ctx, dir)
			require.NoError(t, err)
			require.Equal(t, execute.BaselineMissing, report.Entries[0].Status)
			_, err = runner.AdoptChecksumBaseline(ctx)
			require.ErrorContains(t, err, "version 1")
		})
	}
}
