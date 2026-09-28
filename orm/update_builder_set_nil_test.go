package orm

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestPointerCustomer struct {
	schema.BaseSchema
	ID  int64   `db:"id"`
	Bio *string `db:"bio"`
}

func (TestPointerCustomer) Meta() schema.Meta {
	return schema.Meta{TableName: "pointer_customers"}
}

func (TestPointerCustomer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("bio"),
	}
}

func TestUpdateBuilder_SetNil(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()

	ptrSchema, err := GetModelSchema[TestPointerCustomer]()
	require.NoError(t, err)
	bioField := ptrSchema.GetField("bio")
	require.NotNil(t, bioField)
	bioField.Type = reflect.TypeOf((*string)(nil))

	customerQS, err := NewQuerySet[TestCustomer]("customers")
	require.NoError(t, err)
	customerQS = customerQS.SetDB(database)

	pointerQS, err := NewQuerySet[TestPointerCustomer]("pointer_customers")
	require.NoError(t, err)
	pointerQS = pointerQS.SetDB(database)

	tests := []struct {
		name        string
		setupUB     func() *UpdateBuilder[TestCustomer]
		field       string
		shouldPanic bool
		expectErr   bool
		errContains string
	}{
		{
			name: "non-nullable field with nil does not panic and records error",
			setupUB: func() *UpdateBuilder[TestCustomer] {
				ub, err := NewUpdateBuilder[TestCustomer](customerQS)
				require.NoError(t, err)
				return ub
			},
			field:       "name",
			shouldPanic: false,
			expectErr:   true,
			errContains: "field name cannot be set to nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ub := tt.setupUB()
			assert.NotPanics(t, func() {
				ub.Set(tt.field, nil)
			})
			if tt.expectErr {
				if assert.Error(t, ub.err) {
					assert.Contains(t, ub.err.Error(), "field name cannot be set to nil")
				}
			} else {
				assert.NoError(t, ub.err)
			}
		})
	}

	t.Run("pointer-typed field with nil records no error", func(t *testing.T) {
		ub, err := NewUpdateBuilder[TestPointerCustomer](pointerQS)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			ub.Set("bio", nil)
		})
		assert.NoError(t, ub.err)
		assert.Nil(t, ub.updates["bio"])
	})

	typeAllowedTests := []struct {
		name         string
		expectedType reflect.Type
		shouldAllow  bool
	}{
		{
			name:         "pointer type allowed",
			expectedType: reflect.TypeOf((*string)(nil)),
			shouldAllow:  true,
		},
		{
			name:         "interface type allowed",
			expectedType: reflect.TypeOf((*any)(nil)).Elem(),
			shouldAllow:  true,
		},
		{
			name:         "slice type allowed",
			expectedType: reflect.TypeOf([]string{}),
			shouldAllow:  true,
		},
		{
			name:         "map type allowed",
			expectedType: reflect.TypeOf(map[string]any{}),
			shouldAllow:  true,
		},
		{
			name:         "sql.NullString allowed",
			expectedType: reflect.TypeOf(sql.NullString{}),
			shouldAllow:  true,
		},
		{
			name:         "sql.NullInt64 allowed",
			expectedType: reflect.TypeOf(sql.NullInt64{}),
			shouldAllow:  true,
		},
		{
			name:         "int64 non-nullable rejected",
			expectedType: reflect.TypeOf(int64(0)),
			shouldAllow:  false,
		},
		{
			name:         "string non-nullable rejected",
			expectedType: reflect.TypeOf(""),
			shouldAllow:  false,
		},
	}

	for _, tt := range typeAllowedTests {
		t.Run(tt.name, func(t *testing.T) {
			customSchema := &ModelSchema{
				Fields: []FieldInfo{
					{Name: "test_field", Type: tt.expectedType},
				},
			}
			ub := &UpdateBuilder[TestCustomer]{
				schema:  customSchema,
				updates: make(map[string]interface{}),
			}

			assert.NotPanics(t, func() {
				ub.Set("test_field", nil)
			})

			if tt.shouldAllow {
				assert.NoError(t, ub.err)
				val, exists := ub.updates["test_field"]
				assert.True(t, exists)
				assert.Nil(t, val)
			} else {
				if assert.Error(t, ub.err) {
					assert.Contains(t, ub.err.Error(), "field test_field cannot be set to nil")
				}
			}
		})
	}
}
