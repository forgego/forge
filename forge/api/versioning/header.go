package versioning

import (
	"net/http"
	"strings"
)

// HeaderVersioning determines version from Accept header
// Example: Accept: application/json; version=1
type HeaderVersioning struct {
	// VersionHeader is the header name (default: "Accept")
	VersionHeader string
}

// NewHeaderVersioning creates a new header versioning
func NewHeaderVersioning() *HeaderVersioning {
	return &HeaderVersioning{
		VersionHeader: "Accept",
	}
}

// DetermineVersion extracts version from header
func (v *HeaderVersioning) DetermineVersion(r *http.Request) (string, error) {
	header := r.Header.Get(v.VersionHeader)
	if header == "" {
		return "", nil
	}

	// Parse "application/json; version=1"
	parts := strings.Split(header, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "version=") {
			version := strings.TrimPrefix(part, "version=")
			return strings.TrimSpace(version), nil
		}
	}

	return "", nil
}

// Reverse reverses a URL with version
func (v *HeaderVersioning) Reverse(name string, version string, args ...interface{}) string {
	return ""
}
