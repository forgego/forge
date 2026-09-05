package filters

import (
	"context"
	"fmt"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// LookupFilter allows dynamic lookup selection
type LookupFilter[T any] struct {
	*filter.BaseFilter[T]
	allowedLookups []string
}

// NewLookupFilter creates a new lookup filter
func NewLookupFilter[T any](fieldPath string, allowedLookups []string) *LookupFilter[T] {
	if len(allowedLookups) == 0 {
		allowedLookups = []string{"exact", "contains", "gt", "gte", "lt", "lte", "in", "isnull"}
	}

	return &LookupFilter[T]{
		BaseFilter:     filter.NewBaseFilter[T](fieldPath, "exact"),
		allowedLookups: allowedLookups,
	}
}

// Parse parses a lookup and value (format: "lookup:value" or separate params)
func (f *LookupFilter[T]) Parse(value string) (interface{}, error) {
	// This would parse lookup:value format
	// For now, return the value as-is
	return value, nil
}

// Apply applies the filter with the specified lookup
func (f *LookupFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
	if value == nil {
		return qs, nil
	}

	expr, err := f.ToExpression(f.GetFieldPath(), value)
	if err != nil {
		return nil, err
	}

	return qs.Filter(expr), nil
}

// ToAST converts to AST node
func (f *LookupFilter[T]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, f.GetLookup(), value), nil
}

// ToExpression converts to ORM expression
func (f *LookupFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	lookup := f.GetLookup()

	// Validate lookup is allowed
	allowed := false
	for _, l := range f.allowedLookups {
		if l == lookup {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("lookup '%s' not allowed for field '%s'", lookup, fieldPath)
	}

	// Create expression based on lookup type
	// Use the expression converter to handle different lookup types
	// For now, create a basic comparison expression
	fieldExpr := orm.NewField[interface{}](fieldPath, "")

	// Map lookup to operator
	var op orm.Operator
	switch lookup {
	case "exact":
		op = orm.OpEquals
	case "contains":
		op = orm.OpContains
	case "gt":
		op = orm.OpGreater
	case "gte":
		op = orm.OpGreaterOrEqual
	case "lt":
		op = orm.OpLess
	case "lte":
		op = orm.OpLessOrEqual
	case "in":
		op = orm.OpIn
	case "isnull":
		op = orm.OpIsNull
	default:
		return nil, fmt.Errorf("unsupported lookup: %s", lookup)
	}

	return orm.ComparisonExpression[interface{}]{
		Field: fieldExpr,
		Op:    op,
		Value: value,
	}, nil
}

// GetWidget returns the widget
func (f *LookupFilter[T]) GetWidget() filter.Widget {
	return &LookupWidget{allowedLookups: f.allowedLookups}
}

// GetOptions returns lookup options
func (f *LookupFilter[T]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	options := make([]filter.FilterOption, len(f.allowedLookups))
	for i, lookup := range f.allowedLookups {
		options[i] = filter.FilterOption{
			Label: lookup,
			Value: lookup,
		}
	}
	return options, nil
}

// ValidateValue validates the value
func (f *LookupFilter[T]) ValidateValue(value interface{}) error {
	return nil
}

// LookupWidget is a widget for lookup filters
type LookupWidget struct {
	allowedLookups []string
}

// Type returns the widget type
func (w *LookupWidget) Type() string {
	return "lookup"
}

// Render renders the widget (lookup dropdown + value input)
func (w *LookupWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	html := `<div class="lookup-filter">`
	html += `<select name="` + name + `_lookup" class="form-control">`
	for _, lookup := range w.allowedLookups {
		html += `<option value="` + lookup + `">` + lookup + `</option>`
	}
	html += `</select>`
	html += `<input type="text" name="` + name + `_value" class="form-control">`
	html += `</div>`
	return html, nil
}

// Parse parses the widget value
func (w *LookupWidget) Parse(value string) (interface{}, error) {
	return value, nil
}

