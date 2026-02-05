package state

import (
	generator "github.com/forgego/forge/codegen"
)

// ToModelDefinitions converts state back to ModelDefinitions (for comparison)
func (s *SchemaState) ToModelDefinitions() []*generator.ModelDefinition {
	var defs []*generator.ModelDefinition
	for _, tableState := range s.Tables {
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
		for _, colState := range tableState.Columns {
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
		for _, idxState := range tableState.Indexes {
			def.Meta.Indexes = append(def.Meta.Indexes, generator.IndexDefinition{
				Name:   idxState.Name,
				Fields: idxState.Fields,
				Unique: idxState.Unique,
			})
		}

		// Convert constraints
		for _, constrState := range tableState.Constraints {
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
