package filters

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// RangeValue represents a range filter value
type RangeValue struct {
	Start interface{}
	End   interface{}
}

// RangeFilter filters fields with range queries
type RangeFilter[T any] struct {
	*filter.BaseFilter[T]
}

// NewRangeFilter creates a new range filter
func NewRangeFilter[T any](fieldPath string) *RangeFilter[T] {
	return &RangeFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "range"),
	}
}

// Parse parses a range value (format: "min,max")
func (f *RangeFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Parse "min,max" format
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("range requires exactly 2 values separated by comma, got %d", len(parts))
	}

	minStr := strings.TrimSpace(parts[0])
	maxStr := strings.TrimSpace(parts[1])

	// Try to parse as float64 (works for both int and float)
	min, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid min value '%s': %w", minStr, err)
	}

	max, err := strconv.ParseFloat(maxStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid max value '%s': %w", maxStr, err)
	}

	return RangeValue{
		Start: min,
		End:   max,
	}, nil
}

// Apply applies the filter to a queryset
func (f *RangeFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
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
func (f *RangeFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "range", value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *RangeFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	rangeValue, ok := value.(RangeValue)
	if !ok {
		return nil, fmt.Errorf("range filter value must be RangeValue, got %T", value)
	}

	// Create field expression for float64
	fieldExpr := orm.NewField[float64](fieldPath, "")

	// Convert start/end to float64
	var min, max float64
	switch v := rangeValue.Start.(type) {
	case float64:
		min = v
	case float32:
		min = float64(v)
	case int:
		min = float64(v)
	case int64:
		min = float64(v)
	default:
		return nil, fmt.Errorf("range start value must be numeric, got %T", rangeValue.Start)
	}

	switch v := rangeValue.End.(type) {
	case float64:
		max = v
	case float32:
		max = float64(v)
	case int:
		max = float64(v)
	case int64:
		max = float64(v)
	default:
		return nil, fmt.Errorf("range end value must be numeric, got %T", rangeValue.End)
	}

	return fieldExpr.Range(min, max), nil
}

// GetWidget returns the widget for this filter
func (f *RangeFilter[T]) GetWidget() filter.Widget {
	return &RangeWidget{}
}

// GetOptions returns filter options (not applicable for range filters)
func (f *RangeFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	return nil, nil
}

// ValidateValue validates a filter value
func (f *RangeFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(RangeValue)
	if !ok {
		return fmt.Errorf("range filter value must be a RangeValue, got %T", value)
	}

	return nil
}

// RangeWidget is a widget for range filters
type RangeWidget struct{}

// Type returns the widget type
func (w *RangeWidget) Type() string {
	return "range"
}

// Render renders the widget HTML (two inputs for min and max)
func (w *RangeWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	startValue := ""
	endValue := ""

	if rv, ok := value.(RangeValue); ok {
		if rv.Start != nil {
			startValue = fmt.Sprintf("%v", rv.Start)
		}
		if rv.End != nil {
			endValue = fmt.Sprintf("%v", rv.End)
		}
	}

	html := `<div class="range-filter">`
	html += `<input type="text" name="` + name + `_start" value="` + template.HTMLEscapeString(startValue) + `" placeholder="Min"`
	for k, v := range attrs {
		html += ` ` + k + `="` + template.HTMLEscapeString(v) + `"`
	}
	html += ` class="form-control">`
	html += `<input type="text" name="` + name + `_end" value="` + template.HTMLEscapeString(endValue) + `" placeholder="Max"`
	for k, v := range attrs {
		html += ` ` + k + `="` + template.HTMLEscapeString(v) + `"`
	}
	html += ` class="form-control">`
	html += `</div>`

	return html, nil
}

// Parse parses the widget value (expects "start,end" format)
func (w *RangeWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Parse "start,end" format
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("range widget requires exactly 2 values separated by comma, got %d", len(parts))
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// Try to parse as float64
	start, err := strconv.ParseFloat(startStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start value '%s': %w", startStr, err)
	}

	end, err := strconv.ParseFloat(endStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid end value '%s': %w", endStr, err)
	}

	return RangeValue{
		Start: start,
		End:   end,
	}, nil
}

// DateRangeFilter filters date fields with range queries
type DateRangeFilter[T any] struct {
	*RangeFilter[T]
}

// NewDateRangeFilter creates a new date range filter
func NewDateRangeFilter[T any](fieldPath string) *DateRangeFilter[T] {
	return &DateRangeFilter[T]{
		RangeFilter: NewRangeFilter[T](fieldPath),
	}
}
