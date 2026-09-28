package errors

import (
	"testing"
	"time"

	forgeerrors "github.com/forgego/forge/errors"
)

func TestDatabaseStore_NotImplemented(t *testing.T) {
	store, err := NewDatabaseStore(nil, "")
	if err == nil {
		t.Fatal("expected error from NewDatabaseStore, got nil")
	}
	if !forgeerrors.IsNotImplemented(err) {
		t.Errorf("expected NotImplementedError from NewDatabaseStore, got %v", err)
	}
	if store != nil {
		t.Errorf("expected nil store, got %v", store)
	}

	zeroStore := &DatabaseStore{}
	if err := zeroStore.Set("key", &CachedResponse{}, time.Minute); !forgeerrors.IsNotImplemented(err) {
		t.Errorf("expected NotImplementedError from Set, got %v", err)
	}
	if _, err := zeroStore.Get("key"); !forgeerrors.IsNotImplemented(err) {
		t.Errorf("expected NotImplementedError from Get, got %v", err)
	}
	if err := zeroStore.Delete("key"); !forgeerrors.IsNotImplemented(err) {
		t.Errorf("expected NotImplementedError from Delete, got %v", err)
	}
	if err := zeroStore.Cleanup(); !forgeerrors.IsNotImplemented(err) {
		t.Errorf("expected NotImplementedError from Cleanup, got %v", err)
	}
}
