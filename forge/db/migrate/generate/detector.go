package generate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// ChangeDetector detects changes between current and previous model states
type ChangeDetector interface {
	// DetectChanges compares current models to previous state and returns all changes
	DetectChanges(current, previous []*generator.ModelDefinition) ([]core.Change, error)
}

// Detector is the default implementation of ChangeDetector
type Detector struct{}

// NewDetector creates a new change detector
func NewDetector() ChangeDetector {
	return &Detector{}
}

// DetectChanges compares current models to previous state and returns all changes
func (d *Detector) DetectChanges(current, previous []*generator.ModelDefinition) ([]core.Change, error) {
	var changes []core.Change

	// Build maps for easier lookup
	currentMap := make(map[string]*generator.ModelDefinition)
	previousMap := make(map[string]*generator.ModelDefinition)

	for _, def := range current {
		tableName := getTableName(def)
		currentMap[tableName] = def
	}

	for _, def := range previous {
		tableName := getTableName(def)
		previousMap[tableName] = def
	}

	// Detect table-level changes
	for tableName, currentDef := range currentMap {
		previousDef, exists := previousMap[tableName]
		if !exists {
			// New table - create table and add all indexes/foreign keys
			changes = append(changes, &core.CreateTable{Table: currentDef})

			// Add indexes for new table
			for _, idx := range currentDef.Meta.Indexes {
				changes = append(changes, &core.AddIndex{
					Table: tableName,
					Index: idx,
				})
			}

			// Add foreign keys for new table
			for _, rel := range currentDef.Relations {
				if rel.Type == "ForeignKey" || rel.Type == "OneToOne" {
					// Check if the relation column exists in fields
					hasColumn := false
					for _, field := range currentDef.Fields {
						if field.Name == rel.Name {
							hasColumn = true
							break
						}
					}
					if !hasColumn {
						continue
					}

					targetTable := findTargetTable(rel.To, current)
					if targetTable != "" {
						changes = append(changes, &core.AddForeignKey{
							Table:       tableName,
							Relation:    rel,
							TargetTable: targetTable,
						})
					}
				}
			}
		} else if previousDef != nil {
			// Existing table - detect column, index, constraint changes
			tableChanges, err := d.detectTableChanges(currentDef, previousDef, tableName, current)
			if err != nil {
				return nil, err
			}
			changes = append(changes, tableChanges...)
		}
	}

	// Detect dropped tables
	for tableName := range previousMap {
		if _, exists := currentMap[tableName]; !exists {
			changes = append(changes, &core.DropTable{Table: tableName})
		}
	}

	return changes, nil
}

// detectTableChanges detects changes within a single table
func (d *Detector) detectTableChanges(current, previous *generator.ModelDefinition, tableName string, allDefs []*generator.ModelDefinition) ([]core.Change, error) {
	var changes []core.Change

	// Detect column changes
	columnChanges, err := d.detectColumnChanges(tableName, current.Fields, previous.Fields)
	if err != nil {
		return nil, err
	}
	changes = append(changes, columnChanges...)

	// Detect index changes
	indexChanges, err := d.detectIndexChanges(tableName, current.Meta.Indexes, previous.Meta.Indexes)
	if err != nil {
		return nil, err
	}
	changes = append(changes, indexChanges...)

	// Detect constraint changes
	constraintChanges, err := d.detectConstraintChanges(tableName, current.Meta.Constraints, previous.Meta.Constraints)
	if err != nil {
		return nil, err
	}
	changes = append(changes, constraintChanges...)

	// Detect foreign key changes
	fkChanges, err := d.detectForeignKeyChanges(tableName, current.Relations, previous.Relations, current, previous, allDefs)
	if err != nil {
		return nil, err
	}
	changes = append(changes, fkChanges...)

	return changes, nil
}

