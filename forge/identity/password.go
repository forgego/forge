package identity

import (
	"github.com/forgego/forge/identity/utils"
)

// Re-export password utilities for backward compatibility
const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = utils.DefaultCost
	// MinCost is the minimum bcrypt cost
	MinCost = utils.MinCost
	// MaxCost is the maximum bcrypt cost
	MaxCost = utils.MaxCost
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	return utils.HashPassword(password)
}

// HashPasswordWithCost hashes a password with a specific cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	return utils.HashPasswordWithCost(password, cost)
}

// CheckPassword checks if a password matches a hash
func CheckPassword(password, hash string) bool {
	return utils.CheckPassword(password, hash)
}

// CheckPasswordHash is an alias for CheckPassword (Django-style naming)
func CheckPasswordHash(password, hash string) bool {
	return utils.CheckPasswordHash(password, hash)
}

// NeedsRehash checks if a password hash needs to be rehashed (e.g., cost changed)
func NeedsRehash(hash string, cost int) bool {
	return utils.NeedsRehash(hash, cost)
}
