package orm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An OR group chained with another Filter must stay grouped:
// Filter(Or(a, b)).Filter(c) means (a OR b) AND c, not a OR (b AND c).
// The admin list relies on this to combine search with list filters.
func TestFilter_OrGroupKeepsPrecedenceWhenChained(t *testing.T) {
	database := setupRelationTestDB(t)
	defer database.Close()
	_, err := database.Exec(`CREATE TABLE aggregate_test_items (id INTEGER PRIMARY KEY, amount REAL, kind TEXT)`)
	require.NoError(t, err)
	_, err = database.Exec(`INSERT INTO aggregate_test_items (id, amount, kind) VALUES (1, 5, 'a'), (2, 50, 'b'), (3, 50, 'c')`)
	require.NoError(t, err)

	base, err := NewQuerySet[aggregateTestItem]("aggregate_test_items")
	require.NoError(t, err)
	qs := base.SetDB(database)
	ctx := context.Background()

	withBool, err := qs.Filter(Or(F("kind").Eq("a"), F("kind").Eq("b"))).Filter(F("amount").Gt(10.0)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), withBool)

	withQ, err := qs.Filter(NewQ(F("kind").Eq("a")).Or(NewQ(F("kind").Eq("b")))).Filter(F("amount").Gt(10.0)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), withQ)

	// Negated groups were already wrapped; keep them correct too.
	negated, err := qs.Filter(NewQ(F("kind").Eq("a")).Or(NewQ(F("kind").Eq("b"))).Not()).Filter(F("amount").Gt(10.0)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), negated)
}
