package filters

import (
	"context"
	"fmt"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// CharFilter filters string fields
type CharFilter[T any] struct {
	*filter.BaseFilter[T]
	lookups []string // Allowed lookups
}

// NewCharFilter creates a new char filter
func NewCharFilter[T any](fieldPath string) *CharFilter[T] {
	return &CharFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "exact"),
		lookups:    []string{"exact", "iexact", "contains", "icontains", "startswith", "istartswith", "endswith", "iendswith", "in"},
	}
}

// Lookup sets allowed lookups for this filter
func (f *CharFilter[T]) Lookup(lookups ...string) *CharFilter[T] {
	f.lookups = lookups
	return f
}

// Contains sets the filter to use contains lookup
func (f *CharFilter[T]) Contains() *CharFilter[T] {
	f.lookups = []string{"contains"}
	return f
}

// IContains sets the filter to use case-insensitive contains lookup
func (f *CharFilter[T]) IContains() *CharFilter[T] {
	f.lookups = []string{"icontains"}
	return f
}

// Parse parses a query parameter value
func (f *CharFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}
	return value, nil
}

// Apply applies the filter to a queryset
func (f *CharFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
	if value == nil {
		return qs, nil
	}

	strValue, ok := value.(string)
	if !ok {
		return nil, filter.NewFilterError(f.GetFieldPath(), f.GetLookup(), value, "value must be a string", nil)
	}

	// Create expression based on lookup type
	expr, err := f.ToExpression(f.GetFieldPath(), strValue)
	if err != nil {
		return nil, err
	}

	return qs.Filter(expr), nil
}

// ToAST converts the filter value to an AST node
func (f *CharFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, f.GetLookup(), value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *CharFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	strValue, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("char filter value must be a string, got %T", value)
	}

	lookup := f.GetLookup()

	// Create field expression for string
	fieldExpr := orm.NewField[string](fieldPath, "")

	// Convert lookup to operator and create comparison
	switch lookup {
	case "exact":
		return fieldExpr.Eq(strValue), nil
	case "iexact":
		return fieldExpr.IExact(strValue), nil
	case "contains":
		return fieldExpr.Contains(strValue), nil
	case "icontains":
		return fieldExpr.IContains(strValue), nil
	case "startswith":
		return fieldExpr.StartsWith(strValue), nil
	case "istartswith":
		// Use IContains with pattern for case-insensitive startswith
		return fieldExpr.IContains(strValue), nil // Will be handled by dialect adapter
	case "endswith":
		return fieldExpr.EndsWith(strValue), nil
	case "iendswith":
		// Use IContains with pattern for case-insensitive endswith
		return fieldExpr.IContains(strValue), nil // Will be handled by dialect adapter
	case "in":
		// Convert string slice to []string
		if strSlice, ok := value.([]string); ok {
			return fieldExpr.In(strSlice...), nil
		}
		return nil, fmt.Errorf("in lookup requires []string value")
	default:
		return nil, fmt.Errorf("unsupported lookup for char filter: %s", lookup)
	}
}

// GetWidget returns the widget for this filter
func (f *CharFilter[T]) GetWidget() filter.Widget {
	return &filter.DefaultWidget{}
}

// GetOptions returns filter options (not applicable for char filters)
func (f *CharFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	return nil, nil
}

// ValidateValue validates a filter value
func (f *CharFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(string)
	if !ok {
		return fmt.Errorf("char filter value must be a string, got %T", value)
	}

	return nil
}

// StartsWith creates a filter that matches strings starting with the value
func (f *CharFilter[T]) StartsWith() *CharFilter[T] {
	f.lookups = []string{"startswith"}
	return f
}

// EndsWith creates a filter that matches strings ending with the value
func (f *CharFilter[T]) EndsWith() *CharFilter[T] {
	f.lookups = []string{"endswith"}
	return f
}

// Exact creates a filter that matches exact strings
func (f *CharFilter[T]) Exact() *CharFilter[T] {
	f.lookups = []string{"exact"}
	return f
}

// IExact creates a filter that matches exact strings (case-insensitive)
func (f *CharFilter[T]) IExact() *CharFilter[T] {
	f.lookups = []string{"iexact"}
	return f
}
