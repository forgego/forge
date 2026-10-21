package errors

import (
	"bytes"
	"testing"
	"time"
)

func TestIdempotencyKeyReusesCachedResponse(t *testing.T) {
	key := ExtractIdempotencyKey("Idempotency-Key", "retry-123")
	if !key.Valid || key.Key == "retry-123" {
		t.Fatalf("unexpected key: %+v", key)
	}
	again := ExtractIdempotencyKey("Idempotency-Key", "retry-123")
	if again.Key != key.Key {
		t.Fatal("same request key must map to same store key")
	}
	store := NewInMemoryStore()
	response := &CachedResponse{StatusCode: 201, Body: []byte(`{"id":1}`)}
	if err := store.Set(key.Key, response, time.Minute); err != nil {
		t.Fatal(err)
	}
	replay, err := store.Get(again.Key)
	if err != nil || replay.StatusCode != 201 || !bytes.Equal(replay.Body, response.Body) {
		t.Fatalf("replay = %+v, %v", replay, err)
	}
}
