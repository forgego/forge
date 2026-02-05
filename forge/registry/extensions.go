package registry

import (
	"fmt"
	"sync"
)

// ExtensionRegistry maintains a registry of all extensions
type ExtensionRegistry struct {
	extensions map[string]Extension
	mu         sync.RWMutex
}

// Extension is the interface that all extensions must implement
type Extension interface {
	Name() string
	Version() string
}

// FieldExtension extends field functionality
type FieldExtension interface {
	Extension
	ExtendField(field interface{}) error
}

// RelationExtension extends relation functionality
type RelationExtension interface {
	Extension
	ExtendRelation(relation interface{}) error
}

// QuerySetExtension extends QuerySet functionality
type QuerySetExtension interface {
	Extension
	ExtendQuerySet(queryset interface{}) error
}

// ManagerExtension extends Manager functionality
type ManagerExtension interface {
	Extension
	ExtendManager(manager interface{}) error
}

// AdminExtension extends Admin functionality
type AdminExtension interface {
	Extension
	ExtendAdmin(admin interface{}) error
}

var globalExtensionRegistry = &ExtensionRegistry{
	extensions: make(map[string]Extension),
}

// RegisterExtension registers an extension
func RegisterExtension(ext Extension) error {
	globalExtensionRegistry.mu.Lock()
	defer globalExtensionRegistry.mu.Unlock()

	if _, exists := globalExtensionRegistry.extensions[ext.Name()]; exists {
		return fmt.Errorf("extension %s is already registered", ext.Name())
	}

	globalExtensionRegistry.extensions[ext.Name()] = ext
	return nil
}

// GetExtension retrieves an extension
func GetExtension(name string) (Extension, error) {
	globalExtensionRegistry.mu.RLock()
	defer globalExtensionRegistry.mu.RUnlock()

	ext, exists := globalExtensionRegistry.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s is not registered", name)
	}

	return ext, nil
}

// GetAllExtensions returns all registered extensions
func GetAllExtensions() map[string]Extension {
	globalExtensionRegistry.mu.RLock()
	defer globalExtensionRegistry.mu.RUnlock()

	result := make(map[string]Extension)
	for k, v := range globalExtensionRegistry.extensions {
		result[k] = v
	}

	return result
}

