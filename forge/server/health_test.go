package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"


	"github.com/stretchr/testify/assert"
)

func TestHealthHandlers(t *testing.T) {
	// Let's reset the health checkers since they're global to avoid cross-test pollution
	t.Cleanup(func() {
		globalRegistry.mu.Lock()
		globalRegistry.checkers = make(map[string]HealthChecker)
		globalRegistry.mu.Unlock()
	})

	t.Run("HealthHandler with no checkers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler := HealthHandler()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])
	})

	t.Run("HealthHandler with passing checker", func(t *testing.T) {
		RegisterHealthCheckFunc("test_db", func(ctx context.Context) error {
			return nil
		})

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler := HealthHandler()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])

		checks, ok := resp["checks"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "healthy", checks["test_db"])
	})

	t.Run("HealthHandler with failing checker", func(t *testing.T) {
		RegisterHealthCheckFunc("fail_test", func(ctx context.Context) error {
			return fmt.Errorf("connection failed")
		})

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler := HealthHandler()
		handler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "unhealthy", resp["status"])

		checks, ok := resp["checks"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "unhealthy: connection failed", checks["fail_test"])
	})

	t.Run("ReadinessHandler with passing checker", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		// Unregister fail_test
		UnregisterHealthCheck("fail_test")

		handler := ReadinessHandler()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "ready", resp["status"])
	})

	t.Run("ReadinessHandler with failing checker", func(t *testing.T) {
		RegisterHealthCheckFunc("fail_ready", func(ctx context.Context) error {
			return fmt.Errorf("not initialized")
		})

		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()

		handler := ReadinessHandler()
		handler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "not ready", resp["status"])
	})

	t.Run("LivenessHandler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()

		handler := LivenessHandler()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "alive", resp["status"])
	})

	t.Run("SimpleHealthHandler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/simple", nil)
		w := httptest.NewRecorder()

		handler := SimpleHealthHandler()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("Uptime calculation", func(t *testing.T) {
		// Just call it to verify it returns a non-negative duration
		uptime := GetUptime()
		assert.True(t, uptime >= 0)
	})
}

func TestHealthCheckRegistration(t *testing.T) {
	// Reset state
	t.Cleanup(func() {
		globalRegistry.mu.Lock()
		globalRegistry.checkers = make(map[string]HealthChecker)
		globalRegistry.mu.Unlock()
	})

	// Test Registration and Unregistration
	RegisterHealthCheckFunc("test1", func(ctx context.Context) error { return nil })

	globalRegistry.mu.RLock()
	_, exists := globalRegistry.checkers["test1"]
	globalRegistry.mu.RUnlock()
	assert.True(t, exists)

	UnregisterHealthCheck("test1")

	globalRegistry.mu.RLock()
	_, exists = globalRegistry.checkers["test1"]
	globalRegistry.mu.RUnlock()
	assert.False(t, exists)
}

func TestCheckHealthMethod(t *testing.T) {
	called := false
	fn := HealthCheckFunc(func(ctx context.Context) error {
		called = true
		return nil
	})

	err := fn.CheckHealth(context.Background())
	assert.NoError(t, err)
	assert.True(t, called)
}
