package orm

import (
	"context"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySetGet_EmptyResultReturnsTypedNotFound(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	qs, err := NewQuerySet[TestCompany]("companies")
	require.NoError(t, err)
	_, err = qs.SetDB(database).Filter(F("id").Eq(int64(999))).Get(context.Background())
	require.Error(t, err)
	assert.True(t, forgeerrors.IsNotFound(err))
	assert.Equal(t, "companies matching query does not exist", err.Error())
}
