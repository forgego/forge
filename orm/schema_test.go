package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetModelSchema(t *testing.T) {
	t.Run("get schema for test model", func(t *testing.T) {
		schema, err := GetModelSchema[testModel]()
		require.NoError(t, err)
		assert.NotNil(t, schema)
	})
}

func TestNewFieldAccessor(t *testing.T) {
	t.Run("create field accessor", func(t *testing.T) {
		fa, err := NewFieldAccessor[testModel]()
		require.NoError(t, err)
		assert.NotNil(t, fa)
	})
}

func TestFieldAccessor_Field(t *testing.T) {
	fa, err := NewFieldAccessor[testModel]()
	require.NoError(t, err)

	t.Run("get string field", func(t *testing.T) {
		// Field method requires type parameter - use FieldFor helper instead
		field, err := FieldFor[testModel, string](fa, "name")
		assert.NoError(t, err)
		assert.NotNil(t, field)
	})

	t.Run("get float64 field", func(t *testing.T) {
		field, err := FieldFor[testModel, float64](fa, "price")
		assert.NoError(t, err)
		assert.NotNil(t, field)
	})

	t.Run("get int64 field", func(t *testing.T) {
		field, err := FieldFor[testModel, int64](fa, "id")
		assert.NoError(t, err)
		assert.NotNil(t, field)
	})

	t.Run("get bool field", func(t *testing.T) {
		field, err := FieldFor[testModel, bool](fa, "available")
		assert.NoError(t, err)
		assert.NotNil(t, field)
	})
}

func TestFieldFor(t *testing.T) {
	fa, err := NewFieldAccessor[testModel]()
	require.NoError(t, err)

	t.Run("get field with FieldFor helper", func(t *testing.T) {
		field, err := FieldFor[testModel, string](fa, "name")
		assert.NoError(t, err)
		assert.NotNil(t, field)
	})
}

func TestModelSchema_TableName(t *testing.T) {
	schema, err := GetModelSchema[testModel]()
	require.NoError(t, err)

	assert.NotEmpty(t, schema.TableName)
}

func TestModelSchema_Fields(t *testing.T) {
	schema, err := GetModelSchema[testModel]()
	require.NoError(t, err)

	assert.NotNil(t, schema.Fields)
}
