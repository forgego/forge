package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryOptimizer_OptimizeNil(t *testing.T) {
	opt := NewQueryOptimizer()
	plan, err := opt.Optimize(nil)
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, "none", plan.Strategy)
}

func TestQueryOptimizer_TwoConditionAST(t *testing.T) {
	opt := NewQueryOptimizer()
	ast := NewAndNode(
		NewFieldNode("username", "exact", "alice"),
		NewFieldNode("email", "contains", "@example.com"),
	)
	plan, err := opt.Optimize(ast)
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.NotEmpty(t, plan.Strategy)
	require.NotEqual(t, "none", plan.Strategy)
}
