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

// FilterSet optimizes generated queries
func TestFilterSetOptimizesGeneratedQueries(t *testing.T) {
	opt := NewQueryOptimizer()

	// Shallow relation with low cost produces a join strategy
	shallowAST := NewFieldNode("author__name", "exact", "Alice")
	shallowPlan, err := opt.Optimize(shallowAST)
	require.NoError(t, err)
	require.Equal(t, "join", shallowPlan.Strategy)
	require.Contains(t, shallowPlan.SQLPreview, "[join strategy]")

	// Deep relation path (depth > maxJoinDepth 3) optimizes to an exists strategy
	deepAST := NewFieldNode("organization__department__team__leader__name", "exact", "Bob")
	deepPlan, err := opt.Optimize(deepAST)
	require.NoError(t, err)
	require.Equal(t, "exists", deepPlan.Strategy)
	require.Contains(t, deepPlan.SQLPreview, "[exists strategy]")
}
