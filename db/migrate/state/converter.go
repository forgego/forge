package state

import (
	"sort"

	generator "github.com/forgego/forge/codegen"
)

// sortedKeys returns a map's keys in order so converted definitions, and the
// down migrations rendered from them, are deterministic.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ToModelDefinitions converts state back to ModelDefinitions (for comparison)
func (s *SchemaState) ToModelDefinitions() []*generator.ModelDefinition {
	var defs []*generator.ModelDefinition
	for _, tableName := range sortedKeys(s.Tables) {
		tableState := s.Tables[tableName]
		if tableState.Name == "schema_migrations" {
			continue
		}
		def := &generator.ModelDefinition{
			Name:      tableState.Name,
			Fields:    []generator.FieldDefinition{},
			Relations: []generator.RelationDefinition{},
			Meta: generator.MetaDefinition{
				TableName:   tableState.Name,
				Indexes:     []generator.IndexDefinition{},
				Constraints: []generator.ConstraintDefinition{},
			},
		}

		// Convert columns to fields
		for _, colName := range sortedKeys(tableState.Columns) {
			colState := tableState.Columns[colName]
			field := generator.FieldDefinition{
				Name:          colState.Name,
				Type:          colState.Type,
				GoType:        colState.GoType,
				Required:      colState.Required,
				PrimaryKey:    colState.PrimaryKey,
				AutoIncrement: colState.AutoIncrement,
				Default:       colState.Default,
				Options:       colState.Options,
			}
			if field.Options == nil {
				field.Options = make(map[string]interface{})
			}
			if colState.Unique {
				field.Options["unique"] = true
			}
			def.Fields = append(def.Fields, field)
		}

		// Convert indexes
		for _, idxName := range sortedKeys(tableState.Indexes) {
			idxState := tableState.Indexes[idxName]
			def.Meta.Indexes = append(def.Meta.Indexes, generator.IndexDefinition{
				Name:   idxState.Name,
				Fields: idxState.Fields,
				Unique: idxState.Unique,
			})
		}

		// Convert foreign keys so regeneration does not re-add recorded ones.
		// To holds the target table and the actions hold their SQL form.
		for _, fkName := range sortedKeys(tableState.ForeignKeys) {
			fkState := tableState.ForeignKeys[fkName]
			def.Relations = append(def.Relations, generator.RelationDefinition{
				Name: fkState.Column,
				Type: "ForeignKey",
				To:   fkState.TargetTable,
				Options: map[string]interface{}{
					"on_delete": fkState.OnDelete,
					"on_update": fkState.OnUpdate,
				},
			})
		}

		// Convert constraints
		for _, constrName := range sortedKeys(tableState.Constraints) {
			constrState := tableState.Constraints[constrName]
			def.Meta.Constraints = append(def.Meta.Constraints, generator.ConstraintDefinition{
				Name:      constrState.Name,
				Type:      constrState.Type,
				Condition: constrState.Condition,
				Fields:    constrState.Fields,
			})
		}

		defs = append(defs, def)
	}
	return defs
}
