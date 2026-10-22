package migrations

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

func migrationVersion(t *testing.T, path string) (int, bool) {
	t.Helper()
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	var version int
	var dirty bool
	if err := raw.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	return version, dirty
}
func setMigrationState(t *testing.T, path string, version int, dirty bool) {
	t.Helper()
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_, err = raw.Exec("CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, dirty BOOLEAN NOT NULL)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)", version, dirty)
	if err != nil {
		t.Fatal(err)
	}
}
func executeMigration(t *testing.T, command interface {
	Definition() *cobra.Command
	Execute(*core.Context, []string) error
}, dbPath string, flags []string, args ...string) error {
	t.Helper()
	def := command.Definition()
	if err := def.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return command.Execute(&core.Context{Cmd: def, Config: migrationConfig(dbPath)}, args)
}

// fake uses the selected files to record a version without executing SQL.
func TestFakeMarksSelectedMigrationAppliedWithoutExecutingSQL(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "chosen")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, migrations, "000001", "widgets")
	err := executeMigration(t, NewFakeCommand(), dbPath, []string{"--path", migrations}, "1")
	if err != nil {
		t.Fatal(err)
	}
	if sqliteTableExists(t, dbPath, "widgets") {
		t.Fatal("fake executed the up SQL")
	}
	version, dirty := migrationVersion(t, dbPath)
	if version != 1 || dirty {
		t.Fatalf("state = %d dirty=%v", version, dirty)
	}
}

// existing initial tables are adopted without executing their CREATE TABLE.
func TestFakeInitialAdoptsExistingInitialTable(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, migrations, "000001", "widgets")
	raw, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec("CREATE TABLE widgets (id INTEGER PRIMARY KEY)")
	raw.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := executeMigration(t, NewFakeCommand(), dbPath, []string{"--path", migrations, "--fake-initial"}); err != nil {
		t.Fatal(err)
	}
	version, dirty := migrationVersion(t, dbPath)
	if version != 1 || dirty {
		t.Fatalf("state = %d dirty=%v", version, dirty)
	}
}

// force clears dirty state at the requested version using the chosen directory.
func TestForceSetsCleanVersionFromChosenMigrationDirectory(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "chosen")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, migrations, "000001", "widgets")
	writeMigrationPair(t, migrations, "000002", "gadgets")
	setMigrationState(t, dbPath, 1, true)
	if err := executeMigration(t, NewForceCommand(), dbPath, []string{"--path", migrations}, "2"); err != nil {
		t.Fatal(err)
	}
	version, dirty := migrationVersion(t, dbPath)
	if version != 2 || dirty {
		t.Fatalf("state = %d dirty=%v", version, dirty)
	}
}

// migrate exposes the migration management commands.
func TestMigrateGroupsMigrationCommands(t *testing.T) {
	group := NewMigrateGroup()
	if group.Definition().Use != "migrate" {
		t.Fatal("wrong group name")
	}
	names := map[string]bool{}
	for _, c := range group.Commands() {
		names[strings.Fields(c.Definition().Use)[0]] = true
	}
	for _, name := range []string{"up", "rollback", "lint", "makemigrations", "fake", "force", "recover", "baseline"} {
		if !names[name] {
			t.Fatalf("missing %s", name)
		}
	}
}

// rollback reverts the newest applied migration and retains older tables.
func TestRollbackRevertsNewestAppliedMigration(t *testing.T) {
	root := t.TempDir()
	migrations := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "app.db")
	writeMigrationPair(t, migrations, "000001", "first")
	writeMigrationPair(t, migrations, "000002", "second")
	if err := executeMigration(t, NewUpCommand(), dbPath, []string{"--path", migrations}); err != nil {
		t.Fatal(err)
	}
	if err := executeMigration(t, NewRollbackCommand(), dbPath, []string{"--path", migrations}); err != nil {
		t.Fatal(err)
	}
	if !sqliteTableExists(t, dbPath, "first") || sqliteTableExists(t, dbPath, "second") {
		t.Fatal("rollback did not retain first and remove second")
	}
}
