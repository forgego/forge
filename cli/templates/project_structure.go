package templates

import "os"

// ProjectTemplate represents a project structure template
type ProjectTemplate string

const (
	// TemplateSimple is the simple Django-like app-based structure
	TemplateSimple ProjectTemplate = "simple"
	// TemplateAdvanced is the advanced hybrid structure with domain/infra
	TemplateAdvanced ProjectTemplate = "advanced"
)

// ProjectStructure defines the directory structure for a project template
type ProjectStructure struct {
	Template    ProjectTemplate
	Directories []string
	Files       []FileTemplate
}

// FileTemplate defines a file to be generated
type FileTemplate struct {
	Path        string
	Template    string
	Permissions os.FileMode
}

// GetSimpleStructure returns the simple template structure
func GetSimpleStructure() *ProjectStructure {
	return &ProjectStructure{
		Template: TemplateSimple,
		Directories: []string{
			"cmd/server",
			"app",
			"migrations",
			"config",
			"static",
			"templates",
		},
		Files: []FileTemplate{
			{
				Path:        "cmd/server/main.go",
				Template:    "simple_main.go.tmpl",
				Permissions: 0644,
			},
			{
				Path:        "go.mod",
				Template:    "go_mod.tmpl",
				Permissions: 0644,
			},
			{
				Path:        "README.md",
				Template:    "readme_simple.tmpl",
				Permissions: 0644,
			},
		},
	}
}

// GetAdvancedStructure returns the advanced template structure
func GetAdvancedStructure() *ProjectStructure {
	return &ProjectStructure{
		Template: TemplateAdvanced,
		Directories: []string{
			"cmd/server",
			"app",
			"domain",
			"infra",
			"pkg",
			"migrations",
			"config",
			"static",
			"templates",
		},
		Files: []FileTemplate{
			{
				Path:        "cmd/server/main.go",
				Template:    "advanced_main.go.tmpl",
				Permissions: 0644,
			},
			{
				Path:        "go.mod",
				Template:    "go_mod.tmpl",
				Permissions: 0644,
			},
			{
				Path:        "README.md",
				Template:    "readme_advanced.tmpl",
				Permissions: 0644,
			},
		},
	}
}

// GetStructure returns the structure for a given template
func GetStructure(template ProjectTemplate) *ProjectStructure {
	switch template {
	case TemplateAdvanced:
		return GetAdvancedStructure()
	default:
		return GetSimpleStructure()
	}
}