// detectColumnChanges detects column-level changes
func (d *Detector) detectColumnChanges(tableName string, current, previous []generator.FieldDefinition) ([]core.Change, error) {
	var changes []core.Change

	currentMap := make(map[string]generator.FieldDefinition)
	previousMap := make(map[string]generator.FieldDefinition)

	for _, field := range current {
		currentMap[field.Name] = field
	}

	for _, field := range previous {
		previousMap[field.Name] = field
	}

	// Detect new columns
	for name, field := range currentMap {
		if _, exists := previousMap[name]; !exists {
			changes = append(changes, &core.AddColumn{
				Table:  tableName,
				Column: field,
			})
		}
	}

	// Detect dropped columns
	for name := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, &core.DropColumn{
				Table:      tableName,
				ColumnName: name,
			})
		}
	}

	// Detect modified columns
	for name, currentField := range currentMap {
		if previousField, exists := previousMap[name]; exists {
			if d.fieldChanged(currentField, previousField) {
				changes = append(changes, &core.ModifyColumn{
					Table:     tableName,
					OldColumn: previousField,
					NewColumn: currentField,
				})
			}
		}
	}

	return changes, nil
}

// fieldChanged checks if a field has changed
func (d *Detector) fieldChanged(current, previous generator.FieldDefinition) bool {
	if current.Type != previous.Type {
		return true
	}
	if current.Required != previous.Required {
		return true
	}
	if current.PrimaryKey != previous.PrimaryKey {
		return true
	}
	if current.AutoIncrement != previous.AutoIncrement {
		return true
	}
	// Use reflect.DeepEqual for default value comparison
	if !reflect.DeepEqual(current.Default, previous.Default) {
		return true
	}
	// Use reflect.DeepEqual for options map comparison
	if !reflect.DeepEqual(current.Options, previous.Options) {
		return true
	}
	return false
}

// detectIndexChanges detects index changes
func (d *Detector) detectIndexChanges(tableName string, current, previous []generator.IndexDefinition) ([]core.Change, error) {
	var changes []core.Change

	currentMap := make(map[string]generator.IndexDefinition)
	previousMap := make(map[string]generator.IndexDefinition)

	for _, idx := range current {
		idxName := idx.Name
		if idxName == "" {
			idxName = fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Fields, "_"))
		}
		currentMap[idxName] = idx
	}

	for _, idx := range previous {
		idxName := idx.Name
		if idxName == "" {
			idxName = fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Fields, "_"))
		}
		previousMap[idxName] = idx
	}

	// Detect new indexes
	for name, idx := range currentMap {
		if _, exists := previousMap[name]; !exists {
			changes = append(changes, &core.AddIndex{
				Table: tableName,
				Index: idx,
			})
		} else {
			// Check if index changed
			prevIdx := previousMap[name]
			if d.indexChanged(idx, prevIdx) {
				changes = append(changes, &core.ModifyIndex{
					Table:    tableName,
					OldIndex: prevIdx,
					NewIndex: idx,
				})
			}
		}
	}

	// Detect dropped indexes
	for name := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, &core.DropIndex{
				Table:     tableName,
				IndexName: name,
			})
		}
	}

	return changes, nil
}

// indexChanged checks if an index has changed
func (d *Detector) indexChanged(current, previous generator.IndexDefinition) bool {
	if current.Unique != previous.Unique {
		return true
	}
	if len(current.Fields) != len(previous.Fields) {
		return true
	}
	for i, field := range current.Fields {
		if field != previous.Fields[i] {
			return true
		}
	}
	return false
}

