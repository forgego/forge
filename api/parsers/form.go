package parsers

import (
	"io"
	"net/url"
)

// FormParser parses form data
type FormParser struct{}

// NewFormParser creates a new form parser
func NewFormParser() *FormParser {
	return &FormParser{}
}

// Parse parses form data from a reader
func (p *FormParser) Parse(r io.Reader, v interface{}) error {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Parse form data
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	// Convert to map[string]interface{}
	if m, ok := v.(*map[string]interface{}); ok {
		*m = make(map[string]interface{})
		for k, v := range values {
			if len(v) == 1 {
				(*m)[k] = v[0]
			} else {
				(*m)[k] = v
			}
		}
	}

	return nil
}

// MediaType returns the form media type
func (p *FormParser) MediaType() string {
	return "application/x-www-form-urlencoded"
}
