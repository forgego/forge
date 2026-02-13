package api

import (
	"context"
	"net/http"

	"github.com/forgego/forge/filter"
)

// FilterSetIntegration integrates FilterSet with API ViewSets
type FilterSetIntegration[T any] struct {
	filterSet *filter.FilterSet[T]
	parser    *filter.Parser
}

// NewFilterSetIntegration creates a new filter set integration
func NewFilterSetIntegration[T any](fs *filter.FilterSet[T]) *FilterSetIntegration[T] {
	return &FilterSetIntegration[T]{
		filterSet: fs,
		parser:    filter.NewParser(filter.WithSecurity(fs.GetSecurityConfig())),
	}
}

// ApplyToViewSet applies filters to a ViewSet's List method
func (fsi *FilterSetIntegration[T]) ApplyToViewSet(ctx context.Context, r *http.Request, qs interface{}) (interface{}, error) {
	// Convert to AST
	ast, err := fsi.parser.ParseFilterNode(r, nil)
	if err != nil {
		return nil, err
	}

	// Apply to queryset
	if ast != nil {
		fsi.filterSet.SetAST(ast)
		qs, err = fsi.filterSet.ApplyAST(ctx, ast)
		if err != nil {
			return nil, err
		}
	}

	return qs, nil
}

// GetFilterMetadata returns filter metadata for API responses
func (fsi *FilterSetIntegration[T]) GetFilterMetadata(r *http.Request) map[string]interface{} {
	ctx := r.Context()
	metadata := make(map[string]interface{})

	// Get available filters from FilterSet
	filterMetadata, err := fsi.filterSet.GetMetadata(ctx)
	if err == nil && filterMetadata != nil {
		available := make(map[string]interface{})
		for name, info := range filterMetadata.AvailableFilters {
			available[name] = map[string]interface{}{
				"type":      info.Type,
				"lookups":   info.Lookups,
				"field":     info.FieldPath,
				"label":     info.Label,
				"widget":    info.WidgetType,
			}
		}
		metadata["available"] = available
		metadata["operators"] = filterMetadata.AvailableOperators
		metadata["field_types"] = filterMetadata.FieldTypes
	} else {
		metadata["available"] = make(map[string]interface{})
	}

	// Get applied filters
	applied := make(map[string]interface{})
	filters, _ := fsi.parser.ParseQueryParams(r, nil)
	for k, v := range filters {
		applied[k] = v
	}
	metadata["applied"] = applied

	return metadata
}

// EnhanceViewSetResponse enhances a ViewSet response with filter metadata
func (fsi *FilterSetIntegration[T]) EnhanceViewSetResponse(data map[string]interface{}, r *http.Request) map[string]interface{} {
	filterMetadata := fsi.GetFilterMetadata(r)
	data["filters"] = filterMetadata
	return data
}

