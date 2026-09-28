package orm

import (
	"context"
	stderrors "errors"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errValidationFlagTargetInvalid = stderrors.New("validation flag target is invalid")

type validationFlagTarget struct {
	schema.BaseSchema
	ID               int64  `db:"id"`
	Name             string `db:"name"`
	validationCalls  *int
	invalidateOnSave bool
}

func (validationFlagTarget) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (validationFlagTarget) Meta() schema.Meta {
	return schema.Meta{TableName: "validation_flag_targets"}
}

func (m *validationFlagTarget) Validate() error {
	if m.validationCalls != nil {
		*m.validationCalls++
	}
	if m.Name == "invalid" {
		return errValidationFlagTargetInvalid
	}
	return nil
}

func (m *validationFlagTarget) BeforeSave(context.Context) error {
	if m.invalidateOnSave {
		m.Name = "invalid"
	}
	return nil
}

type validationFlagCreateParent struct {
	schema.BaseSchema
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	targetManager *Manager[validationFlagTarget]
	targetCalls   *int
}

func (validationFlagCreateParent) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (validationFlagCreateParent) Meta() schema.Meta {
	return schema.Meta{TableName: "validation_flag_create_parents"}
}

func (m *validationFlagCreateParent) BeforeCreate(ctx context.Context) error {
	return m.targetManager.Create(ctx, &validationFlagTarget{Name: "invalid", validationCalls: m.targetCalls})
}

type validationFlagUpdateParent struct {
	schema.BaseSchema
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	targetManager *Manager[validationFlagTarget]
	targetCalls   *int
}

func (validationFlagUpdateParent) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

func (validationFlagUpdateParent) Meta() schema.Meta {
	return schema.Meta{TableName: "validation_flag_update_parents"}
}

func (m *validationFlagUpdateParent) BeforeSave(ctx context.Context) error {
	return m.targetManager.Update(ctx, &validationFlagTarget{ID: 1, Name: "invalid", validationCalls: m.targetCalls})
}

func setupValidationContextDB(t *testing.T) *db.DB {
	t.Helper()

	database, err := db.NewDB(filepath.Join(t.TempDir(), "validation_context.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	_, err = database.Exec(`
		CREATE TABLE validation_flag_targets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
		CREATE TABLE validation_flag_create_parents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
		CREATE TABLE validation_flag_update_parents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
	`)
	require.NoError(t, err)

	return database
}

func TestManager_Create_ValidationFlagDoesNotLeakIntoHooks(t *testing.T) {
	database := setupValidationContextDB(t)
	targetManager, err := NewManagerWithDB[validationFlagTarget]("validation_flag_targets", database)
	require.NoError(t, err)
	parentManager, err := NewManagerWithDB[validationFlagCreateParent]("validation_flag_create_parents", database)
	require.NoError(t, err)

	targetValidationCalls := 0
	parent := &validationFlagCreateParent{
		Name:          "parent",
		targetManager: targetManager,
		targetCalls:   &targetValidationCalls,
	}
	err = parentManager.Create(WithModelValidationCompleted(context.Background()), parent)
	require.ErrorIs(t, err, errValidationFlagTargetInvalid)
	assert.Equal(t, 1, targetValidationCalls)

	var targetCount int
	require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM validation_flag_targets").Scan(&targetCount))
	assert.Zero(t, targetCount)
}

func TestManager_Update_ValidationFlagDoesNotLeakIntoHooks(t *testing.T) {
	database := setupValidationContextDB(t)
	targetManager, err := NewManagerWithDB[validationFlagTarget]("validation_flag_targets", database)
	require.NoError(t, err)
	parentManager, err := NewManagerWithDB[validationFlagUpdateParent]("validation_flag_update_parents", database)
	require.NoError(t, err)

	_, err = database.Exec("INSERT INTO validation_flag_targets (id, name) VALUES (1, 'valid')")
	require.NoError(t, err)
	_, err = database.Exec("INSERT INTO validation_flag_update_parents (id, name) VALUES (1, 'original')")
	require.NoError(t, err)

	targetValidationCalls := 0
	parent := &validationFlagUpdateParent{
		ID:            1,
		Name:          "changed",
		targetManager: targetManager,
		targetCalls:   &targetValidationCalls,
	}
	err = parentManager.Update(WithModelValidationCompleted(context.Background()), parent)
	require.ErrorIs(t, err, errValidationFlagTargetInvalid)
	assert.Equal(t, 1, targetValidationCalls)

	var targetName string
	require.NoError(t, database.QueryRow("SELECT name FROM validation_flag_targets WHERE id = 1").Scan(&targetName))
	assert.Equal(t, "valid", targetName)

	var parentName string
	require.NoError(t, database.QueryRow("SELECT name FROM validation_flag_update_parents WHERE id = 1").Scan(&parentName))
	assert.Equal(t, "original", parentName)
}

func TestManager_Create_ValidationFlagStillRunsValidate(t *testing.T) {
	database := setupValidationContextDB(t)
	manager, err := NewManagerWithDB[validationFlagTarget]("validation_flag_targets", database)
	require.NoError(t, err)

	unmarkedValidationCalls := 0
	err = manager.Create(context.Background(), &validationFlagTarget{Name: "invalid", validationCalls: &unmarkedValidationCalls})
	require.ErrorIs(t, err, errValidationFlagTargetInvalid)
	assert.Equal(t, 1, unmarkedValidationCalls)

	markedValidationCalls := 0
	err = manager.Create(
		WithModelValidationCompleted(context.Background()),
		&validationFlagTarget{Name: "invalid", validationCalls: &markedValidationCalls},
	)
	require.ErrorIs(t, err, errValidationFlagTargetInvalid)
	assert.Equal(t, 1, markedValidationCalls)
}

func TestManager_MarkedValidationRejectsBeforeSaveMutation(t *testing.T) {
	database := setupValidationContextDB(t)
	manager, err := NewManagerWithDB[validationFlagTarget]("validation_flag_targets", database)
	require.NoError(t, err)

	t.Run("create", func(t *testing.T) {
		calls := 0
		target := &validationFlagTarget{Name: "valid", validationCalls: &calls, invalidateOnSave: true}
		err := manager.Create(WithModelValidationCompleted(context.Background()), target)
		require.ErrorIs(t, err, errValidationFlagTargetInvalid)
		assert.Equal(t, 1, calls)

		var count int
		require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM validation_flag_targets").Scan(&count))
		assert.Zero(t, count)
	})

	t.Run("update", func(t *testing.T) {
		_, err := database.Exec("INSERT INTO validation_flag_targets (id, name) VALUES (2, 'original')")
		require.NoError(t, err)
		calls := 0
		target := &validationFlagTarget{ID: 2, Name: "valid", validationCalls: &calls, invalidateOnSave: true}
		err = manager.Update(WithModelValidationCompleted(context.Background()), target)
		require.ErrorIs(t, err, errValidationFlagTargetInvalid)
		assert.Equal(t, 1, calls)

		var name string
		require.NoError(t, database.QueryRow("SELECT name FROM validation_flag_targets WHERE id = 2").Scan(&name))
		assert.Equal(t, "original", name)
	})
}
