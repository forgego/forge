package generate

import (
	"os"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
)

func TestSquashMigrations_NotImplemented(t *testing.T) {
	dir := t.TempDir()
	squasher := NewSquasher(dir)
	err := squasher.SquashMigrations("000001", "000002", "x")
	if err == nil {
		t.Fatal("expected error from SquashMigrations, got nil")
	}
	if !forgeerrors.IsNotImplemented(err) {
		t.Fatalf("expected NotImplementedError, got: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files in temp dir, found %d", len(entries))
	}
}
