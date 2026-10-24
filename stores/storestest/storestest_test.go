package storestest_test

import (
	"context"
	"testing"

	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/require"
)

func TestOpenMemory(t *testing.T) {
	database := storestest.OpenMemory(t)
	require.NotNil(t, database)
	require.NoError(t, stores.CheckMigrated(context.Background(), database))
}
