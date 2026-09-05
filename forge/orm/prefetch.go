package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/forgego/forge/db/dialect"
)

// prefetch handles prefetching of related objects
func (qs *BaseQuerySet[T]) prefetch(ctx context.Context, results []*T) error {
	if len(results) == 0 || len(qs.prefetchRelated) == 0 {
		return nil
	}

	for _, path := range qs.prefetchRelated {
		rel := qs.schema.GetRelation(path)
		if rel == nil {
			continue
		}

		if rel.Type == RelationForeignKey {
			if err := qs.prefetchForeignKey(ctx, results, rel); err != nil {
				return err
			}
		} else if rel.Type == RelationManyToMany {
			if err := qs.prefetchManyToMany(ctx, results, rel); err != nil {
				return err
			}
		}
	}
	return nil
}

// prefetchForeignKey handles prefetching for ForeignKey relations
func (qs *BaseQuerySet[T]) prefetchForeignKey(ctx context.Context, results []*T, rel *RelationInfo) error {
	// 1. Collect foreign keys
	ids := make([]interface{}, 0, len(results))
	idMap := make(map[interface{}]bool)

	// Determine FK field name on source model
	// rel.FieldName is the struct field name of the relation (e.g. "User")
	// We need the FK field (e.g. "UserID")
	// RelationInfo doesn't explicitly store "SourceField", but we can try to infer or check schema
	fkField := ""
	for _, f := range qs.schema.Fields {
		if f.StructFieldName == rel.Name+"ID" {
			fkField = f.StructFieldName
			break
		}
		// Try to match by type if it's a pointer to the relation type? No, simple naming convention for now.
	}

	// Fallback: Check if there's a field with "ID" suffix
	if fkField == "" {
		fkField = rel.Name + "ID"
	}

	// Iterate results to collect IDs
	for _, result := range results {
		val := reflect.ValueOf(result).Elem()
		field := val.FieldByName(fkField)
		if !field.IsValid() {
			continue
		}

		id := field.Interface()
		// Check for zero value
		if isZero(id) {
			continue
		}

		if !idMap[id] {
			ids = append(ids, id)
			idMap[id] = true
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// 2. Query target model
	targetSchema, err := GetModelSchemaByName(rel.TargetModel)
	if err != nil {
		return err
	}

	targetResults, err := qs.fetchByIDs(ctx, targetSchema, ids)
	if err != nil {
		return err
	}

	// 3. Map results
	resultMap := make(map[interface{}]reflect.Value)
	for _, res := range targetResults {
		pk := getPKValue(res, targetSchema)
		resultMap[pk] = res
	}

	// 4. Assign back
	for _, result := range results {
		val := reflect.ValueOf(result).Elem()
		fkFieldVal := val.FieldByName(fkField)
		if !fkFieldVal.IsValid() {
			continue
		}

		id := fkFieldVal.Interface()
		if target, ok := resultMap[id]; ok {
			relField := val.FieldByName(rel.Name)
			if relField.IsValid() && relField.CanSet() {
				// Handle pointer
				if relField.Kind() == reflect.Ptr {
					// target is already a pointer (fetchByIDs returns pointers)
					relField.Set(target)
				} else {
					relField.Set(target.Elem())
				}
			}
		}
	}

	return nil
}

// prefetchManyToMany handles prefetching for ManyToMany relations
func (qs *BaseQuerySet[T]) prefetchManyToMany(ctx context.Context, results []*T, rel *RelationInfo) error {
	// 1. Collect source IDs
	sourceIDs := make([]interface{}, 0, len(results))
	for _, result := range results {
		val := reflect.ValueOf(result).Elem()
		pkField := val.FieldByName(qs.schema.GetField(qs.schema.PrimaryKey).StructFieldName)
		if pkField.IsValid() {
			sourceIDs = append(sourceIDs, pkField.Interface())
		}
	}

	if len(sourceIDs) == 0 {
		return nil
	}

	// 2. Query Through Table
	db, err := qs.getDB(ctx)
	if err != nil {
		return err
	}

	// Get dialect for placeholder generation
	d, err := qs.getDialect()
	if err != nil {
		return err
	}

	throughTable := rel.Through
	if throughTable == "" {
		// Infer through table name
		// For MVP, simplistic: table1_table2 sorted?
		// Or assume standard naming if not provided.
		// For now, fail if not provided, but usually schema builder should have set default.
		return fmt.Errorf("through table not defined for M2M relation %s", rel.Name)
	}

	// Infer columns
	// source_id, target_id
	// This assumes standard naming: {model}_id
	// Need proper column names from schema or relation info.
	// RelationInfo assumes we know.
	// Let's assume simplistic for MVP or use what we have.
	targetSchema, err := GetModelSchemaByName(rel.TargetModel)
	if err != nil {
		return err
	}

	// Handle singularization if needed.
	// Usually standard is singular_id.
	// Let's try to be smart or rely on user providing explicit Through table with columns.
	// For now, assume {singular_table}_id.
	sourceCol := strings.TrimSuffix(qs.table, "s") + "_id"
	targetCol := strings.TrimSuffix(targetSchema.TableName, "s") + "_id"

	query := fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s IN (%s)",
		EscapeIdentifier(sourceCol),
		EscapeIdentifier(targetCol),
		EscapeIdentifier(throughTable),
		EscapeIdentifier(sourceCol),
		d.BuildPlaceholders(len(sourceIDs)),
	)

	rows, err := db.QueryContext(ctx, query, sourceIDs...)
	if err != nil {
		return fmt.Errorf("failed to query through table: %w", err)
	}
	defer rows.Close()

	// Map sourceID -> []targetID
	relationMap := make(map[interface{}][]interface{})
	allTargetIDs := make([]interface{}, 0)
	targetIDMap := make(map[interface{}]bool)

	for rows.Next() {
		var sID, tID interface{}
		if err := rows.Scan(&sID, &tID); err != nil {
			return err
		}
		relationMap[sID] = append(relationMap[sID], tID)
		if !targetIDMap[tID] {
			allTargetIDs = append(allTargetIDs, tID)
			targetIDMap[tID] = true
		}
	}

	if len(allTargetIDs) == 0 {
		return nil
	}

	// 3. Fetch Targets
	targetResults, err := qs.fetchByIDs(ctx, targetSchema, allTargetIDs)
	if err != nil {
		return err
	}

	// Map targetID -> TargetObject
	targetMap := make(map[interface{}]reflect.Value)
	for _, res := range targetResults {
		pk := getPKValue(res, targetSchema)
		targetMap[pk] = res
	}

	// 4. Assign back
	for _, result := range results {
		val := reflect.ValueOf(result).Elem()
		pkField := val.FieldByName(qs.schema.GetField(qs.schema.PrimaryKey).StructFieldName)
		pk := pkField.Interface()

		if targetIDs, ok := relationMap[pk]; ok {
			relField := val.FieldByName(rel.Name)
			if relField.IsValid() && relField.CanSet() {
				// Expect slice
				if relField.Kind() == reflect.Slice {
					// Create slice
					sliceType := relField.Type()
					elemType := sliceType.Elem()
					isPtr := elemType.Kind() == reflect.Ptr

					newSlice := reflect.MakeSlice(sliceType, 0, len(targetIDs))

					for _, tID := range targetIDs {
						if target, ok := targetMap[tID]; ok {
							if isPtr {
								newSlice = reflect.Append(newSlice, target)
							} else {
								newSlice = reflect.Append(newSlice, target.Elem())
							}
						}
					}
					relField.Set(newSlice)
				}
			}
		}
	}

	return nil
}

// fetchByIDs fetches objects by IDs dynamically
func (qs *BaseQuerySet[T]) fetchByIDs(ctx context.Context, schema *ModelSchema, ids []interface{}) ([]reflect.Value, error) {
	db, err := qs.getDB(ctx)
	if err != nil {
		return nil, err
	}

	// Get dialect for placeholder generation
	d, err := qs.getDialect()
	if err != nil {
		return nil, err
	}

	// Build SQL
	// Build placeholders using dialect
	placeholders := d.BuildPlaceholders(len(ids))

	// We need to SELECT *
	// We can reuse BaseQuerySet logic if we could instantiate it, but we can't easily.
	// So we construct SQL manually.

	fields := make([]string, 0, len(schema.Fields))
	for _, f := range schema.Fields {
		fields = append(fields, EscapeIdentifier(f.DBColumn))
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s IN (%s)",
		strings.Join(fields, ", "),
		EscapeIdentifier(schema.TableName),
		EscapeIdentifier(schema.PrimaryKey),
		placeholders,
	)

	rows, err := db.QueryContext(ctx, query, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []reflect.Value
	columns, _ := rows.Columns()

	for rows.Next() {
		// Create instance
		instanceVal := reflect.New(schema.ModelType) // *Model
		if schema.ModelType.Kind() == reflect.Ptr {
			// ModelType is usually *User, so New gives **User
			// We want *User
			instanceVal = reflect.New(schema.ModelType.Elem())
		}

		// Scan
		scanArgs := make([]interface{}, len(columns))
		elem := instanceVal.Elem()

		for i, col := range columns {
			fieldInfo := schema.GetField(col)
			if fieldInfo != nil {
				f := elem.FieldByName(fieldInfo.StructFieldName)
				if f.IsValid() && f.CanSet() {
					scanArgs[i] = f.Addr().Interface()
				} else {
					var dump interface{}
					scanArgs[i] = &dump
				}
			} else {
				var dump interface{}
				scanArgs[i] = &dump
			}
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		results = append(results, instanceVal)
	}

	return results, nil
}

// buildPlaceholdersWithDialect generates placeholders using the provided dialect.
// This is the preferred way to build placeholders for database-agnostic SQL.
func buildPlaceholdersWithDialect(d dialect.Dialect, n int) string {
	return d.BuildPlaceholders(n)
}

// buildPlaceholders generates PostgreSQL-style placeholders ($1, $2, etc.).
//
// Deprecated: Use buildPlaceholdersWithDialect or dialect.BuildPlaceholders() instead.
// This function remains for backward compatibility but will be removed in v3.0.
func buildPlaceholders(n int) string {
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}

func isZero(v interface{}) bool {
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

func getPKValue(val reflect.Value, schema *ModelSchema) interface{} {
	// val is *Model
	elem := val.Elem()
	pkField := schema.GetField(schema.PrimaryKey)
	if pkField != nil {
		return elem.FieldByName(pkField.StructFieldName).Interface()
	}
	return nil
}
