package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/forgego/forge/utils"
)

// scanRows scans rows into model instances
func (qs *BaseQuerySet[T]) scanRows(rows *sql.Rows) ([]*T, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Build column-to-field mapping from schema
	fieldMap := qs.buildFieldMap(columns)

	var results []*T
	for rows.Next() {
		instance := new(T)
		scanArgs, postScan := qs.prepareScanArgs(instance, columns, fieldMap)

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Post-process related fields
		if postScan != nil {
			postScan()
		}

		results = append(results, instance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

type scanField struct {
	fieldInfo    *FieldInfo
	relationName string
}

// buildFieldMap creates a mapping from column names to field info
func (qs *BaseQuerySet[T]) buildFieldMap(columns []string) map[string]scanField {
	fieldMap := make(map[string]scanField)
	for _, col := range columns {
		// Check related columns first (Relation__Column)
		if strings.Contains(col, "__") {
			parts := strings.SplitN(col, "__", 2)
			relName := parts[0]
			relCol := parts[1]

			rel := qs.schema.GetRelation(relName)
			if rel == nil {
				continue
			}

			if !qs.joinMap[relName] && !qs.joinMap[rel.Name] && !qs.joinMap[strings.ToLower(rel.Name)] && !qs.joinMap[rel.TargetModel] && !qs.joinMap[strings.ToLower(rel.TargetModel)] {
				// Might be a manual alias or something else, skip relation logic
				continue
			}

			targetSchema, err := GetModelSchemaByName(rel.TargetModel)
			if err != nil {
				continue
			}

			// Find field in target schema
			var targetField *FieldInfo
			for i := range targetSchema.Fields {
				f := &targetSchema.Fields[i]
				if strings.EqualFold(f.DBColumn, relCol) || strings.EqualFold(f.Name, relCol) {
					targetField = f
					break
				}
			}

			if targetField != nil {
				fieldMap[col] = scanField{
					fieldInfo:    targetField,
					relationName: rel.Name,
				}
				continue
			}
		}

		for i := range qs.schema.Fields {
			field := &qs.schema.Fields[i]
			if strings.EqualFold(field.DBColumn, col) || strings.EqualFold(field.Name, col) {
				fieldMap[col] = scanField{fieldInfo: field}
				break
			}
		}
	}
	return fieldMap
}

// prepareScanArgs prepares scan arguments using schema
func (qs *BaseQuerySet[T]) prepareScanArgs(instance *T, columns []string, fieldMap map[string]scanField) ([]interface{}, func()) {
	scanArgs := make([]interface{}, len(columns))
	instanceValue := reflect.ValueOf(instance).Elem()

	// Map relationName -> map[fieldName]holder
	relatedHolders := make(map[string]map[string]*interface{})
	// Track optional local fields scanned into holders
	type localHolder struct {
		field  reflect.Value
		holder *interface{}
	}
	localHolders := make([]localHolder, 0)

	for i, col := range columns {
		info, ok := fieldMap[col]
		if !ok {
			var val interface{}
			scanArgs[i] = &val
			continue
		}

		if info.relationName != "" {
			// Related field
			holder := new(interface{})
			scanArgs[i] = holder

			if relatedHolders[info.relationName] == nil {
				relatedHolders[info.relationName] = make(map[string]*interface{})
			}
			relatedHolders[info.relationName][info.fieldInfo.Name] = holder
		} else {
			// Local field
			var field reflect.Value
			if info.fieldInfo.StructFieldName != "" {
				field = instanceValue.FieldByName(info.fieldInfo.StructFieldName)
			} else {
				field = instanceValue.FieldByName(info.fieldInfo.Name)
				if !field.IsValid() {
					field = instanceValue.FieldByName(utils.ToPascal(info.fieldInfo.Name))
				}
			}
			if field.IsValid() && field.CanSet() {
				// For optional fields, scan into a holder to tolerate NULLs
				if !info.fieldInfo.Required && field.Kind() != reflect.Ptr {
					holder := new(interface{})
					scanArgs[i] = holder
					localHolders = append(localHolders, localHolder{field: field, holder: holder})
				} else {
					scanArgs[i] = field.Addr().Interface()
				}
			} else {
				var val interface{}
				scanArgs[i] = &val
			}
		}
	}

	postScan := func() {
		// Populate optional local fields from holders
		for _, lh := range localHolders {
			if lh.holder == nil || *lh.holder == nil {
				continue
			}
			setFieldValue(lh.field, *lh.holder)
		}

		for relName, fields := range relatedHolders {
			rel := qs.schema.GetRelation(relName)
			if rel == nil {
				continue
			}

			targetSchema, err := GetModelSchemaByName(rel.TargetModel)
			if err != nil {
				continue
			}

			// Check PK
			pkField := targetSchema.PrimaryKey
			// Find holder for PK
			var pkFieldName string
			for _, f := range targetSchema.Fields {
				if f.DBColumn == pkField {
					pkFieldName = f.Name
					break
				}
			}

			pkHolder, ok := fields[pkFieldName]
			if !ok || pkHolder == nil || *pkHolder == nil {
				continue // No PK -> relation is nil
			}

			// Locate relation field on instance struct
			relField := instanceValue.FieldByName(relName)
			if !relField.IsValid() {
				relField = instanceValue.FieldByName(utils.ToPascal(relName))
			}
			if !relField.IsValid() && rel != nil {
				relField = instanceValue.FieldByName(rel.TargetModel)
				if !relField.IsValid() {
					relField = instanceValue.FieldByName(utils.ToPascal(rel.TargetModel))
				}
			}
			if !relField.IsValid() {
				trimmed := strings.TrimSuffix(strings.TrimSuffix(relName, "_id"), "ID")
				relField = instanceValue.FieldByName(utils.ToPascal(trimmed))
			}

			if relField.IsValid() && relField.CanSet() {
				var nestedVal reflect.Value
				if relField.Kind() == reflect.Ptr {
					val := reflect.New(relField.Type().Elem())
					relField.Set(val)
					nestedVal = val.Elem()
				} else if relField.Kind() == reflect.Struct {
					nestedVal = relField
				} else {
					continue
				}

				// Populate fields
				for fName, holder := range fields {
					if holder == nil || *holder == nil {
						continue
					}

					fInfo := targetSchema.GetField(fName)
					if fInfo == nil {
						continue
					}

					structField := nestedVal.FieldByName(fInfo.Name)
					if fInfo.StructFieldName != "" {
						structField = nestedVal.FieldByName(fInfo.StructFieldName)
					}
					if !structField.IsValid() {
						structField = nestedVal.FieldByName(utils.ToPascal(fInfo.Name))
					}

					if structField.IsValid() && structField.CanSet() {
						setFieldValue(structField, *holder)
					}
				}
			}
		}
	}

	return scanArgs, postScan
}

// Helper to set field value with type conversion
func setFieldValue(field reflect.Value, value interface{}) {
	val := reflect.ValueOf(value)
	if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
		return
	}

	raw := value
	if b, ok := value.([]byte); ok {
		raw = string(b)
	}

	switch field.Kind() {
	case reflect.String:
		if s, ok := raw.(string); ok {
			field.SetString(s)
		}
	case reflect.Bool:
		switch v := raw.(type) {
		case string:
			if parsed, err := strconv.ParseBool(v); err == nil {
				field.SetBool(parsed)
			}
		case int64:
			field.SetBool(v != 0)
		case int32:
			field.SetBool(v != 0)
		case int:
			field.SetBool(v != 0)
		case float64:
			field.SetBool(v != 0)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := raw.(type) {
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(parsed)
			}
		case float64:
			field.SetInt(int64(v))
		case int:
			field.SetInt(int64(v))
		case int32:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v := raw.(type) {
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
				field.SetUint(parsed)
			}
		case float64:
			field.SetUint(uint64(v))
		case int:
			if v >= 0 {
				field.SetUint(uint64(v))
			}
		case int64:
			if v >= 0 {
				field.SetUint(uint64(v))
			}
		case uint64:
			field.SetUint(v)
		}
	case reflect.Float32, reflect.Float64:
		switch v := raw.(type) {
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				field.SetFloat(parsed)
			}
		case float64:
			field.SetFloat(v)
		case float32:
			field.SetFloat(float64(v))
		case int:
			field.SetFloat(float64(v))
		case int64:
			field.SetFloat(float64(v))
		}
	}
}
