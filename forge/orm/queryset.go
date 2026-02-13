package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/forgego/forge/utils"
)

// QuerySet is the type-safe QuerySet interface
type QuerySet[T any] interface {
	// SetDB sets the database connection
	SetDB(db interface{}) QuerySet[T]
	// Filtering - accepts new Expression interface
	Filter(expr Expression) QuerySet[T]
	Exclude(expr Expression) QuerySet[T]

	// Ordering
	OrderBy(fields ...any) QuerySet[T]
	Reverse() QuerySet[T]

	// Limiting
	Limit(n int) QuerySet[T]
	Offset(n int) QuerySet[T]
	Distinct(fields ...any) QuerySet[T]

	// Field Selection
	Select(fields ...any) QuerySet[T]
	Only(fields ...any) QuerySet[T]
	Defer(fields ...any) QuerySet[T]

	// Relations
	SelectRelated(relations ...any) QuerySet[T]
	PrefetchRelated(relations ...any) QuerySet[T]

	// Aggregation
	Aggregate(aggs ...Aggregate) QuerySet[T]
	Annotate(anns ...AnnotationExpr) QuerySet[T]

	// Values - type-safe return types
	Values(fields ...any) ValuesQuerySet[T]
	ValuesList(fields ...any) ValuesListQuerySet[T]

	// Project - type-safe projection to a different type
	// Note: Use Project[T, P](qs) function instead of method

	// Execution
	All(ctx context.Context) ([]*T, error)
	Get(ctx context.Context) (*T, error)
	First(ctx context.Context) (*T, error)
	Last(ctx context.Context) (*T, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context) (bool, error)

	// Updates - type-safe
	Update(ctx context.Context, updates UpdateMap) (int64, error)
	UpdateBuilder() (*UpdateBuilder[T], error)
	BulkUpdate(ctx context.Context, updates []UpdateMap) error

	// Deletes
	Delete(ctx context.Context) (int64, error)

	// Set Operations
	Union(other QuerySet[T]) QuerySet[T]
	Intersection(other QuerySet[T]) QuerySet[T]
	Difference(other QuerySet[T]) QuerySet[T]
}

// OrderField represents an ordering field with direction
type OrderField struct {
	Field     string
	Ascending bool
}

// GetFieldPath returns the field path for ordering
func (o OrderField) GetFieldPath() string {
	return o.Field
}

// IsAscending returns whether the ordering is ascending
func (o OrderField) IsAscending() bool {
	return o.Ascending
}

// Asc creates an ascending order field
func Asc(field string) OrderField {
	return OrderField{
		Field:     field,
		Ascending: true,
	}
}

// Desc creates a descending order field
func Desc(field string) OrderField {
	return OrderField{
		Field:     field,
		Ascending: false,
	}
}

// NewOrderField creates an order field with explicit direction.
//
// Deprecated: Use Asc(field) or Desc(field) instead for clarity. NewOrderField will be removed in v3.0.
// Migration:
//
//	// Old
//	order := orm.NewOrderField("created_at", true)
//	order := orm.NewOrderField("created_at", false)
//	// New
//	order := orm.Asc("created_at")
//	order := orm.Desc("created_at")
func NewOrderField(field string, ascending bool) OrderField {
	if ascending {
		return Asc(field)
	}
	return Desc(field)
}

// BaseQuerySet is the implementation
type BaseQuerySet[T any] struct {
	table           string
	schema          *ModelSchema
	conditions      []Expression
	excludes        []Expression
	orderBy         []OrderField
	limitVal        *int
	offsetVal       *int
	distinctFields  []string
	selectFields    []string
	onlyFields      []string
	deferFields     []string
	selectRelated   []string
	prefetchRelated []string
	preloaded       map[string]bool // Track which relations are preloaded (N+1 prevention)
	joins           []string
	joinMap         map[string]bool
	aggregates      []Aggregate
	annotations     []AnnotationExpr // Using existing AnnotationExpr type
	db              interface{}      // *db.DB
	err             error            // Deferred error from Filter/Exclude validation (checked at execution time)
}

