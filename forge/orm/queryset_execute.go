package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// All executes the query and returns all results
func (qs *BaseQuerySet[T]) All(ctx context.Context) ([]*T, error) {
	if qs.err != nil {
		return nil, qs.err
	}
	sql, args, err := qs.buildSQL()
	if err != nil {
		return nil, err
	}

	db, err := qs.getDB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	results, err := qs.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Prefetch related objects
	if err := qs.prefetch(ctx, results); err != nil {
		return nil, err
	}

	return results, nil
}

// Get retrieves a single model instance from the filtered queryset.
// Returns an error if zero or more than one instance is found.
//
// This is different from Manager.Get() which retrieves by primary key ID.
// Use QuerySet.Get() when filtering, use Manager.Get() when you know the ID.
//
// Use First() if you want the first of many results, or want a different error
// when no instances are found. Get() requires exactly one match.
//
// Example:
//
//	user, err := qs.Filter(User.Email.Eq("john@example.com")).Get(ctx)
//	// Returns error if 0 or >1 users found
func (qs *BaseQuerySet[T]) Get(ctx context.Context) (*T, error) {
	results, err := qs.Limit(2).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", qs.table)
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d", qs.table, len(results))
	}

	return results[0], nil
}

// First retrieves the first model instance from the filtered queryset.
// Returns an error if no instances are found.
//
// This is ordered by the queryset's ordering (via OrderBy()), or natural
// database order if no ordering is specified.
//
// Use Get() if you require exactly one match (errors if 0 or >1).
// Use First() if you want the first of potentially many results.
//
// Example:
//
//	user, err := qs.Filter(User.Age.Gt(18)).OrderBy(User.CreatedAt.Desc()).First(ctx)
//	// Returns first user over 18, ordered by creation date (newest first)
func (qs *BaseQuerySet[T]) First(ctx context.Context) (*T, error) {
	results, err := qs.withDefaultPKOrder().Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", qs.table)
	}

	return results[0], nil
}

// Last retrieves the last object
func (qs *BaseQuerySet[T]) Last(ctx context.Context) (*T, error) {
	reversed := qs.withDefaultPKOrder().Reverse()
	results, err := reversed.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", qs.table)
	}

	return results[0], nil
}

// withDefaultPKOrder orders by primary key when no ordering is set, matching Django.
func (qs *BaseQuerySet[T]) withDefaultPKOrder() *BaseQuerySet[T] {
	if len(qs.orderBy) > 0 {
		return qs
	}
	pkCol := "id"
	if qs.schema != nil && qs.schema.PrimaryKey != "" {
		pkCol = qs.schema.PrimaryKey
	}
	clone := qs.clone()
	clone.orderBy = []OrderField{Asc(pkCol)}
	return clone
}

// Count counts matching records
func (qs *BaseQuerySet[T]) Count(ctx context.Context) (int64, error) {
	if qs.err != nil {
		return 0, qs.err
	}
	db, err := qs.getDB(ctx)
	if err != nil {
		return 0, err
	}

	query, args, err := qs.buildCountSQL()
	if err != nil {
		return 0, err
	}

	var count int64
	err = db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	return count, nil
}

// Exists checks if any records exist
func (qs *BaseQuerySet[T]) Exists(ctx context.Context) (bool, error) {
	if qs.err != nil {
		return false, qs.err
	}
	db, err := qs.getDB(ctx)
	if err != nil {
		return false, err
	}

	query, args, err := qs.buildExistsSQL()
	if err != nil {
		return false, err
	}

	var dummy int
	err = db.QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("exists query failed: %w", err)
	}

	return true, nil
}
