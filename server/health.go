package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthChecker is the interface for health check implementations
type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}

// HealthCheckFunc is a function type that implements HealthChecker
type HealthCheckFunc func(ctx context.Context) error

// CheckHealth implements HealthChecker
func (f HealthCheckFunc) CheckHealth(ctx context.Context) error {
	return f(ctx)
}

// HealthStatus represents the status of a health check
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
	Message   string            `json:"message,omitempty"`
}

// healthRegistry stores registered health checkers
type healthRegistry struct {
	mu        sync.RWMutex
	checkers  map[string]HealthChecker
	startTime time.Time
}

var (
	globalRegistry = &healthRegistry{
		checkers:  make(map[string]HealthChecker),
		startTime: time.Now(),
	}
)

// RegisterHealthCheck registers a health checker with a name
func RegisterHealthCheck(name string, checker HealthChecker) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.checkers[name] = checker
}

// RegisterHealthCheckFunc registers a health check function
func RegisterHealthCheckFunc(name string, fn func(ctx context.Context) error) {
	RegisterHealthCheck(name, HealthCheckFunc(fn))
}

// UnregisterHealthCheck removes a health checker
func UnregisterHealthCheck(name string) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	delete(globalRegistry.checkers, name)
}

// GetUptime returns the server uptime
func GetUptime() time.Duration {
	return time.Since(globalRegistry.startTime)
}

// HealthHandler returns a handler for the health check endpoint
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		globalRegistry.mu.RLock()
		checkers := make(map[string]HealthChecker, len(globalRegistry.checkers))
		for k, v := range globalRegistry.checkers {
			checkers[k] = v
		}
		globalRegistry.mu.RUnlock()

		status := HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		}

		// Run all health checks
		allHealthy := true
		for name, checker := range checkers {
			if err := checker.CheckHealth(ctx); err != nil {
				status.Checks[name] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				status.Checks[name] = "healthy"
			}
		}

		if !allHealthy {
			status.Status = "unhealthy"
		}

		// Add uptime information
		uptime := GetUptime()
		status.Message = fmt.Sprintf("Server uptime: %s", uptime.Round(time.Second))

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}

// ReadinessHandler returns a handler for the readiness check endpoint
// Readiness checks if the service is ready to accept traffic
func ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		globalRegistry.mu.RLock()
		checkers := make(map[string]HealthChecker, len(globalRegistry.checkers))
		for k, v := range globalRegistry.checkers {
			checkers[k] = v
		}
		globalRegistry.mu.RUnlock()

		status := HealthStatus{
			Status:    "ready",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		}

		// Run all readiness checks
		allReady := true
		for name, checker := range checkers {
			if err := checker.CheckHealth(ctx); err != nil {
				status.Checks[name] = "not ready: " + err.Error()
				allReady = false
			} else {
				status.Checks[name] = "ready"
			}
		}

		if !allReady {
			status.Status = "not ready"
		}

		w.Header().Set("Content-Type", "application/json")
		if allReady {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}

// LivenessHandler returns a handler for the liveness check endpoint
// Liveness checks if the service is alive (should be simple and fast)
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := HealthStatus{
			Status:    "alive",
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Server uptime: %s", GetUptime().Round(time.Second)),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}
}

// SimpleHealthHandler returns a simple health check that always returns OK
func SimpleHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
