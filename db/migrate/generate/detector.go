package generate

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/sql"
)

// ChangeDetector detects changes between current and previous model states
type ChangeDetector interface {
	// DetectChanges compares current models to previous state and returns all changes
	DetectChanges(current, previous []*generator.ModelDefinition) ([]core.Change, error)
}

// Detector is the default implementation of ChangeDetector
type Detector struct {
	// driver, when set, lets the detector compare columns by the DDL the
	// driver's builder renders, so a model that still renders the recorded
	// column produces no change.
	driver core.Driver
}

// NewDetector creates a new change detector
func NewDetector() ChangeDetector {
	return &Detector{}
}

// NewDetectorForDriver creates a change detector that compares columns by the
// DDL rendered for driver.
func NewDetectorForDriver(driver core.Driver) ChangeDetector {
	return &Detector{driver: driver}
}

// isForeignKeyRelation reports whether a relation type creates a foreign key
// constraint. Both the fluent (ForeignKey) and functional (ForeignKeyField)
// constructor names are accepted.
func isForeignKeyRelation(relationType string) bool {
	switch relationType {
	case "ForeignKey", "ForeignKeyField", "OneToOne", "OneToOneField":
		return true
	}
	return false
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
	for _, tableName := range sortedKeys(currentMap) {
		currentDef := currentMap[tableName]
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
				if isForeignKeyRelation(rel.Type) {
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

			// Add table constraints for new table
			for _, constr := range currentDef.Meta.Constraints {
				changes = append(changes, &core.AddConstraint{
					Table:      tableName,
					Constraint: constr,
				})
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
	for _, tableName := range sortedKeys(previousMap) {
		prevDef := previousMap[tableName]
		if _, exists := currentMap[tableName]; !exists {
			changes = append(changes, &core.DropTable{
				Table:      tableName,
				Definition: prevDef,
			})
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
	for _, name := range sortedKeys(currentMap) {
		field := currentMap[name]
		if _, exists := previousMap[name]; !exists {
			changes = append(changes, &core.AddColumn{
				Table:  tableName,
				Column: field,
			})
		}
	}

	// Detect dropped columns
	for _, name := range sortedKeys(previousMap) {
		prevField := previousMap[name]
		if _, exists := currentMap[name]; !exists {
			col := prevField
			changes = append(changes, &core.DropColumn{
				Table:      tableName,
				ColumnName: name,
				Column:     &col,
			})
		}
	}

	// Detect modified columns
	for _, name := range sortedKeys(currentMap) {
		currentField := currentMap[name]
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
	if d.sameColumnDDL(current, previous) {
		return false
	}
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
	// Only compare schema-relevant options that affect DDL
	schemaOptionKeys := []string{
		"max_length",
		"max_digits",
		"decimal_places",
		"db_type",
		"db_column",
		"db_default",
		"generated",
		"generated_expr",
		"generated_stored",
	}
	for _, key := range schemaOptionKeys {
		currVal, currHas := current.Options[key]
		prevVal, prevHas := previous.Options[key]
		if currHas != prevHas {
			if currHas && isZeroOrNil(currVal) && !prevHas {
				continue
			}
			if prevHas && isZeroOrNil(prevVal) && !currHas {
				continue
			}
			return true
		}
		if currHas && prevHas && !reflect.DeepEqual(currVal, prevVal) {
			return true
		}
	}
	return false
}

func isZeroOrNil(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case int:
		return val == 0
	case string:
		return val == ""
	case bool:
		return !val
	default:
		return false
	}
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
	for _, name := range sortedKeys(currentMap) {
		idx := currentMap[name]
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
	for _, name := range sortedKeys(previousMap) {
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

// sameColumnDDL reports whether both fields render to the same column DDL for
// the detector's driver. Fields reconstructed from migration files carry their
// recorded SQL type, so an unchanged model compares equal even when its builder
// name (StringField) differs from the name inferred from SQL (String).
func (d *Detector) sameColumnDDL(current, previous generator.FieldDefinition) bool {
	if d.driver == "" {
		return false
	}
	currentDDL, err := sql.ColumnDefinition(d.driver, canonicalColumn(current))
	if err != nil {
		return false
	}
	previousDDL, err := sql.ColumnDefinition(d.driver, canonicalColumn(previous))
	if err != nil {
		return false
	}
	return normalizeDDL(currentDDL) == normalizeDDL(previousDDL)
}

// normalizeDDL upper-cases DDL and collapses its whitespace outside quoted
// literals and identifiers, which it keeps exactly, so a default that changes
// only in case or spacing still compares unequal.
func normalizeDDL(ddl string) string {
	var b strings.Builder
	var quote rune
	pendingSpace := false
	for _, r := range ddl {
		if quote != 0 {
			b.WriteRune(r)
			if r == quote {
				quote = 0
			}
			continue
		}
		if unicode.IsSpace(r) {
			pendingSpace = b.Len() > 0
			continue
		}
		if pendingSpace {
			b.WriteByte(' ')
			pendingSpace = false
		}
		if r == '\'' || r == '"' {
			quote = r
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// canonicalColumn states the nullability AutoNowAdd implies explicitly, so the
// builder renders NOT NULL in the same position as for a column read back from
// SQL, where it is an ordinary required field.
func canonicalColumn(field generator.FieldDefinition) generator.FieldDefinition {
	if autoNowAdd, ok := field.Options["auto_now_add"].(bool); ok && autoNowAdd && !field.PrimaryKey {
		field.Required = true
	}
	return field
}

// detectForeignKeyChanges detects foreign key changes
func (d *Detector) detectForeignKeyChanges(tableName string, current, previous []generator.RelationDefinition, currentDef, previousDef *generator.ModelDefinition, allDefs []*generator.ModelDefinition) ([]core.Change, error) {
	var changes []core.Change

	currentMap := make(map[string]generator.RelationDefinition)
	previousMap := make(map[string]generator.RelationDefinition)

	for _, rel := range current {
		if isForeignKeyRelation(rel.Type) {
			currentMap[rel.Name] = rel
		}
	}

	for _, rel := range previous {
		if isForeignKeyRelation(rel.Type) {
			previousMap[rel.Name] = rel
		}
	}

	// Find target table names using all model definitions
	targetTableMap := make(map[string]string)
	for _, rel := range current {
		if isForeignKeyRelation(rel.Type) {
			targetTable := findTargetTable(rel.To, allDefs)
			if targetTable != "" {
				targetTableMap[rel.Name] = targetTable
			}
		}
	}

	// Detect new foreign keys
	for _, name := range sortedKeys(currentMap) {
		rel := currentMap[name]
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
			if d.relationChanged(rel, prevRel, targetTableMap[name], resolveTargetTable(prevRel.To, allDefs)) {
				targetTable := targetTableMap[name]
				if targetTable == "" {
					// Skip if target table not found
					continue
				}
				changes = append(changes, &core.ModifyForeignKey{
					Table:          tableName,
					OldFK:          prevRel,
					NewFK:          rel,
					TargetTable:    targetTable,
					OldTargetTable: resolveTargetTable(prevRel.To, allDefs),
				})
			}
		}
	}

	// Detect dropped foreign keys
	for _, name := range sortedKeys(previousMap) {
		if _, exists := currentMap[name]; !exists {
			prevRel := previousMap[name]
			changes = append(changes, &core.DropForeignKey{
				Table:       tableName,
				FKName:      fmt.Sprintf("fk_%s_%s", tableName, name),
				Relation:    &prevRel,
				TargetTable: resolveTargetTable(prevRel.To, allDefs),
			})
		}
	}

	return changes, nil
}

// relationChanged checks if a relation has changed. Targets are compared as
// table names and referential actions as the SQL they render to, because
// relations reconstructed from migration files carry table names and SQL
// actions while model relations carry model names and cascade constants.
func (d *Detector) relationChanged(current, previous generator.RelationDefinition, currentTarget, previousTarget string) bool {
	if currentTarget != previousTarget {
		return true
	}
	currentOnDelete, _ := current.Options["on_delete"].(string)
	previousOnDelete, _ := previous.Options["on_delete"].(string)
	if sql.CascadeAction(currentOnDelete) != sql.CascadeAction(previousOnDelete) {
		return true
	}
	currentOnUpdate, _ := current.Options["on_update"].(string)
	previousOnUpdate, _ := previous.Options["on_update"].(string)
	return sql.CascadeAction(currentOnUpdate) != sql.CascadeAction(previousOnUpdate)
}

// resolveTargetTable returns the table for a relation target given as either a
// model name or, for relations reconstructed from migrations, a table name.
func resolveTargetTable(target string, allDefs []*generator.ModelDefinition) string {
	if table := findTargetTable(target, allDefs); table != "" {
		return table
	}
	return target
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

	// Detect new constraints, and changed ones, which are dropped and re-added
	for _, name := range sortedKeys(currentMap) {
		constr := currentMap[name]
		prevConstr, exists := previousMap[name]
		if exists && !constraintChanged(constr, prevConstr) {
			continue
		}
		if exists {
			changes = append(changes, &core.DropConstraint{
				Table:          tableName,
				ConstraintName: name,
				Constraint:     &prevConstr,
			})
		}
		changes = append(changes, &core.AddConstraint{
			Table:      tableName,
			Constraint: constr,
		})
	}

	// Detect dropped constraints
	for _, name := range sortedKeys(previousMap) {
		if _, exists := currentMap[name]; !exists {
			prevConstr := previousMap[name]
			changes = append(changes, &core.DropConstraint{
				Table:          tableName,
				ConstraintName: name,
				Constraint:     &prevConstr,
			})
		}
	}

	return changes, nil
}

// constraintChanged reports whether a constraint's type, condition or fields
// changed. Conditions compare as normalized SQL, so a constraint read back
// from a migration file equals the model's unchanged one.
func constraintChanged(current, previous generator.ConstraintDefinition) bool {
	if !strings.EqualFold(current.Type, previous.Type) {
		return true
	}
	if normalizeDDL(current.Condition) != normalizeDDL(previous.Condition) {
		return true
	}
	return !reflect.DeepEqual(nonNil(current.Fields), nonNil(previous.Fields))
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// sortedKeys returns a map's keys in order, so detected changes, and the
// migration SQL written from them, do not depend on map iteration order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
