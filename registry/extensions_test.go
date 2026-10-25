package registry

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// resetGlobalExtensionRegistry backs up the current globalExtensionRegistry map,
// replaces it with an empty map, and registers a t.Cleanup callback to restore
// the original map when the test finishes.
func resetGlobalExtensionRegistry(t *testing.T) {
	t.Helper()
	globalExtensionRegistry.mu.Lock()
	saved := make(map[string]Extension, len(globalExtensionRegistry.extensions))
	for k, v := range globalExtensionRegistry.extensions {
		saved[k] = v
	}
	globalExtensionRegistry.extensions = make(map[string]Extension)
	globalExtensionRegistry.mu.Unlock()

	t.Cleanup(func() {
		globalExtensionRegistry.mu.Lock()
		globalExtensionRegistry.extensions = saved
		globalExtensionRegistry.mu.Unlock()
	})
}

// baseTestExtension is a basic test implementation of Extension.
type baseTestExtension struct {
	name    string
	version string
}

func (e *baseTestExtension) Name() string    { return e.name }
func (e *baseTestExtension) Version() string { return e.version }

// testFieldExtension implements FieldExtension.
type testFieldExtension struct {
	baseTestExtension
}

func (e *testFieldExtension) ExtendField(field interface{}) error { return nil }

// testRelationExtension implements RelationExtension.
type testRelationExtension struct {
	baseTestExtension
}

func (e *testRelationExtension) ExtendRelation(relation interface{}) error { return nil }

// testQuerySetExtension implements QuerySetExtension.
type testQuerySetExtension struct {
	baseTestExtension
}

func (e *testQuerySetExtension) ExtendQuerySet(queryset interface{}) error { return nil }

// testManagerExtension implements ManagerExtension.
type testManagerExtension struct {
	baseTestExtension
}

func (e *testManagerExtension) ExtendManager(manager interface{}) error { return nil }

// testAdminExtension implements AdminExtension.
type testAdminExtension struct {
	baseTestExtension
}

func (e *testAdminExtension) ExtendAdmin(admin interface{}) error { return nil }

func TestRegisterExtension_Success(t *testing.T) {
	tests := []struct {
		name        string
		ext         Extension
		wantName    string
		wantVersion string
	}{
		{
			name:        "base extension",
			ext:         &baseTestExtension{name: "audit-logger", version: "1.0.0"},
			wantName:    "audit-logger",
			wantVersion: "1.0.0",
		},
		{
			name:        "field extension",
			ext:         &testFieldExtension{baseTestExtension: baseTestExtension{name: "encrypted-field", version: "0.2.1"}},
			wantName:    "encrypted-field",
			wantVersion: "0.2.1",
		},
		{
			name:        "relation extension",
			ext:         &testRelationExtension{baseTestExtension: baseTestExtension{name: "soft-delete-rel", version: "1.5.0"}},
			wantName:    "soft-delete-rel",
			wantVersion: "1.5.0",
		},
		{
			name:        "queryset extension",
			ext:         &testQuerySetExtension{baseTestExtension: baseTestExtension{name: "tenant-filter", version: "2.0.0-beta"}},
			wantName:    "tenant-filter",
			wantVersion: "2.0.0-beta",
		},
		{
			name:        "manager extension",
			ext:         &testManagerExtension{baseTestExtension: baseTestExtension{name: "cache-manager", version: "3.1.4"}},
			wantName:    "cache-manager",
			wantVersion: "3.1.4",
		},
		{
			name:        "admin extension",
			ext:         &testAdminExtension{baseTestExtension: baseTestExtension{name: "custom-dashboard", version: "0.9.0"}},
			wantName:    "custom-dashboard",
			wantVersion: "0.9.0",
		},
		{
			name:        "extension with dots and underscores",
			ext:         &baseTestExtension{name: "org.forge.ext_v2", version: "2.1.0"},
			wantName:    "org.forge.ext_v2",
			wantVersion: "2.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalExtensionRegistry(t)

			if err := RegisterExtension(tt.ext); err != nil {
				t.Fatalf("RegisterExtension(%q) unexpected error: %v", tt.ext.Name(), err)
			}

			got, err := GetExtension(tt.wantName)
			if err != nil {
				t.Fatalf("GetExtension(%q) unexpected error: %v", tt.wantName, err)
			}
			if got != tt.ext {
				t.Errorf("GetExtension(%q) got instance %p, want %p", tt.wantName, got, tt.ext)
			}
			if got.Name() != tt.wantName {
				t.Errorf("got.Name() = %q, want %q", got.Name(), tt.wantName)
			}
			if got.Version() != tt.wantVersion {
				t.Errorf("got.Version() = %q, want %q", got.Version(), tt.wantVersion)
			}
		})
	}
}

