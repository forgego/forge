package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	forgedb "github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
)

// InMemoryStore is an in-memory idempotency store for tests only; production uses the database store.
type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]*cachedEntry
}

type cachedEntry struct {
	response *CachedResponse
	expires  time.Time
}

// NewInMemoryStore creates a new in-memory store for tests only; production uses the database store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: make(map[string]*cachedEntry),
	}
}

var (
	_ IdempotencyStore = (*InMemoryStore)(nil)
	_ IdempotencyStore = (*DatabaseStore)(nil)
)

// Get retrieves a cached response
func (s *InMemoryStore) Get(key string) (*CachedResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.store[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	// Check expiration
	if time.Now().After(entry.expires) {
		delete(s.store, key)
		return nil, fmt.Errorf("key expired")
	}

	return entry.response, nil
}

// Set stores a cached response
func (s *InMemoryStore) Set(key string, response *CachedResponse, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = &cachedEntry{
		response: response,
		expires:  time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a cached response
func (s *InMemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, key)
	return nil
}

// RedisStore is a Redis-based idempotency store (for production/distributed systems)
// To use Redis, add a Redis client library such as github.com/go-redis/redis/v8
type RedisStore struct {
	client    interface{} // Redis client interface
	keyPrefix string
}

// RedisClient interface for Redis operations
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Del(key string) error
}

// NewRedisStore creates a new Redis store
// Example: client := redis.NewClient(&redis.Options{Addr: addr, DB: db})
//
//	store := NewRedisStore(client, "idempotency:")
func NewRedisStore(client interface{}, keyPrefix string) (*RedisStore, error) {
	if keyPrefix == "" {
		keyPrefix = "idempotency:"
	}

	return &RedisStore{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

// Get retrieves a cached response from Redis
func (s *RedisStore) Get(key string) (*CachedResponse, error) {
	fullKey := s.keyPrefix + key

	// Skeleton implementation - requires Redis client library
	// With go-redis: data, err := s.client.(*redis.Client).Get(ctx, fullKey).Bytes()
	_ = fullKey
	return nil, fmt.Errorf("redis client not configured - add go-redis library")
}

// Set stores a cached response in Redis
func (s *RedisStore) Set(key string, response *CachedResponse, ttl time.Duration) error {
	fullKey := s.keyPrefix + key

	// Serialize response
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Skeleton implementation - requires Redis client library
	// With go-redis: return s.client.(*redis.Client).Set(ctx, fullKey, data, ttl).Err()
	_ = fullKey
	_ = data
	return fmt.Errorf("redis client not configured - add go-redis library")
}

// Delete removes a cached response from Redis
func (s *RedisStore) Delete(key string) error {
	fullKey := s.keyPrefix + key

	// Skeleton implementation - requires Redis client library
	// With go-redis: return s.client.(*redis.Client).Del(ctx, fullKey).Err()
	_ = fullKey
	return fmt.Errorf("redis client not configured - add go-redis library")
}

// DatabaseStore is a database-backed idempotency store. Call EnsureSchema before use.
type DatabaseStore struct {
	db      *sql.DB
	table   string
	dialect dialect.Dialect
}

// NewDatabaseStore creates a new database store
func NewDatabaseStore(database interface{}, table string) (*DatabaseStore, error) {
	var conn *sql.DB
	var d dialect.Dialect
	switch value := database.(type) {
	case *sql.DB:
		conn = value
	case *forgedb.DB:
		if value != nil {
			conn = value.DB
			d = value.Dialect()
		}
	}
	if conn == nil {
		return nil, fmt.Errorf("idempotency database must be a non-nil *sql.DB or *db.DB")
	}
	if table == "sqlite" || table == "sqlite3" {
		d = dialect.NewSQLiteDialect()
		table = "idempotency_keys"
	}
	if table == "postgres" || table == "postgresql" {
		d = dialect.NewPostgreSQLDialect()
		table = "idempotency_keys"
	}
	if table == "" {
		table = "idempotency_keys"
	}
	if !validSQLIdentifier(table) {
		return nil, fmt.Errorf("invalid idempotency table name %q", table)
	}
	if d == nil {
		d = dialect.NewPostgreSQLDialect()
	}
	return &DatabaseStore{db: conn, table: table, dialect: d}, nil
}

func validSQLIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// EnsureSchema creates the idempotency table if it does not already exist.
func (s *DatabaseStore) EnsureSchema(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("idempotency database is not configured")
	}
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (key TEXT PRIMARY KEY, response_json TEXT, expires_at TIMESTAMP NOT NULL)`, s.table))
	return err
}

func (s *DatabaseStore) placeholder(n int) string { return s.dialect.Placeholder(n) }

// Claim atomically reserves a key until ttl elapses. Expired reservations can be reclaimed.
func (s *DatabaseStore) Claim(key string, ttl time.Duration) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("idempotency database is not configured")
	}
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	query := fmt.Sprintf(`INSERT INTO "%s" (key, response_json, expires_at) VALUES (%s, NULL, %s) ON CONFLICT (key) DO UPDATE SET response_json = NULL, expires_at = excluded.expires_at WHERE "%s".expires_at <= %s`, s.table, s.placeholder(1), s.placeholder(2), s.table, s.placeholder(3))
	result, err := s.db.Exec(query, key, time.Now().Add(ttl), time.Now())
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

// Get retrieves a cached response from database
func (s *DatabaseStore) Get(key string) (*CachedResponse, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("idempotency database is not configured")
	}
	var payload sql.NullString
	var expires time.Time
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	err := s.db.QueryRow(fmt.Sprintf(`SELECT response_json, expires_at FROM "%s" WHERE key = %s`, s.table, s.placeholder(1)), key).Scan(&payload, &expires)
	if err != nil {
		return nil, err
	}
	if time.Now().After(expires) {
		// table is validated by validSQLIdentifier; values are bound parameters.
		// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
		if _, err := s.db.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE key = %s`, s.table, s.placeholder(1)), key); err != nil {
			return nil, fmt.Errorf("delete expired idempotency key: %w", err)
		}
		return nil, fmt.Errorf("key expired")
	}
	if !payload.Valid {
		return nil, fmt.Errorf("response not completed")
	}
	var response CachedResponse
	if err := json.Unmarshal([]byte(payload.String), &response); err != nil {
		return nil, fmt.Errorf("decode cached response: %w", err)
	}
	return &response, nil
}

// Set stores a cached response in database
func (s *DatabaseStore) Set(key string, response *CachedResponse, ttl time.Duration) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("idempotency database is not configured")
	}
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	query := fmt.Sprintf(`INSERT INTO "%s" (key, response_json, expires_at) VALUES (%s, %s, %s) ON CONFLICT (key) DO UPDATE SET response_json = excluded.response_json, expires_at = excluded.expires_at`, s.table, s.placeholder(1), s.placeholder(2), s.placeholder(3))
	_, err = s.db.Exec(query, key, string(data), time.Now().Add(ttl))
	return err
}

// Delete removes a cached response from database
func (s *DatabaseStore) Delete(key string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("idempotency database is not configured")
	}
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	_, err := s.db.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE key = %s`, s.table, s.placeholder(1)), key)
	return err
}

// Cleanup removes expired entries from database
func (s *DatabaseStore) Cleanup() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("idempotency database is not configured")
	}
	// table is validated by validSQLIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	_, err := s.db.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE expires_at <= %s`, s.table, s.placeholder(1)), time.Now())
	return err
}
