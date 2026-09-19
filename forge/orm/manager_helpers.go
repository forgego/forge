package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/forgego/forge/schema"
)

// ExecuteHooks executes model hooks in the correct order
func ExecuteHooks(ctx context.Context, instance interface{}, hookType string) error {
	// Check if instance implements Schema interface
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return nil // No hooks if not a schema
	}

	hooks := schemaInstance.Hooks()
	if hooks == nil {
		return nil
	}

	switch hookType {
	case "BeforeCreate":
		if hooks.BeforeSave != nil {
			if err := hooks.BeforeSave(ctx, instance); err != nil {
				return fmt.Errorf("BeforeSave hook failed: %w", err)
			}
		}
		if hooks.BeforeCreate != nil {
			if err := hooks.BeforeCreate(ctx, instance); err != nil {
				return fmt.Errorf("BeforeCreate hook failed: %w", err)
			}
		}
	case "AfterCreate":
		if hooks.AfterCreate != nil {
			if err := hooks.AfterCreate(ctx, instance); err != nil {
				return fmt.Errorf("AfterCreate hook failed: %w", err)
			}
		}
		if hooks.AfterSave != nil {
			if err := hooks.AfterSave(ctx, instance); err != nil {
				return fmt.Errorf("AfterSave hook failed: %w", err)
			}
		}
	case "BeforeUpdate":
		if hooks.BeforeSave != nil {
			if err := hooks.BeforeSave(ctx, instance); err != nil {
				return fmt.Errorf("BeforeSave hook failed: %w", err)
			}
		}
		if hooks.BeforeUpdate != nil {
			if err := hooks.BeforeUpdate(ctx, instance); err != nil {
				return fmt.Errorf("BeforeUpdate hook failed: %w", err)
			}
		}
	case "AfterUpdate":
		if hooks.AfterUpdate != nil {
			if err := hooks.AfterUpdate(ctx, instance); err != nil {
				return fmt.Errorf("AfterUpdate hook failed: %w", err)
			}
		}
		if hooks.AfterSave != nil {
			if err := hooks.AfterSave(ctx, instance); err != nil {
				return fmt.Errorf("AfterSave hook failed: %w", err)
			}
		}
	case "BeforeDelete":
		if hooks.BeforeDelete != nil {
			if err := hooks.BeforeDelete(ctx, instance); err != nil {
				return fmt.Errorf("BeforeDelete hook failed: %w", err)
			}
		}
	case "AfterDelete":
		if hooks.AfterDelete != nil {
			if err := hooks.AfterDelete(ctx, instance); err != nil {
				return fmt.Errorf("AfterDelete hook failed: %w", err)
			}
		}
	}

	return nil
}

// ValidateInstance runs Clean validation hook if available
func ValidateInstance(instance interface{}) error {
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return nil
	}

	hooks := schemaInstance.Hooks()
	if hooks == nil || hooks.Clean == nil {
		return nil
	}

	if err := hooks.Clean(instance); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// BuildInsertSQL builds an INSERT SQL statement from a model instance using default primary key column "id"
// and PostgreSQL placeholders.
func BuildInsertSQL(instance interface{}, tableName string) (sql string, values []interface{}, columns []string, err error) {
	return BuildInsertSQLForPK(instance, tableName, "id")
}

