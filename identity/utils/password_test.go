package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_And_IsHashed(t *testing.T) {
	plain := "my-secret-password"

	assert.False(t, IsHashed(plain))

	hash, err := HashPassword(plain)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, IsHashed(hash))

	// Re-hashing an existing hash must return the same hash (no double-hashing)
	hashAgain, err := HashPassword(hash)
	require.NoError(t, err)
	assert.Equal(t, hash, hashAgain)

	// CheckPassword works with original plain
	assert.True(t, CheckPassword(plain, hash))
	assert.False(t, CheckPassword("wrong-password", hash))
}
