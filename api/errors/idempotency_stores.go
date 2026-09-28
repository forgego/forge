package errors

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	forgeerrors "github.com/forgego/forge/errors"
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

// DatabaseStore is a database-based idempotency store
// NOT IMPLEMENTED: see NewDatabaseStore.
type DatabaseStore struct{}

// NewDatabaseStore creates a new database store
func NewDatabaseStore(_ interface{}, _ string) (*DatabaseStore, error) {
	return nil, forgeerrors.NewNotImplementedError("idempotency DatabaseStore")
}

// Get retrieves a cached response from database
func (s *DatabaseStore) Get(key string) (*CachedResponse, error) {
	return nil, forgeerrors.NewNotImplementedError("idempotency DatabaseStore")
}

// Set stores a cached response in database
func (s *DatabaseStore) Set(key string, response *CachedResponse, ttl time.Duration) error {
	return forgeerrors.NewNotImplementedError("idempotency DatabaseStore")
}

// Delete removes a cached response from database
func (s *DatabaseStore) Delete(key string) error {
	return forgeerrors.NewNotImplementedError("idempotency DatabaseStore")
}

// Cleanup removes expired entries from database
func (s *DatabaseStore) Cleanup() error {
	return forgeerrors.NewNotImplementedError("idempotency DatabaseStore")
}
