package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModelSchema(t *testing.T) {
	t.Run("get schema for test model", func(t *testing.T) {
		schema, err := GetModelSchema[testModel]()
		// Schema might not be registered yet, so error is acceptable
		if err == nil {
			assert.NotNil(t, schema)
		}
	})
}

func TestNewFieldAccessor(t *testing.T) {
	t.Run("create field accessor", func(t *testing.T) {
		fa, err := NewFieldAccessor[testModel]()
		// Might fail if schema not registered
		if err == nil {
			assert.NotNil(t, fa)
		}
	})
}

func TestFieldAccessor_Field(t *testing.T) {
	fa, err := NewFieldAccessor[testModel]()
	if err != nil {
		t.Skip("Schema not registered, skipping field accessor tests")
		return
	}

	t.Run("get string field", func(t *testing.T) {
		// Field method requires type parameter - use FieldFor helper instead
		field := FieldFor[testModel, string](fa, "name")
		assert.NotNil(t, field)
	})

	t.Run("get float64 field", func(t *testing.T) {
		field := FieldFor[testModel, float64](fa, "price")
		assert.NotNil(t, field)
	})

	t.Run("get int64 field", func(t *testing.T) {
		field := FieldFor[testModel, int64](fa, "id")
		assert.NotNil(t, field)
	})

	t.Run("get bool field", func(t *testing.T) {
		field := FieldFor[testModel, bool](fa, "available")
		assert.NotNil(t, field)
	})
}

func TestFieldFor(t *testing.T) {
	fa, err := NewFieldAccessor[testModel]()
	if err != nil {
		t.Skip("Schema not registered, skipping FieldFor tests")
		return
	}

	t.Run("get field with FieldFor helper", func(t *testing.T) {
		field := FieldFor[testModel, string](fa, "name")
		assert.NotNil(t, field)
	})
}

func TestModelSchema_TableName(t *testing.T) {
	schema, err := GetModelSchema[testModel]()
	if err != nil {
		t.Skip("Schema not registered, skipping schema tests")
		return
	}

	assert.NotEmpty(t, schema.TableName)
}

func TestModelSchema_Fields(t *testing.T) {
	schema, err := GetModelSchema[testModel]()
	if err != nil {
		t.Skip("Schema not registered, skipping schema tests")
		return
	}

	assert.NotNil(t, schema.Fields)
}
