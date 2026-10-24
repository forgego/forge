package migrations

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

func TestConcurrentApplyIsSerialised(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	opts := testhelpers.LocalPostgresOpts(t.Name())
	opts.DBName = fmt.Sprintf("lifecycle_%d", time.Now().UnixNano())
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
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(lifecycleModelsV1), 0o644))
	out, err := forgeRun(t, "makemigrations", "initial", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)

	type applyResult struct {
		out string
		err error
	}

	results := make([]applyResult, 2)
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			resOut, resErr := forgeRun(t, "migrate", "up", "--path", migrationsDir)
			results[idx] = applyResult{out: resOut, err: resErr}
		}(i)
	}

	close(start)
	wg.Wait()

	// The advisory lock serialises the runs: the second waits, then finds
	// nothing pending. Both succeed and the migration is applied once.
	for i, res := range results {
		require.NoError(t, res.err, "apply %d: %s", i, res.out)
	}

	// Afterwards the schema and the migration version are correct.
	state := readMigrationState(t, conn)
	require.Equal(t, uint(1), state.version)
	require.False(t, state.dirty)

	testhelpers.AssertTableExists(ctx, t, conn, "postgres", "customers")
	testhelpers.AssertTableExists(ctx, t, conn, "postgres", "orders")
}
