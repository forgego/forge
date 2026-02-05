package filter

import (
	"context"
	"fmt"
	"time"
)

// SavedFilter represents a persisted filter
type SavedFilter struct {
	ID            string
	OwnerID       string
	Name          string
	Description   string
	AST           *FilterNode
	DialectHint   string
	SchemaVersion string
	Public        bool
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SampleCount   int64
	EstimatedCost int
}

// FilterStorage is the interface for storing filters
type FilterStorage interface {
	Save(filter *SavedFilter) error
	Load(id string) (*SavedFilter, error)
	Delete(id string) error
	List(ownerID string) ([]*SavedFilter, error)
	Update(filter *SavedFilter) error
}

// InMemoryFilterStorage is an in-memory filter storage
type InMemoryFilterStorage struct {
	filters map[string]*SavedFilter
}

// NewInMemoryFilterStorage creates a new in-memory storage
func NewInMemoryFilterStorage() *InMemoryFilterStorage {
	return &InMemoryFilterStorage{
		filters: make(map[string]*SavedFilter),
	}
}

// Save saves a filter
func (s *InMemoryFilterStorage) Save(filter *SavedFilter) error {
	if filter.ID == "" {
		return fmt.Errorf("filter ID cannot be empty")
	}

	filter.CreatedAt = time.Now()
	filter.UpdatedAt = time.Now()
	s.filters[filter.ID] = filter
	return nil
}

// Load loads a filter by ID
func (s *InMemoryFilterStorage) Load(id string) (*SavedFilter, error) {
	filter, ok := s.filters[id]
	if !ok {
		return nil, fmt.Errorf("filter not found: %s", id)
	}
	return filter, nil
}

// Delete deletes a filter
func (s *InMemoryFilterStorage) Delete(id string) error {
	delete(s.filters, id)
	return nil
}

// List lists filters for an owner
func (s *InMemoryFilterStorage) List(ownerID string) ([]*SavedFilter, error) {
	filters := make([]*SavedFilter, 0)
	for _, filter := range s.filters {
		if filter.OwnerID == ownerID || filter.Public {
			filters = append(filters, filter)
		}
	}
	return filters, nil
}

// Update updates a filter
func (s *InMemoryFilterStorage) Update(filter *SavedFilter) error {
	if _, ok := s.filters[filter.ID]; !ok {
		return fmt.Errorf("filter not found: %s", filter.ID)
	}

	filter.UpdatedAt = time.Now()
	s.filters[filter.ID] = filter
	return nil
}

// SaveFilter saves a filter AST as a persisted filter
func SaveFilter[T any](fs *FilterSet[T], name, description string, storage FilterStorage) (*SavedFilter, error) {
	ast := fs.GetAST()
	if ast == nil {
		return nil, fmt.Errorf("no filter AST to save")
	}

	// Generate ID (in production, use UUID)
	id := fmt.Sprintf("filter_%d", time.Now().UnixNano())

	filter := &SavedFilter{
		ID:            id,
		Name:          name,
		Description:   description,
		AST:           ast.Clone(),
		DialectHint:   "postgres", // Would be determined from DB
		SchemaVersion: "1.0",       // Would be from schema
		Public:        false,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := storage.Save(filter); err != nil {
		return nil, err
	}

	return filter, nil
}

// LoadFilter loads a saved filter
func LoadFilter[T any](id string, storage FilterStorage) (*FilterSet[T], error) {
	saved, err := storage.Load(id)
	if err != nil {
		return nil, err
	}

	fs, err := NewFilterSet[T]()
	if err != nil {
		return nil, err
	}

	fs.SetAST(saved.AST)
	return fs, nil
}

// PreviewFilter previews a filter with sample count
func PreviewFilter[T any](fs *FilterSet[T], ctx context.Context, limit int) (int64, error) {
	ast := fs.GetAST()
	if ast == nil {
		return 0, nil
	}

	// Apply filter and get count
	qs, err := fs.ApplyAST(ctx, ast)
	if err != nil {
		return 0, err
	}

	// Get limited count
	qs = qs.Limit(int(limit))
	count, err := qs.Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}
