package docker

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
	Host      string // Database host (e.g., "192.168.132.50")
	Port      string // Database port (e.g., "5432")
	UseDirect bool   // If true, connect directly without creating container
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
		User:        "testuser",
		Password:    "testpass",
		DBName:      "testdb",
		WaitTimeout: 30 * time.Second,
	}

	// Check for direct connection environment variables
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		// Default to localhost if not set
		host = "localhost"
	}

	opts.UseDirect = true
	opts.Host = host
	opts.Port = os.Getenv("POSTGRES_PORT")
	if opts.Port == "" {
		opts.Port = "5432"
	}
	opts.User = os.Getenv("POSTGRES_USER")
	if opts.User == "" {
		opts.User = "postgres"
	}
	opts.Password = os.Getenv("POSTGRES_PASSWORD")
	if opts.Password == "" {
		opts.Password = "123" // Default to local password
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
	return opts

	// Use DOCKER_HOST from environment if set
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost != "" {
		opts.DockerHost = dockerHost
	}

	// Always generate unique DB name for Docker tests to avoid conflicts
	if testName != "" {
		opts.DBName = fmt.Sprintf("testdb_%s_%d", sanitizeTestName(testName), time.Now().UnixNano())
	} else {
		opts.DBName = fmt.Sprintf("testdb_%d", time.Now().UnixNano())
	}

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
	// Limit length to 50 chars (PostgreSQL identifier limit is 63, but we want some room)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		opts.User, opts.Password, host, port, opts.DBName)
	fmt.Printf("[DEBUG] DSN: %s\n", dsn)

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
		return nil, "", nil, fmt.Errorf("could not connect to postgres after %d retries. Last error: %w. DSN: %s. Container may not be ready or port mapping incorrect", maxRetries, lastErr, dsn)
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
	host := opts.Host
	if host == "" {
		host = "localhost"
	}
	port := opts.Port
	if port == "" {
		port = "5432"
	}
	user := opts.User
	if user == "" {
		user = "postgres"
	}
	password := opts.Password
	if password == "" {
		password = "postgres"
	}
	dbName := opts.DBName
	if dbName == "" {
		dbName = "testdb"
	}

	fmt.Printf("[DEBUG] Connecting directly to PostgreSQL at %s:%s\n", host, port)

	// First, connect to default "postgres" database to create the test database if needed
	defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		user, password, host, port)

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
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName)

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
		return nil, "", nil, fmt.Errorf("could not connect to postgres after %d retries. Last error: %w. DSN: %s", maxRetries, err, dsn)
	}

	cleanup := func() error {
		// Close main connection first
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing DB connection: %v\n", err)
		}

		// Connect to default DB to drop test DB
		defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
			user, password, host, port)
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
