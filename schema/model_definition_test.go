package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleModel struct {
	BaseSchema
}

func (sampleModel) Fields() []Field {
	return []Field{
		Int64Field("id", DBIndex()),
		StringField("title", Required()),
	}
}

func (sampleModel) Relations() []Relation {
	return []Relation{
		{
			Name:     "author",
			To:       "User",
			Type:     RelationForeignKey,
			OnDelete: CascadeCASCADE,
		},
	}
}

func (sampleModel) Meta() Meta {
	return Meta{
		TableName:         "sample_models",
		VerboseName:       "Sample Model",
		VerboseNamePlural: "Sample Models",
	}
}

func (sampleModel) Hooks() *ModelHooks {
	return NewModelHooks().
		WithBeforeCreate(func(ctx context.Context, instance interface{}) error { return nil }).
		WithAfterCreate(func(ctx context.Context, instance interface{}) error { return nil }).
		WithBeforeUpdate(func(ctx context.Context, instance interface{}) error { return nil }).
		WithAfterUpdate(func(ctx context.Context, instance interface{}) error { return nil }).
		WithBeforeSave(func(ctx context.Context, instance interface{}) error { return nil }).
		WithAfterSave(func(ctx context.Context, instance interface{}) error { return nil }).
		WithBeforeDelete(func(ctx context.Context, instance interface{}) error { return nil }).
		WithAfterDelete(func(ctx context.Context, instance interface{}) error { return nil }).
		WithClean(func(instance interface{}) error { return nil })
}

// Models define fields and relations
func TestModelDefinesFieldsAndRelations(t *testing.T) {
	var model Schema = sampleModel{}
	require.NotNil(t, model)

	fields := model.Fields()
	require.Len(t, fields, 2)
	assert.Equal(t, "id", fields[0].Name)
	assert.Equal(t, "title", fields[1].Name)

	relations := model.Relations()
	require.Len(t, relations, 1)
	assert.Equal(t, "author", relations[0].Name)
	assert.Equal(t, "User", relations[0].To)
	assert.Equal(t, RelationForeignKey, relations[0].Type)

	meta := model.Meta()
	assert.Equal(t, "sample_models", meta.TableName)
	assert.Equal(t, "Sample Model", meta.VerboseName)

	hooks := model.Hooks()
	require.NotNil(t, hooks)
}

// Models support all listed lifecycle hooks
func TestModelLifecycleHooksConfiguredAndCallable(t *testing.T) {
	called := make(map[string]bool)

	hooks := NewModelHooks().
		WithBeforeCreate(func(ctx context.Context, inst interface{}) error {
			called["BeforeCreate"] = true
			return nil
		}).
		WithAfterCreate(func(ctx context.Context, inst interface{}) error {
			called["AfterCreate"] = true
			return nil
		}).
		WithBeforeUpdate(func(ctx context.Context, inst interface{}) error {
			called["BeforeUpdate"] = true
			return nil
		}).
		WithAfterUpdate(func(ctx context.Context, inst interface{}) error {
			called["AfterUpdate"] = true
			return nil
		}).
		WithBeforeSave(func(ctx context.Context, inst interface{}) error {
			called["BeforeSave"] = true
			return nil
		}).
		WithAfterSave(func(ctx context.Context, inst interface{}) error {
			called["AfterSave"] = true
			return nil
		}).
		WithBeforeDelete(func(ctx context.Context, inst interface{}) error {
			called["BeforeDelete"] = true
			return nil
		}).
		WithAfterDelete(func(ctx context.Context, inst interface{}) error {
			called["AfterDelete"] = true
			return nil
		}).
		WithClean(func(inst interface{}) error {
			called["Clean"] = true
			return nil
		})

	ctx := context.Background()
	dummy := &sampleModel{}

	require.NoError(t, hooks.BeforeCreate(ctx, dummy))
	require.NoError(t, hooks.AfterCreate(ctx, dummy))
	require.NoError(t, hooks.BeforeUpdate(ctx, dummy))
	require.NoError(t, hooks.AfterUpdate(ctx, dummy))
	require.NoError(t, hooks.BeforeSave(ctx, dummy))
	require.NoError(t, hooks.AfterSave(ctx, dummy))
	require.NoError(t, hooks.BeforeDelete(ctx, dummy))
	require.NoError(t, hooks.AfterDelete(ctx, dummy))
	require.NoError(t, hooks.Clean(dummy))

	expectedHooks := []string{
		"BeforeCreate", "AfterCreate",
		"BeforeUpdate", "AfterUpdate",
		"BeforeSave", "AfterSave",
		"BeforeDelete", "AfterDelete",
		"Clean",
	}
	for _, name := range expectedHooks {
		assert.True(t, called[name], "expected hook %s to be invoked", name)
	}
}
