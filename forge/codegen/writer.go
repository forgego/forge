package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/forgego/forge/utils"
)

// Writer writes generated code to files
type Writer struct {
	templates map[string]*template.Template
}

// NewWriter creates a new writer
func NewWriter() *Writer {
	return &Writer{
		templates: make(map[string]*template.Template),
	}
}

// WriteCombined writes all generated code to a single gen.go file
func (w *Writer) WriteCombined(definitions []*ModelDefinition, outputDir string) error {
	if len(definitions) == 0 {
		return nil
	}

	// Get package name from first model
	packageName := definitions[0].Package

	// Create template functions
	t := template.New("combined").Funcs(template.FuncMap{
		"ToLower":  strings.ToLower,
		"ToSnake":  utils.ToSnake,
		"ToCamel":  utils.ToCamel,
		"ToPascal": utils.ToPascal,
	})

	t, err := t.Parse(combinedTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse combined template: %w", err)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Package": packageName,
		"Models":  definitions,
	}

	// Create gen.go file in the output directory
	filename := filepath.Join(outputDir, "gen.go")
	return w.writeTemplate(t, data, filename)
}

// writeTemplate writes a template to a file
func (w *Writer) writeTemplate(t *template.Template, data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
