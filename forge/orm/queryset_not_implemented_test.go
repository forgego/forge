package orm

import (
	"context"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySet_SetOperationsNotImplemented(t *testing.T) {
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

func TestQuerySet_AggregateNotImplemented(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	ctx := context.Background()

	tests := []struct {
		name      string
		aggregate Aggregate
	}{
		{
			name:      "Count aggregate chained into All returns NotImplementedError",
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
