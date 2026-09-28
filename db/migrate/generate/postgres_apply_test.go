package generate_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/forgego/forge/db/migrate/core"
)

// applyOnPostgres applies every migration in dir up, then down, then up again
// in a scratch schema of FORGE_TEST_DATABASE_URL, proving the generated SQL is
// valid PostgreSQL. It does nothing for other drivers, and when no database is
// configured or reachable.
func applyOnPostgres(t *testing.T, driver core.Driver, dir string) {
	t.Helper()
	if !driver.IsPostgreSQL() {
		return
	}
	url := os.Getenv("FORGE_TEST_DATABASE_URL")
	if url == "" {
		return
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Logf("PostgreSQL unreachable, not applying migrations: %v", err)
		return
	}
	defer conn.Close()
	schema := fmt.Sprintf("forge_gen_%d", time.Now().UnixNano())
	for _, stmt := range []string{"CREATE SCHEMA " + schema, "SET search_path TO " + schema} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Logf("PostgreSQL unusable, not applying migrations: %v", err)
			return
		}
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()

	var ups, downs []string
	for _, name := range migrationFiles(t, dir) {
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			ups = append(ups, name)
		case strings.HasSuffix(name, ".down.sql"):
			downs = append([]string{name}, downs...)
		}
	}
	exec := func(names []string) {
		for _, name := range names {
			content, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := conn.ExecContext(ctx, string(content)); err != nil {
				t.Fatalf("PostgreSQL rejects %s: %v\n%s", name, err, content)
			}
		}
	}
	exec(ups)
	exec(downs)
	exec(ups)
}
