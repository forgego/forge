package errors

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// CachedResponse represents a cached response for idempotency
type CachedResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Timestamp  time.Time
}

// IdempotencyStore is the interface for idempotency key storage
type IdempotencyStore interface {
	Get(key string) (*CachedResponse, error)
	Set(key string, response *CachedResponse, ttl time.Duration) error
	Delete(key string) error
}

// IdempotencyKey extracts and validates idempotency keys
type IdempotencyKey struct {
	Key   string
	Valid bool
}

// ExtractIdempotencyKey extracts the idempotency key from a request
func ExtractIdempotencyKey(headerName, key string) *IdempotencyKey {
	if key == "" {
		return &IdempotencyKey{Valid: false}
	}

	// Validate key format (should be a valid string, not too long)
	if len(key) > 255 {
		return &IdempotencyKey{Valid: false}
	}

	// Hash the key for storage (optional, for security)
	hashedKey := hashKey(key)

	return &IdempotencyKey{
		Key:   hashedKey,
		Valid: true,
	}
}

// hashKey hashes an idempotency key
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// ValidateNestingDepth validates that nested idempotency keys don't exceed max depth
func ValidateNestingDepth(key string, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("idempotency key nesting depth exceeded: %d > %d", currentDepth, maxDepth)
	}
	return nil
}
