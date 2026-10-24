package api

import (
	"strings"

	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/schema"
)

// MaxFilterConditions is the maximum number of filter conditions permitted in a single request.
const MaxFilterConditions = 20

// CodeFilterTooComplex is the error code returned when filter conditions exceed MaxFilterConditions.
const CodeFilterTooComplex = "filter_too_complex"

// FilterableFields returns the schema names of fields safe for generated list
// filters. Non-serializable fields are excluded.
func FilterableFields(model interface{}) []string {
	s, ok := model.(schema.Schema)
	if !ok {
		return nil
	}
	fields := s.Fields()
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.Serialize && field.Name != "" {
			result = append(result, field.Name)
		}
	}
	return result
}

func listControlParameter(key string) bool {
	switch key {
	case "page", "page_size", "limit", "offset", "cursor", "ordering", "search", "format":
		return true
	default:
		return false
	}
}

func filterLookupAllowed(lookup string) bool {
	switch lookup {
	case "exact", "ne", "gt", "gte", "lt", "lte", "contains", "icontains", "startswith", "endswith", "in", "isnull":
		return true
	default:
		return false
	}
}

func filterNameAllowed(model interface{}, name string, allowed []string) bool {
	for _, candidate := range allowed {
		if strings.EqualFold(name, candidate) {
			return true
		}
		_, field, ok := resolveQueryField(model, candidate)
		if !ok {
			continue
		}
		for _, alias := range resolvedFieldNames(model, field) {
			if strings.EqualFold(name, alias) {
				return true
			}
		}
	}
	return false
}

func filterParameterError(key, message string) error {
	return exceptions.NewValidationError(map[string][]string{key: {"invalid filter: " + message}})
}
