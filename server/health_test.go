package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		assert.Equal(t, "unhealthy", checks["fail_test"])
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

// Regression for #290: a failing check's error text (which may carry DSNs,
// hosts or driver details) must not reach the public health responses, but
// must still be logged for operators.
func TestHealthResponsesDoNotLeakCheckErrors(t *testing.T) {
	globalRegistry.mu.Lock()
	saved := globalRegistry.checkers
	globalRegistry.checkers = make(map[string]HealthChecker)
	globalRegistry.mu.Unlock()
	t.Cleanup(func() {
		globalRegistry.mu.Lock()
		globalRegistry.checkers = saved
		globalRegistry.mu.Unlock()
	})

	const secret = "dial tcp 10.1.2.3:5432: password=hunter2"
	RegisterHealthCheckFunc("database", func(ctx context.Context) error {
		return errors.New(secret)
	})

	var logBuf bytes.Buffer
	prevOut, prevFlags := stdlog.Writer(), stdlog.Flags()
	stdlog.SetOutput(&logBuf)
	stdlog.SetFlags(0)
	t.Cleanup(func() {
		stdlog.SetOutput(prevOut)
		stdlog.SetFlags(prevFlags)
	})

	cases := []struct {
		path, status string
		handler      http.HandlerFunc
	}{
		{"/health", "unhealthy", HealthHandler()},
		{"/health/ready", "not ready", ReadinessHandler()},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			logBuf.Reset()
			w := httptest.NewRecorder()
			tc.handler(w, httptest.NewRequest(http.MethodGet, tc.path, nil))

			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
			assert.NotContains(t, w.Body.String(), "hunter2")
			assert.NotContains(t, w.Body.String(), "10.1.2.3")

			var resp struct {
				Status string            `json:"status"`
				Checks map[string]string `json:"checks"`
			}
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tc.status, resp.Status)
			assert.Equal(t, tc.status, resp.Checks["database"])

			assert.Contains(t, logBuf.String(), `"database"`)
			assert.Contains(t, logBuf.String(), strconv.Quote(secret))
		})
	}
}
