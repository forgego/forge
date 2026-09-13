package filters

import (
	"context"
	"fmt"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// Choice represents a choice option
type Choice struct {
	Label string
	Value interface{}
}

// ChoiceFilter filters fields with predefined choices
type ChoiceFilter[T any] struct {
	*filter.BaseFilter[T]
	choices []Choice
}

// NewChoiceFilter creates a new choice filter
func NewChoiceFilter[T any](fieldPath string, choices []Choice) *ChoiceFilter[T] {
	return &ChoiceFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "exact"),
		choices:    choices,
	}
}

// Parse parses a query parameter value
func (f *ChoiceFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Validate that the value is in the choices
	for _, choice := range f.choices {
		if fmt.Sprintf("%v", choice.Value) == value {
			return choice.Value, nil
		}
	}

	return nil, fmt.Errorf("invalid choice value: %s", value)
}

// Apply applies the filter to a queryset
func (f *ChoiceFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
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
func (f *ChoiceFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "exact", value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *ChoiceFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	// Choice filters use exact match with interface{} type for flexibility
	fieldExpr := orm.NewField[interface{}](fieldPath, "")

	return orm.ComparisonExpression[interface{}]{
		Field: fieldExpr,
		Op:    orm.OpEquals,
		Value: value,
	}, nil
}

// GetWidget returns the widget for this filter
func (f *ChoiceFilter[T]) GetWidget() filter.Widget {
	return &ChoiceWidget{choices: f.choices}
}

// GetOptions returns filter options
func (f *ChoiceFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	options := make([]filter.FilterOption, len(f.choices))
	for i, choice := range f.choices {
		options[i] = filter.FilterOption{
			Label: choice.Label,
			Value: choice.Value,
		}
	}
	return options, nil
}

// ValidateValue validates a filter value
func (f *ChoiceFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	// Check if value is in choices
	for _, choice := range f.choices {
		if choice.Value == value {
			return nil
		}
	}

	return fmt.Errorf("value %v is not a valid choice", value)
}

// ChoiceWidget is a widget for choice filters
type ChoiceWidget struct {
	choices []Choice
}

// Type returns the widget type
func (w *ChoiceWidget) Type() string {
	return "select"
}

// Render renders the widget HTML
func (w *ChoiceWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	html := fmt.Sprintf(`<select name="%s"`, name)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`
	html += `<option value="">All</option>`

	for _, choice := range w.choices {
		html += fmt.Sprintf(`<option value="%v"`, choice.Value)
		if fmt.Sprintf("%v", choice.Value) == fmt.Sprintf("%v", value) {
			html += ` selected`
		}
		html += fmt.Sprintf(`>%s</option>`, choice.Label)
	}

	html += `</select>`
	return html, nil
}

// Parse parses the widget value
func (w *ChoiceWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	for _, choice := range w.choices {
		if fmt.Sprintf("%v", choice.Value) == value {
			return choice.Value, nil
		}
	}

	return nil, fmt.Errorf("invalid choice value: %s", value)
}

// MultipleChoiceFilter filters fields with multiple choice selection
type MultipleChoiceFilter[T any] struct {
	*ChoiceFilter[T]
}

// NewMultipleChoiceFilter creates a new multiple choice filter
func NewMultipleChoiceFilter[T any](fieldPath string, choices []Choice) *MultipleChoiceFilter[T] {
	return &MultipleChoiceFilter[T]{
		ChoiceFilter: NewChoiceFilter[T](fieldPath, choices),
	}
}

// Parse parses comma-separated values
func (f *MultipleChoiceFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	parts := splitCommaSeparated(value)
	results := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		val, err := f.ChoiceFilter.Parse(part)
		if err != nil {
			return nil, err
		}
		if val != nil {
			results = append(results, val)
		}
	}

	return results, nil
}

// ToAST converts to AST with "in" lookup
func (f *MultipleChoiceFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "in", value), nil
}

func splitCommaSeparated(s string) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range s {
		if char == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
