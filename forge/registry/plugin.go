package registry

import (
	"fmt"
	"reflect"
)

// Plugin is the main interface that all plugins must implement
// This allows third parties to extend models, admin, and APIs
type Plugin interface {
	// Name returns the plugin name
	Name() string
	// Version returns the plugin version
	Version() string
	// Install is called when the plugin is registered
	Install() error
}

// ModelPlugin extends model functionality
type ModelPlugin interface {
	Plugin
	// ExtendModel is called for each registered model
	ExtendModel(model interface{}) error
	// GetModelFields returns additional fields to add to models
	GetModelFields(modelName string) []interface{}
	// GetModelRelations returns additional relations to add to models
	GetModelRelations(modelName string) []interface{}
	// GetModelHooks returns additional hooks to add to models
	GetModelHooks(modelName string) interface{}
}

// AdminPlugin extends admin functionality
type AdminPlugin interface {
	Plugin
	// ExtendAdmin is called for each admin model
	ExtendAdmin(adminModel interface{}) error
	// GetAdminActions returns custom admin actions
	GetAdminActions(modelName string) []AdminAction
	// GetAdminFilters returns custom admin filters
	GetAdminFilters(modelName string) []AdminFilter
	// GetAdminWidgets returns custom form widgets
	GetAdminWidgets(modelName string) map[string]AdminWidget
}

// APIPlugin extends API functionality
type APIPlugin interface {
	Plugin
	// ExtendAPI is called for each API viewset
	ExtendAPI(viewset interface{}) error
	// GetAPIActions returns custom API actions
	GetAPIActions(resourceName string) []APIAction
	// GetAPIFilters returns custom API filters
	GetAPIFilters(resourceName string) []APIFilter
	// GetAPISerializers returns custom serializers
	GetAPISerializers(resourceName string) map[string]interface{}
}

// AdminAction represents a custom admin action
type AdminAction struct {
	Handler     func(ids []interface{}) error
	Name        string
	Label       string
	Description string
}

// AdminFilter represents a custom admin filter
type AdminFilter struct {
	Handler func(value interface{}) interface{}
	Name    string
	Label   string
	Type    string
	Choices []FilterChoice
}

// FilterChoice represents a filter choice option
type FilterChoice struct {
	Value string
	Label string
}

// AdminWidget represents a custom form widget
type AdminWidget struct {
	Type     string   // "select", "textarea", "datepicker", etc.
	Template string   // HTML template for the widget
	Scripts  []string // JavaScript files to include
	Styles   []string // CSS files to include
}

// APIAction represents a custom API action
type APIAction struct {
	Handler func(w, r interface{}) error
	Name    string
	Method  string
	Path    string
}

// APIFilter represents a custom API filter
type APIFilter struct {
	Handler func(value interface{}) interface{}
	Name    string
	Type    string
}

// PluginRegistry maintains a registry of all plugins
type PluginRegistry struct {
	plugins      map[string]Plugin
	modelPlugins []ModelPlugin
	adminPlugins []AdminPlugin
	apiPlugins   []APIPlugin
}

var globalPluginRegistry = &PluginRegistry{
	plugins:      make(map[string]Plugin),
	modelPlugins: []ModelPlugin{},
	adminPlugins: []AdminPlugin{},
	apiPlugins:   []APIPlugin{},
}

// RegisterPlugin registers a plugin and applies its extensions
func RegisterPlugin(plugin Plugin) error {
	if _, exists := globalPluginRegistry.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s is already registered", plugin.Name())
	}

	globalPluginRegistry.plugins[plugin.Name()] = plugin

	// Call Install (no arguments for new plugin interface)
	if err := plugin.Install(); err != nil {
		return fmt.Errorf("plugin %s installation failed: %w", plugin.Name(), err)
	}

	// Register by type
	if modelPlugin, ok := plugin.(ModelPlugin); ok {
		globalPluginRegistry.modelPlugins = append(globalPluginRegistry.modelPlugins, modelPlugin)
		applyModelExtensions(modelPlugin)
	}

	if adminPlugin, ok := plugin.(AdminPlugin); ok {
		globalPluginRegistry.adminPlugins = append(globalPluginRegistry.adminPlugins, adminPlugin)
		applyAdminExtensions(adminPlugin)
	}

	if apiPlugin, ok := plugin.(APIPlugin); ok {
		globalPluginRegistry.apiPlugins = append(globalPluginRegistry.apiPlugins, apiPlugin)
		applyAPIExtensions(apiPlugin)
	}

	return nil
}

// GetPlugin retrieves a plugin by name
func GetPlugin(name string) (Plugin, error) {
	plugin, exists := globalPluginRegistry.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s is not registered", name)
	}
	return plugin, nil
}

// GetAllPlugins returns all registered plugins
func GetAllPlugins() map[string]Plugin {
	result := make(map[string]Plugin)
	for k, v := range globalPluginRegistry.plugins {
		result[k] = v
	}
	return result
}

// GetModelPlugins returns all model plugins
func GetModelPlugins() []ModelPlugin {
	return globalPluginRegistry.modelPlugins
}

// GetAdminPlugins returns all admin plugins
func GetAdminPlugins() []AdminPlugin {
	return globalPluginRegistry.adminPlugins
}

// GetAPIPlugins returns all API plugins
func GetAPIPlugins() []APIPlugin {
	return globalPluginRegistry.apiPlugins
}

