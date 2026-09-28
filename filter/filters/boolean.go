package filters

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// BooleanFilter filters boolean fields
type BooleanFilter[T any] struct {
	*filter.BaseFilter[T]
}

// NewBooleanFilter creates a new boolean filter
func NewBooleanFilter[T any](fieldPath string) *BooleanFilter[T] {
	return &BooleanFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "exact"),
	}
}

// Parse parses a query parameter value
func (f *BooleanFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Parse boolean string
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "true" || value == "1" || value == "yes" || value == "on" {
		return true, nil
	}
	if value == "false" || value == "0" || value == "no" || value == "off" {
		return false, nil
	}

	// Try parsing as integer
	if num, err := strconv.ParseInt(value, 10, 64); err == nil {
		return num != 0, nil
	}

	return nil, fmt.Errorf("invalid boolean value: %s", value)
}

// Apply applies the filter to a queryset
func (f *BooleanFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
	if value == nil {
		return qs, nil
	}

	boolValue, ok := value.(bool)
	if !ok {
		return nil, filter.NewFilterError(f.GetFieldPath(), f.GetLookup(), value, "value must be a boolean", nil)
	}

	expr, err := f.ToExpression(f.GetFieldPath(), boolValue)
	if err != nil {
		return nil, err
	}

	return qs.Filter(expr), nil
}

// ToAST converts the filter value to an AST node
func (f *BooleanFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "exact", value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *BooleanFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	boolValue, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("boolean filter value must be a boolean, got %T", value)
	}

	// Create field expression for bool
	fieldExpr := orm.NewField[bool](fieldPath, "")

	// Boolean filters use exact match
	return fieldExpr.Eq(boolValue), nil
}

// GetWidget returns the widget for this filter
func (f *BooleanFilter[T]) GetWidget() filter.Widget {
	// Boolean filters typically use a select/dropdown widget
	return &BooleanWidget{}
}

// GetOptions returns filter options for boolean (Yes/No)
func (f *BooleanFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	return []filter.FilterOption{
		{Label: "Yes", Value: true},
		{Label: "No", Value: false},
	}, nil
}

// ValidateValue validates a filter value
func (f *BooleanFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("boolean filter value must be a boolean, got %T", value)
	}

	return nil
}

// BooleanWidget is a widget for boolean filters
type BooleanWidget struct{}

// Type returns the widget type
func (w *BooleanWidget) Type() string {
	return "select"
}

// Render renders the widget HTML
func (w *BooleanWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	html := fmt.Sprintf(`<select name="%s"`, name)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`
	html += `<option value="">All</option>`
	html += `<option value="true"`

	if value == true {
		html += ` selected`
	}
	html += `>Yes</option>`
	html += `<option value="false"`
	if value == false {
		html += ` selected`
	}
	html += `>No</option>`
	html += `</select>`

	return html, nil
}

// Parse parses the widget value
func (w *BooleanWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	value = strings.ToLower(strings.TrimSpace(value))
	if value == "true" || value == "1" || value == "yes" {
		return true, nil
	}
	if value == "false" || value == "0" || value == "no" {
		return false, nil
	}

	return nil, fmt.Errorf("invalid boolean value: %s", value)
}
