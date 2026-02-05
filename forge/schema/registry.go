package schema

import (
	"fmt"
	"sync"
)

// FieldBuilderFactory is a function that creates a new field builder for a given field name
type FieldBuilderFactory func(name string) interface{}

var (
	fieldRegistry = make(map[string]FieldBuilderFactory)
	registryMutex sync.RWMutex
)

// RegisterFieldType registers a custom field type that can be used by third-party packages
// or users to extend the schema system.
//
// Example:
//
//	RegisterFieldType("custom_type", func(name string) interface{} {
//	    return &CustomFieldBuilder{UnifiedFieldBuilder: newUnifiedFieldBuilder(name, TypeString)}}
//	})
func RegisterFieldType(typeName string, factory FieldBuilderFactory) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if typeName == "" {
		return fmt.Errorf("field type name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("field builder factory cannot be nil")
	}

	if _, exists := fieldRegistry[typeName]; exists {
		return fmt.Errorf("field type %q is already registered", typeName)
	}

	fieldRegistry[typeName] = factory
	return nil
}

// UnregisterFieldType removes a registered field type
func UnregisterFieldType(typeName string) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	delete(fieldRegistry, typeName)
}

// GetFieldBuilderFactory retrieves the factory function for a registered field type
func GetFieldBuilderFactory(typeName string) (FieldBuilderFactory, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	factory, exists := fieldRegistry[typeName]
	if !exists {
		return nil, fmt.Errorf("field type %q is not registered", typeName)
	}

	return factory, nil
}

// IsFieldTypeRegistered checks if a field type is registered
func IsFieldTypeRegistered(typeName string) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	_, exists := fieldRegistry[typeName]
	return exists
}

// ListRegisteredFieldTypes returns a list of all registered custom field types
func ListRegisteredFieldTypes() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	types := make([]string, 0, len(fieldRegistry))
	for typeName := range fieldRegistry {
		types = append(types, typeName)
	}
	return types
}

// CustomFieldBuilder is an interface that custom field builders should implement
// to work seamlessly with the schema system
type CustomFieldBuilder interface {
	// Build returns the final Field
	Build() Field
	// GetFieldType returns the FieldType for this builder
	GetFieldType() FieldType
}

// NewFieldBuilder creates a new field builder instance for a registered field type
// This allows dynamic creation of field builders at runtime
func NewFieldBuilder(typeName, fieldName string) (CustomFieldBuilder, error) {
	factory, err := GetFieldBuilderFactory(typeName)
	if err != nil {
		return nil, err
	}

	builder := factory(fieldName)
	if builder == nil {
		return nil, fmt.Errorf("factory for field type %q returned nil", typeName)
	}

	customBuilder, ok := builder.(CustomFieldBuilder)
	if !ok {
		return nil, fmt.Errorf("field builder for type %q does not implement CustomFieldBuilder interface", typeName)
	}

	return customBuilder, nil
}