// BuildInsertSQLForPK builds an INSERT SQL statement from a model instance with a custom primary key column
// and optional placeholder function.
func BuildInsertSQLForPK(instance interface{}, tableName string, pkColumn string, placeholder ...func(int) string) (sql string, values []interface{}, columns []string, err error) {
	if pkColumn == "" {
		pkColumn = "id"
	}
	ph := defaultPlaceholder
	if len(placeholder) > 0 && placeholder[0] != nil {
		ph = placeholder[0]
	}

	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	if instanceValue.Kind() != reflect.Struct {
		return "", nil, nil, fmt.Errorf("instance must be a struct")
	}

	// Get schema to understand fields
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return "", nil, nil, fmt.Errorf("instance must implement schema.Schema")
	}

	// Iterate schema fields first, then get values only for those fields
	schemaFields := schemaInstance.Fields()

	var insertColumns []string
	var insertPlaceholders []string
	var insertValues []interface{}
	columnIndex := 1

	if len(schemaFields) > 0 {
		for _, schemaField := range schemaFields {
			if schemaField.PrimaryKey && schemaField.AutoIncrement {
				continue
			}
			fieldValue, err := getSchemaFieldValue(instance, schemaField)
			if err != nil {
				continue
			}
			fieldValueReflect := reflect.ValueOf(fieldValue)
			if !schemaField.Required && fieldValueReflect.IsZero() {
				continue
			}
			columnName := schemaField.DBColumn
			if columnName == "" {
				columnName = schemaField.Name
			}
			insertColumns = append(insertColumns, columnName)
			insertPlaceholders = append(insertPlaceholders, ph(columnIndex))
			insertValues = append(insertValues, fieldValue)
			columnIndex++
		}
	} else {
		// Fallback: derive columns from struct fields and tags
		typ := instanceValue.Type()
		for i := 0; i < typ.NumField(); i++ {
			sf := typ.Field(i)
			// Skip unexported
			if sf.PkgPath != "" {
				continue
			}
			tag := sf.Tag.Get("db")
			col := tag
			if col == "" || col == "-" {
				col = strings.ToLower(sf.Name)
			}
			// Skip id auto-increment
			if col == "id" || col == pkColumn {
				continue
			}
			val := instanceValue.Field(i)
			if !val.IsValid() || !val.CanInterface() {
				continue
			}
			if isZeroValue(val) {
				continue
			}
			insertColumns = append(insertColumns, col)
			insertPlaceholders = append(insertPlaceholders, ph(columnIndex))
			insertValues = append(insertValues, val.Interface())
			columnIndex++
		}
	}
	if len(insertColumns) == 0 {
		return "", nil, nil, fmt.Errorf("no fields to insert")
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
		EscapeIdentifier(tableName),
		strings.Join(EscapeIdentifierList(insertColumns), ", "),
		strings.Join(insertPlaceholders, ", "),
		EscapeIdentifier(pkColumn),
	)

	return insertSQL, insertValues, insertColumns, nil
}

func getSchemaFieldValue(instance interface{}, field schema.Field) (interface{}, error) {
	resolved, ok := schema.ResolveField(instance, field)
	if !ok {
		return nil, fmt.Errorf("field %s not found", field.Name)
	}
	value := reflect.ValueOf(instance)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil, fmt.Errorf("field %s is not accessible", field.Name)
		}
		value = value.Elem()
	}
	for _, index := range resolved.StructField.Index {
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return nil, fmt.Errorf("field %s is not accessible", field.Name)
			}
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %s is not accessible", field.Name)
		}
		value = value.Field(index)
	}
	if !value.IsValid() || !value.CanInterface() {
		return nil, fmt.Errorf("field %s is not accessible", field.Name)
	}
	return value.Interface(), nil
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

// BuildUpdateSQL builds an UPDATE SQL statement from a model instance
func BuildUpdateSQL(instance interface{}, tableName, idField string, placeholder ...func(int) string) (string, []interface{}, error) {
	ph := defaultPlaceholder
	if len(placeholder) > 0 && placeholder[0] != nil {
		ph = placeholder[0]
	}

	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	if instanceValue.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("instance must be a struct")
	}

	// Get schema to understand fields
	schemaInstance, ok := instance.(schema.Schema)
	if !ok {
		return "", nil, fmt.Errorf("instance must implement schema.Schema")
	}

	// Iterate schema fields first, then get values only for those fields
	schemaFields := schemaInstance.Fields()

	var setParts []string
	var values []interface{}
	paramIndex := 1
	var idValue interface{}

	for _, schemaField := range schemaFields {
		// Get column name from schema (DBColumn or Name)
		columnName := schemaField.DBColumn
		if columnName == "" {
			columnName = schemaField.Name
		}

		// Check if this is the ID field
		if strings.EqualFold(schemaField.Name, idField) || strings.EqualFold(columnName, idField) {
			// Get ID value
			val, err := getSchemaFieldValue(instance, schemaField)
			if err == nil {
				idValue = val
			}
			continue
		}

		// Skip primary key
		if schemaField.PrimaryKey {
			continue
		}

		// Get field value using helper function
		fieldValue, err := getSchemaFieldValue(instance, schemaField)
		if err != nil {
			// Field not found or not accessible - skip it
			continue
		}

		// Include field in UPDATE
		setParts = append(setParts, fmt.Sprintf("%s = %s", EscapeIdentifier(columnName), ph(paramIndex)))
		values = append(values, fieldValue)
		paramIndex++
	}

	if len(setParts) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	if idValue == nil {
		return "", nil, fmt.Errorf("id field not found or is zero")
	}

	// Add ID to the end for WHERE clause
	values = append(values, idValue)
	sql := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = %s",
		EscapeIdentifier(tableName),
		strings.Join(setParts, ", "),
		EscapeIdentifier(idField),
		ph(paramIndex),
	)

	return sql, values, nil
}