func TestRegisterExtension_DuplicatesAndInvalidNames(t *testing.T) {
	tests := []struct {
		name        string
		initial     Extension
		duplicate   Extension
		wantErrText string
	}{
		{
			name:        "duplicate same instance",
			initial:     &baseTestExtension{name: "dup-instance", version: "1.0.0"},
			duplicate:   &baseTestExtension{name: "dup-instance", version: "1.0.0"},
			wantErrText: "extension dup-instance is already registered",
		},
		{
			name:        "duplicate name with different version",
			initial:     &baseTestExtension{name: "dup-version", version: "1.0.0"},
			duplicate:   &baseTestExtension{name: "dup-version", version: "2.0.0"},
			wantErrText: "extension dup-version is already registered",
		},
		{
			name:        "duplicate name across different extension types",
			initial:     &testFieldExtension{baseTestExtension: baseTestExtension{name: "poly-ext", version: "1.0.0"}},
			duplicate:   &testAdminExtension{baseTestExtension: baseTestExtension{name: "poly-ext", version: "1.0.0"}},
			wantErrText: "extension poly-ext is already registered",
		},
		{
			name:        "empty name accepted initially then duplicate fails",
			initial:     &baseTestExtension{name: "", version: "1.0.0"},
			duplicate:   &baseTestExtension{name: "", version: "2.0.0"},
			wantErrText: "extension  is already registered",
		},
		{
			name:        "whitespace name accepted initially then duplicate fails",
			initial:     &baseTestExtension{name: "   ", version: "1.0.0"},
			duplicate:   &baseTestExtension{name: "   ", version: "2.0.0"},
			wantErrText: "extension     is already registered",
		},
		{
			name:        "special characters in name accepted initially then duplicate fails",
			initial:     &baseTestExtension{name: "ext/@#$!/v1", version: "1.0.0"},
			duplicate:   &baseTestExtension{name: "ext/@#$!/v1", version: "2.0.0"},
			wantErrText: "extension ext/@#$!/v1 is already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalExtensionRegistry(t)

			if err := RegisterExtension(tt.initial); err != nil {
				t.Fatalf("initial RegisterExtension(%q) failed: %v", tt.initial.Name(), err)
			}

			err := RegisterExtension(tt.duplicate)
			if err == nil {
				t.Fatalf("expected error registering duplicate %q, got nil", tt.duplicate.Name())
			}
			if err.Error() != tt.wantErrText {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErrText)
			}

			// Ensure initial registration is still preserved
			registered, getErr := GetExtension(tt.initial.Name())
			if getErr != nil {
				t.Fatalf("GetExtension(%q) failed after rejected duplicate: %v", tt.initial.Name(), getErr)
			}
			if registered != tt.initial {
				t.Errorf("registered = %p, want original %p", registered, tt.initial)
			}
		})
	}
}

