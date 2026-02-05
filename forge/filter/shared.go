package filter

import (
	"context"
	"fmt"
)

// SharedFilter represents a filter that can be shared between API and Admin
type SharedFilter[T any] struct {
	*FilterSet[T]
	ID          string
	Name        string
	Description string
	Public      bool
	OwnerID     string
}

// NewSharedFilter creates a new shared filter
func NewSharedFilter[T any](fs *FilterSet[T], id, name, description, ownerID string) *SharedFilter[T] {
	return &SharedFilter[T]{
		FilterSet:   fs,
		ID:          id,
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Public:      false,
	}
}

// MakePublic makes the filter public
func (sf *SharedFilter[T]) MakePublic() *SharedFilter[T] {
	sf.Public = true
	return sf
}

// ApplyToQueryset applies the shared filter to a queryset
func (sf *SharedFilter[T]) ApplyToQueryset(ctx context.Context, qs interface{}) (interface{}, error) {
	// This would apply the filter AST to the queryset
	ast := sf.GetAST()
	if ast == nil {
		return qs, nil
	}

	// Apply AST
	qs, err := sf.ApplyAST(ctx, ast)
	if err != nil {
		return nil, fmt.Errorf("failed to apply shared filter: %w", err)
	}

	return qs, nil
}

// RBACFilterStorage extends FilterStorage with RBAC
type RBACFilterStorage interface {
	FilterStorage
	CanAccess(userID, filterID string) (bool, error)
	SetAccess(filterID, userID string, canAccess bool) error
}