// BuildDeleteSQL builds a DELETE SQL statement
func BuildDeleteSQL(tableName, idField string, idValue interface{}, placeholder ...func(int) string) (string, []interface{}) {
	ph := defaultPlaceholder
	if len(placeholder) > 0 && placeholder[0] != nil {
		ph = placeholder[0]
	}
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", EscapeIdentifier(tableName), EscapeIdentifier(idField), ph(1))
	return sql, []interface{}{idValue}
}

// ExecuteInsertTx executes an INSERT statement using the provided DBTX and dialect, returning the generated ID.
func ExecuteInsertTx(ctx context.Context, dbtx DBTX, d dialect.Dialect, sql string, args []interface{}) (int64, error) {
	if dbtx == nil {
		return 0, fmt.Errorf("database execution handle is nil")
	}

	var id int64
	err := dbtx.QueryRowContext(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	return id, nil
}

// ExecuteInsert executes an INSERT statement and returns the generated ID.
func ExecuteInsert(ctx context.Context, database *db.DB, sql string, args []interface{}) (int64, error) {
	if database == nil || database.DB == nil {
		return 0, fmt.Errorf("database connection not set")
	}
	return ExecuteInsertTx(ctx, database.DB, database.Dialect(), database.RebindPlaceholders(sql), args)
}

// BuildBulkInsertSQL builds a bulk INSERT SQL statement for multiple instances using default primary key column "id"
// and PostgreSQL placeholders.
func BuildBulkInsertSQL(instances []interface{}, tableName string) (sql string, values []interface{}, columns []string, err error) {
	return BuildBulkInsertSQLForPK(instances, tableName, "id")
}

// BuildBulkInsertSQLForPK builds a bulk INSERT SQL statement for multiple instances with a custom primary key column
// and optional placeholder function.
func BuildBulkInsertSQLForPK(instances []interface{}, tableName string, pkColumn string, placeholder ...func(int) string) (sql string, values []interface{}, columns []string, err error) {
	if len(instances) == 0 {
		return "", nil, nil, fmt.Errorf("no instances to insert")
	}
	if pkColumn == "" {
		pkColumn = "id"
	}
	ph := defaultPlaceholder
	if len(placeholder) > 0 && placeholder[0] != nil {
		ph = placeholder[0]
	}

	// Use first instance to determine columns
	firstInstance := instances[0]
	_, _, columns, err = BuildInsertSQLForPK(firstInstance, tableName, pkColumn, ph)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to build insert SQL for first instance: %w", err)
	}

	if len(columns) == 0 {
		return "", nil, nil, fmt.Errorf("no columns to insert")
	}

	// Build VALUES clause for all instances
	var valueClauses []string
	var allValues []interface{}
	paramIndex := 1

	for _, instance := range instances {
		// Get values for this instance
		_, instanceValues, instanceColumns, err := BuildInsertSQLForPK(instance, tableName, pkColumn, ph)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to build insert SQL for instance: %w", err)
		}
		if len(instanceColumns) != len(columns) || strings.Join(instanceColumns, ",") != strings.Join(columns, ",") {
			return "", nil, nil, fmt.Errorf("bulk insert requires consistent columns across instances: expected [%s], got [%s]",
				strings.Join(columns, ", "),
				strings.Join(instanceColumns, ", "),
			)
		}

		// Build placeholders for this row
		var placeholders []string
		for range instanceValues {
			placeholders = append(placeholders, ph(paramIndex))
			paramIndex++
		}
		valueClauses = append(valueClauses, "("+strings.Join(placeholders, ", ")+")")
		allValues = append(allValues, instanceValues...)
	}

	// Build final SQL: INSERT INTO table (cols) VALUES (row1), (row2), ... RETURNING id
	escapedColumns := make([]string, len(columns))
	for i, col := range columns {
		escapedColumns[i] = EscapeIdentifier(col)
	}

	sql = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s RETURNING %s",
		EscapeIdentifier(tableName),
		strings.Join(escapedColumns, ", "),
		strings.Join(valueClauses, ", "),
		EscapeIdentifier(pkColumn),
	)

	return sql, allValues, columns, nil
}

