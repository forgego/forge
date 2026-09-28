package filters

import (
	"context"
	"fmt"
	"strconv"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// NumberFilter filters numeric fields
type NumberFilter[T any] struct {
	*filter.BaseFilter[T]
	lookups []string
}

// NewNumberFilter creates a new number filter
func NewNumberFilter[T any](fieldPath string) *NumberFilter[T] {
	return &NumberFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "exact"),
		lookups:    []string{"exact", "gt", "gte", "lt", "lte", "range", "in"},
	}
}

// Lookup sets allowed lookups
func (f *NumberFilter[T]) Lookup(lookups ...string) *NumberFilter[T] {
	f.lookups = lookups
	return f
}

// Range sets the filter to use range lookup
func (f *NumberFilter[T]) Range() *NumberFilter[T] {
	f.lookups = []string{"range"}
	return f
}

// GreaterThan sets the filter to use greater than lookup
func (f *NumberFilter[T]) GreaterThan() *NumberFilter[T] {
	f.lookups = []string{"gt"}
	return f
}

// LessThan sets the filter to use less than lookup
func (f *NumberFilter[T]) LessThan() *NumberFilter[T] {
	f.lookups = []string{"lt"}
	return f
}

// Parse parses a query parameter value
func (f *NumberFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Try to parse as float64
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number: %w", err)
	}

	return num, nil
}

// Apply applies the filter to a queryset
func (f *NumberFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
	if value == nil {
		return qs, nil
	}

	expr, err := f.ToExpression(f.GetFieldPath(), value)
	if err != nil {
		return nil, err
	}

	return qs.Filter(expr), nil
}

// ToAST converts the filter value to an AST node
func (f *NumberFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, f.GetLookup(), value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *NumberFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	lookup := f.GetLookup()

	// Try to convert value to float64 for numeric operations
	var numValue float64
	switch v := value.(type) {
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	default:
		return nil, fmt.Errorf("number filter value must be numeric, got %T", value)
	}

	// Create field expression for float64
	fieldExpr := orm.NewField[float64](fieldPath, "")

	// Convert lookup to operator and create comparison
	switch lookup {
	case "exact":
		return fieldExpr.Eq(numValue), nil
	case "gt":
		return fieldExpr.Gt(numValue), nil
	case "gte":
		return fieldExpr.Gte(numValue), nil
	case "lt":
		return fieldExpr.Lt(numValue), nil
	case "lte":
		return fieldExpr.Lte(numValue), nil
	case "range":
		// Range expects [min, max] slice
		if rangeValues, ok := value.([]float64); ok && len(rangeValues) == 2 {
			return fieldExpr.Range(rangeValues[0], rangeValues[1]), nil
		}
		return nil, fmt.Errorf("range lookup requires []float64 with 2 values")
	case "in":
		// Convert to []float64
		if numSlice, ok := value.([]float64); ok {
			return fieldExpr.In(numSlice...), nil
		}
		return nil, fmt.Errorf("in lookup requires []float64 value")
	default:
		return nil, fmt.Errorf("unsupported lookup for number filter: %s", lookup)
	}
}

// GetWidget returns the widget for this filter
func (f *NumberFilter[T]) GetWidget() filter.Widget {
	return &filter.DefaultWidget{}
}

// GetOptions returns filter options (not applicable for number filters)
func (f *NumberFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	return nil, nil
}

// ValidateValue validates a filter value
func (f *NumberFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	switch value.(type) {
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32, float64:
		return nil
	default:
		return fmt.Errorf("number filter value must be a number, got %T", value)
	}
}
