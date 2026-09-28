package orm

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// columnFor resolves a field or column key to the database column name.
func (qs *BaseQuerySet[T]) columnFor(key string) (string, error) {
	if qs == nil || qs.schema == nil {
		return key, nil
	}
	field := qs.schema.GetField(key)
	if field == nil {
		target := qs.table
		if target == "" && qs.schema != nil {
			target = qs.schema.TableName
		}
		return "", fmt.Errorf("field %s not found on %s", key, target)
	}
	return field.DBColumn, nil
}

// Update performs a bulk update
func (qs *BaseQuerySet[T]) Update(ctx context.Context, updates UpdateMap) (int64, error) {
	if qs.err != nil {
		return 0, qs.err
	}
	if len(updates) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	db, err := qs.getDB(ctx)
	if err != nil {
		return 0, err
	}

	builder := qs.newSQLBuilder()
	var joinCalled bool
	builder.SetJoinResolver(func(parts []string) (string, string, error) {
		joinCalled = true
		return "", "", fmt.Errorf("filtering by related fields is not supported in update")
	})

	// Build SET clause
	var setParts []string

	keys := make([]string, 0, len(updates))
	for fieldName := range updates {
		keys = append(keys, fieldName)
	}
	sort.Strings(keys)

	for _, fieldName := range keys {
		value := updates[fieldName]
		col, err := qs.columnFor(fieldName)
		if err != nil {
			return 0, err
		}
		escapedField := EscapeIdentifier(col)

		// Check if value is an Expression
		if expr, ok := value.(Expression); ok {
			// Build expression SQL
			exprSQL, _, err := expr.ToSQL(builder)
			if err != nil {
				if joinCalled {
					return 0, fmt.Errorf("filtering by related fields is not supported in update")
				}
				return 0, fmt.Errorf("failed to build expression SQL for field %s: %w", fieldName, err)
			}
			setParts = append(setParts, fmt.Sprintf("%s = %s", escapedField, exprSQL))
		} else {
			// Regular value
			placeholder := builder.AddArg(value)
			setParts = append(setParts, fmt.Sprintf("%s = %s", escapedField, placeholder))
		}
	}

	// Build WHERE clause
	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		if joinCalled {
			return 0, fmt.Errorf("filtering by related fields is not supported in update")
		}
		return 0, whereErr
	}
	if joinCalled {
		return 0, fmt.Errorf("filtering by related fields is not supported in update")
	}

	// Combine all args
	allArgs := builder.Args()

	// Build SQL
	// Table name is quoted with EscapeIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	updateSQL := fmt.Sprintf("UPDATE %s SET %s", EscapeIdentifier(qs.table), strings.Join(setParts, ", "))
	if whereClause != "" {
		updateSQL += " " + whereClause
	}

	// Execute
	result, err := db.ExecContext(ctx, updateSQL, allArgs...)
	if err != nil {
		return 0, fmt.Errorf("update query failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BulkUpdate performs bulk updates by applying each UpdateMap entry sequentially.
// Each entry in the updates slice is applied as a separate UPDATE statement against
// the current QuerySet filters. For true single-statement bulk updates with
// different values per row, use raw SQL.
func (qs *BaseQuerySet[T]) BulkUpdate(ctx context.Context, updates []UpdateMap) error {
	if qs.err != nil {
		return qs.err
	}
	if len(updates) == 0 {
		return nil
	}
	for i, update := range updates {
		if _, err := qs.Update(ctx, update); err != nil {
			return fmt.Errorf("BulkUpdate failed at index %d: %w", i, err)
		}
	}
	return nil
}

// Delete performs a bulk delete
func (qs *BaseQuerySet[T]) Delete(ctx context.Context) (int64, error) {
	if qs.err != nil {
		return 0, qs.err
	}
	db, err := qs.getDB(ctx)
	if err != nil {
		return 0, err
	}

	builder := qs.newSQLBuilder()
	var joinCalled bool
	builder.SetJoinResolver(func(parts []string) (string, string, error) {
		joinCalled = true
		return "", "", fmt.Errorf("filtering by related fields is not supported in delete")
	})

	// Build WHERE clause
	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		if joinCalled {
			return 0, fmt.Errorf("filtering by related fields is not supported in delete")
		}
		return 0, whereErr
	}
	if joinCalled {
		return 0, fmt.Errorf("filtering by related fields is not supported in delete")
	}

	// Build SQL
	// Table name is quoted with EscapeIdentifier; values are bound parameters.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query, go.lang.security.audit.database.string-formatted-query
	deleteSQL := fmt.Sprintf("DELETE FROM %s", EscapeIdentifier(qs.table))
	if whereClause != "" {
		deleteSQL += " " + whereClause
	}

	args := builder.Args()

	// Execute
	result, err := db.ExecContext(ctx, deleteSQL, args...)
	if err != nil {
		return 0, fmt.Errorf("delete query failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
