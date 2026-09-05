package testutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var testDBCounter uint64

func generateTestDBName() string {
	seq := atomic.AddUint64(&testDBCounter, 1)
	return fmt.Sprintf("test_identity_%d_%d_%d", os.Getpid(), time.Now().UnixNano(), seq)
}

func isDuplicateDatabaseError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "42P04"
}

func createTestDatabase(ctx context.Context, defaultDB *sql.DB) (string, error) {
	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		dbName := generateTestDBName()
		query := fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(dbName))
		if _, err := defaultDB.ExecContext(ctx, query); err != nil {
			if isDuplicateDatabaseError(err) {
				continue
			}
			return "", err
		}
		return dbName, nil
	}

	return "", fmt.Errorf("failed to create unique test database after %d attempts", maxAttempts)
}

// SetupTestDB creates a Postgres database for testing
func SetupTestDB(t *testing.T) *db.DB {
	// Postgres connection info
	host := "127.0.0.1"
	port := "5432"
	user := "postgres"
	password := "123"

	// Connect to default DB to create test DB
	defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		user, password, host, port)
	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		t.Skipf("PostgreSQL not available: %v. Skipping identity DB tests.", err)
	}
	defer defaultDB.Close()
	if err := defaultDB.Ping(); err != nil {
		t.Skipf("PostgreSQL not available: %v. Skipping identity DB tests.", err)
	}

	// Create a unique database (retrying if a name collision occurs under parallel test startup)
	dbName, err := createTestDatabase(context.Background(), defaultDB)
	require.NoError(t, err)

	// Connect to test DB
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName)
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		t.Skipf("PostgreSQL not available: %v. Skipping identity DB tests.", err)
	}

	testDB := &db.DB{DB: sqlDB, Driver: "postgres"}

	// Run migrations (Postgres syntax)
	// Note: Postgres uses SERIAL for auto-increment, TRUE/FALSE for booleans
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(150) UNIQUE NOT NULL,
			email VARCHAR(254) UNIQUE NOT NULL,
			password VARCHAR(128) NOT NULL,
			first_name VARCHAR(150),
			last_name VARCHAR(150),
			bio TEXT,
			website VARCHAR(255),
			location VARCHAR(255),
			avatar VARCHAR(255),
			phone_number VARCHAR(20),
			phone_verified BOOLEAN DEFAULT FALSE,
			timezone VARCHAR(50),
			locale VARCHAR(10),
			language VARCHAR(10),
			is_active BOOLEAN DEFAULT TRUE,
			is_staff BOOLEAN DEFAULT FALSE,
			is_superuser BOOLEAN DEFAULT FALSE,
			is_locked BOOLEAN DEFAULT FALSE,
			email_verified BOOLEAN DEFAULT FALSE,
			password_changed_at TIMESTAMP,
			password_expires_at TIMESTAMP,
			must_change_password BOOLEAN DEFAULT FALSE,
			locked_at TIMESTAMP,
			locked_reason VARCHAR(255),
			failed_login_attempts INTEGER DEFAULT 0,
			last_failed_login_at TIMESTAMP,
			email_verified_at TIMESTAMP,
			date_joined TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

		CREATE TABLE IF NOT EXISTS user_sessions (
			id SERIAL PRIMARY KEY,
			session_key VARCHAR(255) UNIQUE NOT NULL,
			user_id INTEGER NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			last_activity TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			is_remember_me BOOLEAN DEFAULT FALSE
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_expire_date ON user_sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);

		CREATE TABLE IF NOT EXISTS email_verification_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(254) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			verified_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_email_tokens_token ON email_verification_tokens(token);

		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			used_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_password_tokens_token ON password_reset_tokens(token);
	`)
	require.NoError(t, err)

	// Register cleanup
	t.Cleanup(func() {
		testDB.Close()
		// Reconnect to default to drop
		cleanupDB, err := sql.Open("postgres", defaultDSN)
		if err != nil {
			return
		}
		defer cleanupDB.Close()
		// Use WITH (FORCE) to drop even if connections remain
		query := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pq.QuoteIdentifier(dbName))
		_, _ = cleanupDB.ExecContext(context.Background(), query)
	})

	return testDB
}
