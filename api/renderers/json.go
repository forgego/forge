package renderers

import (
	"encoding/json"
	"io"
)

// JSONRenderer renders data as JSON
type JSONRenderer struct{}

// NewJSONRenderer creates a new JSON renderer
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

// Render renders data to JSON bytes
func (r *JSONRenderer) Render(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// MediaType returns the JSON media type
func (r *JSONRenderer) MediaType() string {
	return "application/json"
}

// RenderToWriter renders data directly to a writer
func (r *JSONRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	return json.NewEncoder(w).Encode(data)
}
