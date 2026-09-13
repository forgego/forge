package versioning

import (
	"net/http"
	"strings"
)

// URLPathVersioning determines version from URL path
// Example: /api/v1/users/ -> version "v1"
type URLPathVersioning struct {
	// VersionPrefix is the prefix in the URL (default: "v")
	VersionPrefix string
}

// NewURLPathVersioning creates a new URL path versioning
func NewURLPathVersioning() *URLPathVersioning {
	return &URLPathVersioning{
		VersionPrefix: "v",
	}
}

// DetermineVersion extracts version from URL path
func (v *URLPathVersioning) DetermineVersion(r *http.Request) (string, error) {
	path := r.URL.Path

	// Look for /api/v1/ pattern
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, v.VersionPrefix) {
			// Extract version number
			version := strings.TrimPrefix(part, v.VersionPrefix)
			if version != "" {
				return version, nil
			}
		}
		// Also check for /v1/ pattern
		if i > 0 && strings.HasPrefix(part, v.VersionPrefix) {
			version := strings.TrimPrefix(part, v.VersionPrefix)
			if version != "" {
				return version, nil
			}
		}
	}

	return "", nil
}

// Reverse reverses a URL with version
func (v *URLPathVersioning) Reverse(name string, version string, args ...interface{}) string {
	// Simplified - full implementation would use URL reversing
	return ""
}
