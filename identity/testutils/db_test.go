package testutils

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTestDBName_Unique(t *testing.T) {
	const iterations = 200
	seen := make(map[string]struct{}, iterations)

	for i := 0; i < iterations; i++ {
		name := generateTestDBName()
		assert.Contains(t, name, "test_identity_")
		_, exists := seen[name]
		assert.False(t, exists, "duplicate generated DB name: %s", name)
		seen[name] = struct{}{}
	}
}

func TestIsDuplicateDatabaseError(t *testing.T) {
	assert.True(t, isDuplicateDatabaseError(&pq.Error{Code: "42P04"}))
	assert.False(t, isDuplicateDatabaseError(&pq.Error{Code: "23505"}))
	assert.False(t, isDuplicateDatabaseError(nil))
}
