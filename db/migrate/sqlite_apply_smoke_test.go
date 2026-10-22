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

// The supported in-process boundary includes applying and reversing a basic table migration.
func TestGeneratedSQLiteMigrationRoundTripInMemory(t *testing.T) {
	models := t.TempDir()
	migrations := t.TempDir()
	writeTestModel(t, models)
	if err := GenerateForDriver("create_items", models, migrations, core.DriverSQLite); err != nil {
		t.Fatal(err)
	}
	database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	runner, err := db.NewMigrationRunner(database, migrations)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	ctx := context.Background()
	if err := runner.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, "INSERT INTO items (name) VALUES (?)", "sample"); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := database.QueryRowContext(ctx, "SELECT name FROM items").Scan(&name); err != nil || name != "sample" {
		t.Fatalf("read inserted item: name=%q err=%v", name, err)
	}
	if err := runner.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='items'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("items table after rollback: count=%d err=%v", count, err)
	}
}
