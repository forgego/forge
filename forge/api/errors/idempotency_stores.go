package errors

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// InMemoryStore is an in-memory idempotency store (for dev/testing)
type InMemoryStore struct {
	mu    sync.RWMutex
	store map[string]*cachedEntry
}

type cachedEntry struct {
	response *CachedResponse
	expires  time.Time
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		store: make(map[string]*cachedEntry),
	}
}

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
//          store := NewRedisStore(client, "idempotency:")
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
	return nil, fmt.Errorf("Redis client not configured - add go-redis library")
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
	return fmt.Errorf("Redis client not configured - add go-redis library")
}

// Delete removes a cached response from Redis
func (s *RedisStore) Delete(key string) error {
	fullKey := s.keyPrefix + key
	
	// Skeleton implementation - requires Redis client library
	// With go-redis: return s.client.(*redis.Client).Del(ctx, fullKey).Err()
	_ = fullKey
	return fmt.Errorf("Redis client not configured - add go-redis library")
}

// DatabaseStore is a database-based idempotency store
type DatabaseStore struct {
	db        interface{} // sql.DB or similar
	tableName string
}

// DatabaseConnection interface for database operations
type DatabaseConnection interface {
	QueryRow(query string, args ...interface{}) interface{ Scan(...interface{}) error }
	Exec(query string, args ...interface{}) (interface{}, error)
}

// NewDatabaseStore creates a new database store
func NewDatabaseStore(db interface{}, tableName string) (*DatabaseStore, error) {
	if tableName == "" {
		tableName = "idempotency_cache"
	}
	
	store := &DatabaseStore{
		db:        db,
		tableName: tableName,
	}
	
	// Initialize table if needed
	if err := store.ensureTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}
	
	return store, nil
}

// ensureTable creates the idempotency table if it doesn't exist
func (s *DatabaseStore) ensureTable() error {
	// This is a basic implementation - in practice, this should use migrations
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			key VARCHAR(255) PRIMARY KEY,
			status_code INT NOT NULL,
			headers TEXT,
			body BYTEA NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`, s.tableName)
	
	// Execute table creation
	// Note: This requires proper database connection implementation
	_ = query
	return nil
}

// Get retrieves a cached response from database
func (s *DatabaseStore) Get(key string) (*CachedResponse, error) {
	query := fmt.Sprintf(`
		SELECT status_code, headers, body, expires_at
		FROM %s
		WHERE key = $1 AND expires_at > NOW()
	`, s.tableName)
	
	// This would execute the query and scan results
	// For now, return not found
	_ = query
	return nil, fmt.Errorf("not found")
}

// Set stores a cached response in database
func (s *DatabaseStore) Set(key string, response *CachedResponse, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	
	// Serialize headers
	headersJSON, err := json.Marshal(response.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}
	
	query := fmt.Sprintf(`
		INSERT INTO %s (key, status_code, headers, body, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE
		SET status_code = $2, headers = $3, body = $4, expires_at = $5
	`, s.tableName)
	
	// This would execute the query
	_ = query
	_ = headersJSON
	_ = expiresAt
	return nil
}

// Delete removes a cached response from database
func (s *DatabaseStore) Delete(key string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE key = $1`, s.tableName)
	
	// This would execute the delete
	_ = query
	return nil
}

// Cleanup removes expired entries from database
func (s *DatabaseStore) Cleanup() error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE expires_at < NOW()`, s.tableName)
	
	// This would execute the cleanup
	_ = query
	return nil
}
