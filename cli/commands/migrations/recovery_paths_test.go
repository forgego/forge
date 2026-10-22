package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

// baseline reads the selected migration directory when adopting checksums.
func TestBaselineAdoptsChecksumsFromChosenDirectory(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "chosen")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, dir, "000001", "widgets")
	setMigrationState(t, dbPath, 1, false)
	c := NewBaselineCommand()
	def := c.Definition()
	def.SetContext(context.Background())
	if err := def.ParseFlags([]string{"--path", dir, "--adopt"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	def.SetOut(&out)
	if err := c.Execute(&core.Context{Cmd: def, Config: migrationConfig(dbPath)}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Adopted versions: [1]") || !strings.Contains(out.String(), "1 verified") {
		t.Fatalf("baseline: %s", out.String())
	}
}

// recover uses selected files for advice and cleans only the selected version.
func TestRecoverUsesChosenDirectoryAndVersion(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "chosen")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, dir, "000001", "first")
	writeMigrationPair(t, dir, "000002", "second")
	setMigrationState(t, dbPath, 1, true)
	raw, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (2, true)")
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}
	c := NewRecoverCommand()
	def := c.Definition()
	if err := def.ParseFlags([]string{"--path", dir, "--clean", "--version", "1"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	def.SetOut(&out)
	if err := c.Execute(&core.Context{Cmd: def, Config: migrationConfig(dbPath)}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Migration 1 has been marked as clean") {
		t.Fatalf("recovery output: %s", out.String())
	}
	raw, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	for version, want := range map[int]bool{1: false, 2: true} {
		var dirty bool
		if err := raw.QueryRow("SELECT dirty FROM schema_migrations WHERE version=?", version).Scan(&dirty); err != nil {
			t.Fatal(err)
		}
		if dirty != want {
			t.Fatalf("version %d dirty=%v", version, dirty)
		}
	}
	// The selected path determines the prior-version advice when no cleaning is requested.
	c = NewRecoverCommand()
	def = c.Definition()
	if err := def.ParseFlags([]string{"--path", dir}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	def.SetOut(&out)
	if err := c.Execute(&core.Context{Cmd: def, Config: migrationConfig(dbPath)}, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "forge migrate force 1") {
		t.Fatalf("path advice: %s", out.String())
	}
}

// the verbose option is reserved; its help no longer promises warnings.
func TestMakeMigrationsVerboseHelpDescribesReservedOption(t *testing.T) {
	flag := NewMakeMigrationsCommand().Definition().Flags().Lookup("verbose")
	if flag == nil || !strings.Contains(flag.Usage, "Reserved") {
		t.Fatalf("verbose help: %v", flag)
	}
}
