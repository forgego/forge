package orm

import (
	"context"
	"fmt"
	"reflect"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/errors"
	"github.com/forgego/forge/schema"
)

// Manager provides type-safe CRUD operations
type Manager[T any] struct {
	tableName string
	schema    *ModelSchema
	db        *db.DB
}

// NewManager creates a new manager
func NewManager[T any](tableName string) (*Manager[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = schema.TableName
	}

	return &Manager[T]{
		tableName: tableName,
		schema:    schema,
	}, nil
}

// NewManagerWithDB creates a new manager with a database connection.
// Returns a ConfigurationError if the database connection is nil.
func NewManagerWithDB[T any](tableName string, db *db.DB) (*Manager[T], error) {
	if db == nil {
		return nil, errors.NewConfigurationError("database connection is required", "db")
	}

	manager, err := NewManager[T](tableName)
	if err != nil {
		return nil, err
	}

	manager.SetDB(db)
	return manager, nil
}

// SetDB sets the database connection
func (m *Manager[T]) SetDB(database *db.DB) {
	m.db = database
}

// FieldAccessor returns a field accessor for type-safe field operations.
func (m *Manager[T]) FieldAccessor() (*FieldAccessor[T], error) {
	return NewFieldAccessor[T]()
}

// GetFieldAccessor returns a field accessor for this model.
// Deprecated: Use FieldAccessor() instead.
func (m *Manager[T]) GetFieldAccessor() (*FieldAccessor[T], error) {
	return m.FieldAccessor()
}

// Filter returns a QuerySet for filtering
func (m *Manager[T]) Filter(expr Expression) (QuerySet[T], error) {
	qs, err := NewQuerySet[T](m.tableName)
	if err != nil {
		return nil, err
	}

	if m.db != nil {
		qs = qs.SetDB(m.db)
	}

	return qs.Filter(expr), nil
}

// Get retrieves a single model instance by its primary key ID.
func (m *Manager[T]) Get(ctx context.Context, id int64) (*T, error) {
	if m.db == nil {
		return nil, errors.NewConfigurationError("database connection not set", "db")
	}

	// Prefer schema primary key if available.
	idFieldName := m.schema.PrimaryKey
	if idFieldName == "" {
		// Try common ID variants.
		for _, candidate := range []string{"id", "ID", "Id"} {
			if m.schema.GetField(candidate) != nil {
				idFieldName = candidate
				break
			}
		}
	}
	if idFieldName == "" {
		return nil, fmt.Errorf("primary key field not found for %s", m.tableName)
	}

	qs, err := m.Filter(F(idFieldName).Eq(id))
	if err != nil {
		return nil, err
	}

	return qs.Get(ctx)
}

// All returns all model instances
func (m *Manager[T]) All(ctx context.Context) ([]*T, error) {
	qs, err := NewQuerySet[T](m.tableName)
	if err != nil {
		return nil, err
	}

	if m.db != nil {
		qs = qs.SetDB(m.db)
	}

	return qs.All(ctx)
}

// Count returns the number of model instances
func (m *Manager[T]) Count(ctx context.Context) (int64, error) {
	qs, err := NewQuerySet[T](m.tableName)
	if err != nil {
		return 0, err
	}

	if m.db != nil {
		qs = qs.SetDB(m.db)
	}

	return qs.Count(ctx)
}

// Create creates a new model instance
func (m *Manager[T]) Create(ctx context.Context, instance *T) error {
	if m.db == nil {
		return errors.NewConfigurationError("database connection not set", "db")
	}

	if err := m.runHooks(ctx, instance, "BeforeCreate"); err != nil {
		return err
	}
	if err := m.runHooks(ctx, instance, "BeforeSave"); err != nil {
		return err
	}
	if err := m.validate(instance); err != nil {
		return err
	}

	// Build and execute INSERT
	sql, args, _, err := BuildInsertSQL(instance, m.tableName)
	if err != nil {
		return fmt.Errorf("failed to build insert SQL: %w", err)
	}

	id, err := ExecuteInsert(ctx, m.db, sql, args)
	if err != nil {
		return err
	}

	m.setID(instance, id)

	if err := m.runHooks(ctx, instance, "AfterCreate"); err != nil {
		return err
	}
	return m.runHooks(ctx, instance, "AfterSave")
}

// BulkCreate creates multiple model instances efficiently using a single INSERT statement
func (m *Manager[T]) BulkCreate(ctx context.Context, instances []*T) error {
	if m.db == nil {
		return errors.NewConfigurationError("database connection not set", "db")
	}

	if len(instances) == 0 {
		return nil
	}

	// Pre-process all instances
	for _, instance := range instances {
		if err := m.runHooks(ctx, instance, "BeforeCreate"); err != nil {
			return err
		}
		if err := m.runHooks(ctx, instance, "BeforeSave"); err != nil {
			return err
		}
		if err := m.validate(instance); err != nil {
			return err
		}
	}

	// Convert []*T to []interface{}
	instancesInterface := make([]interface{}, len(instances))
	for i, instance := range instances {
		instancesInterface[i] = instance
	}

	// Build and execute bulk INSERT
	sql, args, _, err := BuildBulkInsertSQL(instancesInterface, m.tableName)
	if err != nil {
		return fmt.Errorf("failed to build bulk insert SQL: %w", err)
	}

	ids, err := ExecuteBulkInsert(ctx, m.db, sql, args)
	if err != nil {
		return err
	}

	// Set IDs
	for i, instance := range instances {
		if i < len(ids) {
			m.setID(instance, ids[i])
		}
	}

	// Post-process all instances
	for _, instance := range instances {
		if err := m.runHooks(ctx, instance, "AfterCreate"); err != nil {
			return err
		}
		if err := m.runHooks(ctx, instance, "AfterSave"); err != nil {
			return err
		}
	}

	return nil
}

