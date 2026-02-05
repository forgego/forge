package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = bcrypt.DefaultCost
	// MinCost is the minimum bcrypt cost
	MinCost = bcrypt.MinCost
	// MaxCost is the maximum bcrypt cost
	MaxCost = bcrypt.MaxCost
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultCost)
}

// HashPasswordWithCost hashes a password with a specific cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword checks if a password matches a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckPasswordHash is an alias for CheckPassword (Django-style naming)
func CheckPasswordHash(password, hash string) bool {
	return CheckPassword(password, hash)
}

// NeedsRehash checks if a password hash needs to be rehashed (e.g., cost changed)
func NeedsRehash(hash string, cost int) bool {
	actualCost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true // If we can't determine cost, assume it needs rehashing
	}
	return actualCost != cost
}
