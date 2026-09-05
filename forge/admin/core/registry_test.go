package core

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/forgego/forge/admin/components"
	"github.com/forgego/forge/db"
)

// mockAdmin implements AdminInterface for testing
type mockAdmin struct {
	name string
}

func (m *mockAdmin) ModelName() string                                               { return m.name }
func (m *mockAdmin) ModelType() reflect.Type                                         { return reflect.TypeOf(struct{}{}) }
func (m *mockAdmin) GetMetadata(ctx context.Context, user interface{}) (*Metadata, error) { return nil, nil }
func (m *mockAdmin) SetDB(database *db.DB)                                           {}
func (m *mockAdmin) ManagerInterface() interface{}                                   { return nil }
func (m *mockAdmin) ConfigInterface() interface{}                                    { return nil }
func (m *mockAdmin) GetHistory(ctx context.Context, objectID string) ([]LogEntry, error)  { return nil, nil }
func (m *mockAdmin) LogAction(ctx context.Context, user interface{}, objectID string, repr string, action ActionType, changes string) error {
	return nil
}
func (m *mockAdmin) HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return true
}
func (m *mockAdmin) HasAddPermission(ctx context.Context, user interface{}) bool {
	return true
}
func (m *mockAdmin) HasChangePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return true
}
func (m *mockAdmin) HasDeletePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return true
}
func (m *mockAdmin) HasModulePermission(ctx context.Context, user interface{}) bool {
	return true
}
func (m *mockAdmin) PageType() string { return "list" }
func (m *mockAdmin) ListObjects(ctx context.Context, params ListParams) (*PaginatedResponse, error) {
	return nil, nil
}
func (m *mockAdmin) GetObject(ctx context.Context, id interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockAdmin) CreateObject(ctx context.Context, data map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockAdmin) UpdateObject(ctx context.Context, id interface{}, data map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockAdmin) DeleteObject(ctx context.Context, id interface{}) error {
	return nil
}
func (m *mockAdmin) ExecuteAction(ctx context.Context, actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m *mockAdmin) Autocomplete(ctx context.Context, query string, limit int) ([]AutocompleteItem, error) {
	return nil, nil
}

// mockPlugin implements Plugin for testing
type mockPlugin struct {
	id   string
	name string
}

func (m *mockPlugin) ID() string                                        { return m.id }
func (m *mockPlugin) Name() string                                      { return m.name }
func (m *mockPlugin) Init(ctx context.Context, site interface{}) error { return nil }
func (m *mockPlugin) GetPages() map[string]components.Component         { return nil }
func (m *mockPlugin) GetMenuItems() []MenuItem                          { return nil }

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if registry.admins == nil {
		t.Error("Registry.admins is nil")
	}
	if registry.plugins == nil {
		t.Error("Registry.plugins is nil")
	}
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name      string
		adminName string
		expectErr bool
	}{
		{
			name:      "register new admin",
			adminName: "user",
			expectErr: false,
		},
		{
			name:      "register another admin",
			adminName: "product",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			admin := &mockAdmin{name: tt.adminName}
			err := registry.Register(admin)
			if tt.expectErr {
				if err == nil {
					t.Error("Register() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Register() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	registry := NewRegistry()
	admin1 := &mockAdmin{name: "user"}
	admin2 := &mockAdmin{name: "user"}

	err := registry.Register(admin1)
	if err != nil {
		t.Errorf("First Register() error = %v", err)
	}

	err = registry.Register(admin2)
	if err == nil {
		t.Error("Second Register() expected error for duplicate, got nil")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	admin := &mockAdmin{name: "user"}
	registry.Register(admin)

	tests := []struct {
		name      string
		modelName string
		expectErr bool
	}{
		{
			name:      "get existing admin",
			modelName: "user",
			expectErr: false,
		},
		{
			name:      "get non-existing admin",
			modelName: "nonexistent",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.Get(tt.modelName)
			if tt.expectErr {
				if err == nil {
					t.Error("Get() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Get() error = %v, want nil", err)
				}
				if result == nil {
					t.Error("Get() returned nil")
				}
			}
		})
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()
	
	// Empty registry
	all := registry.GetAll()
	if len(all) != 0 {
		t.Errorf("GetAll() on empty registry = %d items, want 0", len(all))
	}

	// Add admins
	registry.Register(&mockAdmin{name: "user"})
	registry.Register(&mockAdmin{name: "product"})

	all = registry.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() = %d items, want 2", len(all))
	}

	// Verify it's a copy
	all["modified"] = nil
	if registry.Has("modified") {
		t.Error("GetAll() should return a copy, not the original map")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockAdmin{name: "user"})

	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{"existing", "user", true},
		{"non-existing", "product", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := registry.Has(tt.modelName); result != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.modelName, result, tt.expected)
			}
		})
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockAdmin{name: "user"})

	// Unregister existing
	err := registry.Unregister("user")
	if err != nil {
		t.Errorf("Unregister() error = %v", err)
	}
	if registry.Has("user") {
		t.Error("Unregister() failed to remove admin")
	}

	// Unregister non-existing
	err = registry.Unregister("nonexistent")
	if err == nil {
		t.Error("Unregister() expected error for non-existing admin, got nil")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()
	
	if count := registry.Count(); count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	registry.Register(&mockAdmin{name: "user"})
	if count := registry.Count(); count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}

	registry.Register(&mockAdmin{name: "product"})
	if count := registry.Count(); count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestRegistry_ModelNames(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockAdmin{name: "user"})
	registry.Register(&mockAdmin{name: "product"})

	names := registry.ModelNames()
	if len(names) != 2 {
		t.Errorf("ModelNames() length = %d, want 2", len(names))
	}

	// Check that names are present
	nameMap := make(map[string]bool)
	for _, n := range names {
		nameMap[n] = true
	}
	if !nameMap["user"] || !nameMap["product"] {
		t.Errorf("ModelNames() = %v, want [user, product]", names)
	}
}

