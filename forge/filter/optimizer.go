package filter

import (
	"fmt"
)

// QueryOptimizer optimizes filter queries
type QueryOptimizer struct {
	maxJoinDepth int
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		maxJoinDepth: 3,
	}
}

// QueryPlan represents an optimized query plan
type QueryPlan struct {
	Strategy      string  // "join", "exists", "subquery"
	EstimatedCost int
	EstimatedRows int64
	SQLPreview    string
}

// Optimize optimizes a filter AST and returns a query plan
func (o *QueryOptimizer) Optimize(ast *FilterNode) (*QueryPlan, error) {
	if ast == nil {
		return &QueryPlan{
			Strategy:      "none",
			EstimatedCost: 0,
			EstimatedRows: 0,
		}, nil
	}

	// Analyze the AST to determine the best strategy
	cost := o.estimateCost(ast)
	depth := o.getMaxDepth(ast)

	// Choose strategy based on depth and cost
	strategy := o.chooseStrategy(depth, cost)

	return &QueryPlan{
		Strategy:      strategy,
		EstimatedCost: cost,
		EstimatedRows: 0, // Would be calculated based on statistics
		SQLPreview:    o.generateSQLPreview(ast, strategy),
	}, nil
}

// estimateCost estimates the cost of a filter AST
func (o *QueryOptimizer) estimateCost(node *FilterNode) int {
	if node == nil {
		return 0
	}

	cost := 0

	switch node.Op {
	case OpField:
		// Base cost for a field filter
		cost = 1
		// Check if it's a relation path (contains __)
		if contains(node.Field, "__") {
			// Add cost for each relation level
			parts := splitPath(node.Field)
			cost += len(parts) - 1 // Each relation adds 1 point
		}

	case OpAnd:
		// AND is cheap, just sum children
		for _, child := range node.Children {
			cost += o.estimateCost(child)
		}

	case OpOr:
		// OR is more expensive, especially with many branches
		for _, child := range node.Children {
			cost += o.estimateCost(child)
		}
		cost += len(node.Children) // Add penalty for OR branches

	case OpNot:
		// NOT adds some overhead
		cost = o.estimateCost(node.Children[0]) + 1
	}

	return cost
}

// getMaxDepth gets the maximum relation depth in the AST
func (o *QueryOptimizer) getMaxDepth(node *FilterNode) int {
	if node == nil {
		return 0
	}

	maxDepth := 0

	switch node.Op {
	case OpField:
		if contains(node.Field, "__") {
			parts := splitPath(node.Field)
			depth := len(parts) - 1
			if depth > maxDepth {
				maxDepth = depth
			}
		}

	case OpAnd, OpOr:
		for _, child := range node.Children {
			depth := o.getMaxDepth(child)
			if depth > maxDepth {
				maxDepth = depth
			}
		}

	case OpNot:
		depth := o.getMaxDepth(node.Children[0])
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// chooseStrategy chooses the best query strategy
func (o *QueryOptimizer) chooseStrategy(depth, cost int) string {
	// Simple heuristic:
	// - Use JOIN for shallow relations (depth <= 2) and low cost
	// - Use EXISTS for deeper relations or higher cost
	// - Use subquery for very complex queries

	if depth > o.maxJoinDepth {
		return "exists"
	}

	if cost > 20 {
		return "subquery"
	}

	if depth > 1 && cost > 10 {
		return "exists"
	}

	return "join"
}

// generateSQLPreview generates a SQL preview for the query plan
func (o *QueryOptimizer) generateSQLPreview(ast *FilterNode, strategy string) string {
	// This would generate actual SQL preview
	// For now, return a placeholder
	return fmt.Sprintf("SELECT * FROM table WHERE ... [%s strategy]", strategy)
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitPath(path string) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range path {
		if char == '_' && len(current) > 0 && current[len(current)-1] == '_' {
			// Double underscore, treat as single underscore in field name
			current = current[:len(current)-1] + string(char)
		} else if char == '_' && len(current) > 0 {
			// Single underscore, might be separator
			// For simplicity, assume __ is always separator
			if len(current) > 0 {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}
