package renderers

import (
	"go.yaml.in/yaml/v3"
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
	return yaml.Marshal(data)
}

// MediaType returns the YAML media type
func (r *YAMLRenderer) MediaType() string {
	return "application/x-yaml"
}

// RenderToWriter renders data directly to a writer
func (r *YAMLRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	encoded, err := r.Render(data)
	if err != nil {
		return err
	}
	_, err = w.Write(encoded)
	return err
}
