package filter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/forgego/forge/db"
)

var (
	filterIDCounter        atomic.Uint64
	defaultFilterStorageMu sync.RWMutex
	defaultFilterStorage   FilterStorage
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

// InMemoryFilterStorage is an in-memory filter storage for tests only; production uses the database store.
type InMemoryFilterStorage struct {
	mu      sync.RWMutex
	filters map[string]*SavedFilter
}

// NewInMemoryFilterStorage creates a new in-memory storage for tests only; production uses the database store.
func NewInMemoryFilterStorage() *InMemoryFilterStorage {
	return &InMemoryFilterStorage{
		filters: make(map[string]*SavedFilter),
	}
}

// Save saves a filter
func (s *InMemoryFilterStorage) Save(filter *SavedFilter) error {
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}
	if filter.ID == "" {
		return fmt.Errorf("filter ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	filter.CreatedAt = time.Now()
	filter.UpdatedAt = time.Now()
	s.filters[filter.ID] = filter
	return nil
}

// Load loads a filter by ID
func (s *InMemoryFilterStorage) Load(id string) (*SavedFilter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filter, ok := s.filters[id]
	if !ok {
		return nil, fmt.Errorf("filter not found: %s", id)
	}
	return filter, nil
}

// Delete deletes a filter
func (s *InMemoryFilterStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.filters, id)
	return nil
}

// List lists filters for an owner
func (s *InMemoryFilterStorage) List(ownerID string) ([]*SavedFilter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}
	if filter.ID == "" {
		return fmt.Errorf("filter ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.filters[filter.ID]; !ok {
		return fmt.Errorf("filter not found: %s", filter.ID)
	}

	filter.UpdatedAt = time.Now()
	s.filters[filter.ID] = filter
	return nil
}

// DatabaseFilterStorage is a database-backed FilterStorage using the forge_saved_filters table.
type DatabaseFilterStorage struct {
	db *db.DB
}

// DBFilterStorage is an alias for DatabaseFilterStorage.
type DBFilterStorage = DatabaseFilterStorage

// NewDatabaseFilterStorage creates a new database-backed filter storage.
func NewDatabaseFilterStorage(database *db.DB) *DatabaseFilterStorage {
	return &DatabaseFilterStorage{db: database}
}

// NewDBFilterStorage creates a new database-backed filter storage.
func NewDBFilterStorage(database *db.DB) *DatabaseFilterStorage {
	return NewDatabaseFilterStorage(database)
}

// DB returns the underlying database connection.
func (s *DatabaseFilterStorage) DB() *db.DB {
	return s.db
}

var (
	_ FilterStorage = (*DatabaseFilterStorage)(nil)
	_ FilterStorage = (*InMemoryFilterStorage)(nil)
)

// SetDefaultFilterStorage sets the package default filter storage.
func SetDefaultFilterStorage(s FilterStorage) {
	defaultFilterStorageMu.Lock()
	defer defaultFilterStorageMu.Unlock()
	defaultFilterStorage = s
}

// GetDefaultFilterStorage returns the package default filter storage.
func GetDefaultFilterStorage() FilterStorage {
	defaultFilterStorageMu.RLock()
	defer defaultFilterStorageMu.RUnlock()
	return defaultFilterStorage
}

// Save saves a filter to the database.
func (s *DatabaseFilterStorage) Save(filter *SavedFilter) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database is not configured")
	}
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}
	if filter.ID == "" {
		return fmt.Errorf("filter ID cannot be empty")
	}

	now := time.Now().UTC()
	if filter.CreatedAt.IsZero() {
		filter.CreatedAt = now
	}
	filter.UpdatedAt = now

	if filter.DialectHint == "" && s.db.Driver != "" {
		filter.DialectHint = s.db.Driver
	}
	if filter.SchemaVersion == "" {
		filter.SchemaVersion = "1.0"
	}
	if filter.Version == 0 {
		filter.Version = 1
	}

	var astJSON string
	if filter.AST != nil {
		data, err := json.Marshal(filter.AST)
		if err != nil {
			return fmt.Errorf("failed to marshal filter AST: %w", err)
		}
		astJSON = string(data)
	}

	query := `INSERT INTO forge_saved_filters (
		id, owner_id, name, description, ast, dialect_hint,
		schema_version, public, version, created_at, updated_at,
		sample_count, estimated_cost
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (id) DO UPDATE SET
		owner_id = excluded.owner_id,
		name = excluded.name,
		description = excluded.description,
		ast = excluded.ast,
		dialect_hint = excluded.dialect_hint,
		schema_version = excluded.schema_version,
		public = excluded.public,
		version = excluded.version,
		updated_at = excluded.updated_at,
		sample_count = excluded.sample_count,
		estimated_cost = excluded.estimated_cost`

	_, err := s.db.ExecContext(context.Background(), query,
		filter.ID,
		filter.OwnerID,
		filter.Name,
		filter.Description,
		astJSON,
		filter.DialectHint,
		filter.SchemaVersion,
		filter.Public,
		filter.Version,
		filter.CreatedAt,
		filter.UpdatedAt,
		filter.SampleCount,
		filter.EstimatedCost,
	)
	return err
}

