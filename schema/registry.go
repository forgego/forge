package schema

import (
	"fmt"
	"sync"
)

// FieldFactory is a function that creates a new field for a given field name.
type FieldFactory func(name string) Field

var (
	fieldRegistry = make(map[string]FieldFactory)
	registryMutex sync.RWMutex
)

// RegisterFieldType registers a custom field type that can be used by third-party packages
// or users to extend the schema system.
//
// Example:
//
//	RegisterFieldType("custom_type", func(name string) Field {
//	    return String(name)
//	})
func RegisterFieldType(typeName string, factory FieldFactory) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if typeName == "" {
		return fmt.Errorf("field type name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("field factory cannot be nil")
	}

	if _, exists := fieldRegistry[typeName]; exists {
		return fmt.Errorf("field type %q is already registered", typeName)
	}

	fieldRegistry[typeName] = factory
	return nil
}

// UnregisterFieldType removes a registered field type.
func UnregisterFieldType(typeName string) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	delete(fieldRegistry, typeName)
}

// GetFieldFactory retrieves the factory function for a registered field type.
func GetFieldFactory(typeName string) (FieldFactory, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	factory, exists := fieldRegistry[typeName]
	if !exists {
		return nil, fmt.Errorf("field type %q is not registered", typeName)
	}

	return factory, nil
}

// IsFieldTypeRegistered checks if a field type is registered.
func IsFieldTypeRegistered(typeName string) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	_, exists := fieldRegistry[typeName]
	return exists
}

// ListRegisteredFieldTypes returns a list of all registered custom field types.
func ListRegisteredFieldTypes() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	types := make([]string, 0, len(fieldRegistry))
	for typeName := range fieldRegistry {
		types = append(types, typeName)
	}
	return types
}

// NewField creates a new field instance for a registered field type.
func NewField(typeName, fieldName string) (Field, error) {
	factory, err := GetFieldFactory(typeName)
	if err != nil {
		return Field{}, err
	}

	return factory(fieldName), nil
}
