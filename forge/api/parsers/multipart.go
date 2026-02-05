package parsers

import (
	"io"
	"mime/multipart"
)

// MultiPartParser parses multipart/form-data
type MultiPartParser struct {
	// MaxMemory is the maximum memory to use for parsing (default: 32MB)
	MaxMemory int64
}

// NewMultiPartParser creates a new multipart parser
func NewMultiPartParser() *MultiPartParser {
	return &MultiPartParser{
		MaxMemory: 32 << 20, // 32MB
	}
}

// Parse parses multipart data from a reader
func (p *MultiPartParser) Parse(r io.Reader, v interface{}) error {
	// Create multipart reader
	reader := multipart.NewReader(r, "")

	// Parse form
	form, err := reader.ReadForm(p.MaxMemory)
	if err != nil {
		return err
	}
	defer form.RemoveAll()

	// Convert to map[string]interface{}
	if m, ok := v.(*map[string]interface{}); ok {
		*m = make(map[string]interface{})
		for k, v := range form.Value {
			if len(v) == 1 {
				(*m)[k] = v[0]
			} else {
				(*m)[k] = v
			}
		}
		// Handle files
		for k, files := range form.File {
			if len(files) == 1 {
				(*m)[k] = files[0]
			} else {
				(*m)[k] = files
			}
		}
	}

	return nil
}

// MediaType returns the multipart media type
func (p *MultiPartParser) MediaType() string {
	return "multipart/form-data"
}
