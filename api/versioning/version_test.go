package versioning

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLPathVersioning_DetermineVersion(t *testing.T) {
	versioning := NewURLPathVersioning()

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users/", "1"},
		{"/api/v2/products/", "2"},
		{"/v1/users/", "1"},
		{"/api/users/", ""},
		{"/users/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			version, err := versioning.DetermineVersion(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}

func TestQueryParameterVersioning_DetermineVersion(t *testing.T) {
	versioning := NewQueryParameterVersioning()

	req := httptest.NewRequest("GET", "/api/users/?version=2", nil)
	version, err := versioning.DetermineVersion(req)

	assert.NoError(t, err)
	assert.Equal(t, "2", version)
}

func TestQueryParameterVersioning_NoVersion(t *testing.T) {
	versioning := NewQueryParameterVersioning()

	req := httptest.NewRequest("GET", "/api/users/", nil)
	version, err := versioning.DetermineVersion(req)

	assert.NoError(t, err)
	assert.Equal(t, "", version)
}

func TestHeaderVersioning_DetermineVersion(t *testing.T) {
	versioning := NewHeaderVersioning()

	req := httptest.NewRequest("GET", "/api/users/", nil)
	req.Header.Set("Accept", "application/json; version=3")

	version, err := versioning.DetermineVersion(req)

	assert.NoError(t, err)
	assert.Equal(t, "3", version)
}

func TestHeaderVersioning_NoVersion(t *testing.T) {
	versioning := NewHeaderVersioning()

	req := httptest.NewRequest("GET", "/api/users/", nil)
	req.Header.Set("Accept", "application/json")

	version, err := versioning.DetermineVersion(req)

	assert.NoError(t, err)
	assert.Equal(t, "", version)
}

func TestVersioningList_DetermineVersion(t *testing.T) {
	list := VersioningList{
		NewURLPathVersioning(),
		NewQueryParameterVersioning(),
		NewHeaderVersioning(),
	}

	// Test URL path versioning (first in list)
	req := httptest.NewRequest("GET", "/api/v1/users/", nil)
	version, err := list.DetermineVersion(req)

	assert.NoError(t, err)
	assert.Equal(t, "1", version)

	// Test query parameter versioning (fallback)
	req2 := httptest.NewRequest("GET", "/api/users/?version=2", nil)
	version2, err := list.DetermineVersion(req2)

	assert.NoError(t, err)
	assert.Equal(t, "2", version2)

	// Test header versioning (fallback)
	req3 := httptest.NewRequest("GET", "/api/users/", nil)
	req3.Header.Set("Accept", "application/json; version=3")
	version3, err := list.DetermineVersion(req3)

	assert.NoError(t, err)
	assert.Equal(t, "3", version3)
}
