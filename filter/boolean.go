package filter

// Boolean tree composition is already implemented in filterset.go
// This file contains additional helper functions for boolean operations

// And combines multiple filter nodes with AND
func And(nodes ...*FilterNode) *FilterNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	return NewAndNode(nodes...)
}

// Or combines multiple filter nodes with OR
func Or(nodes ...*FilterNode) *FilterNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	return NewOrNode(nodes...)
}

// Not negates a filter node
func Not(node *FilterNode) *FilterNode {
	if node == nil {
		return nil
	}
	return NewNotNode(node)
}

// CombineNodes combines nodes with a given operation
func CombineNodes(op FilterOp, nodes ...*FilterNode) *FilterNode {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}

	switch op {
	case OpAnd:
		return NewAndNode(nodes...)
	case OpOr:
		return NewOrNode(nodes...)
	default:
		// Default to AND
		return NewAndNode(nodes...)
	}
}
