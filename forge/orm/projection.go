package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

// ProjectionQuerySet provides type-safe projection operations
type ProjectionQuerySet[T any, P any] struct {
	base      QuerySet[T]
	projection *Projection[P]
}

// Projection represents a type-safe projection
type Projection[P any] struct {
	fields    []string
	fieldTypes map[string]reflect.Type
	targetType reflect.Type
}

// Project creates a type-safe projection QuerySet
// Usage: User.Objects.Project[UserProjection]().All(ctx)
func Project[T any, P any](qs QuerySet[T]) *ProjectionQuerySet[T, P] {
	// Get target type information
	var zero P
	targetType := reflect.TypeOf(zero)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	// Extract field names from projection type
	fields := make([]string, 0, targetType.NumField())
	fieldTypes := make(map[string]reflect.Type)

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		// Check for db tag for column name
		if dbTag := field.Tag.Get("db"); dbTag != "" && dbTag != "-" {
			fieldName = dbTag
		}

		fields = append(fields, fieldName)
		fieldTypes[fieldName] = field.Type
	}

	projection := &Projection[P]{
		fields:     fields,
		fieldTypes: fieldTypes,
		targetType: targetType,
	}

	return &ProjectionQuerySet[T, P]{
		base:       qs,
		projection: projection,
	}
}

// All executes the projection query and returns typed results
func (pqs *ProjectionQuerySet[T, P]) All(ctx context.Context) ([]*P, error) {
	// Get the base QuerySet's database connection
	baseQS, ok := pqs.base.(*BaseQuerySet[T])
	if !ok {
		return nil, fmt.Errorf("projection requires BaseQuerySet")
	}

	var db *sql.DB
	db, err := baseQS.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Build SQL with only selected fields
	sqlBuilder := NewSQLBuilder()
	
	// Build SELECT clause with projection fields
	selectClause := "SELECT "
	for i, field := range pqs.projection.fields {
		if i > 0 {
			selectClause += ", "
		}
		selectClause += EscapeIdentifier(field)
	}
	selectClause += " FROM " + EscapeIdentifier(baseQS.table)

	// Build WHERE clause from conditions
	whereClause := ""
	args := []interface{}{}

	if len(baseQS.conditions) > 0 {
		whereParts := []string{}
		for _, expr := range baseQS.conditions {
			sql, sqlArgs, err := expr.ToSQL(sqlBuilder)
			if err != nil {
				return nil, fmt.Errorf("failed to build SQL: %w", err)
			}
			whereParts = append(whereParts, sql)
			args = append(args, sqlArgs...)
		}
		if len(whereParts) > 0 {
			whereClause = " WHERE " + joinWithAnd(whereParts)
		}
	}

	// Build ORDER BY
	orderClause := ""
	if len(baseQS.orderBy) > 0 {
		orderParts := []string{}
		for _, order := range baseQS.orderBy {
			dir := "ASC"
			if !order.Ascending {
				dir = "DESC"
			}
			orderParts = append(orderParts, EscapeIdentifier(order.Field)+" "+dir)
		}
		orderClause = " ORDER BY " + joinStrings(orderParts, ", ")
	}

	// Build LIMIT/OFFSET
	limitClause := ""
	if baseQS.limitVal != nil {
		limitClause = fmt.Sprintf(" LIMIT %d", *baseQS.limitVal)
	}
	if baseQS.offsetVal != nil {
		limitClause += fmt.Sprintf(" OFFSET %d", *baseQS.offsetVal)
	}

	// Execute query
	query := selectClause + whereClause + orderClause + limitClause
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute projection query: %w", err)
	}
	defer rows.Close()

	// Scan results into projection type
	results := []*P{}
	for rows.Next() {
		// Create new instance of projection type
		projectionPtr := reflect.New(pqs.projection.targetType)
		projectionValue := projectionPtr.Elem()

		// Create scan targets
		scanTargets := make([]interface{}, len(pqs.projection.fields))
		for i, fieldName := range pqs.projection.fields {
			fieldType := pqs.projection.fieldTypes[fieldName]
			fieldValue := projectionValue.FieldByName(getFieldNameFromColumn(fieldName, pqs.projection.targetType))
			
			if !fieldValue.IsValid() {
				return nil, fmt.Errorf("field %s not found in projection type", fieldName)
			}

			// Create pointer to field for scanning
			if fieldValue.CanAddr() {
				scanTargets[i] = fieldValue.Addr().Interface()
			} else {
				// For unaddressable fields, create a temporary
				temp := reflect.New(fieldType).Elem()
				scanTargets[i] = temp.Addr().Interface()
			}
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("failed to scan projection row: %w", err)
		}

		// Copy scanned values back to projection
		for i, fieldName := range pqs.projection.fields {
			fieldValue := projectionValue.FieldByName(getFieldNameFromColumn(fieldName, pqs.projection.targetType))
			if fieldValue.IsValid() && fieldValue.CanSet() {
				scannedValue := reflect.ValueOf(scanTargets[i]).Elem()
				if scannedValue.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(scannedValue)
				}
			}
		}

		results = append(results, projectionPtr.Interface().(*P))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projection rows: %w", err)
	}

	return results, nil
}

// getFieldNameFromColumn finds the struct field name from a column name
func getFieldNameFromColumn(columnName string, structType reflect.Type) string {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		// Check db tag
		if dbTag := field.Tag.Get("db"); dbTag == columnName {
			return field.Name
		}

		// Check if field name matches (case-insensitive)
		if toSnakeCase(field.Name) == columnName {
			return field.Name
		}
	}

	// Fallback: try exact match
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Name == columnName {
			return field.Name
		}
	}

	return columnName
}

// Helper functions
func joinWithAnd(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += " AND " + part
	}
	return result
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, part := range parts[1:] {
		result += sep + part
	}
	return result
}

// toSnakeCase converts CamelCase to snake_case (simplified)
func toSnakeCase(s string) string {
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}
