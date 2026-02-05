package filters

import (
	"net/http"
	"reflect"
)

// SearchFilter provides full-text search across multiple fields
type SearchFilter struct {
	// SearchFields are the fields to search in
	SearchFields []string
	// SearchParam is the query parameter name (default: "search")
	SearchParam string
}

// NewSearchFilter creates a new search filter
func NewSearchFilter(searchFields []string) *SearchFilter {
	return &SearchFilter{
		SearchFields: searchFields,
		SearchParam:  "search",
	}
}

// FilterQueryset filters a queryset based on search query
func (f *SearchFilter) FilterQueryset(r *http.Request, queryset interface{}) interface{} {
	searchQuery := r.URL.Query().Get(f.SearchParam)
	if searchQuery == "" || len(f.SearchFields) == 0 {
		return queryset
	}

	// Apply search using reflection
	qsValue := reflect.ValueOf(queryset)
	if !qsValue.IsValid() {
		return queryset
	}

	// Try to call Filter method with search conditions
	// This is a simplified implementation - full version would use QueryExpr
	searchMethod := qsValue.MethodByName("Search")
	if searchMethod.IsValid() {
		results := searchMethod.Call([]reflect.Value{
			reflect.ValueOf(searchQuery),
			reflect.ValueOf(f.SearchFields),
		})
		if len(results) > 0 {
			return results[0].Interface()
		}
	}

	return queryset
}

// GetSchema returns the filter schema
func (f *SearchFilter) GetSchema(r *http.Request, view interface{}) map[string]interface{} {
	return map[string]interface{}{
		f.SearchParam: map[string]interface{}{
			"type":        "string",
			"description": "Search query",
		},
	}
}