// NewQuerySet creates a new QuerySet
func NewQuerySet[T any](tableName string) (QuerySet[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = schema.TableName
	}

	return &BaseQuerySet[T]{
		table:      tableName,
		schema:     schema,
		conditions: []Expression{},
		excludes:   []Expression{},
		orderBy:    []OrderField{},
		preloaded:  make(map[string]bool),
		joins:      []string{},
		joinMap:    make(map[string]bool),
	}, nil
}

// SetDB sets the database connection
func (qs *BaseQuerySet[T]) SetDB(db interface{}) QuerySet[T] {
	qs.db = db
	return qs
}

// getDB retrieves the database connection
func (qs *BaseQuerySet[T]) getDB(ctx context.Context) (*sql.DB, error) {
	if qs.db != nil {
		return GetSQLDB(qs.db)
	}
	return nil, fmt.Errorf("database connection not set on QuerySet")
}

// clone creates a deep copy
func (qs *BaseQuerySet[T]) clone() *BaseQuerySet[T] {
	clone := &BaseQuerySet[T]{
		table:           qs.table,
		schema:          qs.schema,
		conditions:      append([]Expression{}, qs.conditions...),
		excludes:        append([]Expression{}, qs.excludes...),
		orderBy:         append([]OrderField{}, qs.orderBy...),
		limitVal:        qs.limitVal,
		offsetVal:       qs.offsetVal,
		distinctFields:  append([]string{}, qs.distinctFields...),
		selectFields:    append([]string{}, qs.selectFields...),
		onlyFields:      append([]string{}, qs.onlyFields...),
		deferFields:     append([]string{}, qs.deferFields...),
		selectRelated:   append([]string{}, qs.selectRelated...),
		prefetchRelated: append([]string{}, qs.prefetchRelated...),
		preloaded:       make(map[string]bool),
		joins:           append([]string{}, qs.joins...),
		joinMap:         make(map[string]bool),
		aggregates:      append([]Aggregate{}, qs.aggregates...),
		annotations:     append([]AnnotationExpr{}, qs.annotations...),
		db:              qs.db,
		err:             qs.err,
	}
	// Copy preloaded map
	for k, v := range qs.preloaded {
		clone.preloaded[k] = v
	}
	// Copy joinMap
	for k, v := range qs.joinMap {
		clone.joinMap[k] = v
	}
	return clone
}

// Filter adds a filter condition.
// Validation errors are deferred and returned when the QuerySet is executed
// (e.g. via All, Get, Count). This avoids panics while preserving the
// chainable API.
func (qs *BaseQuerySet[T]) Filter(expr Expression) QuerySet[T] {
	// Check for nil queryset or schema
	if qs == nil || qs.schema == nil {
		clone := qs.clone()
		clone.err = fmt.Errorf("cannot filter: queryset or schema is nil")
		return clone
	}

	// Validate expression -- store error instead of panicking
	if err := expr.Resolve(qs.schema); err != nil {
		clone := qs.clone()
		if clone.err == nil {
			clone.err = fmt.Errorf("invalid filter expression: %w", err)
		}
		return clone
	}

	clone := qs.clone()
	clone.conditions = append(clone.conditions, expr)
	return clone
}

// Exclude adds an exclude condition.
// Validation errors are deferred and returned when the QuerySet is executed.
func (qs *BaseQuerySet[T]) Exclude(expr Expression) QuerySet[T] {
	// Validate expression -- store error instead of panicking
	if err := expr.Resolve(qs.schema); err != nil {
		clone := qs.clone()
		if clone.err == nil {
			clone.err = fmt.Errorf("invalid exclude expression: %w", err)
		}
		return clone
	}

	clone := qs.clone()
	clone.excludes = append(clone.excludes, expr)
	return clone
}

