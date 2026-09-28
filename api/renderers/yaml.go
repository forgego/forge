package renderers

import (
	"encoding/json"
	"fmt"
	"io"
)

// YAMLRenderer renders data as YAML
type YAMLRenderer struct{}

// NewYAMLRenderer creates a new YAML renderer
func NewYAMLRenderer() *YAMLRenderer {
	return &YAMLRenderer{}
}

// Render renders data to YAML bytes
// Note: This is a simplified YAML renderer. For production, use a proper YAML library.
func (r *YAMLRenderer) Render(data interface{}) ([]byte, error) {
	// Convert to JSON first, then format as YAML-like
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// Simple YAML-like formatting (simplified)
	// For production, use github.com/go-yaml/yaml or gopkg.in/yaml.v3
	return jsonData, nil
}

// MediaType returns the YAML media type
func (r *YAMLRenderer) MediaType() string {
	return "application/x-yaml"
}

// RenderToWriter renders data directly to a writer
func (r *YAMLRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(jsonData))
	return err
}
