package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeOrdering_KeepsKnownFields(t *testing.T) {
	valid := map[string]bool{"name": true, "created_at": true, "id": true, "pk": true}

	out := sanitizeOrdering([]string{"name", "-created_at", "id"}, valid)
	assert.Equal(t, []string{"name", "-created_at", "id"}, out)
}

func TestSanitizeOrdering_DropsUnknownAndUnsafe(t *testing.T) {
	valid := map[string]bool{"name": true}

	out := sanitizeOrdering([]string{
		"nonexistent_field", // unknown -> dropped
		"name; DROP TABLE x", // unsafe chars -> dropped
		"name)",              // unsafe chars -> dropped
		"  ",                 // blank -> dropped
		"-name",              // valid desc
	}, valid)
	assert.Equal(t, []string{"-name"}, out)
}

func TestSanitizeOrdering_EmptyWhenNothingValid(t *testing.T) {
	valid := map[string]bool{"name": true}

	assert.Empty(t, sanitizeOrdering(nil, valid))
	assert.Empty(t, sanitizeOrdering([]string{"nope"}, valid))
}

func TestIsSafeOrderField(t *testing.T) {
	assert.True(t, isSafeOrderField("name"))
	assert.True(t, isSafeOrderField("created_at"))
	assert.True(t, isSafeOrderField("a.b"))
	assert.False(t, isSafeOrderField(""))
	assert.False(t, isSafeOrderField("name desc"))
	assert.False(t, isSafeOrderField("name;DROP"))
	assert.False(t, isSafeOrderField("name)"))
	assert.False(t, isSafeOrderField("-name"))
}