func TestGetExtension(t *testing.T) {
	tests := []struct {
		name       string
		setup      []Extension
		lookupName string
		wantExt    Extension
		wantErr    string
	}{
		{
			name: "existing extension found",
			setup: []Extension{
				&baseTestExtension{name: "auth-provider", version: "1.0.0"},
			},
			lookupName: "auth-provider",
			wantExt:    &baseTestExtension{name: "auth-provider", version: "1.0.0"},
			wantErr:    "",
		},
		{
			name: "lookup among multiple registered extensions",
			setup: []Extension{
				&baseTestExtension{name: "ext-a", version: "1.0"},
				&baseTestExtension{name: "ext-b", version: "2.0"},
				&baseTestExtension{name: "ext-c", version: "3.0"},
			},
			lookupName: "ext-b",
			wantExt:    &baseTestExtension{name: "ext-b", version: "2.0"},
			wantErr:    "",
		},
		{
			name:       "missing extension returns not found error",
			setup:      nil,
			lookupName: "nonexistent",
			wantExt:    nil,
			wantErr:    "extension nonexistent is not registered",
		},
		{
			name: "missing name when other extensions exist",
			setup: []Extension{
				&baseTestExtension{name: "ext-1", version: "1.0"},
			},
			lookupName: "ext-2",
			wantExt:    nil,
			wantErr:    "extension ext-2 is not registered",
		},
		{
			name:       "empty name not registered returns error",
			setup:      nil,
			lookupName: "",
			wantExt:    nil,
			wantErr:    "extension  is not registered",
		},
		{
			name: "case-sensitive lookup returns not found for different casing",
			setup: []Extension{
				&baseTestExtension{name: "MyExtension", version: "1.0.0"},
			},
			lookupName: "myextension",
			wantExt:    nil,
			wantErr:    "extension myextension is not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalExtensionRegistry(t)

			var expectedInstance Extension
			for _, ext := range tt.setup {
				if err := RegisterExtension(ext); err != nil {
					t.Fatalf("setup RegisterExtension(%q) failed: %v", ext.Name(), err)
				}
				if tt.wantExt != nil && ext.Name() == tt.wantExt.Name() {
					expectedInstance = ext
				}
			}

			got, err := GetExtension(tt.lookupName)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("GetExtension(%q) expected error, got nil", tt.lookupName)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("GetExtension(%q) error = %q, want %q", tt.lookupName, err.Error(), tt.wantErr)
				}
				if got != nil {
					t.Errorf("GetExtension(%q) got %v, want nil", tt.lookupName, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetExtension(%q) unexpected error: %v", tt.lookupName, err)
			}
			if got != expectedInstance {
				t.Errorf("GetExtension(%q) = %p, want %p", tt.lookupName, got, expectedInstance)
			}
		})
	}
}

func TestGetAllExtensions(t *testing.T) {
	tests := []struct {
		name       string
		registered []Extension
	}{
		{
			name:       "empty registry returns empty map",
			registered: nil,
		},
		{
			name: "single extension registered",
			registered: []Extension{
				&baseTestExtension{name: "ext-one", version: "1.0.0"},
			},
		},
		{
			name: "multiple extensions of different types registered",
			registered: []Extension{
				&baseTestExtension{name: "base-ext", version: "1.0.0"},
				&testFieldExtension{baseTestExtension: baseTestExtension{name: "field-ext", version: "1.1.0"}},
				&testRelationExtension{baseTestExtension: baseTestExtension{name: "relation-ext", version: "1.2.0"}},
				&testQuerySetExtension{baseTestExtension: baseTestExtension{name: "queryset-ext", version: "1.3.0"}},
				&testManagerExtension{baseTestExtension: baseTestExtension{name: "manager-ext", version: "1.4.0"}},
				&testAdminExtension{baseTestExtension: baseTestExtension{name: "admin-ext", version: "1.5.0"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalExtensionRegistry(t)

			for _, ext := range tt.registered {
				if err := RegisterExtension(ext); err != nil {
					t.Fatalf("RegisterExtension(%q) failed: %v", ext.Name(), err)
				}
			}

			all := GetAllExtensions()
			if all == nil {
				t.Fatalf("GetAllExtensions() returned nil map")
			}
			if len(all) != len(tt.registered) {
				t.Fatalf("GetAllExtensions() returned %d entries, want %d", len(all), len(tt.registered))
			}

			for _, ext := range tt.registered {
				got, exists := all[ext.Name()]
				if !exists {
					t.Errorf("GetAllExtensions() missing expected extension %q", ext.Name())
					continue
				}
				if got != ext {
					t.Errorf("GetAllExtensions()[%q] = %p, want %p", ext.Name(), got, ext)
				}
			}
		})
	}
}

func TestGetAllExtensions_MapIsolation(t *testing.T) {
	resetGlobalExtensionRegistry(t)

	original := &baseTestExtension{name: "original-ext", version: "1.0.0"}
	if err := RegisterExtension(original); err != nil {
		t.Fatalf("RegisterExtension failed: %v", err)
	}

	m1 := GetAllExtensions()
	if len(m1) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(m1))
	}

	// Mutate the returned map
	m1["extra-key"] = &baseTestExtension{name: "extra-key", version: "1.0.0"}
	delete(m1, "original-ext")

	// Verify that a second call returns an unmutated map
	m2 := GetAllExtensions()
	if len(m2) != 1 {
		t.Errorf("GetAllExtensions() affected by mutation; len = %d, want 1", len(m2))
	}
	if _, exists := m2["original-ext"]; !exists {
		t.Errorf("original-ext was removed from internal registry by mutating returned map")
	}
	if _, exists := m2["extra-key"]; exists {
		t.Errorf("extra-key leaked into internal registry by mutating returned map")
	}

	// Verify GetExtension still functions correctly
	got, err := GetExtension("original-ext")
	if err != nil || got != original {
		t.Errorf("GetExtension(original-ext) returned %v, %v; want %p, nil", got, err, original)
	}
	if _, err := GetExtension("extra-key"); err == nil {
		t.Errorf("GetExtension(extra-key) expected error, got nil")
	}
}

func TestExtensionRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	resetGlobalExtensionRegistry(t)

	const numWriters = 50
	const numReaders = 50

	names := make([]string, numWriters)
	for i := 0; i < numWriters; i++ {
		names[i] = fmt.Sprintf("ext_concurrent_%d", i)
	}

	var startWg sync.WaitGroup
	startWg.Add(1)

	var writersWg sync.WaitGroup
	writersWg.Add(numWriters)

	var stopReaders atomic.Bool
	var readersWg sync.WaitGroup
	readersWg.Add(numReaders)

	// Launch writer goroutines
	for i := 0; i < numWriters; i++ {
		go func(idx int) {
			defer writersWg.Done()
			startWg.Wait()
			ext := &baseTestExtension{name: names[idx], version: "1.0.0"}
			if err := RegisterExtension(ext); err != nil {
				t.Errorf("RegisterExtension(%s) unexpected error: %v", names[idx], err)
			}
		}(i)
	}

	// Launch reader goroutines
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer readersWg.Done()
			startWg.Wait()
			for !stopReaders.Load() {
				_ = GetAllExtensions()
				targetName := names[readerID%numWriters]
				_, _ = GetExtension(targetName)
				runtime.Gosched()
			}
		}(i)
	}

	// Release all goroutines simultaneously
	startWg.Done()

	// Wait for writers to complete
	writersWg.Wait()

	// Stop readers and wait
	stopReaders.Store(true)
	readersWg.Wait()

	// Verify all extensions were successfully registered and are readable
	all := GetAllExtensions()
	if len(all) != numWriters {
		t.Fatalf("GetAllExtensions() returned %d entries, want %d", len(all), numWriters)
	}

	for _, name := range names {
		ext, err := GetExtension(name)
		if err != nil {
			t.Errorf("GetExtension(%s) error: %v", name, err)
		}
		if ext == nil || ext.Name() != name {
			t.Errorf("GetExtension(%s) returned %v", name, ext)
		}
	}
}

func TestExtensionRegistry_ConcurrentSameNameRegistration(t *testing.T) {
	resetGlobalExtensionRegistry(t)

	const numGoroutines = 10
	const duplicateName = "race-conflict-ext"

	var startWg sync.WaitGroup
	startWg.Add(1)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errs := make([]error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			startWg.Wait()
			ext := &baseTestExtension{name: duplicateName, version: fmt.Sprintf("1.0.%d", idx)}
			errs[idx] = RegisterExtension(ext)
		}(i)
	}

	// Release all goroutines at the same time
	startWg.Done()
	wg.Wait()

	successCount := 0
	failureCount := 0
	expectedErrMsg := fmt.Sprintf("extension %s is already registered", duplicateName)

	for _, err := range errs {
		if err == nil {
			successCount++
		} else {
			failureCount++
			if err.Error() != expectedErrMsg {
				t.Errorf("unexpected error message: %q, want %q", err.Error(), expectedErrMsg)
			}
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 registration to succeed, got %d", successCount)
	}
	if failureCount != numGoroutines-1 {
		t.Fatalf("expected %d registrations to fail, got %d", numGoroutines-1, failureCount)
	}

	registered, err := GetExtension(duplicateName)
	if err != nil {
		t.Fatalf("GetExtension(%q) failed: %v", duplicateName, err)
	}
	if registered.Name() != duplicateName {
		t.Errorf("registered.Name() = %q, want %q", registered.Name(), duplicateName)
	}
}