// Load loads a filter by ID from the database.
func (s *DatabaseFilterStorage) Load(id string) (*SavedFilter, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}

	query := `SELECT id, owner_id, name, description, ast, dialect_hint,
		schema_version, public, version, created_at, updated_at,
		sample_count, estimated_cost
	FROM forge_saved_filters
	WHERE id = $1`

	var f SavedFilter
	var astJSON string
	err := s.db.QueryRowContext(context.Background(), query, id).Scan(
		&f.ID,
		&f.OwnerID,
		&f.Name,
		&f.Description,
		&astJSON,
		&f.DialectHint,
		&f.SchemaVersion,
		&f.Public,
		&f.Version,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.SampleCount,
		&f.EstimatedCost,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("filter not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	if astJSON != "" {
		var node FilterNode
		if err := json.Unmarshal([]byte(astJSON), &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filter AST: %w", err)
		}
		f.AST = &node
	}

	return &f, nil
}

// Delete deletes a filter by ID from the database.
func (s *DatabaseFilterStorage) Delete(id string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database is not configured")
	}

	query := `DELETE FROM forge_saved_filters WHERE id = $1`
	_, err := s.db.ExecContext(context.Background(), query, id)
	return err
}

// List lists filters for an owner (and public filters) from the database.
func (s *DatabaseFilterStorage) List(ownerID string) ([]*SavedFilter, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}

	query := `SELECT id, owner_id, name, description, ast, dialect_hint,
		schema_version, public, version, created_at, updated_at,
		sample_count, estimated_cost
	FROM forge_saved_filters
	WHERE (owner_id = $1 OR public = true)
	ORDER BY created_at, id`

	rows, err := s.db.QueryContext(context.Background(), query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filters := make([]*SavedFilter, 0)
	for rows.Next() {
		var f SavedFilter
		var astJSON string
		if err := rows.Scan(
			&f.ID,
			&f.OwnerID,
			&f.Name,
			&f.Description,
			&astJSON,
			&f.DialectHint,
			&f.SchemaVersion,
			&f.Public,
			&f.Version,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.SampleCount,
			&f.EstimatedCost,
		); err != nil {
			return nil, err
		}

		if astJSON != "" {
			var node FilterNode
			if err := json.Unmarshal([]byte(astJSON), &node); err != nil {
				return nil, fmt.Errorf("failed to unmarshal filter AST: %w", err)
			}
			f.AST = &node
		}

		filters = append(filters, &f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return filters, nil
}

// Update updates a filter in the database.
func (s *DatabaseFilterStorage) Update(filter *SavedFilter) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database is not configured")
	}
	if filter == nil {
		return fmt.Errorf("filter cannot be nil")
	}
	if filter.ID == "" {
		return fmt.Errorf("filter ID cannot be empty")
	}

	filter.UpdatedAt = time.Now().UTC()

	var astJSON string
	if filter.AST != nil {
		data, err := json.Marshal(filter.AST)
		if err != nil {
			return fmt.Errorf("failed to marshal filter AST: %w", err)
		}
		astJSON = string(data)
	}

	query := `UPDATE forge_saved_filters SET
		owner_id = $2,
		name = $3,
		description = $4,
		ast = $5,
		dialect_hint = $6,
		schema_version = $7,
		public = $8,
		version = $9,
		updated_at = $10,
		sample_count = $11,
		estimated_cost = $12
	WHERE id = $1`

	res, err := s.db.ExecContext(context.Background(), query,
		filter.ID,
		filter.OwnerID,
		filter.Name,
		filter.Description,
		astJSON,
		filter.DialectHint,
		filter.SchemaVersion,
		filter.Public,
		filter.Version,
		filter.UpdatedAt,
		filter.SampleCount,
		filter.EstimatedCost,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("filter not found: %s", filter.ID)
	}
	return nil
}

// SaveFilter saves a filter AST as a persisted filter
func SaveFilter[T any](fs *FilterSet[T], name, description string, storage FilterStorage) (*SavedFilter, error) {
	if storage == nil {
		storage = GetDefaultFilterStorage()
	}
	if storage == nil {
		return nil, fmt.Errorf("filter storage is nil")
	}

	ast := fs.GetAST()
	if ast == nil {
		return nil, fmt.Errorf("no filter AST to save")
	}

	// Generate unique ID
	id := fmt.Sprintf("filter_%d_%d", time.Now().UnixNano(), filterIDCounter.Add(1))

	filter := &SavedFilter{
		ID:            id,
		Name:          name,
		Description:   description,
		AST:           ast.Clone(),
		DialectHint:   "postgres", // Would be determined from DB
		SchemaVersion: "1.0",      // Would be from schema
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
	if storage == nil {
		storage = GetDefaultFilterStorage()
	}
	if storage == nil {
		return nil, fmt.Errorf("filter storage is nil")
	}

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
	qs = qs.Limit(limit)
	count, err := qs.Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}
