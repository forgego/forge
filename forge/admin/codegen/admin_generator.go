package codegen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// AdminGenerator generates admin code from templates
type AdminGenerator struct {
	templates map[string]*template.Template
	outputDir string
}

// NewAdminGenerator creates a new admin generator
func NewAdminGenerator(outputDir string) *AdminGenerator {
	return &AdminGenerator{
		templates: make(map[string]*template.Template),
		outputDir: outputDir,
	}
}

// GenerateAdmin generates admin code for a model
func (ag *AdminGenerator) GenerateAdmin(modelName, packageName string, fields []FieldInfo) error {
	// Generate admin file
	adminCode, err := ag.generateAdminFile(modelName, packageName, fields)
	if err != nil {
		return fmt.Errorf("failed to generate admin file: %w", err)
	}

	// Write to file
	outputPath := filepath.Join(ag.outputDir, fmt.Sprintf("%s_admin.go", strings.ToLower(modelName)))
	return ag.writeFormattedFile(outputPath, adminCode)
}

// FieldInfo contains information about a model field
type FieldInfo struct {
	Name        string
	Type        string
	Tag         string
	IsRequired  bool
	IsReadOnly  bool
	VerboseName string
	HelpText    string
}

// generateAdminFile generates admin code
func (ag *AdminGenerator) generateAdminFile(modelName, packageName string, fields []FieldInfo) (string, error) {
	tmpl := `package {{.PackageName}}

import (
	admincore "github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// Register{{.ModelName}}Admin registers {{.ModelName}} with admin
// This function should be called during application initialization
func Register{{.ModelName}}Admin(
	schemaInstance schema.Schema,
	manager *orm.Manager[{{.ModelName}}],
) (*admin.Admin[{{.ModelName}}], error) {
	config := &admin.Config[{{.ModelName}}]{
		VerboseName:       "{{.ModelName}}",
		VerboseNamePlural: "{{.ModelName}}s",
		ListPerPage:       20,
		{{if .Fields}}
		ListDisplay: []interface{}{
			{{range .Fields}}"{{.Name}}",
			{{end}}
		},
		SearchFields: []interface{}{
			{{range .SearchFields}}"{{.}}",
			{{end}}
		},
		{{end}}
	}

	return admin.Register[{{.ModelName}}](schemaInstance, manager, config)
}
`

	t := template.Must(template.New("admin").Parse(tmpl))

	var searchFields []string
	for _, field := range fields {
		if field.Type == "string" || field.Type == "text" {
			searchFields = append(searchFields, field.Name)
		}
	}

	data := map[string]interface{}{
		"PackageName":  packageName,
		"ModelName":    modelName,
		"Fields":       fields,
		"SearchFields": searchFields,
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// writeFormattedFile writes formatted Go code to a file
func (ag *AdminGenerator) writeFormattedFile(path string, code string) error {
	// Format the code
	formatted, err := format.Source([]byte(code))
	if err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	return os.WriteFile(path, formatted, 0644)
}