// Update updates an existing model instance
func (m *Manager[T]) Update(ctx context.Context, instance *T) error {
	if m.db == nil {
		return errors.NewConfigurationError("database connection not set", "db")
	}

	id, err := m.getID(instance)
	if err != nil {
		return err
	}
	if id == 0 {
		return errors.NewInvalidInputError("id", "ID must be non-zero for update")
	}

	if err := m.runHooks(ctx, instance, "BeforeUpdate"); err != nil {
		return err
	}
	if err := m.runHooks(ctx, instance, "BeforeSave"); err != nil {
		return err
	}
	if err := m.validate(instance); err != nil {
		return err
	}

	pkColumn := m.primaryKeyColumn()
	sql, args, err := BuildUpdateSQL(instance, m.tableName, pkColumn)
	if err != nil {
		return fmt.Errorf("failed to build update SQL: %w", err)
	}

	rowsAffected, err := ExecuteUpdate(ctx, m.db, sql, args)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("model", id)
	}

	if err := m.runHooks(ctx, instance, "AfterUpdate"); err != nil {
		return err
	}
	return m.runHooks(ctx, instance, "AfterSave")
}

// Save saves a model (create or update)
func (m *Manager[T]) Save(ctx context.Context, instance *T) error {
	id, _ := m.getID(instance)
	if id != 0 {
		return m.Update(ctx, instance)
	}
	return m.Create(ctx, instance)
}

// Delete deletes a model instance
func (m *Manager[T]) Delete(ctx context.Context, instance *T) error {
	if m.db == nil {
		return errors.NewConfigurationError("database connection not set", "db")
	}

	id, err := m.getID(instance)
	if err != nil {
		return err
	}
	if id == 0 {
		return errors.NewInvalidInputError("id", "ID must be non-zero for delete")
	}

	if err := m.runHooks(ctx, instance, "BeforeDelete"); err != nil {
		return err
	}

	pkColumn := m.primaryKeyColumn()
	sql, args := BuildDeleteSQL(m.tableName, pkColumn, id)

	rowsAffected, err := ExecuteDelete(ctx, m.db, sql, args)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("model", id)
	}

	return m.runHooks(ctx, instance, "AfterDelete")
}

// UpdateFields updates specific fields type-safely
func (m *Manager[T]) UpdateFields(ctx context.Context, id int64, updates UpdateMap) error {
	// Validate all fields exist
	for fieldName := range updates {
		fieldInfo := m.schema.GetField(fieldName)
		if fieldInfo == nil {
			return fmt.Errorf("field %s not found", fieldName)
		}
	}

	// Resolve primary key field without relying on typed FieldFor, which can
	// panic when models don't expose the expected "ID" Go field name.
	idFieldName := m.schema.PrimaryKey
	if idFieldName == "" {
		for _, candidate := range []string{"id", "ID", "Id"} {
			if m.schema.GetField(candidate) != nil {
				idFieldName = candidate
				break
			}
		}
	}
	if idFieldName == "" {
		return fmt.Errorf("primary key field not found for %s", m.tableName)
	}

	qs, err := m.Filter(F(idFieldName).Eq(id))
	if err != nil {
		return err
	}

	ub, err := qs.UpdateBuilder()
	if err != nil {
		return err
	}

	for fieldName, value := range updates {
		ub.updates[fieldName] = value
	}

	_, err = ub.Execute(ctx)
	return err
}

// primaryKeyColumn returns the database column name for the primary key.
// It uses the schema's PrimaryKey if set, otherwise falls back to common
// ID field name variants.
func (m *Manager[T]) primaryKeyColumn() string {
	if m.schema.PrimaryKey != "" {
		return m.schema.PrimaryKey
	}
	for _, candidate := range []string{"id", "ID", "Id"} {
		if m.schema.GetField(candidate) != nil {
			return candidate
		}
	}
	return "id" // ultimate fallback
}

// Helper methods

func (m *Manager[T]) getID(instance *T) (int64, error) {
	if modelWithID, ok := any(instance).(ModelWithID); ok {
		return modelWithID.GetID(), nil
	}

	idValue, err := GetIDValue(instance, "id")
	if err != nil {
		return 0, fmt.Errorf("failed to get ID: %w", err)
	}

	if id, ok := idValue.(int64); ok {
		return id, nil
	}
	return 0, fmt.Errorf("ID must be int64, got %T", idValue)
}

