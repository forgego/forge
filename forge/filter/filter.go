package filter

import (
	"context"
	"fmt"

	"github.com/forgego/forge/orm"
)

// Filter is the base interface for all filter types
type Filter[T any] interface {
	// Parse parses a query parameter value into the filter's internal format
	Parse(value string) (interface{}, error)

	// Apply applies the filter to a queryset
	Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error)

	// ToAST converts the filter value to an AST node
	ToAST(fieldPath string, value interface{}) (*FilterNode, error)

	// ToExpression converts the filter value to an ORM expression
	ToExpression(fieldPath string, value interface{}) (orm.Expression, error)

	// GetWidget returns the admin UI widget for this filter
	GetWidget() Widget

	// GetOptions returns filter options for UI (for choice filters, etc.)
	GetOptions(ctx context.Context, qs orm.QuerySet[T]) ([]FilterOption, error)

	// GetFieldPath returns the field path this filter operates on
	GetFieldPath() string

	// GetLookup returns the lookup type this filter uses
	GetLookup() string

	// ValidateValue validates a filter value
	ValidateValue(value interface{}) error

	// Label returns the human-readable label for this filter
	Label() string
}

// BaseFilter provides common functionality for all filters
type BaseFilter[T any] struct {
	fieldPath string
	lookup    string
	label     string
	helpText  string
	required  bool
}

// NewBaseFilter creates a new base filter
func NewBaseFilter[T any](fieldPath, lookup string) *BaseFilter[T] {
	return &BaseFilter[T]{
		fieldPath: fieldPath,
		lookup:    lookup,
	}
}

// GetFieldPath returns the field path
func (f *BaseFilter[T]) GetFieldPath() string {
	return f.fieldPath
}

// GetLookup returns the lookup type
func (f *BaseFilter[T]) GetLookup() string {
	return f.lookup
}

// SetLabel sets the filter label
func (f *BaseFilter[T]) SetLabel(label string) *BaseFilter[T] {
	f.label = label
	return f
}

// SetHelpText sets the help text
func (f *BaseFilter[T]) SetHelpText(text string) *BaseFilter[T] {
	f.helpText = text
	return f
}

// SetRequired sets whether the filter is required
func (f *BaseFilter[T]) SetRequired(required bool) *BaseFilter[T] {
	f.required = required
	return f
}

// Label returns the filter label
func (f *BaseFilter[T]) Label() string {
	if f.label != "" {
		return f.label
	}
	return f.fieldPath
}

// FilterOption represents an option for choice filters
type FilterOption struct {
	Label string
	Value interface{}
	Count int64
}

// Widget is the interface for admin UI widgets
type Widget interface {
	Type() string
	Render(name string, value interface{}, attrs map[string]string) (string, error)
	Parse(value string) (interface{}, error)
}

// DefaultWidget is a basic text input widget
type DefaultWidget struct{}

// Type returns the widget type
func (w *DefaultWidget) Type() string {
	return "text"
}

// Render renders the widget HTML
func (w *DefaultWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	valueStr := ""
	if value != nil {
		valueStr = fmt.Sprintf("%v", value)
	}

	html := fmt.Sprintf(`<input type="text" name="%s" value="%s"`, name, valueStr)
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	html += ` class="form-control">`

	return html, nil
}

// Parse parses the widget value
func (w *DefaultWidget) Parse(value string) (interface{}, error) {
	return value, nil
}

// FilterError represents a filter-specific error
type FilterError struct {
	Field   string
	Lookup  string
	Value   interface{}
	Message string
	Err     error
}

func (e *FilterError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("filter error on %s__%s: %s", e.Field, e.Lookup, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("filter error on %s__%s: %v", e.Field, e.Lookup, e.Err)
	}
	return fmt.Sprintf("filter error on %s__%s", e.Field, e.Lookup)
}

func (e *FilterError) Unwrap() error {
	return e.Err
}

// NewFilterError creates a new filter error
func NewFilterError(field, lookup string, value interface{}, message string, err error) *FilterError {
	return &FilterError{
		Field:   field,
		Lookup:  lookup,
		Value:   value,
		Message: message,
		Err:     err,
	}
}