// OrderBy sets ordering - accepts both OrderField (string) and OrderFieldExpr[T] (type-safe)
func (qs *BaseQuerySet[T]) OrderBy(fields ...any) QuerySet[T] {
	clone := qs.clone()
	for _, field := range fields {
		orderField := OrderField{
			Field:     extractOrderFieldPath(field),
			Ascending: extractOrderFieldAscending(field),
		}
		clone.orderBy = append(clone.orderBy, orderField)
	}
	return clone
}

// extractOrderFieldPath extracts the field path from an OrderFieldSpec
func extractOrderFieldPath(field any) string {
	// Use type assertion to OrderFieldSpec interface
	if spec, ok := field.(OrderFieldSpec); ok {
		return spec.GetFieldPath()
	}
	// Allow FieldPath (e.g. orm.Field[T]) directly
	if fp, ok := field.(FieldPath); ok {
		return fp.Path()
	}
	// Allow plain string field names, including "-field" prefix
	if s, ok := field.(string); ok {
		return strings.TrimPrefix(s, "-")
	}
	return ""
}

// extractOrderFieldAscending extracts the ascending flag from an OrderFieldSpec
func extractOrderFieldAscending(field any) bool {
	// Use type assertion to OrderFieldSpec interface
	if spec, ok := field.(OrderFieldSpec); ok {
		return spec.IsAscending()
	}
	// FieldPath defaults to ascending
	if _, ok := field.(FieldPath); ok {
		return true
	}
	// String supports "-" prefix for descending
	if s, ok := field.(string); ok {
		return !strings.HasPrefix(s, "-")
	}
	return true
}

// Reverse reverses the current ordering
func (qs *BaseQuerySet[T]) Reverse() QuerySet[T] {
	clone := qs.clone()
	for i := range clone.orderBy {
		clone.orderBy[i].Ascending = !clone.orderBy[i].Ascending
	}
	return clone
}

// Limit sets the limit
func (qs *BaseQuerySet[T]) Limit(n int) QuerySet[T] {
	clone := qs.clone()
	clone.limitVal = &n
	return clone
}

// Offset sets the offset
func (qs *BaseQuerySet[T]) Offset(n int) QuerySet[T] {
	clone := qs.clone()
	clone.offsetVal = &n
	return clone
}

// Distinct sets distinct fields - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Distinct(fields ...any) QuerySet[T] {
	clone := qs.clone()
	if len(fields) > 0 {
		paths := make([]string, len(fields))
		for i, field := range fields {
			paths[i] = ExtractPathFromAny(field)
		}
		clone.distinctFields = paths
	} else {
		clone.distinctFields = []string{"*"}
	}
	return clone
}

// Select sets fields to select - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Select(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return clone
}

// Only sets fields to only load - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Only(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.onlyFields = paths
	return clone
}

// Defer sets fields to defer loading - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Defer(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.deferFields = paths
	return clone
}

// SelectRelated adds relations to select - accepts both string and RelationExpression
func (qs *BaseQuerySet[T]) SelectRelated(relations ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(relations))
	for i, relation := range relations {
		paths[i] = extractRelationPathFromAny(relation)
	}
	clone.selectRelated = append(clone.selectRelated, paths...)
	return clone
}

// PrefetchRelated adds relations to prefetch - accepts both string and RelationExpression
// Marks relations as preloaded to prevent N+1 queries
func (qs *BaseQuerySet[T]) PrefetchRelated(relations ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(relations))
	for i, relation := range relations {
		path := extractRelationPathFromAny(relation)
		paths[i] = path
		// Mark as preloaded
		clone.preloaded[path] = true
	}
	clone.prefetchRelated = append(clone.prefetchRelated, paths...)
	return clone
}

// Aggregate adds aggregate expressions
func (qs *BaseQuerySet[T]) Aggregate(aggs ...Aggregate) QuerySet[T] {
	clone := qs.clone()
	clone.aggregates = append(clone.aggregates, aggs...)
	return clone
}

// Annotate adds annotation expressions
func (qs *BaseQuerySet[T]) Annotate(anns ...AnnotationExpr) QuerySet[T] {
	clone := qs.clone()
	clone.annotations = append(clone.annotations, anns...)
	return clone
}

