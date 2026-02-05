package filter

import (
	"context"
	"fmt"
	"sync"

	"github.com/forgego/forge/orm"
)

var (
	customFilters   = make(map[string]*CustomFilterHandler)
	customFiltersMu sync.RWMutex
)

// CustomFilterHandler represents a custom filter handler
type CustomFilterHandler struct {
	ID          string
	Name        string
	Handler     func(value interface{}) (orm.Expression, error)
	AllowedRoles []string
	Cost        int
}

// RegisterCustom registers a custom filter handler
func RegisterCustom(name string, handler *CustomFilterHandler) error {
	if name == "" {
		return fmt.Errorf("custom filter name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("custom filter handler cannot be nil")
	}
	if handler.ID == "" {
		return fmt.Errorf("custom filter handler ID cannot be empty")
	}
	if handler.Handler == nil {
		return fmt.Errorf("custom filter handler function cannot be nil")
	}

	customFiltersMu.Lock()
	defer customFiltersMu.Unlock()

	if _, exists := customFilters[name]; exists {
		return fmt.Errorf("custom filter '%s' already registered", name)
	}

	customFilters[name] = handler
	return nil
}

// GetCustomFilter gets a custom filter handler by name
func GetCustomFilter(name string) (*CustomFilterHandler, bool) {
	customFiltersMu.RLock()
	defer customFiltersMu.RUnlock()

	handler, ok := customFilters[name]
	return handler, ok
}

// ListCustomFilters returns all registered custom filters
func ListCustomFilters() []string {
	customFiltersMu.RLock()
	defer customFiltersMu.RUnlock()

	names := make([]string, 0, len(customFilters))
	for name := range customFilters {
		names = append(names, name)
	}
	return names
}

// CustomFilter is a filter that uses a custom handler
type CustomFilter[T any] struct {
	*BaseFilter[T]
	handlerName string
	handler     *CustomFilterHandler
}

// NewCustomFilter creates a new custom filter
func NewCustomFilter[T any](fieldPath, handlerName string) (*CustomFilter[T], error) {
	handler, ok := GetCustomFilter(handlerName)
	if !ok {
		return nil, fmt.Errorf("custom filter handler '%s' not found", handlerName)
	}

	return &CustomFilter[T]{
		BaseFilter:  NewBaseFilter[T](fieldPath, "custom"),
		handlerName: handlerName,
		handler:     handler,
	}, nil
}

// Apply applies the custom filter
func (f *CustomFilter[T]) Apply(ctx context.Context, qs orm.QuerySet[T], value interface{}) (orm.QuerySet[T], error) {
	if value == nil {
		return qs, nil
	}

	expr, err := f.handler.Handler(value)
	if err != nil {
		return nil, fmt.Errorf("custom filter handler error: %w", err)
	}

	return qs.Filter(expr), nil
}

// ToExpression converts to ORM expression using the custom handler
func (f *CustomFilter[T]) ToExpression(fieldPath string, value interface{}) (orm.Expression, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot create expression for nil value")
	}

	return f.handler.Handler(value)
}
