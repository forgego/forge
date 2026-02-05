package core

import (
	"fmt"
	"sync"
)

// Registry manages all registered admin instances
type Registry struct {
	admins  map[string]AdminInterface
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		admins:  make(map[string]AdminInterface),
		plugins: make(map[string]Plugin),
	}
}

// Register registers an admin instance
func (r *Registry) Register(admin AdminInterface) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := admin.ModelName()
	if _, exists := r.admins[name]; exists {
		return fmt.Errorf("admin for model %q already registered", name)
	}

	r.admins[name] = admin
	return nil
}

// Get retrieves an admin by model name
func (r *Registry) Get(modelName string) (AdminInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	admin, ok := r.admins[modelName]
	if !ok {
		return nil, fmt.Errorf("admin for model %q not found", modelName)
	}

	return admin, nil
}

// GetAll returns all registered admins
func (r *Registry) GetAll() map[string]AdminInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]AdminInterface, len(r.admins))
	for k, v := range r.admins {
		result[k] = v
	}

	return result
}

// Has checks if an admin is registered
func (r *Registry) Has(modelName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.admins[modelName]
	return ok
}

// Unregister removes an admin from the registry
func (r *Registry) Unregister(modelName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.admins[modelName]; !ok {
		return fmt.Errorf("admin for model %q not found", modelName)
	}

	delete(r.admins, modelName)
	return nil
}

// Count returns the number of registered admins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.admins)
}

// ModelNames returns all registered model names
func (r *Registry) ModelNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.admins))
	for name := range r.admins {
		names = append(names, name)
	}

	return names
}

// RegisterPlugin registers a plugin
func (r *Registry) RegisterPlugin(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID()
	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin %q already registered", id)
	}

	r.plugins[id] = p
	return nil
}

// GetPlugin retrieves a plugin by ID
func (r *Registry) GetPlugin(id string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", id)
	}

	return p, nil
}

// GetAllPlugins returns all registered plugins
func (r *Registry) GetAllPlugins() map[string]Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Plugin, len(r.plugins))
	for k, v := range r.plugins {
		result[k] = v
	}

	return result
}

// Global registry instance
var globalRegistry = NewRegistry()

// GetGlobalRegistry returns the global registry
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// Register registers an admin with the global registry
func Register[T any](admin *Admin[T]) error {
	return globalRegistry.Register(admin)
}
