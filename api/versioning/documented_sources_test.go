package versioning

import (
	"net/http/httptest"
	"testing"
)

// All three documented request locations supply a version.
func TestVersioningAcceptsHeaderQueryAndPath(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/v3/products?version=2", nil)
	request.Header.Set("Accept", "application/json; version=1")
	for _, tc := range []struct {
		name   string
		scheme Version
		want   string
	}{
		{"header", NewHeaderVersioning(), "1"},
		{"query", NewQueryParameterVersioning(), "2"},
		{"path", NewURLPathVersioning(), "3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.scheme.DetermineVersion(request)
			if err != nil || got != tc.want {
				t.Fatalf("version=%q, err=%v", got, err)
			}
		})
	}
}
