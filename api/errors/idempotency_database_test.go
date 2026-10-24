package errors

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	forgedb "github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
	_ "github.com/mattn/go-sqlite3"
)

func newSQLiteIdempotencyStore(t *testing.T) *DatabaseStore {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(8)
	t.Cleanup(func() { db.Close() })
	store, err := NewDatabaseStore(db, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestDatabaseStoreClaimAndReplay(t *testing.T) {
	store := newSQLiteIdempotencyStore(t)
	won, err := store.Claim("key", time.Minute)
	if err != nil || !won {
		t.Fatalf("Claim = %v, %v", won, err)
	}
	won, err = store.Claim("key", time.Minute)
	if err != nil || won {
		t.Fatalf("duplicate Claim = %v, %v", won, err)
	}
	response := &CachedResponse{StatusCode: 201, Headers: map[string]string{"x": "y"}, Body: []byte("ok"), Timestamp: time.Now()}
	if err := store.Set("key", response, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCode != response.StatusCode || string(got.Body) != "ok" || got.Headers["x"] != "y" {
		t.Fatalf("replay mismatch: %#v", got)
	}
}

func TestDatabaseStoreConcurrentClaim(t *testing.T) {
	store := newSQLiteIdempotencyStore(t)
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			won, err := store.Claim("shared", time.Minute)
			if err != nil {
				t.Errorf("Claim: %v", err)
				return
			}
			if won {
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("successful claims = %d, want 1", wins.Load())
	}
}

func TestDatabaseStoreReclaimsExpiredClaim(t *testing.T) {
	store := newSQLiteIdempotencyStore(t)
	won, err := store.Claim("key", -time.Second)
	if err != nil || !won {
		t.Fatalf("initial Claim = %v, %v", won, err)
	}
	won, err = store.Claim("key", time.Minute)
	if err != nil || !won {
		t.Fatalf("reclaim = %v, %v", won, err)
	}
}

func TestDatabaseStoreGetReturnsExpiredCleanupError(t *testing.T) {
	store := newSQLiteIdempotencyStore(t)
	if _, err := store.Claim("key", -time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`CREATE TRIGGER reject_expired_delete BEFORE DELETE ON "idempotency_keys" BEGIN SELECT RAISE(ABORT, 'delete rejected'); END`); err != nil {
		t.Fatal(err)
	}

	response, err := store.Get("key")
	if response != nil || err == nil || !strings.Contains(err.Error(), "delete expired idempotency key") || !strings.Contains(err.Error(), "delete rejected") {
		t.Fatalf("Get = (%v, %v), want wrapped expired-key deletion error", response, err)
	}
}

func TestDatabaseStoreFrameworkMigrationsSQLite(t *testing.T) {
	database, err := forgedb.NewDBWithDriver("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := stores.Migrate(context.Background(), database); err != nil {
		t.Fatalf("stores.Migrate: %v", err)
	}

	// Create store from *forgedb.DB without calling EnsureSchema
	store, err := NewDatabaseStore(database, "")
	if err != nil {
		t.Fatalf("NewDatabaseStore: %v", err)
	}

	key := "test-migrated-key"
	resp := &CachedResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"status":"ok"}`),
		Timestamp:  time.Now(),
	}

	if err := store.Set(key, resp, 10*time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.StatusCode != 200 || string(got.Body) != `{"status":"ok"}` {
		t.Fatalf("unexpected cached response: %#v", got)
	}

	if err := store.Delete(key); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := store.Get(key); err == nil {
		t.Fatalf("expected error getting deleted key")
	}
}
