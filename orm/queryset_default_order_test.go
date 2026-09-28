package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySet_FirstLastDefaultToPrimaryKeyOrder(t *testing.T) {
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
