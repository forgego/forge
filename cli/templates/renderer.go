package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	ProjectName     string
	AppName         string
	ModelName       string
	ModelNamePlural string
	TableName       string
	HandlerName     string
	HandlerPath     string
	Method          string
	ServiceName     string
	DomainName      string
	EntityName      string
	InfraName       string
	ClientName      string
	ResourceName    string
	PackageName     string
	MigrationName   string
	Timestamp       string
	DatabaseType    string
}

// RenderTemplate renders a template with the given data
func RenderTemplate(templateName string, data TemplateData) ([]byte, error) {
	// First check for user-overridable template
	userTemplatePath := filepath.Join(".forge", "templates", templateName)
	if _, err := os.Stat(userTemplatePath); err == nil {
		content, err := os.ReadFile(userTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read user template: %w", err)
		}
		return renderTemplateContent(string(content), data)
	}

	// Fallback to embedded template
	content, err := ReadTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	return renderTemplateContent(string(content), data)
}

// renderTemplateContent renders template content with data
func renderTemplateContent(content string, data TemplateData) ([]byte, error) {
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// WriteTemplateFile writes a rendered template to a file
func WriteTemplateFile(filePath string, templateName string, data TemplateData, perm os.FileMode) error {
	content, err := RenderTemplate(templateName, data)
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, content, perm)
}
