package parsers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
)

// ErrMissingBoundary is returned when multipart boundary cannot be determined.
var ErrMissingBoundary = errors.New("missing multipart boundary")

// MultiPartParser parses multipart/form-data
type MultiPartParser struct {
	// MaxMemory is the maximum memory to use for parsing (default: 32MB)
	MaxMemory int64
	// ContentType is an optional Content-Type header containing the boundary parameter
	ContentType string
	// Boundary is an optional explicit multipart boundary string
	Boundary string
}

// NewMultiPartParser creates a new multipart parser
func NewMultiPartParser() *MultiPartParser {
	return &MultiPartParser{
		MaxMemory: 32 << 20, // 32MB
	}
}

// Parse parses multipart data from a reader
func (p *MultiPartParser) Parse(r io.Reader, v interface{}) error {
	boundary, bodyReader, err := p.resolveBoundary(r)
	if err != nil {
		return err
	}

	reader := multipart.NewReader(bodyReader, boundary)
	form, err := reader.ReadForm(p.MaxMemory)
	if err != nil {
		return fmt.Errorf("failed to parse multipart form: %w", err)
	}
	defer form.RemoveAll()

	p.populateValues(form, v)
	return nil
}

func (p *MultiPartParser) resolveBoundary(r io.Reader) (string, io.Reader, error) {
	if p.Boundary != "" {
		return p.Boundary, r, nil
	}

	ct := p.ContentType
	if ct == "" {
		if cr, ok := r.(interface{ ContentType() string }); ok {
			ct = cr.ContentType()
		}
	}

	if ct != "" {
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			return "", nil, fmt.Errorf("invalid content type: %w", err)
		}
		boundary, ok := params["boundary"]
		if !ok || boundary == "" {
			return "", nil, fmt.Errorf("content type missing boundary: %w", ErrMissingBoundary)
		}
		return boundary, r, nil
	}

	return p.extractBoundaryFromStream(r)
}

func (p *MultiPartParser) extractBoundaryFromStream(r io.Reader) (string, io.Reader, error) {
	bufReader := bufio.NewReader(r)
	prefix, err := bufReader.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, fmt.Errorf("failed to read multipart stream: %w", err)
	}

	trimmed := strings.TrimRight(string(prefix), "\r\n")
	if !strings.HasPrefix(trimmed, "--") || len(trimmed) <= 2 {
		return "", nil, ErrMissingBoundary
	}

	boundary := strings.TrimPrefix(trimmed, "--")
	boundary = strings.TrimSuffix(boundary, "--")
	fullReader := io.MultiReader(bytes.NewReader(prefix), bufReader)
	return boundary, fullReader, nil
}

func (p *MultiPartParser) populateValues(form *multipart.Form, v interface{}) {
	m, ok := v.(*map[string]interface{})
	if !ok {
		return
	}

	res := make(map[string]interface{})
	for k, vals := range form.Value {
		if len(vals) == 1 {
			res[k] = vals[0]
		} else {
			res[k] = vals
		}
	}
	for k, files := range form.File {
		if len(files) == 1 {
			res[k] = files[0]
		} else {
			res[k] = files
		}
	}
	*m = res
}

// MediaType returns the multipart media type
func (p *MultiPartParser) MediaType() string {
	return "multipart/form-data"
}
