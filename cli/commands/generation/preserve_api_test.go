package generation

import (
	"github.com/forgego/forge/cli/core"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateWithoutAPIFlagRetainsEarlierAPIFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "models.go"), []byte(testModelSource), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(withAPI bool) {
		cmd := NewGenerateCommand().Definition()
		if err := cmd.Flags().Set("models", dir); err != nil {
			t.Fatal(err)
		}
		if err := cmd.Flags().Set("output", dir); err != nil {
			t.Fatal(err)
		}
		if withAPI {
			if err := cmd.Flags().Set("api", "true"); err != nil {
				t.Fatal(err)
			}
		}
		if err := NewGenerateCommand().Execute(&core.Context{Cmd: cmd}, nil); err != nil {
			t.Fatal(err)
		}
	}
	run(true)
	before, err := os.ReadFile(filepath.Join(dir, "api_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	run(false)
	after, err := os.ReadFile(filepath.Join(dir, "api_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("generation without --api changed earlier API output")
	}
}
