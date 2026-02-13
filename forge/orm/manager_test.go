package orm

import (
	"context"
	"strings"
	"testing"

	"github.com/forgego/forge/errors"
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

func TestNewManagerWithDB(t *testing.T) {
	t.Run("nil db returns ConfigurationError", func(t *testing.T) {
		manager, err := NewManagerWithDB[testModel]("test_table", nil)
		require.Error(t, err)
		assert.Nil(t, manager)

		// Verify it's a ConfigurationError
		var configErr *errors.ConfigurationError
		require.ErrorAs(t, err, &configErr)
		assert.Equal(t, "db", configErr.Field)
		assert.Contains(t, configErr.Message, "database connection is required")
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

func TestManager_Get_WithoutDB_ReturnsConfigurationError(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()
	_, err = manager.Get(ctx, 1)
	require.Error(t, err)

	// Verify it's a ConfigurationError, not NotImplementedError
	var configErr *errors.ConfigurationError
	require.ErrorAs(t, err, &configErr)
	assert.Equal(t, "db", configErr.Field)
	assert.Contains(t, configErr.Message, "database connection not set")
}

func TestManager_Create_WithoutDB_ReturnsConfigurationError(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()
	instance := &testModel{ID: 0, Name: "test"}
	err = manager.Create(ctx, instance)
	require.Error(t, err)

	// Verify it's a ConfigurationError, not NotImplementedError
	var configErr *errors.ConfigurationError
	require.ErrorAs(t, err, &configErr)
	assert.Equal(t, "db", configErr.Field)
}

func TestManager_Update_WithoutDB_ReturnsConfigurationError(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()
	instance := &testModel{ID: 1, Name: "test"}
	err = manager.Update(ctx, instance)
	require.Error(t, err)

	// Verify it's a ConfigurationError, not NotImplementedError
	var configErr *errors.ConfigurationError
	require.ErrorAs(t, err, &configErr)
	assert.Equal(t, "db", configErr.Field)
}

func TestManager_Delete_WithoutDB_ReturnsConfigurationError(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()
	instance := &testModel{ID: 1, Name: "test"}
	err = manager.Delete(ctx, instance)
	require.Error(t, err)

	// Verify it's a ConfigurationError, not NotImplementedError
	var configErr *errors.ConfigurationError
	require.ErrorAs(t, err, &configErr)
	assert.Equal(t, "db", configErr.Field)
}

func TestManager_BulkCreate_WithoutDB_ReturnsConfigurationError(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()
	instances := []*testModel{{ID: 0, Name: "test1"}, {ID: 0, Name: "test2"}}
	err = manager.BulkCreate(ctx, instances)
	require.Error(t, err)

	// Verify it's a ConfigurationError, not NotImplementedError
	var configErr *errors.ConfigurationError
	require.ErrorAs(t, err, &configErr)
	assert.Equal(t, "db", configErr.Field)
}

func TestManager_ErrorMessagesAreDescriptive(t *testing.T) {
	manager, err := NewManager[testModel]("test_table")
	require.NoError(t, err)

	ctx := context.Background()

	// Test Get error message
	_, err = manager.Get(ctx, 1)
	require.Error(t, err)
	errStr := err.Error()
	if !strings.Contains(errStr, "configuration error") {
		t.Errorf("Expected error to contain 'configuration error', got: %s", errStr)
	}
	if !strings.Contains(errStr, "db") {
		t.Errorf("Expected error to contain field 'db', got: %s", errStr)
	}
}
