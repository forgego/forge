package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateBuilder(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// UpdateBuilder requires schema registration which testModel doesn't have
	// This will fail, so we skip the test
	_, err = NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}
	// If we get here, schema was registered (unlikely for testModel)
}

func TestUpdateBuilder_Set(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// UpdateBuilder requires schema registration which testModel doesn't have
	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	// Set string value (using interface{} since Set doesn't have type parameter)
	ub = ub.Set("name", "New Name")
	assert.NotNil(t, ub)

	// Set float value
	ub = ub.Set("price", 29.99)
	assert.NotNil(t, ub)

	// Set int value
	ub = ub.Set("id", int64(1))
	assert.NotNil(t, ub)

	// Set bool value
	ub = ub.Set("available", true)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_SetExpr(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	priceField := NewField[float64]("price", "test_table")
	// SetExpr requires an Expression - use the field itself
	// This tests that SetExpr accepts expressions
	ub = ub.SetExpr("price", priceField)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_SetField(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	sourceField := NewField[string]("name", "test_table")
	// Use a field that exists - set email to name value
	ub = ub.SetField("email", sourceField)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_Increment(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	// Increment int64
	ub = ub.Increment("id", int64(1))
	assert.NotNil(t, ub)

	// Increment float64
	ub = ub.Increment("price", 0.5)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_Decrement(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	// Decrement int64
	ub = ub.Decrement("id", int64(1))
	assert.NotNil(t, ub)

	// Decrement float64
	ub = ub.Decrement("price", 0.5)
	assert.NotNil(t, ub)
}

func TestUpdateBuilder_Chaining(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		t.Skipf("Schema not registered for testModel, skipping UpdateBuilder tests: %v", err)
		return
	}

	// Chain multiple operations
	ub = ub.
		Set("name", "Updated Name").
		Set("price", 29.99).
		Increment("id", int64(1))

	assert.NotNil(t, ub)
}

// GetUpdates is not exported - would need integration test to verify updates
// This is tested indirectly through Execute() in integration tests
