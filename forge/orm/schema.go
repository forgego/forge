package orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/forgego/forge/schema"
)

// ModelSchema contains metadata about a model
type ModelSchema struct {
	TableName  string
	Fields     []FieldInfo
	Relations  []RelationInfo
	Indexes    []IndexInfo
	PrimaryKey string
	ModelType  reflect.Type
}

// FieldInfo contains field metadata
type FieldInfo struct {
	Name       string
	DBColumn   string
	Type       reflect.Type
	Required   bool
	PrimaryKey bool
	Unique     bool
	ForeignKey *RelationInfo
	StructFieldName string
}

// RelationInfo contains relation metadata
type RelationInfo struct {
	Name        string
	Type        RelationType
	TargetModel string
	TargetField string
	OnDelete    string
	OnUpdate    string
	RelatedName string
	FieldName   string // The field name on this model
	Through     string // Through table for ManyToMany
}

// RelationType represents the type of relation
type RelationType string

const (
	RelationForeignKey RelationType = "ForeignKey"
	RelationOneToOne   RelationType = "OneToOne"
	RelationManyToMany RelationType = "ManyToMany"
)

// IndexInfo contains index metadata
type IndexInfo struct {
	Name    string
	Fields  []string
	Unique  bool
	Partial string // Partial index condition
}

// GetField retrieves a field by name
func (ms *ModelSchema) GetField(name string) *FieldInfo {
	// First try exact match (case-sensitive)
	for i := range ms.Fields {
		if ms.Fields[i].Name == name || ms.Fields[i].DBColumn == name {
			return &ms.Fields[i]
		}
	}

	// If exact match fails, try case-insensitive match for convenience
	// as sometimes the field might be queried by PascalCase but stored as camelCase or snake_case
	lowerName := strings.ToLower(name)
	for i := range ms.Fields {
		if strings.ToLower(ms.Fields[i].Name) == lowerName || strings.ToLower(ms.Fields[i].DBColumn) == lowerName {
			return &ms.Fields[i]
		}
	}

	return nil
}

// GetRelation retrieves a relation by name
func (ms *ModelSchema) GetRelation(name string) *RelationInfo {
	for i := range ms.Relations {
		if ms.Relations[i].Name == name {
			return &ms.Relations[i]
		}
	}
	return nil
}

// BuildModelSchema builds a ModelSchema from a schema.Schema instance
func BuildModelSchema(schemaInstance schema.Schema) (*ModelSchema, error) {
	ms := &ModelSchema{
		Fields:    []FieldInfo{},
		Relations: []RelationInfo{},
		Indexes:   []IndexInfo{},
	}

	// Get model type
	instanceValue := reflect.ValueOf(schemaInstance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}
	ms.ModelType = instanceValue.Type()

	// Get table name from meta if available
	meta := schemaInstance.Meta()
	if meta.TableName != "" {
		ms.TableName = meta.TableName
	}

	// Build map of DBColumn -> StructFieldName
	structFieldMap := make(map[string]string)
	if ms.ModelType.Kind() == reflect.Struct {
		for i := 0; i < ms.ModelType.NumField(); i++ {
			field := ms.ModelType.Field(i)
			// Skip unexported
			if field.PkgPath != "" {
				continue
			}

			// Check db tag
			dbTag := field.Tag.Get("db")
			if dbTag != "" {
				structFieldMap[dbTag] = field.Name
			}

			// Also map lower case name
			structFieldMap[strings.ToLower(field.Name)] = field.Name
		}
	}

	// Build fields from schema
	fields := schemaInstance.Fields()
	for _, field := range fields {
		fieldInfo := FieldInfo{
			Name:       field.Name,
			DBColumn:   field.DBColumn,
			Type:       getGoType(field.Type),
			Required:   field.Required,
			PrimaryKey: field.PrimaryKey,
			Unique:     field.Unique,
		}

		if field.DBColumn == "" {
			fieldInfo.DBColumn = field.Name
		}

		// Resolve StructFieldName
		if name, ok := structFieldMap[fieldInfo.DBColumn]; ok {
			fieldInfo.StructFieldName = name
		} else if name, ok := structFieldMap[strings.ToLower(field.Name)]; ok {
			fieldInfo.StructFieldName = name
		}

		if field.PrimaryKey {
			ms.PrimaryKey = fieldInfo.DBColumn
		}

		ms.Fields = append(ms.Fields, fieldInfo)
	}

	// Build relations from schema
	relations := schemaInstance.Relations()
	for _, rel := range relations {
		relInfo := RelationInfo{
			Name:        rel.Name,
			TargetModel: rel.To,
			FieldName:   rel.Name, // Use relation name as field name
			OnDelete:    string(rel.OnDelete),
			OnUpdate:    string(rel.OnUpdate),
			RelatedName: rel.RelatedName,
			Through:     rel.Through,
		}

		// Determine relation type
		switch rel.Type {
		case schema.RelationForeignKey:
			relInfo.Type = RelationForeignKey
		case schema.RelationOneToOne:
			relInfo.Type = RelationOneToOne
		case schema.RelationManyToMany:
			relInfo.Type = RelationManyToMany
		}

		ms.Relations = append(ms.Relations, relInfo)
	}

	// Build indexes from meta
	meta = schemaInstance.Meta()
	for _, idx := range meta.Indexes {
		indexInfo := IndexInfo{
			Name:   idx.Name,
			Fields: idx.Fields,
			Unique: idx.Unique,
		}
		ms.Indexes = append(ms.Indexes, indexInfo)
	}

	return ms, nil
}

