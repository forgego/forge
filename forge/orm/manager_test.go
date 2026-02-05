package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("create manager", func(t *testing.T) {
		manager, err := NewManager[testModel]("test_table")
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, "test_table", manager.tableName)
	})

	t.Run("create manager with empty table name", func(t *testing.T) {
		manager, err := NewManager[testModel]("")
		// Should use schema table name if available
		if err == nil {
			assert.NotNil(t, manager)
		}
	})
}

func TestManager_SetDB(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	// SetDB should not panic even with nil
	manager.SetDB(nil)
	assert.NotNil(t, manager)
}

func TestManager_GetFieldAccessor(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	fa, err := manager.FieldAccessor()
	require.NoError(t, err)
	assert.NotNil(t, fa)
}

func TestManager_Filter(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	priceField := NewField[float64]("price", "test_table")
	expr := priceField.Gt(10.0)

	qs, err := manager.Filter(expr)
	require.NoError(t, err)
	assert.NotNil(t, qs)
}

func TestManager_Filter_WithoutDB(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	priceField := NewField[float64]("price", "test_table")
	expr := priceField.Gt(10.0)

	// Should work even without DB set
	qs, err := manager.Filter(expr)
	require.NoError(t, err)
	assert.NotNil(t, qs)
}
