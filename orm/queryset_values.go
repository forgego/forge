package orm

import (
	"context"
	"fmt"
)

// ValuesQuerySet interface for values queries
type ValuesQuerySet[T any] interface {
	All(ctx context.Context) ([]map[string]interface{}, error)
	Get(ctx context.Context) (map[string]interface{}, error)
	First(ctx context.Context) (map[string]interface{}, error)
}

// ValuesListQuerySet interface for values_list queries
type ValuesListQuerySet[T any] interface {
	All(ctx context.Context) ([][]interface{}, error)
	Get(ctx context.Context) ([]interface{}, error)
	First(ctx context.Context) ([]interface{}, error)
	Flat(ctx context.Context) ([]interface{}, error)
}

// BaseValuesQuerySet implementation
type BaseValuesQuerySet[T any] struct {
	base *BaseQuerySet[T]
}

func (vqs *BaseValuesQuerySet[T]) All(ctx context.Context) ([]map[string]interface{}, error) {
	if vqs.base.err != nil {
		return nil, vqs.base.err
	}

	db, err := vqs.base.getDB(ctx)
	if err != nil {
		return nil, err
	}

	sql, args, err := vqs.base.buildSQL()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create scan destination
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		for i := range scanArgs {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create map
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

func (vqs *BaseValuesQuerySet[T]) Get(ctx context.Context) (map[string]interface{}, error) {
	if vqs.base.err != nil {
		return nil, vqs.base.err
	}

	fields := make([]any, len(vqs.base.selectFields))
	for i, f := range vqs.base.selectFields {
		fields[i] = f
	}
	results, err := vqs.base.Limit(2).Values(fields...).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", vqs.base.table)
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d", vqs.base.table, len(results))
	}

	return results[0], nil
}

func (vqs *BaseValuesQuerySet[T]) First(ctx context.Context) (map[string]interface{}, error) {
	if vqs.base.err != nil {
		return nil, vqs.base.err
	}

	fields := make([]any, len(vqs.base.selectFields))
	for i, f := range vqs.base.selectFields {
		fields[i] = f
	}
	results, err := vqs.base.Limit(1).Values(fields...).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", vqs.base.table)
	}

	return results[0], nil
}

// BaseValuesListQuerySet implementation
type BaseValuesListQuerySet[T any] struct {
	base *BaseQuerySet[T]
}

func (vls *BaseValuesListQuerySet[T]) All(ctx context.Context) ([][]interface{}, error) {
	if vls.base.err != nil {
		return nil, vls.base.err
	}

	db, err := vls.base.getDB(ctx)
	if err != nil {
		return nil, err
	}

	sql, args, err := vls.base.buildSQL()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results [][]interface{}
	for rows.Next() {
		// Create scan destination
		scanArgs := make([]interface{}, len(columns))
		values := make([]interface{}, len(columns))
		for i := range scanArgs {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create tuple (slice)
		tuple := make([]interface{}, len(values))
		copy(tuple, values)
		results = append(results, tuple)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

func (vls *BaseValuesListQuerySet[T]) Get(ctx context.Context) ([]interface{}, error) {
	if vls.base.err != nil {
		return nil, vls.base.err
	}

	// Create a new values list query set with limit
	limited := vls.base.clone()
	limit := 2
	limited.limitVal = &limit

	fields := make([]any, len(limited.selectFields))
	for i, f := range limited.selectFields {
		fields[i] = f
	}
	results, err := limited.ValuesList(fields...).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", vls.base.table)
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d", vls.base.table, len(results))
	}

	return results[0], nil
}

func (vls *BaseValuesListQuerySet[T]) First(ctx context.Context) ([]interface{}, error) {
	if vls.base.err != nil {
		return nil, vls.base.err
	}

	// Create a new values list query set with limit
	limited := vls.base.clone()
	limit := 1
	limited.limitVal = &limit

	fields := make([]any, len(limited.selectFields))
	for i, f := range limited.selectFields {
		fields[i] = f
	}
	results, err := limited.ValuesList(fields...).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", vls.base.table)
	}

	return results[0], nil
}

func (vls *BaseValuesListQuerySet[T]) Flat(ctx context.Context) ([]interface{}, error) {
	if vls.base.err != nil {
		return nil, vls.base.err
	}

	// For flat, we expect exactly one field
	if len(vls.base.selectFields) != 1 {
		return nil, fmt.Errorf("Flat() requires exactly one field, got %d", len(vls.base.selectFields))
	}

	results, err := vls.All(ctx)
	if err != nil {
		return nil, err
	}

	// Extract first element from each tuple
	flat := make([]interface{}, len(results))
	for i, tuple := range results {
		if len(tuple) > 0 {
			flat[i] = tuple[0]
		}
	}

	return flat, nil
}
