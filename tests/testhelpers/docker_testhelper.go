package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// PostgresOpts contains options for starting a Postgres container
type PostgresOpts struct {
	Version     string
	User        string
	Password    string
	DBName      string
	Env         []string
	WaitTimeout time.Duration
	DockerHost  string // Optional: Docker endpoint (e.g., "tcp://localhost:2375" or "unix:///var/run/docker.sock")

	// Direct connection options (if set, skips container creation)
	Host      string // Database host (e.g., "127.0.0.1")
	Port      string // Database port (e.g., "5432")
	UseDirect bool   // If true, connect directly without creating container
	RawQuery  string // URL query parameters from FORGE_TEST_DATABASE_URL (e.g. "sslmode=require&connect_timeout=5")
}

// DSN returns the PostgreSQL connection string for opts.
// When FORGE_TEST_DATABASE_URL query parameters are present in opts.RawQuery, they are preserved.
// Otherwise, it defaults to sslmode=disable.
func (opts PostgresOpts) DSN() string {
	applyPostgresPrecedence(&opts)
	query := "sslmode=disable"
	if opts.RawQuery != "" {
		query = opts.RawQuery
	}

	u := &url.URL{
		Scheme:   "postgres",
		RawQuery: query,
	}
	if opts.User != "" {
		if opts.Password != "" {
			u.User = url.UserPassword(opts.User, opts.Password)
		} else {
			u.User = url.User(opts.User)
		}
	}
	host := strings.Trim(opts.Host, "[]")
	if opts.Port != "" {
		u.Host = net.JoinHostPort(host, opts.Port)
	} else if strings.Contains(host, ":") {
		u.Host = "[" + host + "]"
	} else {
		u.Host = host
	}
	if opts.DBName != "" {
		u.Path = "/" + url.PathEscape(opts.DBName)
	}
	return u.String()
}

// DeriveDSN returns the connection string derived from opts, preserving any query parameters.
func DeriveDSN(opts PostgresOpts) string {
	return opts.DSN()
}

// DirectPostgresDSN returns the PostgreSQL connection string for direct postgres connection, applying precedence rules.
func DirectPostgresDSN(opts PostgresOpts) string {
	return opts.DSN()
}

// DefaultPostgresOpts returns sensible defaults
// Generates a unique database name using timestamp to avoid conflicts
func DefaultPostgresOpts() PostgresOpts {
	// Always generate unique DB name to avoid state pollution
	return DefaultPostgresOptsWithTest("")
}

// DefaultPostgresOptsWithTest returns sensible defaults with unique DB name for test
// If POSTGRES_HOST is set, uses direct connection. Otherwise defaults to localhost:5432 with postgres/123
func DefaultPostgresOptsWithTest(testName string) PostgresOpts {
	opts := PostgresOpts{
		Version:     "15",
		User:        "postgres",
		Password:    "123",
		DBName:      "testdb",
		WaitTimeout: 30 * time.Second,
		UseDirect:   true,
		Host:        "127.0.0.1",
		Port:        "5432",
	}

	// Always generate unique DB name to avoid state pollution
	// Use test name if provided, otherwise just use timestamp
	if testName != "" {
		// Create unique DB name: testdb_<testname>_<timestamp>
		opts.DBName = fmt.Sprintf("testdb_%s_%d", sanitizeTestName(testName), time.Now().UnixNano())
	} else {
		// Check if POSTGRES_DB is set (for manual override)
		if dbName := os.Getenv("POSTGRES_DB"); dbName != "" {
			opts.DBName = dbName
		} else {
			// Generate unique name with timestamp
			opts.DBName = fmt.Sprintf("testdb_%d", time.Now().UnixNano())
		}
	}
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
		opts.DockerHost = dockerHost
	}

	opts.DBName = truncateDBName(opts.DBName)
	applyDatabaseURLToOpts(&opts)
	return opts
}

// sanitizeTestName converts test name to valid database name
func sanitizeTestName(name string) string {
	// Replace invalid characters with underscores
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		} else {
			result += "_"
		}
	}
	result = strings.ToLower(result)
	// Limit length to 50 chars (PostgreSQL identifier limit is 63, but we want some room)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func truncateDBName(name string) string {
	const maxDBNameLen = 63
	if len(name) > maxDBNameLen {
		return name[:maxDBNameLen]
	}
	return name
}

// GetDockerEndpoint returns the Docker endpoint to use, trying multiple sources
func GetDockerEndpoint() string {
	// 1. Check DOCKER_HOST environment variable
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
		return dockerHost
	}

	// 2. Return empty string - dockertest will use platform defaults:
	//    - Unix: unix:///var/run/docker.sock
	//    - Windows: npipe:////./pipe/docker_engine
	//    - Remote: tcp://host:port
	return ""
}