// ExecuteBulkInsertTx executes a bulk INSERT statement using the provided DBTX and dialect, returning all generated IDs.
func ExecuteBulkInsertTx(ctx context.Context, dbtx DBTX, d dialect.Dialect, sql string, args []interface{}) ([]int64, error) {
	if dbtx == nil {
		return nil, fmt.Errorf("database execution handle is nil")
	}

	rows, err := dbtx.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("bulk insert failed: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan ID: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ids, nil
}

// ExecuteBulkInsert executes a bulk INSERT statement and returns all generated IDs.
func ExecuteBulkInsert(ctx context.Context, database *db.DB, sql string, args []interface{}) ([]int64, error) {
	if database == nil || database.DB == nil {
		return nil, fmt.Errorf("database connection not set")
	}
	return ExecuteBulkInsertTx(ctx, database.DB, database.Dialect(), database.RebindPlaceholders(sql), args)
}

// ExecuteUpdateTx executes an UPDATE statement using the provided DBTX and dialect, returning rows affected.
func ExecuteUpdateTx(ctx context.Context, dbtx DBTX, d dialect.Dialect, sql string, args []interface{}) (int64, error) {
	if dbtx == nil {
		return 0, fmt.Errorf("database execution handle is nil")
	}

	result, err := dbtx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ExecuteUpdate executes an UPDATE statement and returns rows affected.
func ExecuteUpdate(ctx context.Context, database *db.DB, sql string, args []interface{}) (int64, error) {
	if database == nil || database.DB == nil {
		return 0, fmt.Errorf("database connection not set")
	}
	return ExecuteUpdateTx(ctx, database.DB, database.Dialect(), database.RebindPlaceholders(sql), args)
}

// ExecuteDeleteTx executes a DELETE statement using the provided DBTX and dialect, returning rows affected.
func ExecuteDeleteTx(ctx context.Context, dbtx DBTX, d dialect.Dialect, sql string, args []interface{}) (int64, error) {
	if dbtx == nil {
		return 0, fmt.Errorf("database execution handle is nil")
	}

	result, err := dbtx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ExecuteDelete executes a DELETE statement and returns rows affected.
func ExecuteDelete(ctx context.Context, database *db.DB, sql string, args []interface{}) (int64, error) {
	if database == nil || database.DB == nil {
		return 0, fmt.Errorf("database connection not set")
	}
	return ExecuteDeleteTx(ctx, database.DB, database.Dialect(), database.RebindPlaceholders(sql), args)
}

// GetIDValue extracts the ID value from an instance
func GetIDValue(instance interface{}, idFieldName string) (interface{}, error) {
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	if instanceValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("instance must be a struct")
	}

	var findValue func(v reflect.Value) (reflect.Value, bool, error)
	findValue = func(v reflect.Value) (reflect.Value, bool, error) {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// Recurse into anonymous embedded structs.
			if field.Anonymous {
				switch fieldValue.Kind() {
				case reflect.Struct:
					if nested, ok, err := findValue(fieldValue); err != nil {
						return reflect.Value{}, false, err
					} else if ok {
						return nested, true, nil
					}
				case reflect.Ptr:
					if fieldValue.IsNil() {
						continue
					}
					if fieldValue.Elem().Kind() == reflect.Struct {
						if nested, ok, err := findValue(fieldValue.Elem()); err != nil {
							return reflect.Value{}, false, err
						} else if ok {
							return nested, true, nil
						}
					}
				}
			}

			// Check field name
			if strings.EqualFold(field.Name, idFieldName) {
				if !fieldValue.CanInterface() {
					return reflect.Value{}, false, fmt.Errorf("id field is not accessible")
				}
				return fieldValue, true, nil
			}

			// Check db tag
			if dbTag := field.Tag.Get("db"); dbTag != "" {
				tagParts := strings.Split(dbTag, ",")
				if strings.EqualFold(tagParts[0], idFieldName) {
					if !fieldValue.CanInterface() {
						return reflect.Value{}, false, fmt.Errorf("id field is not accessible")
					}
					return fieldValue, true, nil
				}
			}
		}
		return reflect.Value{}, false, nil
	}

	val, ok, err := findValue(instanceValue)
	if err != nil {
		return nil, err
	}
	if ok {
		return val.Interface(), nil
	}
	return nil, fmt.Errorf("id field '%s' not found", idFieldName)
}
