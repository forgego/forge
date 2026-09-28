package versioning

import "net/http"

// Version is the interface for API versioning schemes
type Version interface {
	// DetermineVersion determines the API version from the request
	DetermineVersion(r *http.Request) (string, error)
	// Reverse reverses a URL with version
	Reverse(name string, version string, args ...interface{}) string
}

// VersioningList is a list of versioning schemes
type VersioningList []Version

// DetermineVersion determines version using the first scheme that succeeds
func (vl VersioningList) DetermineVersion(r *http.Request) (string, error) {
	for _, v := range vl {
		version, err := v.DetermineVersion(r)
		if err == nil && version != "" {
			return version, nil
		}
	}
	return "", nil
}
