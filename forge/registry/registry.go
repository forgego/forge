package registry

import (
	"fmt"
	"reflect"
	"sync"
)

// ModelRegistry maintains a registry of all registered models
type ModelRegistry struct {
	models map[string]*ModelInfo
	mu     sync.RWMutex
}

// ModelInfo contains information about a registered model
type ModelInfo struct {
	ModelType reflect.Type
	Schema    interface{}
	Name      string
}

var globalRegistry = &ModelRegistry{
	models: make(map[string]*ModelInfo),
}

// RegisterModel registers a model in the registry
// This is used for admin auto-generation (Django-style)
func RegisterModel(model interface{}) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	modelName := modelType.Name()
	if _, exists := globalRegistry.models[modelName]; exists {
		return fmt.Errorf("model %s is already registered", modelName)
	}

	globalRegistry.models[modelName] = &ModelInfo{
		Name:      modelName,
		ModelType: modelType,
		Schema:    model,
	}

	return nil
}

// GetModel retrieves a model from the registry
func GetModel(name string) (*ModelInfo, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	info, exists := globalRegistry.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s is not registered", name)
	}

	return info, nil
}

// GetAllModels returns all registered models
func GetAllModels() map[string]*ModelInfo {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	result := make(map[string]*ModelInfo)
	for k, v := range globalRegistry.models {
		result[k] = v
	}

	return result
}

// Global returns the global model registry instance.
// Use this to access the framework's model registry for registration,
// lookup, and iteration over all registered models.
func Global() *ModelRegistry {
	return globalRegistry
}

// GetRegistry returns the global model registry.
//
// Deprecated: Use Global() instead to avoid package name redundancy. GetRegistry will be removed in v3.0.
// Migration:
//   // Old
//   registry := registry.GetRegistry()
//   // New
//   registry := registry.Global()
func GetRegistry() *ModelRegistry {
	return Global()
}
