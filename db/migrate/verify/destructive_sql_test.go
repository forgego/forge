package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintFlagsDropTableAndTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "000001_destructive.up.sql")
	if err := os.WriteFile(path, []byte("DROP TABLE products;\nTRUNCATE orders;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := NewLinter().LintMigration(path)
	if err != nil {
		t.Fatal(err)
	}
	var drop, truncate bool
	for _, r := range results {
		drop = drop || strings.Contains(r.Message, "DROP TABLE")
		truncate = truncate || strings.Contains(r.Message, "TRUNCATE")
	}
	if !drop || !truncate {
		t.Fatalf("lint results = %+v", results)
	}
}
