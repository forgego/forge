package versioning

import "net/http"

// QueryParameterVersioning determines version from query parameter
// Example: /api/users/?version=1 -> version "1"
type QueryParameterVersioning struct {
	// VersionParam is the query parameter name (default: "version")
	VersionParam string
}

// NewQueryParameterVersioning creates a new query parameter versioning
func NewQueryParameterVersioning() *QueryParameterVersioning {
	return &QueryParameterVersioning{
		VersionParam: "version",
	}
}

// DetermineVersion extracts version from query parameter
func (v *QueryParameterVersioning) DetermineVersion(r *http.Request) (string, error) {
	version := r.URL.Query().Get(v.VersionParam)
	return version, nil
}

// Reverse reverses a URL with version
func (v *QueryParameterVersioning) Reverse(name string, version string, args ...interface{}) string {
	// Simplified - full implementation would use URL reversing
	return ""
}