// Values returns a ValuesQuerySet - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Values(fields ...any) ValuesQuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return &BaseValuesQuerySet[T]{base: clone}
}

// ValuesList returns a ValuesListQuerySet - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) ValuesList(fields ...any) ValuesListQuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return &BaseValuesListQuerySet[T]{base: clone}
}

// Project is implemented via the Project function in projection.go
// Usage: Project[User, UserProjection](qs)

// UpdateBuilder returns an UpdateBuilder
func (qs *BaseQuerySet[T]) UpdateBuilder() (*UpdateBuilder[T], error) {
	// Create a wrapper that implements the interface needed by UpdateBuilder
	return NewUpdateBuilderFromQuerySet(qs)
}

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

// buildSQL builds the SQL query
func (qs *BaseQuerySet[T]) buildSQL() (string, []interface{}, error) {
	builder := NewSQLBuilder()

	// Build JOINs first (populates qs.joins and qs.joinMap)
	qs.buildJoinClause(builder)

	// Build SELECT clause
	selectClause := qs.buildSelectClause(builder)

	// Build FROM clause
	fromClause := fmt.Sprintf("FROM %s", EscapeIdentifier(qs.table))

	// Build WHERE clause
	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		return "", nil, whereErr
	}

	// Build ORDER BY clause
	orderByClause := qs.buildOrderByClause(builder)

	// Build LIMIT/OFFSET
	limitClause := qs.buildLimitClause()
	offsetClause := qs.buildOffsetClause()

	// Combine all parts
	parts := []string{selectClause, fromClause}
	if len(qs.joins) > 0 {
		parts = append(parts, strings.Join(qs.joins, " "))
	}
	if whereClause != "" {
		parts = append(parts, whereClause)
	}
	if orderByClause != "" {
		parts = append(parts, orderByClause)
	}
	if limitClause != "" {
		parts = append(parts, limitClause)
	}
	if offsetClause != "" {
		parts = append(parts, offsetClause)
	}

	sql := strings.Join(parts, " ")
	args := builder.Args()

	return sql, args, nil
}

// buildJoinClause builds the JOIN clause
func (qs *BaseQuerySet[T]) buildJoinClause(builder *SQLBuilder) {
	qs.joins = []string{}
	qs.joinMap = make(map[string]bool)

	for _, path := range qs.selectRelated {
		if qs.joinMap[path] {
			continue
		}

		// Resolve path to relation
		// For MVP, support single level: "User"
		rel := qs.schema.GetRelation(path)
		if rel == nil {
			continue
		}

		// Get target schema
		targetSchema, err := GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			continue
		}

		joinTable := EscapeIdentifier(targetSchema.TableName)
		alias := EscapeIdentifier(rel.Name)
		mainTable := EscapeIdentifier(qs.table)

		// Try to find the FK column
		// Strategy:
		// 1. Look for field with DBColumn = rel.Name + "_id" (lowercase)
		// 2. Look for field with Name = rel.Name + "ID"
		fkColumn := ""

		// Try to find field that corresponds to this relation
		for _, f := range qs.schema.Fields {
			// Check common naming conventions
			if f.StructFieldName == rel.Name+"ID" {
				fkColumn = f.DBColumn
				break
			}
			if strings.EqualFold(f.DBColumn, rel.Name+"_id") {
				fkColumn = f.DBColumn
				break
			}
			// Fallback: if field name ends in ID and matches relation prefix
			if strings.HasSuffix(f.Name, "ID") && strings.HasPrefix(f.Name, rel.Name) {
				fkColumn = f.DBColumn
				break
			}
		}

		if fkColumn == "" {
			// Try "user_id" for "User" relation
			guess := strings.ToLower(rel.Name) + "_id"
			if f := qs.schema.GetField(guess); f != nil {
				fkColumn = f.DBColumn
			}
		}

		if fkColumn == "" {
			continue // Could not determine join condition
		}

		// JOIN target_table "Alias" ON main_table.fk = "Alias".pk
		joinSQL := fmt.Sprintf("LEFT OUTER JOIN %s %s ON %s.%s = %s.%s",
			joinTable, alias,
			mainTable, EscapeIdentifier(fkColumn),
			alias, EscapeIdentifier(targetSchema.PrimaryKey))

		qs.joins = append(qs.joins, joinSQL)
		qs.joinMap[path] = true
	}
}

