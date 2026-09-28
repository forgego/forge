package api

import (
	"net/http"
	"strings"

	forgehttp "github.com/forgego/forge/server"
)

// FilterSet provides filtering functionality for API viewsets
type FilterSet struct {
	Filters map[string]FilterFunc
}

// FilterFunc is a function that filters queryset based on a value
type FilterFunc func(value string) interface{} // Returns a filter expression

// NewFilterSet creates a new filter set
func NewFilterSet() *FilterSet {
	return &FilterSet{
		Filters: make(map[string]FilterFunc),
	}
}

// AddFilter adds a filter to the filter set
func (fs *FilterSet) AddFilter(name string, filterFunc FilterFunc) {
	fs.Filters[name] = filterFunc
}

// ApplyFilters applies filters from request query parameters
func (fs *FilterSet) ApplyFilters(r *http.Request) map[string]interface{} {
	filters := make(map[string]interface{})
	query := r.URL.Query()

	for param, values := range query {
		if filterFunc, ok := fs.Filters[param]; ok && len(values) > 0 {
			filters[param] = filterFunc(values[0])
		}
	}

	return filters
}

// GetOrdering extracts ordering parameters from request
func GetOrdering(r *http.Request, defaultOrdering []string) []string {
	orderingParam := forgehttp.GetQueryString(r, "ordering", "")
	if orderingParam == "" {
		return defaultOrdering
	}

	// Split by comma and trim
	fields := strings.Split(orderingParam, ",")
	ordering := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			ordering = append(ordering, field)
		}
	}

	return ordering
}

// GetSearchQuery extracts search query from request
func GetSearchQuery(r *http.Request) string {
	return forgehttp.GetQueryString(r, "search", "")
}

// GetFilterValue gets a filter value from request
func GetFilterValue(r *http.Request, key, defaultValue string) string {
	return forgehttp.GetQueryString(r, key, defaultValue)
}

// GetFilterInt gets a filter integer value from request
func GetFilterInt(r *http.Request, key string, defaultValue int) int {
	return forgehttp.GetQueryInt(r, key, defaultValue)
}

// GetFilterBool gets a filter boolean value from request
func GetFilterBool(r *http.Request, key string, defaultValue bool) bool {
	return forgehttp.GetQueryBool(r, key, defaultValue)
}

// CommonFilterFuncs provides common filter functions
var CommonFilterFuncs = struct {
	Exact      func(value string) interface{}
	IExact     func(value string) interface{}
	Contains   func(value string) interface{}
	IContains  func(value string) interface{}
	StartsWith func(value string) interface{}
	EndsWith   func(value string) interface{}
	In         func(value string) interface{} // Comma-separated values
	Range      func(value string) interface{} // Format: "min,max"
	IsNull     func(value string) interface{}
	IsNotNull  func(value string) interface{}
}{
	Exact: func(value string) interface{} {
		return map[string]interface{}{"exact": value}
	},
	IExact: func(value string) interface{} {
		return map[string]interface{}{"iexact": value}
	},
	Contains: func(value string) interface{} {
		return map[string]interface{}{"contains": value}
	},
	IContains: func(value string) interface{} {
		return map[string]interface{}{"icontains": value}
	},
	StartsWith: func(value string) interface{} {
		return map[string]interface{}{"startswith": value}
	},
	EndsWith: func(value string) interface{} {
		return map[string]interface{}{"endswith": value}
	},
	In: func(value string) interface{} {
		values := strings.Split(value, ",")
		return map[string]interface{}{"in": values}
	},
	Range: func(value string) interface{} {
		parts := strings.Split(value, ",")
		if len(parts) == 2 {
			return map[string]interface{}{"range": []string{parts[0], parts[1]}}
		}
		return nil
	},
	IsNull: func(value string) interface{} {
		if value == "true" || value == "1" {
			return map[string]interface{}{"isnull": true}
		}
		return nil
	},
	IsNotNull: func(value string) interface{} {
		if value == "true" || value == "1" {
			return map[string]interface{}{"isnull": false}
		}
		return nil
	},
}
