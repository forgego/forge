package migrate

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/core"
)

// A basic generated SQLite migration can be applied in process.
func TestGeneratedSQLiteMigrationAppliesBasicTable(t *testing.T) {
	models := t.TempDir()
	migrations := t.TempDir()
	writeTestModel(t, models)
	if err := GenerateForDriver("create_items", models, migrations, core.DriverSQLite); err != nil {
		t.Fatal(err)
	}
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "app.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	runner, err := db.NewMigrationRunner(database, migrations)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	if err := runner.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='items'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("items table count=%d", count)
	}
}
