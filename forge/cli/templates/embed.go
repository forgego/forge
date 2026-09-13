package templates

import (
	"embed"
	"io/fs"
)

//go:embed templates/*
var templateFS embed.FS

// GetTemplateFS returns the embedded template filesystem
func GetTemplateFS() fs.FS {
	return templateFS
}

// ReadTemplate reads a template file from the embedded filesystem
func ReadTemplate(name string) ([]byte, error) {
	return templateFS.ReadFile("templates/" + name)
}

// TemplateExists checks if a template exists
func TemplateExists(name string) bool {
	_, err := templateFS.ReadFile("templates/" + name)
	return err == nil
}
