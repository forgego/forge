package registry

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/forgego/forge/orm"
)

// TypeInfo provides type information for a registered type
type TypeInfo[T any] struct {
	name         string
	modelType    reflect.Type
	fieldAccessor *orm.FieldAccessor[T]
	schema       *orm.ModelSchema
}

// New creates a new instance of the registered type
func (ti *TypeInfo[T]) New() *T {
	return reflect.New(ti.modelType).Interface().(*T)
}

// FieldAccessor returns the field accessor for this type
func (ti *TypeInfo[T]) FieldAccessor() (*orm.FieldAccessor[T], error) {
	if ti.fieldAccessor != nil {
		return ti.fieldAccessor, nil
	}
	return orm.NewFieldAccessor[T]()
}

// Schema returns the model schema
func (ti *TypeInfo[T]) Schema() *orm.ModelSchema {
	return ti.schema
}

// Name returns the type name
func (ti *TypeInfo[T]) Name() string {
	return ti.name
}

// TypeRegistry provides type registration and recovery
type TypeRegistry struct {
	types map[string]interface{} // map[string]*TypeInfo[T]
	mu    sync.RWMutex
}

var globalTypeRegistry = &TypeRegistry{
	types: make(map[string]interface{}),
}

// Register registers a type with the global type registry
func Register[T any](name string) error {
	return RegisterType[T](globalTypeRegistry, name)
}

// Get retrieves type information by name
func Get[T any](name string) (*TypeInfo[T], error) {
	return GetType[T](name)
}

// RegisterType registers a type (generic function wrapper)
func RegisterType[T any](tr *TypeRegistry, name string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// Get model schema
	schema, err := orm.GetModelSchema[T]()
	if err != nil {
		return fmt.Errorf("failed to get model schema: %w", err)
	}

	// Get field accessor
	accessor, err := orm.NewFieldAccessor[T]()
	if err != nil {
		return fmt.Errorf("failed to create field accessor: %w", err)
	}

	// Get model type
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Create type info
	typeInfo := &TypeInfo[T]{
		name:         name,
		modelType:    modelType,
		fieldAccessor: accessor,
		schema:       schema,
	}

	tr.types[name] = typeInfo
	return nil
}

// GetType retrieves type information (generic function, not method)
func GetType[T any](name string) (*TypeInfo[T], error) {
	return GetTypeInternal[T](globalTypeRegistry, name)
}

// Get retrieves type information (non-generic method for internal use)
func (tr *TypeRegistry) getTypeInfo(name string) (interface{}, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	typeInfoInterface, ok := tr.types[name]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", name)
	}
	return typeInfoInterface, nil
}

// GetTypeInternal retrieves type information (internal generic function)
func GetTypeInternal[T any](tr *TypeRegistry, name string) (*TypeInfo[T], error) {
	typeInfoInterface, err := tr.getTypeInfo(name)
	if err != nil {
		return nil, err
	}

	// Type assertion
	typeInfo, ok := typeInfoInterface.(*TypeInfo[T])
	if !ok {
		return nil, fmt.Errorf("type %s is not of type %T", name, (*TypeInfo[T])(nil))
	}

	return typeInfo, nil
}

// GetInternal is deprecated, use GetType instead (internal helper)
func GetInternal[T any](tr *TypeRegistry, name string) (*TypeInfo[T], error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	typeInfoInterface, ok := tr.types[name]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", name)
	}

	// Type assertion
	typeInfo, ok := typeInfoInterface.(*TypeInfo[T])
	if !ok {
		return nil, fmt.Errorf("type %s is not of type %T", name, (*TypeInfo[T])(nil))
	}

	return typeInfo, nil
}

// List returns all registered type names
func (tr *TypeRegistry) List() []string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	names := make([]string, 0, len(tr.types))
	for name := range tr.types {
		names = append(names, name)
	}
	return names
}
