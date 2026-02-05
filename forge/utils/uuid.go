package utils

import (
	"github.com/google/uuid"
)

// NewUUID generates a new UUID
func NewUUID() uuid.UUID {
	return uuid.New()
}

// NewUUIDString generates a new UUID as a string
func NewUUIDString() string {
	return uuid.NewString()
}

// ParseUUID parses a UUID string
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParseUUID parses a UUID string, panics on error
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
