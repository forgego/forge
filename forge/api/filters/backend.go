package filters

import (
	"net/http"
)

// FilterBackend is the interface for filter backends
type FilterBackend interface {
	// FilterQueryset filters a queryset based on the request
	FilterQueryset(r *http.Request, queryset interface{}) interface{}
	// GetSchema returns the filter schema for documentation
	GetSchema(r *http.Request, view interface{}) map[string]interface{}
}

// FilterBackendList is a list of filter backends
type FilterBackendList []FilterBackend

// ApplyFilters applies all filter backends to a queryset
func (fbl FilterBackendList) ApplyFilters(r *http.Request, queryset interface{}) interface{} {
	result := queryset
	for _, backend := range fbl {
		result = backend.FilterQueryset(r, result)
	}
	return result
}
