package orm

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	// schemaNameRegistry maps type names to reflect.Type
	schemaNameRegistry = make(map[string]reflect.Type)
	schemaNameMu      sync.RWMutex
)

// RegisterModelType registers a model type by name
func RegisterModelType(name string, typ reflect.Type) {
	schemaNameMu.Lock()
	defer schemaNameMu.Unlock()
	schemaNameRegistry[name] = typ
}

// GetModelSchemaByName resolves a model schema by type name
func GetModelSchemaByName(typeName string) (*ModelSchema, error) {
	schemaNameMu.RLock()
	typ, ok := schemaNameRegistry[typeName]
	schemaNameMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("model type '%s' not found in registry. Available types: %v", typeName, getRegisteredTypeNames())
	}

	// Get schema from cache or build it
	return GetModelSchemaByType(typ)
}

// GetModelSchemaByType gets or builds a ModelSchema for a reflect.Type
// This is a helper that works with the schema cache
func GetModelSchemaByType(typ reflect.Type) (*ModelSchema, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Check cache
	schemaMu.RLock()
	if schema, ok := schemaCache[typ]; ok {
		schemaMu.RUnlock()
		return schema, nil
	}
	schemaMu.RUnlock()

	// Try to build schema from type
	// This requires the type to implement schema.Schema interface
	instanceValue := reflect.New(typ).Elem()
	
	// Try to get schema interface
	schemaInstance, ok := instanceValue.Interface().(interface {
		Fields() interface{} // Would need actual schema.Schema
		Meta() interface{}
	})
	if !ok {
		return nil, fmt.Errorf("type %v does not implement schema interface. Use GetModelSchema[T]() to build it first", typ)
	}

	// Try to use BuildModelSchema if we can cast to schema.Schema
	// For now, return error - this would need proper schema.Schema interface
	_ = schemaInstance
	return nil, fmt.Errorf("schema for type %v not found in cache. Use GetModelSchema[T]() to build it first", typ)
}

// getRegisteredTypeNames returns all registered type names
func getRegisteredTypeNames() []string {
	schemaNameMu.RLock()
	defer schemaNameMu.RUnlock()

	names := make([]string, 0, len(schemaNameRegistry))
	for name := range schemaNameRegistry {
		names = append(names, name)
	}
	return names
}
