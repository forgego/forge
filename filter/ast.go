package filter

import (
	"encoding/json"
	"fmt"
	"time"
)

// FilterOp represents the operation type in a filter node
type FilterOp string

const (
	OpAnd   FilterOp = "and"
	OpOr    FilterOp = "or"
	OpNot   FilterOp = "not"
	OpField FilterOp = "field"
)

// FilterNode represents a node in the filter AST
type FilterNode struct {
	Op       FilterOp        `json:"op"`
	Field    string          `json:"field,omitempty"`
	Lookup   string          `json:"lookup,omitempty"`
	Value    interface{}     `json:"value,omitempty"`
	Children []*FilterNode   `json:"children,omitempty"`
	Metadata *FilterMetadata `json:"metadata,omitempty"`
}

// FilterMetadata contains metadata about a filter
type FilterMetadata struct {
	Version       string    `json:"version,omitempty"`
	SchemaVersion string    `json:"schema_version,omitempty"`
	EstimatedCost int       `json:"estimated_cost,omitempty"`
	SampleCount   int64     `json:"sample_count,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	OwnerID       string    `json:"owner_id,omitempty"`
}

// NewFieldNode creates a new field filter node
func NewFieldNode(field, lookup string, value interface{}) *FilterNode {
	return &FilterNode{
		Op:     OpField,
		Field:  field,
		Lookup: lookup,
		Value:  value,
	}
}

// NewBooleanNode creates a new boolean operation node
func NewBooleanNode(op FilterOp, children ...*FilterNode) *FilterNode {
	return &FilterNode{
		Op:       op,
		Children: children,
	}
}

// NewAndNode creates an AND node
func NewAndNode(children ...*FilterNode) *FilterNode {
	return NewBooleanNode(OpAnd, children...)
}

// NewOrNode creates an OR node
func NewOrNode(children ...*FilterNode) *FilterNode {
	return NewBooleanNode(OpOr, children...)
}

// NewNotNode creates a NOT node
func NewNotNode(child *FilterNode) *FilterNode {
	return NewBooleanNode(OpNot, child)
}

// AddChild adds a child node to this node
func (n *FilterNode) AddChild(child *FilterNode) {
	if n.Children == nil {
		n.Children = []*FilterNode{}
	}
	n.Children = append(n.Children, child)
}

// IsLeaf returns true if this is a leaf node (field node)
func (n *FilterNode) IsLeaf() bool {
	return n.Op == OpField
}

// IsBoolean returns true if this is a boolean operation node
func (n *FilterNode) IsBoolean() bool {
	return n.Op == OpAnd || n.Op == OpOr || n.Op == OpNot
}

// Clone creates a deep copy of the filter node
func (n *FilterNode) Clone() *FilterNode {
	if n == nil {
		return nil
	}

	clone := &FilterNode{
		Op:     n.Op,
		Field:  n.Field,
		Lookup: n.Lookup,
		Value:  n.Value,
	}

	if n.Children != nil {
		clone.Children = make([]*FilterNode, len(n.Children))
		for i, child := range n.Children {
			clone.Children[i] = child.Clone()
		}
	}

	if n.Metadata != nil {
		clone.Metadata = &FilterMetadata{
			Version:       n.Metadata.Version,
			SchemaVersion: n.Metadata.SchemaVersion,
			EstimatedCost: n.Metadata.EstimatedCost,
			SampleCount:   n.Metadata.SampleCount,
			CreatedAt:     n.Metadata.CreatedAt,
			OwnerID:       n.Metadata.OwnerID,
		}
	}

	return clone
}

// ToJSON serializes the filter node to JSON
func (n *FilterNode) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes a filter node from JSON
func FromJSON(data []byte) (*FilterNode, error) {
	var node FilterNode
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal filter node: %w", err)
	}
	return &node, nil
}

// Validate validates the filter node structure
func (n *FilterNode) Validate() error {
	if n == nil {
		return fmt.Errorf("filter node cannot be nil")
	}

	switch n.Op {
	case OpField:
		if n.Field == "" {
			return fmt.Errorf("field node must have a field name")
		}
		if n.Lookup == "" {
			return fmt.Errorf("field node must have a lookup type")
		}
		if len(n.Children) > 0 {
			return fmt.Errorf("field node cannot have children")
		}

	case OpAnd, OpOr:
		if len(n.Children) < 2 {
			return fmt.Errorf("%s node must have at least 2 children", n.Op)
		}
		for i, child := range n.Children {
			if err := child.Validate(); err != nil {
				return fmt.Errorf("child %d: %w", i, err)
			}
		}

	case OpNot:
		if len(n.Children) != 1 {
			return fmt.Errorf("not node must have exactly 1 child")
		}
		if err := n.Children[0].Validate(); err != nil {
			return fmt.Errorf("not child: %w", err)
		}

	default:
		return fmt.Errorf("unknown filter operation: %s", n.Op)
	}

	return nil
}

// String returns a string representation of the filter node
func (n *FilterNode) String() string {
	if n == nil {
		return "nil"
	}

	switch n.Op {
	case OpField:
		return fmt.Sprintf("%s__%s=%v", n.Field, n.Lookup, n.Value)
	case OpAnd:
		return fmt.Sprintf("AND(%d)", len(n.Children))
	case OpOr:
		return fmt.Sprintf("OR(%d)", len(n.Children))
	case OpNot:
		return "NOT(1)"
	default:
		return fmt.Sprintf("UNKNOWN(%s)", n.Op)
	}
}
