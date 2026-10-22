package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/db/migrate/state"
)

func TestDetectDependenciesRejectsUnreadableMigration(t *testing.T) {
	dir := t.TempDir()
	name := "000001_missing.up.sql"
	if err := os.Symlink(filepath.Join(dir, "gone"), filepath.Join(dir, name)); err != nil {
		t.Fatal(err)
	}
	detector := NewDependencyDetector(dir, state.NewInMemoryState())
	_, err := detector.DetectDependencies(nil, "REFERENCES users")
	if err == nil || !strings.Contains(err.Error(), name) {
		t.Fatalf("dependency error = %v, want file name", err)
	}
}

func TestDetectDependenciesRejectsStateReplayFailure(t *testing.T) {
	dir := t.TempDir()
	name := "000001_bad.up.sql"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("CREATE MATERIALIZED VIEW report AS SELECT 1;"), 0600); err != nil {
		t.Fatal(err)
	}
	detector := NewDependencyDetector(dir, state.NewFileStateLoader(dir))
	_, err := detector.DetectDependencies(nil, "")
	if err == nil || !strings.Contains(err.Error(), name) {
		t.Fatalf("dependency error = %v, want file name", err)
	}
}

func TestGenerateRejectsDependencyReadFailureAfterStateLoad(t *testing.T) {
	models := t.TempDir()
	migrations := t.TempDir()
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(`package models
import "github.com/forgego/forge/schema"
type User struct { schema.BaseSchema; ID int64 }
func (User) Fields() []schema.Field { return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()} }
`), 0600); err != nil {
		t.Fatal(err)
	}
	name := "000001_missing.up.sql"
	if err := os.Symlink(filepath.Join(migrations, "gone"), filepath.Join(migrations, name)); err != nil {
		t.Fatal(err)
	}
	gen, err := NewMigrationGeneratorForDriver(models, migrations, "postgres")
	if err != nil {
		t.Fatal(err)
	}
	gen.stateManager = state.NewInMemoryState()
	err = gen.GenerateMigrations("initial")
	if err == nil || !strings.Contains(err.Error(), name) {
		t.Fatalf("GenerateMigrations error = %v, want dependency read failure", err)
	}
	files, err := filepath.Glob(filepath.Join(migrations, "*initial*"))
	if err != nil || len(files) != 0 {
		t.Fatalf("new migration files = %v, error = %v", files, err)
	}
}
