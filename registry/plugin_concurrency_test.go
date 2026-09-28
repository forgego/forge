package registry

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var testCounter atomic.Int64

type testModelPlugin struct {
	*ModelPluginBase
}

func newTestModelPlugin(name string) *testModelPlugin {
	return &testModelPlugin{
		ModelPluginBase: NewModelPluginBase(name, "1.0.0"),
	}
}

type failingInstallPlugin struct {
	*BasePlugin
}

func (p *failingInstallPlugin) Install() error {
	return errors.New("simulated install error")
}

type selfLookupPlugin struct {
	*BasePlugin
	foundPlugin Plugin
	lookupErr   error
}

func (p *selfLookupPlugin) Install() error {
	p.foundPlugin, p.lookupErr = GetPlugin(p.Name())
	return p.lookupErr
}

func TestPluginRegistry_ConcurrentRegistrationAndReads(t *testing.T) {
	const numWriters = 50
	const numReaders = 50

	runID := testCounter.Add(1)
	names := make([]string, numWriters)
	for i := 0; i < numWriters; i++ {
		names[i] = fmt.Sprintf("%s_%d_plugin_%d", t.Name(), runID, i)
	}

	var startWg sync.WaitGroup
	startWg.Add(1)

	var writersWg sync.WaitGroup
	writersWg.Add(numWriters)

	var stopReaders atomic.Bool
	var readersWg sync.WaitGroup
	readersWg.Add(numReaders)

	// Launch 50 writer goroutines
	for i := 0; i < numWriters; i++ {
		go func(idx int) {
			defer writersWg.Done()
			startWg.Wait()
			p := newTestModelPlugin(names[idx])
			if err := RegisterPlugin(p); err != nil {
				t.Errorf("RegisterPlugin(%s) unexpected error: %v", names[idx], err)
			}
		}(i)
	}

	// Launch 50 reader goroutines
	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer readersWg.Done()
			startWg.Wait()
			for !stopReaders.Load() {
				_ = GetAllPlugins()
				_ = GetModelPlugins()
				targetName := names[readerID%numWriters]
				_, _ = GetPlugin(targetName)
				runtime.Gosched()
			}
		}(i)
	}

	// Release all goroutines simultaneously
	startWg.Done()

	// Wait for all writers to complete
	writersWg.Wait()

	// Stop readers and wait
	stopReaders.Store(true)
	readersWg.Wait()

	// Afterwards, verify all 50 plugins are present
	for _, name := range names {
		p, err := GetPlugin(name)
		if err != nil {
			t.Fatalf("expected plugin %s to be registered, got err: %v", name, err)
		}
		if p == nil || p.Name() != name {
			t.Fatalf("unexpected plugin returned for %s: %v", name, p)
		}
	}
}

func TestPluginRegistry_ConcurrentSameName(t *testing.T) {
	runID := testCounter.Add(1)
	pluginName := fmt.Sprintf("%s_%d_conflict", t.Name(), runID)
	var startWg sync.WaitGroup
	startWg.Add(1)

	var wg sync.WaitGroup
	wg.Add(2)

	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			defer wg.Done()
			startWg.Wait()
			p := NewBasePlugin(pluginName, "1.0.0")
			errs[idx] = RegisterPlugin(p)
		}(i)
	}

	startWg.Done()
	wg.Wait()

	successCount := 0
	failureCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	if successCount != 1 || failureCount != 1 {
		t.Fatalf("expected exactly 1 success and 1 failure for same-name registration, got %d successes and %d failures (errors: %v)",
			successCount, failureCount, errs)
	}

	p, err := GetPlugin(pluginName)
	if err != nil {
		t.Fatalf("expected plugin %s to be registered, got error: %v", pluginName, err)
	}
	if p.Name() != pluginName {
		t.Fatalf("expected plugin name %s, got %s", pluginName, p.Name())
	}
}

func TestPluginRegistry_FailedInstallNotRegistered(t *testing.T) {
	runID := testCounter.Add(1)
	name := fmt.Sprintf("%s_%d_failing", t.Name(), runID)
	failing := &failingInstallPlugin{BasePlugin: NewBasePlugin(name, "1.0.0")}

	err := RegisterPlugin(failing)
	if err == nil {
		t.Fatalf("expected error from RegisterPlugin for failing install, got nil")
	}

	// Plugin whose Install fails is not left registered (GetPlugin returns error)
	if _, err := GetPlugin(name); err == nil {
		t.Fatalf("expected GetPlugin(%s) to return error for uninstalled plugin, got nil", name)
	}

	// And its name can be registered again
	good := NewBasePlugin(name, "1.0.0")
	if err := RegisterPlugin(good); err != nil {
		t.Fatalf("expected re-registration of %s to succeed, got: %v", name, err)
	}

	registered, err := GetPlugin(name)
	if err != nil {
		t.Fatalf("expected GetPlugin(%s) to succeed after re-registration, got: %v", name, err)
	}
	if registered.Name() != name {
		t.Fatalf("expected plugin name %s, got %s", name, registered.Name())
	}
}

func TestPluginRegistry_InstallReentrantGetPlugin(t *testing.T) {
	runID := testCounter.Add(1)
	name := fmt.Sprintf("%s_%d_reentrant", t.Name(), runID)
	p := &selfLookupPlugin{BasePlugin: NewBasePlugin(name, "1.0.0")}

	done := make(chan error, 1)
	go func() {
		done <- RegisterPlugin(p)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RegisterPlugin failed: %v", err)
		}
		if p.foundPlugin == nil {
			t.Fatalf("expected plugin to find itself in Install()")
		}
		if p.foundPlugin.Name() != name {
			t.Fatalf("expected found plugin name %s, got %s", name, p.foundPlugin.Name())
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("deadlock detected: RegisterPlugin did not complete within 2s")
	}
}

func TestPluginRegistry_SliceIsolation(t *testing.T) {
	runID := testCounter.Add(1)
	name := fmt.Sprintf("%s_%d_slice_iso", t.Name(), runID)
	p := newTestModelPlugin(name)
	if err := RegisterPlugin(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	s1 := GetModelPlugins()
	if len(s1) == 0 {
		t.Fatalf("expected at least 1 model plugin")
	}

	targetIdx := len(s1) - 1
	saved := s1[targetIdx]
	t.Cleanup(func() {
		// Restore in case test failed before fix
		globalPluginRegistry.mu.Lock()
		globalPluginRegistry.modelPlugins[targetIdx] = saved
		globalPluginRegistry.mu.Unlock()
	})

	// Mutate slice returned by GetModelPlugins
	s1[targetIdx] = nil

	// Check second call's result
	s2 := GetModelPlugins()
	if s2[targetIdx] == nil {
		t.Fatalf("mutating the slice returned by GetModelPlugins changed a second call's result (slice not isolated)")
	}
}
