package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateBuilder(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_Set(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	// Set string value (using interface{} since Set doesn't have type parameter)
	ub = ub.Set("name", "New Name")
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, "New Name", ub.updates["name"])

	// Set float value
	ub = ub.Set("price", 29.99)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, 29.99, ub.updates["price"])

	// Set int value
	ub = ub.Set("id", int64(1))
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, int64(1), ub.updates["id"])

	// Set bool value
	ub = ub.Set("available", true)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, true, ub.updates["available"])
}

func TestUpdateBuilder_SetExpr(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	priceField := NewField[float64]("price", "test_table")
	// SetExpr requires an Expression - use the field itself
	// This tests that SetExpr accepts expressions
	ub = ub.SetExpr("price", priceField)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, priceField, ub.updates["price"])
}

func TestUpdateBuilder_SetField(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	sourceField := NewField[string]("name", "test_table")
	// Use a field that exists - set email to name value
	ub = ub.SetField("email", sourceField)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, sourceField, ub.updates["email"])
}

func TestUpdateBuilder_Increment(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	// Increment int64
	ub = ub.Increment("id", int64(1))
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.NotNil(t, ub.updates["id"])

	// Increment float64
	ub = ub.Increment("price", 0.5)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.NotNil(t, ub.updates["price"])
}

func TestUpdateBuilder_Decrement(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	// Decrement int64
	ub = ub.Decrement("id", int64(1))
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.NotNil(t, ub.updates["id"])

	// Decrement float64
	ub = ub.Decrement("price", 0.5)
	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.NotNil(t, ub.updates["price"])
}

func TestUpdateBuilder_Chaining(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	require.NoError(t, err)

	// Chain multiple operations
	ub = ub.
		Set("name", "Updated Name").
		Set("price", 29.99).
		Increment("id", int64(1))

	assert.NotNil(t, ub)
	assert.NoError(t, ub.err)
	assert.Equal(t, "Updated Name", ub.updates["name"])
	assert.Equal(t, 29.99, ub.updates["price"])
	assert.NotNil(t, ub.updates["id"])
}

// GetUpdates is not exported - would need integration test to verify updates
// This is tested indirectly through Execute() in integration tests
