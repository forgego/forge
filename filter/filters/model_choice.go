package filters

import (
	"context"
	"fmt"

	"github.com/forgego/forge/filter"
	"github.com/forgego/forge/orm"
)

// ModelChoiceFilter filters foreign key relationships
type ModelChoiceFilter[T any, TRelated any] struct {
	*filter.BaseFilter[T]
	relatedModel string
}

// NewModelChoiceFilter creates a new model choice filter
func NewModelChoiceFilter[T any, TRelated any](fieldPath, relatedModel string) *ModelChoiceFilter[T, TRelated] {
	return &ModelChoiceFilter[T, TRelated]{
		BaseFilter:   filter.NewBaseFilter[T](fieldPath, "exact"),
		relatedModel: relatedModel,
	}
}

// Parse parses a model ID
func (f *ModelChoiceFilter[T, TRelated]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	// Parse as integer ID
	var id int64
	_, err := fmt.Sscanf(value, "%d", &id)
	if err != nil {
		return nil, fmt.Errorf("invalid model ID: %w", err)
	}

	return id, nil
}

// Apply applies the filter
func (f *ModelChoiceFilter[T, TRelated]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
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
func (f *ModelChoiceFilter[T, TRelated]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "exact", value), nil
}

// ToExpression converts to ORM expression
func (f *ModelChoiceFilter[T, TRelated]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	// Model choice filters use exact match on foreign key field
	// The value should be the ID of the related model
	idValue, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("model choice filter value must be an int64 (model ID), got %T", value)
	}

	// Create field expression for int64 (foreign key ID)
	fieldExpr := orm.NewField[int64](fieldPath, "")

	return fieldExpr.Eq(idValue), nil
}

// GetWidget returns the widget
func (f *ModelChoiceFilter[T, TRelated]) GetWidget() filter.Widget {
	return &ModelChoiceWidget{relatedModel: f.relatedModel}
}

// GetOptions returns model choices
func (f *ModelChoiceFilter[T, TRelated]) GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]filter.FilterOption, error) {
	// This would query the related model and return options
	// For now, return empty
	return nil, nil
}

// ValidateValue validates the value
func (f *ModelChoiceFilter[T, TRelated]) ValidateValue(value interface{}) error {
	if value == nil {
		return nil
	}

	_, ok := value.(int64)
	if !ok {
		return fmt.Errorf("model choice filter value must be an int64 (model ID), got %T", value)
	}

	return nil
}

// ModelChoiceWidget is a widget for model choice filters
type ModelChoiceWidget struct {
	relatedModel string
}

// Type returns the widget type
func (w *ModelChoiceWidget) Type() string {
	return "select"
}

// Render renders the widget
func (w *ModelChoiceWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	html := fmt.Sprintf(`<select name="%s"`, name)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`
	html += `<option value="">All</option>`
	// Options would be populated from related model
	html += `</select>`
	return html, nil
}

// Parse parses the widget value
func (w *ModelChoiceWidget) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	var id int64
	_, err := fmt.Sscanf(value, "%d", &id)
	if err != nil {
		return nil, fmt.Errorf("invalid model ID: %w", err)
	}

	return id, nil
}

// MultipleModelChoiceFilter filters multiple foreign key relationships
type MultipleModelChoiceFilter[T any, TRelated any] struct {
	*ModelChoiceFilter[T, TRelated]
}

// NewMultipleModelChoiceFilter creates a new multiple model choice filter
func NewMultipleModelChoiceFilter[T any, TRelated any](fieldPath, relatedModel string) *MultipleModelChoiceFilter[T, TRelated] {
	return &MultipleModelChoiceFilter[T, TRelated]{
		ModelChoiceFilter: NewModelChoiceFilter[T, TRelated](fieldPath, relatedModel),
	}
}

// Parse parses comma-separated model IDs
func (f *MultipleModelChoiceFilter[T, TRelated]) Parse(value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}

	parts := splitCommaSeparated(value)
	ids := make([]int64, 0, len(parts))

	for _, part := range parts {
		var id int64
		_, err := fmt.Sscanf(part, "%d", &id)
		if err != nil {
			return nil, fmt.Errorf("invalid model ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// ToAST converts to AST with "in" lookup
func (f *MultipleModelChoiceFilter[T, TRelated]) ToAST(fieldPath string, value interface{}) (*filter.FilterNode, error) {
	if value == nil {
		return nil, nil
	}

	return filter.NewFieldNode(fieldPath, "in", value), nil
}