func (m *Manager[T]) setID(instance *T, id int64) {
	if modelWithID, ok := any(instance).(ModelWithID); ok {
		modelWithID.SetID(id)
		return
	}

	// Fallback to reflection (including embedded structs)
	instanceValue := reflect.ValueOf(instance).Elem()
	var setInValue func(v reflect.Value) bool
	setInValue = func(v reflect.Value) bool {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			if field.Anonymous {
				switch fieldValue.Kind() {
				case reflect.Struct:
					if setInValue(fieldValue) {
						return true
					}
				case reflect.Ptr:
					if fieldValue.IsNil() {
						continue
					}
					if fieldValue.Elem().Kind() == reflect.Struct {
						if setInValue(fieldValue.Elem()) {
							return true
						}
					}
				}
			}

			for _, name := range []string{"ID", "Id", "id"} {
				if field.Name == name && fieldValue.CanSet() && fieldValue.Kind() == reflect.Int64 {
					fieldValue.SetInt(id)
					return true
				}
			}
		}
		return false
	}

	_ = setInValue(instanceValue)
}

func (m *Manager[T]) validate(instance *T) error {
	// 1. Run interface-based Clean method (defined directly on the struct)
	if validatable, ok := any(instance).(interface{ Clean() error }); ok {
		if err := validatable.Clean(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// 2. Run schema-defined Clean hook (from Schema.Hooks())
	if s, ok := any(instance).(schema.Schema); ok {
		hooks := s.Hooks()
		if hooks != nil && hooks.Clean != nil {
			if err := hooks.Clean(instance); err != nil {
				return fmt.Errorf("schema validation failed: %w", err)
			}
		}
	}

	// 3. Run Validate() if the model implements it (e.g. from generated code)
	if validatable, ok := any(instance).(interface{ Validate() error }); ok {
		if err := validatable.Validate(); err != nil {
			return fmt.Errorf("model validation failed: %w", err)
		}
	}

	return nil
}

func (m *Manager[T]) runHooks(ctx context.Context, instance *T, hookType string) error {
	// 1. Run interface-based hooks (methods defined directly on the model struct)
	var err error
	switch hookType {
	case "BeforeCreate":
		if h, ok := any(instance).(interface{ BeforeCreate(context.Context) error }); ok {
			err = h.BeforeCreate(ctx)
		}
	case "AfterCreate":
		if h, ok := any(instance).(interface{ AfterCreate(context.Context) error }); ok {
			err = h.AfterCreate(ctx)
		}
	case "BeforeUpdate":
		if h, ok := any(instance).(interface{ BeforeUpdate(context.Context) error }); ok {
			err = h.BeforeUpdate(ctx)
		}
	case "AfterUpdate":
		if h, ok := any(instance).(interface{ AfterUpdate(context.Context) error }); ok {
			err = h.AfterUpdate(ctx)
		}
	case "BeforeSave":
		if h, ok := any(instance).(interface{ BeforeSave(context.Context) error }); ok {
			err = h.BeforeSave(ctx)
		}
	case "AfterSave":
		if h, ok := any(instance).(interface{ AfterSave(context.Context) error }); ok {
			err = h.AfterSave(ctx)
		}
	case "BeforeDelete":
		if h, ok := any(instance).(interface{ BeforeDelete(context.Context) error }); ok {
			err = h.BeforeDelete(ctx)
		}
	case "AfterDelete":
		if h, ok := any(instance).(interface{ AfterDelete(context.Context) error }); ok {
			err = h.AfterDelete(ctx)
		}
	}

	if err != nil {
		return fmt.Errorf("%s hook failed: %w", hookType, err)
	}

	// 2. Run schema-defined hooks (functions returned by Schema.Hooks())
	if s, ok := any(instance).(schema.Schema); ok {
		hooks := s.Hooks()
		if hooks != nil {
			var schemaErr error
			switch hookType {
			case "BeforeCreate":
				if hooks.BeforeCreate != nil {
					schemaErr = hooks.BeforeCreate(ctx, instance)
				}
			case "AfterCreate":
				if hooks.AfterCreate != nil {
					schemaErr = hooks.AfterCreate(ctx, instance)
				}
			case "BeforeUpdate":
				if hooks.BeforeUpdate != nil {
					schemaErr = hooks.BeforeUpdate(ctx, instance)
				}
			case "AfterUpdate":
				if hooks.AfterUpdate != nil {
					schemaErr = hooks.AfterUpdate(ctx, instance)
				}
			case "BeforeSave":
				if hooks.BeforeSave != nil {
					schemaErr = hooks.BeforeSave(ctx, instance)
				}
			case "AfterSave":
				if hooks.AfterSave != nil {
					schemaErr = hooks.AfterSave(ctx, instance)
				}
			case "BeforeDelete":
				if hooks.BeforeDelete != nil {
					schemaErr = hooks.BeforeDelete(ctx, instance)
				}
			case "AfterDelete":
				if hooks.AfterDelete != nil {
					schemaErr = hooks.AfterDelete(ctx, instance)
				}
			}
			if schemaErr != nil {
				return fmt.Errorf("%s hook failed: %w", hookType, schemaErr)
			}
		}
	}

	return nil
}
