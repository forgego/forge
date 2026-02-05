package filter

import (
	"context"
	"reflect"
)

// FilterSetMetadata provides metadata about available filters in a FilterSet
type FilterSetMetadata struct {
	AvailableFilters  map[string]FilterInfo
	AvailableOperators map[string][]string
	FieldTypes        map[string]string
}

// FilterInfo contains information about a filter
type FilterInfo struct {
	Type        string
	Lookups     []string
	FieldPath   string
	Label       string
	HelpText    string
	Required    bool
	WidgetType  string
}

// GetMetadata returns filter metadata for a FilterSet
func (fs *FilterSet[T]) GetMetadata(ctx context.Context) (*FilterSetMetadata, error) {
	metadata := &FilterSetMetadata{
		AvailableFilters:  make(map[string]FilterInfo),
		AvailableOperators: make(map[string][]string),
		FieldTypes:         make(map[string]string),
	}

	// Build metadata from registered filters
	for name, f := range fs.filters {
		fieldPath := f.GetFieldPath()
		lookup := f.GetLookup()

		// Get field type from schema
		fieldTypeStr := "unknown"
		var fieldType reflect.Type
		if field, _, err := fs.schema.ResolvePath(fieldPath); err == nil {
			fieldType = field.Type
			fieldTypeStr = field.Type.String()
		}

		// Get allowed lookups for this field
		var allowedLookups []string
		if fieldType != nil {
			allowedLookups = getDefaultLookupsForField(fieldType)
		} else {
			// Try to get from schema, fallback to current lookup
			if lookups, err := fs.schema.GetAllowedLookups(fieldPath); err == nil {
				allowedLookups = lookups
			} else {
				allowedLookups = []string{lookup} // Fallback to current lookup
			}
		}

		info := FilterInfo{
			Type:       getFilterTypeName(f),
			Lookups:    allowedLookups,
			FieldPath:  fieldPath,
			Label:      name,
			WidgetType: f.GetWidget().Type(),
		}

		metadata.AvailableFilters[name] = info
		metadata.AvailableOperators[fieldPath] = allowedLookups
		metadata.FieldTypes[fieldPath] = fieldTypeStr
	}

	return metadata, nil
}

// getFilterTypeName returns a string representation of the filter type
func getFilterTypeName[T any](f Filter[T]) string {
	// Use reflection to get the type name
	typ := reflect.TypeOf(f)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.Name()
}

// getDefaultLookupsForField returns default allowed lookups for a field type
func getDefaultLookupsForField(fieldType reflect.Type) []string {
	switch fieldType.Kind() {
	case reflect.String:
		return []string{"exact", "iexact", "contains", "icontains", "startswith", "istartswith", "endswith", "iendswith", "in", "isnull"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return []string{"exact", "gt", "gte", "lt", "lte", "in", "range", "isnull"}
	case reflect.Bool:
		return []string{"exact", "isnull"}
	default:
		return []string{"exact", "in", "isnull"}
	}
}

// GetAvailableFilters returns a list of available filter names
func (fs *FilterSet[T]) GetAvailableFilters() []string {
	names := make([]string, 0, len(fs.filters))
	for name := range fs.filters {
		names = append(names, name)
	}
	return names
}
