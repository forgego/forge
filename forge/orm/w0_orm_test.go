package orm

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
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

func TestW0_SetOperations(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		operation func(qs, other QuerySet[TestCustomer]) QuerySet[TestCustomer]
	}{
		{
			name: "Union returns NotImplementedError",
			operation: func(qs, other QuerySet[TestCustomer]) QuerySet[TestCustomer] {
				return qs.Union(other)
			},
		},
		{
			name: "Intersection returns NotImplementedError",
			operation: func(qs, other QuerySet[TestCustomer]) QuerySet[TestCustomer] {
				return qs.Intersection(other)
			},
		},
		{
			name: "Difference returns NotImplementedError",
			operation: func(qs, other QuerySet[TestCustomer]) QuerySet[TestCustomer] {
				return qs.Difference(other)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs, err := NewQuerySet[TestCustomer]("customers")
			require.NoError(t, err)
			qs = qs.SetDB(database)

			other, err := NewQuerySet[TestCustomer]("customers")
			require.NoError(t, err)
			other = other.SetDB(database)

			_, err = tt.operation(qs, other).All(ctx)
			require.Error(t, err)
			assert.True(t, forgeerrors.IsNotImplemented(err), "expected NotImplementedError, got %v", err)
		})
	}
}

func TestW0_Aggregate(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		aggregate Aggregate
	}{
		{
			name:      "Count aggregate returns NotImplementedError",
			aggregate: Count("id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs, err := NewQuerySet[TestCustomer]("customers")
			require.NoError(t, err)
			qs = qs.SetDB(database)

			_, err = qs.Aggregate(tt.aggregate).All(ctx)
			require.Error(t, err)
			assert.True(t, forgeerrors.IsNotImplemented(err), "expected NotImplementedError, got %v", err)
		})
	}
}

func TestW0_FirstLastDefaultPKOrdering(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()

	// Insert customers with ids 1,2,3 (names in non-id order)
	_, err := database.Exec(`
		INSERT INTO companies (id, name) VALUES (1, 'Acme Corp');
		INSERT INTO customers (id, name, company_id) VALUES 
			(1, 'Charlie', 1),
			(2, 'Alice', 1),
			(3, 'Bob', 1);
	`)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		query      func(qs QuerySet[TestCustomer]) (*TestCustomer, error)
		expectedID int64
	}{
		{
			name: "unordered First returns customer with lowest primary key (id 1)",
			query: func(qs QuerySet[TestCustomer]) (*TestCustomer, error) {
				return qs.First(ctx)
			},
			expectedID: 1,
		},
		{
			name: "unordered Last returns customer with highest primary key (id 3)",
			query: func(qs QuerySet[TestCustomer]) (*TestCustomer, error) {
				return qs.Last(ctx)
			},
			expectedID: 3,
		},
		{
			name: "OrderBy name Last returns last customer alphabetically (Charlie, id 1)",
			query: func(qs QuerySet[TestCustomer]) (*TestCustomer, error) {
				return qs.OrderBy("name").Last(ctx)
			},
			expectedID: 1,
		},
		{
			name: "OrderBy name First returns first customer alphabetically (Alice, id 2)",
			query: func(qs QuerySet[TestCustomer]) (*TestCustomer, error) {
				return qs.OrderBy("name").First(ctx)
			},
			expectedID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs, err := NewQuerySet[TestCustomer]("customers")
			require.NoError(t, err)
			qs = qs.SetDB(database)

			customer, err := tt.query(qs)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, customer.ID)
		})
	}
}

func TestW0_UpdateBuilderSetNil(t *testing.T) {
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
					assert.Contains(t, ub.err.Error(), tt.errContains)
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
