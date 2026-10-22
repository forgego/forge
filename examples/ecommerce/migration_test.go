package main

import (
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/generate"
)

func TestEcommerceModelsGeneratePostgresMigration(t *testing.T) {
	migrations := t.TempDir()
	generator, err := generate.NewMigrationGeneratorForDriver("app", migrations, core.DriverPostgreSQL)
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.GenerateMigrations("initial"); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(migrations, "*.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected a PostgreSQL migration")
	}
}
