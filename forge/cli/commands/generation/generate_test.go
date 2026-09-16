package generation

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
)

func TestGenerateStrictReportsUnsupportedExpressions(t *testing.T) {
	modelsDir := t.TempDir()
	filename := filepath.Join(modelsDir, "models.go")
	source := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{defaultFields()}
}
func defaultFields() schema.Field { return schema.StringField("name") }
`
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	if output, err := executeGenerate(t, modelsDir, false); err != nil {
		t.Fatalf("non-strict generation failed: %v", err)
	} else if !strings.Contains(output, "warning: "+filename+":5:") {
		t.Fatalf("warning did not name file and line: %q", output)
	}
	if output, err := executeGenerate(t, modelsDir, true); err == nil {
		t.Fatal("strict generation succeeded")
	} else if !strings.Contains(err.Error(), "generation found 1 model expressions") || !strings.Contains(output, "warning: "+filename+":5:") {
		t.Fatalf("strict error/output = %v / %q", err, output)
	}
}

func TestGenerateStrictLeavesPreviousOutputIntact(t *testing.T) {
	modelsDir := t.TempDir()
	filename := filepath.Join(modelsDir, "models.go")
	valid := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.StringField("name")}
}
`
	if err := os.WriteFile(filename, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := executeGenerate(t, modelsDir, false); err != nil {
		t.Fatalf("initial generation failed: %v", err)
	}
	genFile := filepath.Join(modelsDir, "gen.go")
	before, err := os.ReadFile(genFile)
	if err != nil {
		t.Fatal(err)
	}
	unsupported := `package models
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{defaultFields()}
}
func defaultFields() schema.Field { return schema.StringField("name") }
`
	if err := os.WriteFile(filename, []byte(unsupported), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := executeGenerate(t, modelsDir, true); err == nil {
		t.Fatal("strict generation succeeded")
	}
	after, err := os.ReadFile(genFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("strict run truncated previously generated file")
	}
}

func executeGenerate(t *testing.T, modelsDir string, strict bool) (string, error) {
	t.Helper()
	command := NewGenerateCommand().Definition()
	if err := command.Flags().Set("models", modelsDir); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("output", modelsDir); err != nil {
		t.Fatal(err)
	}
	if err := command.Flags().Set("strict", strconv.FormatBool(strict)); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = write
	err = NewGenerateCommand().Execute(&core.Context{Cmd: command}, nil)
	_ = write.Close()
	os.Stderr = original
	output, readErr := io.ReadAll(read)
	_ = read.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(output), err
}
