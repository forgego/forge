package admin

import (
	"fmt"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// DefaultSite is the default admin site
var DefaultSite = NewSite("default")

// Register registers a model with the default admin site
func Register[T any](config *core.Config[T]) (*core.Admin[T], error) {
	return RegisterWithSite(DefaultSite, config)
}

// RegisterWithSite registers a model with a specific admin site
func RegisterWithSite[T any](s *Site, config *core.Config[T]) (*core.Admin[T], error) {
	// Create schema instance from type T
	// T must implement schema.Schema
	var schemaInstance schema.Schema
	var ok bool

	// Try pointer first (common for models)
	schemaInstance, ok = any(new(T)).(schema.Schema)
	if !ok {
		// Try value
		var zero T
		if s, ok := any(zero).(schema.Schema); ok {
			schemaInstance = s
		} else {
			return nil, fmt.Errorf("type %T does not implement schema.Schema", zero)
		}
	}

	// Get table name from schema for manager
	tableName := schemaInstance.Meta().TableName

	manager, err := orm.NewManager[T](tableName)
	if err != nil {
		return nil, err
	}

	// Set DB if available on site
	if s.db != nil {
		manager.SetDB(s.db)
	}

	// Create admin instance
	admin, err := core.NewAdmin(schemaInstance, manager, config)
	if err != nil {
		return nil, err
	}

	// Register with site
	if err := s.registry.Register(admin); err != nil {
		return nil, err
	}

	return admin, nil
}

// NewConfig creates a new configuration with defaults
func NewConfig[T any]() *core.Config[T] {
	return &core.Config[T]{}
}

// Alias core types for easier access
type Config[T any] = core.Config[T]
type Action[T any] = core.Action[T]
type Filter[T any] = core.Filter[T]
type Plugin = core.Plugin
type Field = core.Field
type Method = core.Method
type Fieldset[T any] = core.Fieldset[T]

// NewFieldset creates a new fieldset for admin forms.
func NewFieldset[T any](name string, fields ...string) core.Fieldset[T] {
	return core.NewFieldset[T](name, fields...)
}

// Computed creates a safe reference to a method or computed field
var Computed = core.Computed