func TestRegistry_RegisterPlugin(t *testing.T) {
	registry := NewRegistry()
	plugin := &mockPlugin{id: "analytics", name: "Analytics"}

	err := registry.RegisterPlugin(plugin)
	if err != nil {
		t.Errorf("RegisterPlugin() error = %v", err)
	}

	// Duplicate plugin
	err = registry.RegisterPlugin(plugin)
	if err == nil {
		t.Error("RegisterPlugin() expected error for duplicate, got nil")
	}
}

func TestRegistry_GetPlugin(t *testing.T) {
	registry := NewRegistry()
	plugin := &mockPlugin{id: "analytics", name: "Analytics"}
	registry.RegisterPlugin(plugin)

	// Get existing
	result, err := registry.GetPlugin("analytics")
	if err != nil {
		t.Errorf("GetPlugin() error = %v", err)
	}
	if result == nil {
		t.Error("GetPlugin() returned nil")
	}

	// Get non-existing
	_, err = registry.GetPlugin("nonexistent")
	if err == nil {
		t.Error("GetPlugin() expected error for non-existing, got nil")
	}
}

func TestRegistry_GetAllPlugins(t *testing.T) {
	registry := NewRegistry()
	
	// Empty registry
	all := registry.GetAllPlugins()
	if len(all) != 0 {
		t.Errorf("GetAllPlugins() on empty registry = %d items, want 0", len(all))
	}

	// Add plugins
	registry.RegisterPlugin(&mockPlugin{id: "analytics", name: "Analytics"})
	registry.RegisterPlugin(&mockPlugin{id: "reports", name: "Reports"})

	all = registry.GetAllPlugins()
	if len(all) != 2 {
		t.Errorf("GetAllPlugins() = %d items, want 2", len(all))
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			admin := &mockAdmin{name: string(rune('a' + i%26))}
			registry.Register(admin)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.GetAll()
			registry.Count()
			registry.ModelNames()
		}()
	}

	wg.Wait()
}

func TestGetGlobalRegistry(t *testing.T) {
	registry := GetGlobalRegistry()
	if registry == nil {
		t.Error("GetGlobalRegistry() returned nil")
	}
	
	// Should return the same instance
	registry2 := GetGlobalRegistry()
	if registry != registry2 {
		t.Error("GetGlobalRegistry() should return the same instance")
	}
}