// StartPostgresContainer starts an ephemeral Postgres container using Dockertest
// Or connects directly to an existing database if UseDirect is true
func StartPostgresContainer(ctx context.Context, opts PostgresOpts) (*sql.DB, string, func() error, error) {
	applyPostgresPrecedence(&opts)
	// If UseDirect is true, connect directly to existing database
	if opts.UseDirect {
		return startDirectPostgresConnection(ctx, opts)
	}

	if opts.Version == "" {
		opts.Version = "15"
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 30 * time.Second
	}

	// Get Docker endpoint from options, environment variable, or use default
	dockerEndpoint := opts.DockerHost
	if dockerEndpoint == "" {
		dockerEndpoint = GetDockerEndpoint()
	}

	// Use the calculated endpoint (empty string means use platform default)
	pool, err := dockertest.NewPool(dockerEndpoint)
	if err != nil {
		endpointMsg := dockerEndpoint
		if endpointMsg == "" {
			endpointMsg = "default (platform-specific)"
		}
		return nil, "", nil, fmt.Errorf("could not connect to docker at %s: %w\nHint: Set DOCKER_HOST environment variable (e.g., 'tcp://localhost:2375' for remote, 'unix:///var/run/docker.sock' for Unix, or 'npipe:////./pipe/docker_engine' for Windows)", endpointMsg, err)
	}

	// Set max wait time
	pool.MaxWait = opts.WaitTimeout

	env := []string{
		fmt.Sprintf("POSTGRES_USER=%s", opts.User),
		fmt.Sprintf("POSTGRES_PASSWORD=%s", opts.Password),
		fmt.Sprintf("POSTGRES_DB=%s", opts.DBName),
	}
	env = append(env, opts.Env...)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "postgres",
		Tag:          opts.Version,
		Env:          env,
		ExposedPorts: []string{"5432"}, // PostgreSQL listens on 5432 inside container
		NetworkID:    "",               // Use default bridge network
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		// Use bridge network mode to avoid port conflicts
		config.NetworkMode = "bridge"
		// Map container port 5432 to host port 9091 to avoid conflicts
		config.PortBindings = map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostIP: "127.0.0.1", HostPort: "9091"}},
		}
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("could not start resource: %w", err)
	}

	// Get port - prefer the 0.0.0.0 binding (auto-assigned) as it's more reliable
	// The 127.0.0.1:9091 binding requires additional SSH port forwarding
	var port string
	if resource.Container != nil && resource.Container.NetworkSettings != nil {
		if bindings, ok := resource.Container.NetworkSettings.Ports["5432/tcp"]; ok && len(bindings) > 0 {
			// Prefer 0.0.0.0 binding (auto-assigned random port) as it's accessible via SSH
			for _, binding := range bindings {
				if binding.HostIP == "0.0.0.0" || binding.HostIP == "" {
					port = binding.HostPort
					fmt.Printf("[DEBUG] Using auto-assigned port: %s (from %s)\n", port, binding.HostIP)
					break
				}
			}
			// Fallback to 127.0.0.1 binding if 0.0.0.0 not found
			if port == "" {
				for _, binding := range bindings {
					if binding.HostIP == "127.0.0.1" {
						port = binding.HostPort
						fmt.Printf("[DEBUG] Using 127.0.0.1 port: %s (requires SSH forwarding)\n", port)
						break
					}
				}
			}
		}
	}
	// Final fallback
	if port == "" {
		port = resource.GetPort("5432/tcp")
		if port == "" {
			port = "9091"
		}
	}

	// Log container info for debugging
	fmt.Printf("[DEBUG] Container ID: %s\n", resource.Container.ID)
	fmt.Printf("[DEBUG] Container Name: %s\n", resource.Container.Name)
	fmt.Printf("[DEBUG] Mapped port: %s\n", port)
	if resource.Container != nil && resource.Container.NetworkSettings != nil {
		fmt.Printf("[DEBUG] Port bindings: %+v\n", resource.Container.NetworkSettings.Ports)
	}

	// For remote Docker, we need to use the fixed port (9091) that we bound
	// and ensure SSH tunnel is set up for that port
	host := "127.0.0.1"
	isRemoteDocker := dockerEndpoint != "" && len(dockerEndpoint) >= 4 && (dockerEndpoint[:4] == "tcp:" || dockerEndpoint[:4] == "ssh:")

	if isRemoteDocker {
		// For remote Docker, use the fixed port 9091 that we explicitly bound
		// This requires an SSH tunnel: ssh -L 9091:127.0.0.1:9091 sadra50
		port = "9091"
		fmt.Printf("[DEBUG] Remote Docker detected - using fixed port 9091 (ensure SSH tunnel: ssh -L 9091:127.0.0.1:9091 sadra50)\n")
	} else {
		// For local Docker, use the auto-assigned port
		fmt.Printf("[DEBUG] Local Docker - using port: %s\n", port)
	}

	opts.Host = host
	opts.Port = port
	dsn := opts.DSN()
	fmt.Printf("[DEBUG] DSN: %s\n", redactDSN(dsn))

	cleanup := func() error {
		return pool.Purge(resource)
	}

	// Retry connection
	var db *sql.DB
	retries := 0
	maxRetries := 30
	var lastErr error
	for retries < maxRetries {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			if err = db.PingContext(ctx); err == nil {
				fmt.Printf("[DEBUG] Successfully connected to PostgreSQL after %d retries\n", retries)
				break
			}
			lastErr = err
			db.Close()
		} else {
			lastErr = err
		}
		retries++
		if retries%10 == 0 {
			fmt.Printf("[DEBUG] Retry %d/%d: %v\n", retries, maxRetries, lastErr)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if retries >= maxRetries {
		// Log container info for debugging
		if resource.Container != nil {
			fmt.Printf("[DEBUG] Container ID: %s\n", resource.Container.ID)
			fmt.Printf("[DEBUG] Container Name: %s\n", resource.Container.Name)
			fmt.Printf("[DEBUG] Container Status: %s\n", resource.Container.State.Status)
			fmt.Printf("[DEBUG] Container Ports: %+v\n", resource.Container.NetworkSettings)
		}
		cleanup()
		return nil, "", nil, fmt.Errorf("could not connect to postgres after %d retries. Last error: %w. DSN: %s. Container may not be ready or port mapping incorrect", maxRetries, lastErr, redactDSN(dsn))
	}

	return db, dsn, cleanup, nil
}