// buildSelectClause builds the SELECT clause
func (qs *BaseQuerySet[T]) buildSelectClause(builder *SQLBuilder) string {
	var fields []string

	if len(qs.selectFields) > 0 {
		for _, field := range qs.selectFields {
			fields = append(fields, EscapeIdentifier(field))
		}
	} else if len(qs.onlyFields) > 0 {
		for _, field := range qs.onlyFields {
			fields = append(fields, EscapeIdentifier(field))
		}
	} else {
		fields = []string{"*"}
	}

	// Add annotations to SELECT
	if len(qs.annotations) > 0 {
		for _, ann := range qs.annotations {
			// Build annotation SQL using QueryExpr.ToSQL
			// QueryExpr uses paramIndex, so we need to use builder's paramIndex
			annSQL, annArgs, nextIndex := ann.Expr.ToSQL(builder.paramIndex)
			// Update builder's paramIndex
			builder.paramIndex = nextIndex
			// Add args to builder
			for _, arg := range annArgs {
				builder.args = append(builder.args, arg)
			}
			alias := EscapeIdentifier(ann.Name)
			fields = append(fields, fmt.Sprintf("%s AS %s", annSQL, alias))
		}
	}

	selectClause := "SELECT "
	if len(qs.distinctFields) > 0 {
		selectClause += "DISTINCT "
	}
	selectClause += strings.Join(fields, ", ")

	// Add related fields
	for _, path := range qs.selectRelated {
		if !qs.joinMap[path] {
			continue
		}
		rel := qs.schema.GetRelation(path)
		if rel == nil {
			continue
		}
		targetSchema, err := GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			continue
		}

		// Select all fields from target schema
		for _, f := range targetSchema.Fields {
			alias := EscapeIdentifier(rel.Name)
			col := EscapeIdentifier(f.DBColumn)
			colAlias := EscapeIdentifier(rel.Name + "__" + f.DBColumn)

			selectClause += fmt.Sprintf(", %s.%s AS %s", alias, col, colAlias)
		}
	}

	return selectClause
}

// buildWhereClause builds the WHERE clause.
// Returns the SQL string, arguments, and any error encountered during SQL generation.
func (qs *BaseQuerySet[T]) buildWhereClause(builder *SQLBuilder) (string, []interface{}, error) {
	var parts []string
	var allArgs []interface{}

	// Build conditions
	for _, cond := range qs.conditions {
		sql, args, err := cond.ToSQL(builder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build condition SQL: %w", err)
		}
		parts = append(parts, sql)
		allArgs = append(allArgs, args...)
	}

	// Build excludes (with NOT)
	for _, exclude := range qs.excludes {
		sql, args, err := exclude.ToSQL(builder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build exclude SQL: %w", err)
		}
		parts = append(parts, fmt.Sprintf("NOT (%s)", sql))
		allArgs = append(allArgs, args...)
	}

	if len(parts) == 0 {
		return "", nil, nil
	}

	return "WHERE " + strings.Join(parts, " AND "), allArgs, nil
}

// buildOrderByClause builds the ORDER BY clause
func (qs *BaseQuerySet[T]) buildOrderByClause(builder *SQLBuilder) string {
	if len(qs.orderBy) == 0 {
		return ""
	}

	var parts []string
	for _, field := range qs.orderBy {
		escaped := EscapeIdentifier(field.Field)
		if field.Ascending {
			parts = append(parts, escaped+" ASC")
		} else {
			parts = append(parts, escaped+" DESC")
		}
	}

	return "ORDER BY " + strings.Join(parts, ", ")
}

