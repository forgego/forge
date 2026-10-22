package generation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

// generate scans the requested models and writes to the requested output.
func TestGenerateWritesModelCodeFromChosenModelDirectoryToChosenOutput(t *testing.T) {
	root := t.TempDir()
	models := filepath.Join(root, "custom_models")
	output := filepath.Join(root, "generated")
	if err := os.Mkdir(models, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(output, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(models, "models.go"), []byte(testModelSource), 0600); err != nil {
		t.Fatal(err)
	}
	command := NewGenerateCommand()
	def := command.Definition()
	if err := def.ParseFlags([]string{"--models", models, "--output", output}); err != nil {
		t.Fatal(err)
	}
	if err := command.Execute(&core.Context{Cmd: def}, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(output, "gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Product") {
		t.Fatal("generated code does not contain Product")
	}
}