// detectForeignKeyChanges detects foreign key changes
func (d *Detector) detectForeignKeyChanges(tableName string, current, previous []generator.RelationDefinition, currentDef, previousDef *generator.ModelDefinition, allDefs []*generator.ModelDefinition) ([]core.Change, error) {
	var changes []core.Change

	currentMap := make(map[string]generator.RelationDefinition)
	previousMap := make(map[string]generator.RelationDefinition)

	for _, rel := range current {
		if rel.Type == "ForeignKey" || rel.Type == "OneToOne" {
			currentMap[rel.Name] = rel
		}
	}

	for _, rel := range previous {
		if rel.Type == "ForeignKey" || rel.Type == "OneToOne" {
			previousMap[rel.Name] = rel
		}
	}

	// Find target table names using all model definitions
	targetTableMap := make(map[string]string)
	for _, rel := range current {
		if rel.Type == "ForeignKey" || rel.Type == "OneToOne" {
			targetTable := findTargetTable(rel.To, allDefs)
			if targetTable != "" {
				targetTableMap[rel.Name] = targetTable
			}
		}
	}

	// Detect new foreign keys
	for name, rel := range currentMap {
		// Validate that the relation column exists in the table's fields
		hasColumn := false
		for _, field := range currentDef.Fields {
			if field.Name == rel.Name {
				hasColumn = true
				break
			}
		}
		if !hasColumn {
			// Skip FK if column doesn't exist - this prevents broken foreign keys
			continue
		}

		if _, exists := previousMap[name]; !exists {
			targetTable := targetTableMap[name]
			if targetTable == "" {
				// Skip if target table not found
				continue
			}
			changes = append(changes, &core.AddForeignKey{
				Table:       tableName,
				Relation:    rel,
				TargetTable: targetTable,
			})
		} else {
			// Check if foreign key changed
			prevRel := previousMap[name]
			if d.relationChanged(rel, prevRel) {
				targetTable := targetTableMap[name]
				if targetTable == "" {
					// Skip if target table not found
					continue
				}
				changes = append(changes, &core.ModifyForeignKey{
					Table:       tableName,
					OldFK:       prevRel,
					NewFK:       rel,
					TargetTable: targetTable,
				})
			}
		}
	}

	// Detect dropped foreign keys
	for name := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, &core.DropForeignKey{
				Table:  tableName,
				FKName: fmt.Sprintf("fk_%s_%s", tableName, name),
			})
		}
	}

	return changes, nil
}

// relationChanged checks if a relation has changed
func (d *Detector) relationChanged(current, previous generator.RelationDefinition) bool {
	if current.To != previous.To {
		return true
	}
	currentOnDelete, _ := current.Options["on_delete"].(string)
	previousOnDelete, _ := previous.Options["on_delete"].(string)
	if currentOnDelete != previousOnDelete {
		return true
	}
	currentOnUpdate, _ := current.Options["on_update"].(string)
	previousOnUpdate, _ := previous.Options["on_update"].(string)
	if currentOnUpdate != previousOnUpdate {
		return true
	}
	return false
}

// detectConstraintChanges detects constraint changes
func (d *Detector) detectConstraintChanges(tableName string, current, previous []generator.ConstraintDefinition) ([]core.Change, error) {
	var changes []core.Change

	currentMap := make(map[string]generator.ConstraintDefinition)
	previousMap := make(map[string]generator.ConstraintDefinition)

	for _, constr := range current {
		currentMap[constr.Name] = constr
	}

	for _, constr := range previous {
		previousMap[constr.Name] = constr
	}

	// Detect new constraints
	for name, constr := range currentMap {
		if _, exists := previousMap[name]; !exists {
			changes = append(changes, &core.AddConstraint{
				Table:      tableName,
				Constraint: constr,
			})
		}
	}

	// Detect dropped constraints
	for name := range previousMap {
		if _, exists := currentMap[name]; !exists {
			changes = append(changes, &core.DropConstraint{
				Table:          tableName,
				ConstraintName: name,
			})
		}
	}

	return changes, nil
}

// getTableName gets the table name from a model definition
func getTableName(def *generator.ModelDefinition) string {
	if def.Meta.TableName != "" {
		return def.Meta.TableName
	}
	return fmt.Sprintf("%ss", toSnakeCase(def.Name))
}

// findTargetTable finds the target table name for a relation
func findTargetTable(targetModel string, allDefs []*generator.ModelDefinition) string {
	for _, def := range allDefs {
		if strings.EqualFold(def.Name, targetModel) {
			return getTableName(def)
		}
	}
	return ""
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
}