// StartSQLiteMemory returns an in-memory SQLite connection for tests
func StartSQLiteMemory(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	return db, nil
}

// startDirectPostgresConnection connects directly to an existing PostgreSQL database
func startDirectPostgresConnection(ctx context.Context, opts PostgresOpts) (*sql.DB, string, func() error, error) {
	applyPostgresPrecedence(&opts)

	host := opts.Host
	port := opts.Port
	dbName := opts.DBName

	fmt.Printf("[DEBUG] Connecting directly to PostgreSQL at %s:%s\n", host, port)

	// First, connect to default "postgres" database to create the test database if needed
	defaultDSN := defaultAdminDSN(opts)

	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to open default database connection: %w", err)
	}
	defer defaultDB.Close()

	// Test connection to default database
	if err = defaultDB.PingContext(ctx); err != nil {
		return nil, "", nil, fmt.Errorf("failed to connect to default postgres database: %w", err)
	}

	// Check if test database exists, create if not
	var exists int
	err = defaultDB.QueryRowContext(ctx,
		"SELECT 1 FROM pg_database WHERE datname = $1", dbName).Scan(&exists)
	if err == sql.ErrNoRows {
		// Database doesn't exist, create it
		fmt.Printf("[DEBUG] Database %s does not exist, creating it...\n", dbName)
		_, err = defaultDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		fmt.Printf("[DEBUG] Database %s created successfully\n", dbName)
	} else if err != nil {
		return nil, "", nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Now connect to the test database
	dsn := opts.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection with retries
	retries := 0
	maxRetries := 30
	for retries < maxRetries {
		if err = db.PingContext(ctx); err == nil {
			fmt.Printf("[DEBUG] Successfully connected to PostgreSQL after %d retries\n", retries)
			break
		}
		retries++
		if retries%10 == 0 {
			fmt.Printf("[DEBUG] Retry %d/%d: %v\n", retries, maxRetries, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if retries >= maxRetries {
		db.Close()
		return nil, "", nil, fmt.Errorf("could not connect to postgres after %d retries. Last error: %w. DSN: %s", maxRetries, err, redactDSN(dsn))
	}

	cleanup := func() error {
		// Close main connection first
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing DB connection: %v\n", err)
		}

		// Connect to default DB to drop test DB
		defaultDSN := defaultAdminDSN(opts)
		defaultDB, err := sql.Open("postgres", defaultDSN)
		if err != nil {
			return fmt.Errorf("failed to open default database connection for cleanup: %w", err)
		}
		defer defaultDB.Close()

		// Drop database
		fmt.Printf("[DEBUG] Dropping database %s...\n", dbName)
		// Try with FORCE (Postgres 13+)
		_, err = defaultDB.ExecContext(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName))
		if err != nil {
			// Fallback for older versions or if FORCE syntax fails
			fmt.Printf("[DEBUG] DROP WITH FORCE failed: %v. Retrying without FORCE...\n", err)
			_, err = defaultDB.ExecContext(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		}
		return err
	}

	return db, dsn, cleanup, nil
}

// WaitForDBReady polls the DB connection until ready or timeout
func WaitForDBReady(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := db.PingContext(ctx); err == nil {
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("database not ready after %v", timeout)
			}
		}
	}
}

var dsnPassword = regexp.MustCompile(`(?i)(password\s*=\s*)(?:'(?:[^'\\]|\\.)*'|(?:[^\s\\]|\\.)*)`)

// redactDSN masks passwords before connection strings reach logs or errors.
func redactDSN(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "[redacted invalid DSN]"
		}
		query := u.Query()
		for key := range query {
			if strings.EqualFold(key, "password") {
				query.Set(key, "xxxxx")
			}
		}
		u.RawQuery = query.Encode()
		return u.Redacted()
	}
	return dsnPassword.ReplaceAllString(dsn, "${1}***")
}
