package parsers

import (
	"encoding/json"
	"io"
)

// JSONParser parses JSON data
type JSONParser struct{}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse parses JSON data from a reader
func (p *JSONParser) Parse(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// MediaType returns the JSON media type
func (p *JSONParser) MediaType() string {
	return "application/json"
}
