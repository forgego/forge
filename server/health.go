package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/forgego/forge/db"
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

// HealthRegistry stores registered health checkers for an application.
type HealthRegistry struct {
	mu        sync.RWMutex
	checkers  map[string]HealthChecker
	startTime time.Time
}

// NewHealthRegistry creates an empty health check registry.
func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{
		checkers:  make(map[string]HealthChecker),
		startTime: time.Now(),
	}
}

// Register registers a health checker with a name.
func (r *HealthRegistry) Register(name string, checker HealthChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = checker
}

// RegisterFunc registers a health check function with a name.
func (r *HealthRegistry) RegisterFunc(name string, fn func(ctx context.Context) error) {
	r.Register(name, HealthCheckFunc(fn))
}

// Unregister removes a health checker.
func (r *HealthRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.checkers, name)
}

// Uptime returns the duration since the health registry was created.
func (r *HealthRegistry) Uptime() time.Duration {
	if r == nil {
		return 0
	}
	return time.Since(r.startTime)
}

func (r *HealthRegistry) snapshot() map[string]HealthChecker {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make(map[string]HealthChecker, len(r.checkers))
	for k, v := range r.checkers {
		cp[k] = v
	}
	return cp
}

var (
	defaultHealthRegistry = NewHealthRegistry()
	globalRegistry        = defaultHealthRegistry
)

// mergeCheckers combines two health checker snapshots. A checker in server
// registry wins if both registries define the same name.
func mergeCheckers(defaultCheckers, serverCheckers map[string]HealthChecker) map[string]HealthChecker {
	merged := make(map[string]HealthChecker, len(defaultCheckers)+len(serverCheckers))
	for k, v := range defaultCheckers {
		merged[k] = v
	}
	for k, v := range serverCheckers {
		merged[k] = v
	}
	return merged
}

// RegisterHealthCheck registers a health checker with a name
//
// Deprecated: use (*Server).Health() for per-app checks.
func RegisterHealthCheck(name string, checker HealthChecker) {
	defaultHealthRegistry.Register(name, checker)
}

// RegisterHealthCheckFunc registers a health check function
//
// Deprecated: use (*Server).Health() for per-app checks.
func RegisterHealthCheckFunc(name string, fn func(ctx context.Context) error) {
	defaultHealthRegistry.RegisterFunc(name, fn)
}

// UnregisterHealthCheck removes a health checker
//
// Deprecated: use (*Server).Health() for per-app checks.
func UnregisterHealthCheck(name string) {
	defaultHealthRegistry.Unregister(name)
}

// GetUptime returns the server uptime
func GetUptime() time.Duration {
	return defaultHealthRegistry.Uptime()
}

// logCheckFailure records why a check failed. Values are quoted so a
// check error cannot inject extra log lines.
func logCheckFailure(kind, name string, err error) {
	log.Printf("server: %s check %s failed: %s", kind, strconv.Quote(name), strconv.Quote(err.Error()))
}

// HealthHandler returns a handler for the health check endpoint
func HealthHandler() http.HandlerFunc {
	return healthHandler(defaultHealthRegistry)
}

func healthHandler(r *HealthRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		var checkers map[string]HealthChecker
		if r == defaultHealthRegistry {
			checkers = defaultHealthRegistry.snapshot()
		} else {
			checkers = mergeCheckers(defaultHealthRegistry.snapshot(), r.snapshot())
		}

		status := HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		}

		// Run all health checks
		allHealthy := true
		for name, checker := range checkers {
			if err := checker.CheckHealth(ctx); err != nil {
				// The response is public: report a fixed status and keep
				// the cause, which may name hosts or driver details, in
				// the server log only.
				logCheckFailure("health", name, err)
				status.Checks[name] = "unhealthy"
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
		if r != nil {
			uptime = r.Uptime()
		}
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
	return readinessHandler(nil, defaultHealthRegistry)
}

func readinessHandler(database *db.DB, r *HealthRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		var checkers map[string]HealthChecker
		if r == defaultHealthRegistry {
			checkers = defaultHealthRegistry.snapshot()
		} else {
			checkers = mergeCheckers(defaultHealthRegistry.snapshot(), r.snapshot())
		}

		status := HealthStatus{
			Status:    "ready",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		}

		allReady := true
		if database != nil {
			if err := database.Ping(ctx); err != nil {
				logCheckFailure("readiness", "database", err)
				status.Checks["database"] = "not ready"
				allReady = false
			} else {
				status.Checks["database"] = "ready"
			}
		}

		// Run all readiness checks
		for name, checker := range checkers {
			if err := checker.CheckHealth(ctx); err != nil {
				logCheckFailure("readiness", name, err)
				status.Checks[name] = "not ready"
				allReady = false
			} else if _, ok := status.Checks[name]; !ok {
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