// getGoType converts schema.FieldType to reflect.Type
func getGoType(ft schema.FieldType) reflect.Type {
	switch ft {
	case schema.TypeInt64:
		return reflect.TypeOf(int64(0))
	case schema.TypeInt32:
		return reflect.TypeOf(int32(0))
	case schema.TypeString:
		return reflect.TypeOf("")
	case schema.TypeBool:
		return reflect.TypeOf(false)
	case schema.TypeTime, schema.TypeDate, schema.TypeDateTime:
		return reflect.TypeOf((*interface{})(nil)).Elem() // time.Time - would need import
	case schema.TypeFloat32:
		return reflect.TypeOf(float32(0))
	case schema.TypeFloat64:
		return reflect.TypeOf(float64(0))
	case schema.TypeDecimal:
		return reflect.TypeOf((*interface{})(nil)).Elem() // Would be decimal.Decimal
	default:
		return reflect.TypeOf((*interface{})(nil)).Elem()
	}
}

// Schema cache for performance
var (
	schemaCache = make(map[reflect.Type]*ModelSchema)
	schemaMu    sync.RWMutex
)

// GetModelSchema gets or builds a ModelSchema for a model type
func GetModelSchema[T any]() (*ModelSchema, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Check cache
	schemaMu.RLock()
	if schema, ok := schemaCache[typ]; ok {
		schemaMu.RUnlock()
		return schema, nil
	}
	schemaMu.RUnlock()

	// Build schema
	instanceValue := reflect.New(typ).Elem()
	schemaInstance, ok := instanceValue.Interface().(schema.Schema)
	if !ok {
		return nil, fmt.Errorf("type %v does not implement schema.Schema", typ)
	}

	schema, err := BuildModelSchema(schemaInstance)
	if err != nil {
		return nil, err
	}

	// Cache it
	schemaMu.Lock()
	schemaCache[typ] = schema
	schemaMu.Unlock()

	// Register type name for relation resolution
	RegisterModelType(typ.Name(), typ)

	return schema, nil
}

// FieldAccessor provides type-safe field access for a model
type FieldAccessor[T any] struct {
	schema *ModelSchema
	table  string
}

