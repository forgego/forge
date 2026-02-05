package filters

import (
	"context"
	"fmt"
	"time"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// DateFilter filters date fields
type DateFilter[T any] struct {
	*filter.BaseFilter[T]
	lookups []string
}

// NewDateFilter creates a new date filter
func NewDateFilter[T any](fieldPath string) *DateFilter[T] {
	return &DateFilter[T]{
		BaseFilter: filter.NewBaseFilter[T](fieldPath, "exact"),
		lookups:    []string{"exact", "gt", "gte", "lt", "lte", "range", "year", "month", "day"},
	}
}

// Range sets the filter to use range lookup
func (f *DateFilter[T]) Range() *DateFilter[T] {
	f.lookups = []string{"range"}
	return f
}

// Parse parses a query parameter value
func (f *DateFilter[T]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Try parsing as date
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return nil, fmt.Errorf("invalid date format: %s", value)
}

// Apply applies the filter to a queryset
func (f *DateFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
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
func (f *DateFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, f.GetLookup(), value), nil
}

// ToExpression converts the filter value to an ORM expression
func (f *DateFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	lookup := f.GetLookup()

	// For date filters, we'll use interface{} type since time.Time handling is complex
	fieldExpr := orm.NewField[interface{}](fieldPath, "")

	switch lookup {
	case "exact":
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    orm.OpEquals,
			Value: value,
		}, nil
	case "gt":
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    orm.OpGreater,
			Value: value,
		}, nil
	case "gte":
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    orm.OpGreaterOrEqual,
			Value: value,
		}, nil
	case "lt":
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    orm.OpLess,
			Value: value,
		}, nil
	case "lte":
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    orm.OpLessOrEqual,
			Value: value,
		}, nil
	case "range":
		if rangeValues, ok := value.([]interface{}); ok && len(rangeValues) == 2 {
			return fieldExpr.Range(rangeValues[0], rangeValues[1]), nil
		}
		return nil, fmt.Errorf("range lookup requires []interface{} with 2 values")
	case "year", "month", "day":
		// These use extraction operators
		op := f.lookupToOperatorForDate(lookup)
		return orm.ComparisonExpression[interface{}]{
			Field: fieldExpr,
			Op:    op,
			Value: value,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported lookup for date filter: %s", lookup)
	}
}

// lookupToOperatorForDate converts date-specific lookups to operators
func (f *DateFilter[T]) lookupToOperatorForDate(lookup string) orm.Operator {
	switch lookup {
	case "exact":
		return orm.OpEquals
	case "gt":
		return orm.OpGreater
	case "gte":
		return orm.OpGreaterOrEqual
	case "lt":
		return orm.OpLess
	case "lte":
		return orm.OpLessOrEqual
	case "year":
		return orm.OpYear
	case "month":
		return orm.OpMonth
	case "day":
		return orm.OpDay
	default:
		return orm.OpEquals
	}
}

// GetWidget returns the widget for this filter
func (f *DateFilter[T]) GetWidget() filter.Widget {
	return &DateWidget{}
}

// GetOptions returns filter options (not applicable for date filters)
func (f *DateFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	return nil, nil
}

// ValidateValue validates a filter value
func (f *DateFilter[T]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("date filter value must be a time.Time, got %T", value)
	}

	return nil
}

// DateWidget is a widget for date filters
type DateWidget struct{}

// Type returns the widget type
func (w *DateWidget) Type() string {
	return "date"
}

// Render renders the widget HTML
func (w *DateWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	valueStr := ""
	if value != nil {
		if t, ok := value.(time.Time); ok {
			valueStr = t.Format("2006-01-02")
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}
	}

	html := fmt.Sprintf(`<input type="date" name="%s" value="%s"`, name, valueStr)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`

	return html, nil
}

// Parse parses the widget value
func (w *DateWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	return t, nil
}

// DateTimeFilter filters datetime fields
type DateTimeFilter[T any] struct {
	*DateFilter[T]
}

// NewDateTimeFilter creates a new datetime filter
func NewDateTimeFilter[T any](fieldPath string) *DateTimeFilter[T] {
	return &DateTimeFilter[T]{
		DateFilter: NewDateFilter[T](fieldPath),
	}
}

// DateTimeWidget is a widget for datetime filters
type DateTimeWidget struct{}

// Type returns the widget type
func (w *DateTimeWidget) Type() string {
	return "datetime-local"
}

// Render renders the widget HTML
func (w *DateTimeWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	valueStr := ""
	if value != nil {
		if t, ok := value.(time.Time); ok {
			valueStr = t.Format("2006-01-02T15:04")
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}
	}

	html := fmt.Sprintf(`<input type="datetime-local" name="%s" value="%s"`, name, valueStr)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`

	return html, nil
}

// Parse parses the widget value
func (w *DateTimeWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return nil, fmt.Errorf("invalid datetime format: %s", value)
}