// applyModelExtensions applies model extensions from a plugin
func applyModelExtensions(plugin ModelPlugin) {
	models := GetAllModels()
	for name, modelInfo := range models {
		if err := plugin.ExtendModel(modelInfo.Schema); err != nil {
			// Log error but continue
			fmt.Printf("Warning: plugin %s failed to extend model %s: %v\n", plugin.Name(), name, err)
		}
	}
}

// applyAdminExtensions applies admin extensions from a plugin
func applyAdminExtensions(plugin AdminPlugin) {
	// This would be called when admin models are registered
	// For now, it's a placeholder - full implementation would integrate with admin registry
}

// applyAPIExtensions applies API extensions from a plugin
func applyAPIExtensions(plugin APIPlugin) {
	// This would be called when API viewsets are registered
	// For now, it's a placeholder - full implementation would integrate with API router
}

// BasePlugin provides a base implementation for plugins
type BasePlugin struct {
	name    string
	version string
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name, version string) *BasePlugin {
	return &BasePlugin{
		name:    name,
		version: version,
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *BasePlugin) Version() string {
	return p.version
}

// Install is called when the plugin is registered (can be overridden)
func (p *BasePlugin) Install() error {
	return nil
}

// ModelPluginBase provides a base implementation for model plugins
type ModelPluginBase struct {
	*BasePlugin
}

// NewModelPluginBase creates a new model plugin base
func NewModelPluginBase(name, version string) *ModelPluginBase {
	return &ModelPluginBase{
		BasePlugin: NewBasePlugin(name, version),
	}
}

// ExtendModel extends a model (can be overridden)
func (p *ModelPluginBase) ExtendModel(model interface{}) error {
	return nil
}

// GetModelFields returns additional fields (can be overridden)
func (p *ModelPluginBase) GetModelFields(modelName string) []interface{} {
	return nil
}

// GetModelRelations returns additional relations (can be overridden)
func (p *ModelPluginBase) GetModelRelations(modelName string) []interface{} {
	return nil
}

// GetModelHooks returns additional hooks (can be overridden)
func (p *ModelPluginBase) GetModelHooks(modelName string) interface{} {
	return nil
}

// AdminPluginBase provides a base implementation for admin plugins
type AdminPluginBase struct {
	*BasePlugin
}

// NewAdminPluginBase creates a new admin plugin base
func NewAdminPluginBase(name, version string) *AdminPluginBase {
	return &AdminPluginBase{
		BasePlugin: NewBasePlugin(name, version),
	}
}

// ExtendAdmin extends admin (can be overridden)
func (p *AdminPluginBase) ExtendAdmin(adminModel interface{}) error {
	return nil
}

// GetAdminActions returns custom admin actions (can be overridden)
func (p *AdminPluginBase) GetAdminActions(modelName string) []AdminAction {
	return nil
}

// GetAdminFilters returns custom admin filters (can be overridden)
func (p *AdminPluginBase) GetAdminFilters(modelName string) []AdminFilter {
	return nil
}

// GetAdminWidgets returns custom form widgets (can be overridden)
func (p *AdminPluginBase) GetAdminWidgets(modelName string) map[string]AdminWidget {
	return nil
}

// APIPluginBase provides a base implementation for API plugins
type APIPluginBase struct {
	*BasePlugin
}

// NewAPIPluginBase creates a new API plugin base
func NewAPIPluginBase(name, version string) *APIPluginBase {
	return &APIPluginBase{
		BasePlugin: NewBasePlugin(name, version),
	}
}

// ExtendAPI extends API (can be overridden)
func (p *APIPluginBase) ExtendAPI(viewset interface{}) error {
	return nil
}

// GetAPIActions returns custom API actions (can be overridden)
func (p *APIPluginBase) GetAPIActions(resourceName string) []APIAction {
	return nil
}

// GetAPIFilters returns custom API filters (can be overridden)
func (p *APIPluginBase) GetAPIFilters(resourceName string) []APIFilter {
	return nil
}

// GetAPISerializers returns custom serializers (can be overridden)
func (p *APIPluginBase) GetAPISerializers(resourceName string) map[string]interface{} {
	return nil
}

// Example plugin implementation
type ExampleAuthPlugin struct {
	*ModelPluginBase
	*AdminPluginBase
	*APIPluginBase
}

// NewExampleAuthPlugin creates an example auth plugin
func NewExampleAuthPlugin() *ExampleAuthPlugin {
	base := NewBasePlugin("auth", "1.0.0")
	return &ExampleAuthPlugin{
		ModelPluginBase: &ModelPluginBase{BasePlugin: base},
		AdminPluginBase: &AdminPluginBase{BasePlugin: base},
		APIPluginBase:   &APIPluginBase{BasePlugin: base},
	}
}

// Install installs the auth plugin
func (p *ExampleAuthPlugin) Install() error {
	// Register custom fields, relations, etc.
	return nil
}

// ExtendModel extends models with auth fields
func (p *ExampleAuthPlugin) ExtendModel(model interface{}) error {
	// Add auth-related fields to models
	// This is a simplified example
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	// Example: Add is_active, is_staff fields if not present
	// Full implementation would use schema interface
	return nil
}

// GetAdminActions returns auth-related admin actions
func (p *ExampleAuthPlugin) GetAdminActions(modelName string) []AdminAction {
	if modelName == "User" {
		return []AdminAction{
			{
				Name:        "activate_users",
				Label:       "Activate selected users",
				Description: "Activate the selected users",
				Handler: func(ids []interface{}) error {
					// Implementation
					return nil
				},
			},
		}
	}
	return nil
}