// NewFieldAccessor creates a field accessor for a model type
func NewFieldAccessor[T any]() (*FieldAccessor[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	return &FieldAccessor[T]{
		schema: schema,
		table:  schema.TableName,
	}, nil
}

// Field creates a field expression (use FieldFor for type safety)
// Returns an error if the field is not found
func (fa *FieldAccessor[T]) Field(name string) (FieldExpression[interface{}], error) {
	// Validate field exists
	fieldInfo := fa.schema.GetField(name)
	if fieldInfo == nil {
		return FieldExpression[interface{}]{}, fmt.Errorf("field %s not found on model", name)
	}

	return NewField[interface{}](name, fa.table), nil
}

// AllFieldExpressions returns all field expressions as a map
func (fa *FieldAccessor[T]) AllFieldExpressions() map[string]FieldExpression[interface{}] {
	result := make(map[string]FieldExpression[interface{}])
	for _, field := range fa.schema.Fields {
		result[field.Name] = NewField[interface{}](field.Name, fa.table)
	}
	return result
}

// RelatedField creates a type-safe related field expression
// Use RelatedFieldFor to specify related model and field types
// Returns an error if the relation is not found
func (fa *FieldAccessor[T]) RelatedField(relationName, fieldName string) (FieldExpression[interface{}], error) {
	// Validate relation exists
	rel := fa.schema.GetRelation(relationName)
	if rel == nil {
		return FieldExpression[interface{}]{}, fmt.Errorf("relation %s not found", relationName)
	}

	// Build field path
	fieldPath := relationName + "__" + fieldName
	return NewField[interface{}](fieldPath, fa.table), nil
}

// AllFields returns all field names
func (fa *FieldAccessor[T]) AllFields() []string {
	fields := make([]string, len(fa.schema.Fields))
	for i, f := range fa.schema.Fields {
		fields[i] = f.Name
	}
	return fields
}

// AllRelations returns all relation names
func (fa *FieldAccessor[T]) AllRelations() []string {
	relations := make([]string, len(fa.schema.Relations))
	for i, r := range fa.schema.Relations {
		relations[i] = r.Name
	}
	return relations
}

// ValidatePath validates that a field path exists in the schema
func (ms *ModelSchema) ValidatePath(path string) error {
	_, _, err := ms.ResolvePath(path)
	return err
}

// ResolvePath resolves a field path and returns field info and target schema
// Supports relation traversal: "author__email"
func (ms *ModelSchema) ResolvePath(path string) (*FieldInfo, *ModelSchema, error) {
	parts := splitFieldPathSchema(path)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("empty field path")
	}

	// Check first part exists as field or relation
	field := ms.GetField(parts[0])
	if field != nil {
		// Simple field path
		if len(parts) == 1 {
			return field, ms, nil
		}
		return nil, nil, fmt.Errorf("field %s cannot be traversed further", parts[0])
	}

	// Check if it's a relation
	relation := ms.GetRelation(parts[0])
	if relation == nil {
		return nil, nil, fmt.Errorf("field or relation %s not found", parts[0])
	}

	// If path ends at a relation (no further parts), it's an error
	if len(parts) == 1 {
		return nil, nil, fmt.Errorf("path %s resolves to a relation, not a field", path)
	}

	// Traverse the relation to resolve the remaining path.
	// Look up the target model schema and recurse.
	targetSchema, err := GetModelSchemaByName(relation.TargetModel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve target model %s for relation %s: %w",
			relation.TargetModel, parts[0], err)
	}

	remainingPath := strings.Join(parts[1:], "__")
	return targetSchema.ResolvePath(remainingPath)
}

// GetPathType returns the type information for a field path
func (ms *ModelSchema) GetPathType(path string) (reflect.Type, error) {
	fieldInfo, _, err := ms.ResolvePath(path)
	if err != nil {
		return nil, err
	}
	return fieldInfo.Type, nil
}

// GetAllowedLookups returns allowed lookups for a field (for security whitelisting)
func (ms *ModelSchema) GetAllowedLookups(fieldPath string) ([]string, error) {
	field, _, err := ms.ResolvePath(fieldPath)
	if err != nil {
		return nil, err
	}

	// Return default lookups based on field type
	lookups := getDefaultLookupsForType(field.Type)
	return lookups, nil
}

// getDefaultLookupsForType returns default lookups for a field type
func getDefaultLookupsForType(fieldType reflect.Type) []string {
	// This is a simplified version
	// In practice, this would check the actual reflect.Type
	switch fieldType.Kind() {
	case reflect.String:
		return []string{"exact", "in", "contains", "icontains", "startswith", "endswith"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return []string{"exact", "in", "gt", "gte", "lt", "lte"}
	case reflect.Bool:
		return []string{"exact"}
	default:
		return []string{"exact", "in"}
	}
}

// GetRelationDepth calculates the depth of a relation path
func (ms *ModelSchema) GetRelationDepth(path string) (int, error) {
	parts := splitFieldPathSchema(path)
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid path")
	}

	// Count relations (all parts except the last)
	depth := len(parts) - 1
	return depth, nil
}

// splitFieldPathSchema splits a field path by "__" (schema package version)
func splitFieldPathSchema(path string) []string {
	return strings.Split(path, "__")
}

// Helper to convert field path to SQL (handles double underscore)
func fieldPathToSQL(path string) string {
	// Replace double underscore with single for SQL
	// "author__name" -> "author.name" or JOIN syntax
	parts := strings.Split(path, "__")
	if len(parts) == 1 {
		return EscapeIdentifier(parts[0])
	}
	// For now, just escape - JOIN logic handled elsewhere
	return EscapeIdentifier(strings.Join(parts, "."))
}



