package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
)

// Admin represents a type-safe admin configuration for a model
// It fully integrates with schema, ORM, and filter systems
type Admin[T any] struct {
	// Core dependencies
	schema      schema.Schema
	manager     *orm.Manager[T]
	modelSchema *orm.ModelSchema
	config      *Config[T]

	// Discovered information (cached)
	metadata *Metadata

	// Internal
	name string
}

// NewAdmin creates a new Admin instance
func NewAdmin[T any](
	schemaInstance schema.Schema,
	manager *orm.Manager[T],
	config *Config[T],
) (*Admin[T], error) {
	// Get model schema from ORM
	modelSchema, err := orm.GetModelSchema[T]()
	if err != nil {
		return nil, fmt.Errorf("failed to get model schema: %w", err)
	}

	// Get model name
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	name := typ.Name()

	// Use table name from meta if available
	meta := schemaInstance.Meta()
	if meta.TableName != "" {
		name = meta.TableName
	}

	// Apply defaults to config
	if config == nil {
		config = &Config[T]{}
	}
	applyConfigDefaults(config, schemaInstance)

	admin := &Admin[T]{
		schema:      schemaInstance,
		manager:     manager,
		modelSchema: modelSchema,
		config:      config,
		name:        name,
	}

	return admin, nil
}

// SetDB sets the database connection for the admin manager
func (a *Admin[T]) SetDB(database *db.DB) {
	if a.manager != nil {
		a.manager.SetDB(database)
	}
}

// ModelName returns the name of the model
func (a *Admin[T]) ModelName() string {
	return a.name
}

// ModelType returns the model type
func (a *Admin[T]) ModelType() reflect.Type {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

// Schema returns the schema instance
func (a *Admin[T]) Schema() schema.Schema {
	return a.schema
}

// Manager returns the ORM manager
func (a *Admin[T]) Manager() *orm.Manager[T] {
	return a.manager
}

// ModelSchema returns the ORM model schema
func (a *Admin[T]) ModelSchema() *orm.ModelSchema {
	return a.modelSchema
}

// Config returns the configuration
func (a *Admin[T]) Config() *Config[T] {
	return a.config
}

// Interface implementation for type-agnostic access

func (a *Admin[T]) ManagerInterface() interface{} {
	return a.manager
}

func (a *Admin[T]) ConfigInterface() interface{} {
	return a.config
}

// GetMetadata returns the metadata for this admin
// This is called by API handlers to send to frontend.
// Schema-derived parts are cached; per-request Permissions are computed
// fresh on a copy so concurrent users never see each other's permissions.
func (a *Admin[T]) GetMetadata(ctx context.Context, user interface{}) (*Metadata, error) {
	// Return cached metadata if available
	if a.metadata != nil {
		// Copy: never mutate the shared cache per-user
		meta := *a.metadata
		meta.Permissions = a.getPermissionsMetadata(ctx, user)
		return &meta, nil
	}

	// Build metadata from schema
	meta, err := buildMetadata(a.schema, a.config, a.name)
	if err != nil {
		return nil, err
	}

	// Add permissions
	meta.Permissions = a.getPermissionsMetadata(ctx, user)

	// Add page type
	meta.PageType = a.PageType()

	// Cache it
	a.metadata = meta

	return meta, nil
}

// GetQueryset returns the base queryset for this admin
func (a *Admin[T]) GetQueryset(ctx context.Context) (orm.QuerySet[T], error) {
	// Get base queryset from manager
	// Use pointer to Q for empty filter (all records)
	qs, err := a.manager.Filter(&orm.Q{})
	if err != nil {
		return nil, err
	}

	// Apply custom queryset hook if provided
	if a.config.GetQueryset != nil {
		return a.config.GetQueryset(ctx, a, qs)
	}

	return qs, nil
}

// PageType returns the preferred layout for this admin
func (a *Admin[T]) PageType() string {
	if a.config.PageType != "" {
		return a.config.PageType
	}
	return "list" // Default
}

// applyConfigDefaults applies default configuration from schema
func applyConfigDefaults[T any](config *Config[T], s schema.Schema) {
	meta := s.Meta()

	// Set verbose names if not provided
	if config.VerboseName == "" {
		config.VerboseName = meta.VerboseName
	}
	if config.VerboseNamePlural == "" {
		config.VerboseNamePlural = meta.VerboseNamePlural
	}

	// Set pagination defaults
	if config.ListPerPage == 0 {
		config.ListPerPage = 25
	}
	if config.ListMaxShowAll == 0 {
		config.ListMaxShowAll = 100
	}

	// Set default history manager if not provided
	if config.HistoryManager == nil {
		config.HistoryManager = NewMemoryHistoryManager()
	}
}
