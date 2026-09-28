package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValuesQuerySet_PropagatesErrors(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// Create a filter with an invalid field name to trigger validation error
	invalidField := NewField[string]("non_existent_field", "test_table")
	expr := invalidField.Eq("value")

	// Filter should store the error in the queryset
	filtered := qs.Filter(expr)

	// Values() should return a ValuesQuerySet that wraps the errored queryset
	valuesQS := filtered.Values("id", "name")

	// All() should return the error
	results, err := valuesQS.All(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, results)

	// Get() should return the error
	result, err := valuesQS.Get(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, result)

	// First() should return the error
	result, err = valuesQS.First(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, result)
}

func TestValuesListQuerySet_PropagatesErrors(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// Create a filter with an invalid field name to trigger validation error
	invalidField := NewField[string]("non_existent_field", "test_table")
	expr := invalidField.Eq("value")

	// Filter should store the error in the queryset
	filtered := qs.Filter(expr)

	// ValuesList() should return a ValuesListQuerySet that wraps the errored queryset
	valuesListQS := filtered.ValuesList("id", "name")

	// All() should return the error
	results, err := valuesListQS.All(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, results)

	// Get() should return the error
	result, err := valuesListQS.Get(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, result)

	// First() should return the error
	result, err = valuesListQS.First(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter expression")
	assert.Nil(t, result)
}
