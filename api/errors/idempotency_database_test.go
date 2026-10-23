package errors

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
