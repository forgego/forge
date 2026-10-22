package safepath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileStaysWithinRoot(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "migration.sql")
	if err := os.WriteFile(inside, []byte("SELECT 1"), 0600); err != nil {
		t.Fatal(err)
	}
	content, err := ReadFile(root, inside)
	if err != nil || string(content) != "SELECT 1" {
		t.Fatalf("inside read = %q, %v", content, err)
	}
	outside := filepath.Join(t.TempDir(), "secret.sql")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(root, outside); err == nil {
		t.Fatal("read outside root succeeded")
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape.sql")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := ReadFile(root, filepath.Join(root, "escape.sql")); err == nil {
		t.Fatal("symlink escape succeeded")
	}
}
