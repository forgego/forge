package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ResolverGenerated struct {
	CreatedAt string `json:"createdAt" db:"created_at_col"`
}

type resolverModel struct {
	ResolverGenerated
}

func TestResolveField_ReturnsEveryAliasThroughAnonymousEmbedding(t *testing.T) {
	field := Field{Name: "created_at", DBColumn: "created_at_col"}
	resolved, ok := ResolveField(&resolverModel{}, field)
	require.True(t, ok)
	assert.Equal(t, "CreatedAt", resolved.GoName)
	assert.Equal(t, "createdAt", resolved.JSONName)
	assert.Equal(t, "created_at_col", resolved.DBTag)
	assert.Equal(t, []int{0, 0}, resolved.StructField.Index)
	assert.ElementsMatch(t,
		[]string{"created_at", "created_at_col", "createdAt", "CreatedAt"},
		resolved.Names(),
	)
}
