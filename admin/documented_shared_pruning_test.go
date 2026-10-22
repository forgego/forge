package admin

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/stores/storestest"
)

// A token write prunes expired rows from the shared SQLite store.
func TestAdminSharedTokenWritesPruneExpiredRows(t *testing.T) {
	ctx := context.Background()
	database := storestest.Open(t, "sqlite3", storestest.SQLitePath(t))
	storestest.Migrate(t, database)
	site := NewSite("admin").SetDB(database)
	if err := site.UseDatabaseStores(ctx); err != nil {
		t.Fatal(err)
	}
	_, err := database.ExecContext(ctx, "INSERT INTO forge_admin_tokens (token_hash, username, created_at, expires_at) VALUES ($1,$2,$3,$4)", "expired", "old", time.Now().Add(-2*time.Hour).UTC(), time.Now().Add(-time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := site.apiStores.Tokens.IssueToken(ctx, "new", time.Hour); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM forge_admin_tokens WHERE token_hash='expired'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expired rows remaining=%d", count)
	}
}
