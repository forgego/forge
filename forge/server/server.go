package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/log"
	"go.uber.org/zap"
)

// Server wraps http.Server with framework-specific functionality
type Server struct {
	*http.Server
	router   *Router
	logger   *log.Logger
	config   *config.Config
	settings *config.Settings
}

// NewServer creates a new framework server
func NewServer(cfg *config.Config, settings *config.Settings, logger *log.Logger) (*Server, error) {
	// Create router
	router := NewRouter()

	// Add request ID middleware (should be first)
	router.Use(RequestIDMiddleware())

	// Add error handler middleware (should be early in the stack)
	if logger != nil {
		errorHandlerOpts := DefaultErrorHandlerOptions()
		errorHandlerOpts.Logger = logger.Logger
		router.Use(ErrorHandler(errorHandlerOpts))
	}

	// Add logging middleware if available
	if logger != nil {
		router.Use(log.Middleware(logger))
	}

	// Add request size limit if configured
	if settings.Server.MaxRequestSize > 0 {
		router.Use(RequestSizeLimit(settings.Server.MaxRequestSize))
	}

	// Add session middleware if configured
	if settings.Security.SessionSecret != "" {
		sessionManager := NewSessionManager([]byte(settings.Security.SessionSecret))
		router.Use(sessionManager.Middleware())
	}

	// Add CSRF middleware for non-GET requests
	if settings.Security.CSRFSecretKey != "" {
		csrfProtect := NewCSRF(
			[]byte(settings.Security.CSRFSecretKey),
			DefaultCSRFOptions()...,
		)
		router.Use(csrfProtect.Middleware())
	}

	// Register health check endpoints
	if settings.Server.HealthCheckPath != "" {
		router.Get(settings.Server.HealthCheckPath, HealthHandler())
		router.Get(settings.Server.HealthCheckPath+"/ready", ReadinessHandler())
		router.Get(settings.Server.HealthCheckPath+"/live", LivenessHandler())
	}

	// Register metrics endpoint if enabled
	if settings.Server.MetricsEnabled && settings.Server.MetricsPath != "" {
		router.Get(settings.Server.MetricsPath, MetricsHandler())
	}

	// Register server info endpoint
	router.Get("/info", ServerInfoHandler(settings))

	// Serve static files if configured
	if settings.Server.StaticFilesPath != "" {
		router.Mount("/static", StaticFiles("/static", settings.Server.StaticFilesPath))
	}

	// Enable profiling if configured (dev mode only)
	if settings.Server.EnableProfiling && settings.App.Debug {
		router.Use(Profiler("/debug", true))
	}

	// Create server
	addr := settings.Server.Host + ":" + settings.Server.Port
	server := &Server{
		Server: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  time.Duration(settings.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(settings.Server.WriteTimeout) * time.Second,
		},
		router:   router,
		logger:   logger,
		config:   cfg,
		settings: settings,
	}

	return server, nil
}

// RegisterRoutes registers routes on the server's router
func (s *Server) RegisterRoutes(fn func(*Router)) {
	fn(s.router)
}

// Start starts the server
func (s *Server) Start() error {
	if s.logger != nil {
		s.logger.Info("Starting server",
			zap.String("address", s.Addr),
			zap.String("environment", s.settings.App.Env),
			zap.String("version", s.settings.App.Version),
			zap.Bool("debug", s.settings.App.Debug),
			zap.String("health_check", s.settings.Server.HealthCheckPath),
			zap.Bool("metrics_enabled", s.settings.Server.MetricsEnabled),
		)
	}

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// StartWithGracefulShutdown starts the server with graceful shutdown support
func (s *Server) StartWithGracefulShutdown() error {
	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	select {
	case err := <-serverErr:
		return err
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info("Shutting down server")
	}

	timeout := time.Duration(s.settings.Server.GracefulTimeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.Server.Shutdown(shutdownCtx)
}

// ServerInfoHandler returns a handler for server info endpoint
func ServerInfoHandler(settings *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"name":        settings.App.Name,
			"version":     settings.App.Version,
			"environment": settings.App.Env,
			"debug":       settings.App.Debug,
			"uptime":      GetUptime().String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

// MetricsHandler returns a handler for metrics endpoint
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple metrics response
		// In production, integrate with a proper metrics library like prometheus
		metrics := map[string]interface{}{
			"uptime": GetUptime().String(),
			// Add more metrics as needed
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}
