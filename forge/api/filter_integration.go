package api

import (
	"context"
	"net/http"
	"reflect"

	"github.com/forgego/forge/filter"
)

// ApplyFilterSetToViewSet applies FilterSet filters to a ViewSet's queryset
func ApplyFilterSetToViewSet[T any](ctx context.Context, r *http.Request, qs interface{}, fs *filter.FilterSet[T]) (interface{}, error) {
	if fs == nil || r == nil {
		return qs, nil
	}

	integration := NewFilterSetIntegration(fs)
	
	// Convert queryset to proper type if needed
	qsValue := reflect.ValueOf(qs)
	if !qsValue.IsValid() {
		return qs, nil
	}

	// Apply filters
	filteredQS, err := integration.ApplyToViewSet(ctx, r, qs)
	if err != nil {
		return nil, err
	}

	return filteredQS, nil
}

// EnhanceResponseWithFilters adds filter metadata to API response
func EnhanceResponseWithFilters[T any](data map[string]interface{}, r *http.Request, fs *filter.FilterSet[T]) map[string]interface{} {
	if fs == nil || r == nil {
		return data
	}

	integration := NewFilterSetIntegration(fs)
	return integration.EnhanceViewSetResponse(data, r)
}
