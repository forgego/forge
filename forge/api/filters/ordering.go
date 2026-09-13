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

	ordering := f.parseOrdering(orderingParam)
	if len(ordering) == 0 {
		return queryset
	}

	return f.applyOrdering(queryset, ordering)
}

func (f *OrderingFilter) parseOrdering(param string) []string {
	fields := strings.Split(param, ",")
	var ordering []string

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		if len(f.OrderingFields) > 0 {
			fieldName := strings.TrimPrefix(field, "-")
			if !f.isFieldAllowed(fieldName) {
				continue
			}
		}

		ordering = append(ordering, field)
	}
	return ordering
}

func (f *OrderingFilter) isFieldAllowed(field string) bool {
	for _, allowedField := range f.OrderingFields {
		if allowedField == field {
			return true
		}
	}
	return false
}

func (f *OrderingFilter) applyOrdering(queryset interface{}, ordering []string) interface{} {
	qsValue := reflect.ValueOf(queryset)
	if !qsValue.IsValid() {
		return queryset
	}

	orderByMethod := qsValue.MethodByName("OrderBy")
	if !orderByMethod.IsValid() {
		return queryset
	}

	mType := orderByMethod.Type()
	if mType.IsVariadic() {
		args := make([]reflect.Value, len(ordering))
		for i, field := range ordering {
			args[i] = reflect.ValueOf(field)
		}
		results := orderByMethod.Call(args)
		if len(results) > 0 {
			return results[0].Interface()
		}
	} else if mType.NumIn() == 1 && reflect.TypeOf(ordering).AssignableTo(mType.In(0)) {
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
