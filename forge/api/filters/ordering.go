package filters

import (
	"net/http"
	"reflect"
	"strings"
)

// OrderingFilter provides ordering functionality
type OrderingFilter struct {
	// OrderingFields are the allowed fields for ordering
	OrderingFields []string
	// OrderingParam is the query parameter name (default: "ordering")
	OrderingParam string
}

// NewOrderingFilter creates a new ordering filter
func NewOrderingFilter(orderingFields []string) *OrderingFilter {
	return &OrderingFilter{
		OrderingFields: orderingFields,
		OrderingParam:  "ordering",
	}
}

// FilterQueryset applies ordering to a queryset
func (f *OrderingFilter) FilterQueryset(r *http.Request, queryset interface{}) interface{} {
	orderingParam := r.URL.Query().Get(f.OrderingParam)
	if orderingParam == "" {
		return queryset
	}

	// Parse ordering fields (comma-separated, - prefix for descending)
	fields := strings.Split(orderingParam, ",")
	var ordering []string

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		// Validate field is allowed
		if len(f.OrderingFields) > 0 {
			fieldName := strings.TrimPrefix(field, "-")
			allowed := false
			for _, allowedField := range f.OrderingFields {
				if allowedField == fieldName {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		ordering = append(ordering, field)
	}

	if len(ordering) == 0 {
		return queryset
	}

	// Apply ordering using reflection
	qsValue := reflect.ValueOf(queryset)
	if !qsValue.IsValid() {
		return queryset
	}

	orderByMethod := qsValue.MethodByName("OrderBy")
	if orderByMethod.IsValid() {
		results := orderByMethod.Call([]reflect.Value{
			reflect.ValueOf(ordering),
		})
		if len(results) > 0 {
			return results[0].Interface()
		}
	}

	return queryset
}

// GetSchema returns the filter schema
func (f *OrderingFilter) GetSchema(r *http.Request, view interface{}) map[string]interface{} {
	return map[string]interface{}{
		f.OrderingParam: map[string]interface{}{
			"type":        "string",
			"description": "Ordering fields (comma-separated, - prefix for descending)",
		},
	}
}