// buildLimitClause builds the LIMIT clause
func (qs *BaseQuerySet[T]) buildLimitClause() string {
	if qs.limitVal == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d", *qs.limitVal)
}

// buildOffsetClause builds the OFFSET clause
func (qs *BaseQuerySet[T]) buildOffsetClause() string {
	if qs.offsetVal == nil {
		return ""
	}
	return fmt.Sprintf("OFFSET %d", *qs.offsetVal)
}

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

			if !qs.joinMap[relName] {
				// Might be a manual alias or something else, skip relation logic
				continue
			}

			rel := qs.schema.GetRelation(relName)
			if rel == nil {
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
					relationName: relName,
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

			// Allocate relation struct
			relField := instanceValue.FieldByName(relName)
			// Handle pointer to struct
			if relField.IsValid() && relField.CanSet() && relField.Kind() == reflect.Ptr {
				val := reflect.New(relField.Type().Elem())
				relField.Set(val)

				nestedVal := val.Elem()

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
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d!", qs.table, len(results))
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
	results, err := qs.Limit(1).All(ctx)
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
	reversed := qs.Reverse()
	results, err := reversed.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s matching query does not exist", qs.table)
	}

	return results[0], nil
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

	builder := NewSQLBuilder()
	selectClause := fmt.Sprintf("SELECT COUNT(*) FROM %s", EscapeIdentifier(qs.table))
	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		return 0, whereErr
	}

	sql := selectClause
	if whereClause != "" {
		sql += " " + whereClause
	}

	var count int64
	err = db.QueryRowContext(ctx, sql, builder.Args()...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	return count, nil
}

// Exists checks if any records exist
func (qs *BaseQuerySet[T]) Exists(ctx context.Context) (bool, error) {
	count, err := qs.Count(ctx)
	return count > 0, err
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

	builder := NewSQLBuilder()

	// Build SET clause
	var setParts []string

	for fieldName, value := range updates {
		escapedField := EscapeIdentifier(fieldName)

		// Check if value is an Expression
		if expr, ok := value.(Expression); ok {
			// Build expression SQL
			exprSQL, _, err := expr.ToSQL(builder)
			if err != nil {
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
		return 0, whereErr
	}

	// Combine all args
	allArgs := builder.Args()

	// Build SQL
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

// BulkUpdate performs bulk updates
// Note: This is a planned feature. For now, use Update() in a loop or implement
// custom bulk update logic using raw SQL if needed.
func (qs *BaseQuerySet[T]) BulkUpdate(ctx context.Context, updates []UpdateMap) error {
	return fmt.Errorf("BulkUpdate not yet implemented in QuerySet - use Update() in a loop or raw SQL for now")
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

	builder := NewSQLBuilder()

	// Build WHERE clause
	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		return 0, whereErr
	}

	// Build SQL
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

// Union performs a UNION operation
func (qs *BaseQuerySet[T]) Union(other QuerySet[T]) QuerySet[T] {
	// For now, return a combined query set that will use SQL UNION
	// Full implementation would require SQL query combination
	clone := qs.clone()
	// Mark as union - would need additional state to track this
	// For MVP, return clone with union marker
	return clone
}

// Intersection performs an INTERSECT operation
func (qs *BaseQuerySet[T]) Intersection(other QuerySet[T]) QuerySet[T] {
	// Similar to Union - would need SQL query combination
	clone := qs.clone()
	return clone
}

// Difference performs an EXCEPT operation
func (qs *BaseQuerySet[T]) Difference(other QuerySet[T]) QuerySet[T] {
	// Similar to Union - would need SQL query combination
	clone := qs.clone()
	return clone
}

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
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d!", vqs.base.table, len(results))
	}

	return results[0], nil
}

func (vqs *BaseValuesQuerySet[T]) First(ctx context.Context) (map[string]interface{}, error) {
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
		return nil, fmt.Errorf("get() returned more than one %s -- it returned %d!", vls.base.table, len(results))
	}

	return results[0], nil
}

func (vls *BaseValuesListQuerySet[T]) First(ctx context.Context) ([]interface{}, error) {
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
