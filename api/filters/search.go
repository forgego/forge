package filters

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/forgego/forge/orm"
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

	searchQueries := f.buildSearchQueries(queryset, searchQuery)
	if len(searchQueries) == 0 {
		return queryset
	}

	return f.applySearchFilter(queryset, orm.Or(searchQueries...))
}

func (f *SearchFilter) buildSearchQueries(queryset interface{}, q string) []orm.Expression {
	queries := make([]orm.Expression, 0, len(f.SearchFields))
	for _, field := range f.SearchFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		field, visible := visibleDatabaseField(queryset, field)
		if !visible {
			continue
		}
		queries = append(queries, orm.F(field).IContains(q))
	}
	return queries
}

func (f *SearchFilter) applySearchFilter(queryset interface{}, expr orm.Expression) interface{} {
	qsValue := reflect.ValueOf(queryset)
	if !qsValue.IsValid() {
		return queryset
	}

	filterMethod := qsValue.MethodByName("Filter")
	if !filterMethod.IsValid() {
		return queryset
	}

	results := filterMethod.Call([]reflect.Value{
		reflect.ValueOf(expr),
	})
	if len(results) > 0 {
		return results[0].Interface()
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
