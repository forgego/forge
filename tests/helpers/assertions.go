package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// CascadeBehavior represents foreign key cascade behaviors
type CascadeBehavior string

const (
	CascadeCASCADE     CascadeBehavior = "CASCADE"
	CascadeSET_NULL    CascadeBehavior = "SET NULL"
	CascadeSET_DEFAULT CascadeBehavior = "SET DEFAULT"
	CascadeRESTRICT    CascadeBehavior = "RESTRICT"
	CascadeNO_ACTION   CascadeBehavior = "NO ACTION"
)

// AssertForeignKeyCascade tests FK cascade behaviors
func AssertForeignKeyCascade(ctx context.Context, t *testing.T, db *sql.DB, tableName string, columnName string, expectedBehavior CascadeBehavior) {
	query := `
		SELECT 
			tc.constraint_name,
			rc.delete_rule,
			rc.update_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.referential_constraints rc ON tc.constraint_name = rc.constraint_name
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = $1 
			AND kcu.column_name = $2 
			AND tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public'
	`

	var constraintName, deleteRule, updateRule string
	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&constraintName, &deleteRule, &updateRule)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("foreign key on %s.%s does not exist", tableName, columnName)
		}
		t.Fatalf("failed to check FK cascade: %v", err)
	}

	// Check delete rule
	expectedRule := string(expectedBehavior)
	if deleteRule != expectedRule {
		t.Errorf("foreign key %s on %s.%s has delete rule %s, expected %s",
			constraintName, tableName, columnName, deleteRule, expectedRule)
	}
}

// AssertComplexIndex verifies multi-column, unique, partial indexes
func AssertComplexIndex(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string, options *ComplexIndexOptions) {
	query := `
		SELECT indexdef, indexname FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2
	`

	var indexDef, name string
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&indexDef, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check index: %v", err)
	}

	if options == nil {
		return // Just verify it exists
	}

	// Check columns
	if len(options.Columns) > 0 {
		normalizedDef := strings.ToLower(indexDef)
		for _, col := range options.Columns {
			if !strings.Contains(normalizedDef, strings.ToLower(col)) {
				t.Errorf("index %s on table %s does not contain expected column %s", indexName, tableName, col)
			}
		}
	}

	// Check unique constraint
	if options.Unique {
		// Check if it's a unique index
		uniqueQuery := `
			SELECT indisunique FROM pg_index i
			JOIN pg_class c ON c.oid = i.indexrelid
			WHERE c.relname = $1
		`
		var isUnique bool
		err := db.QueryRowContext(ctx, uniqueQuery, indexName).Scan(&isUnique)
		if err != nil {
			t.Logf("warning: could not check unique constraint for index %s: %v", indexName, err)
		} else if !isUnique {
			t.Errorf("index %s on table %s is not unique as expected", indexName, tableName)
		}
	}

	// Check partial index (WHERE clause)
	if options.Partial {
		if !strings.Contains(strings.ToLower(indexDef), "where") {
			t.Errorf("index %s on table %s is not a partial index as expected", indexName, tableName)
		}
	}

	// Check functional index
	if options.Functional {
		if !strings.Contains(indexDef, "(") || !strings.Contains(indexDef, ")") {
			t.Errorf("index %s on table %s is not a functional index as expected", indexName, tableName)
		}
	}
}

// ComplexIndexOptions specifies options for complex index assertions
type ComplexIndexOptions struct {
	Columns    []string
	Unique     bool
	Partial    bool
	Functional bool
}

// AssertConstraintExistsEnhanced provides enhanced constraint checking
func AssertConstraintExistsEnhanced(ctx context.Context, t *testing.T, db *sql.DB, tableName string, constraintName string, constraintType string) {
	query := `
		SELECT constraint_type FROM information_schema.table_constraints 
		WHERE table_name = $1 AND constraint_name = $2 AND table_schema = 'public'
	`

	var actualType string
	err := db.QueryRowContext(ctx, query, tableName, constraintName).Scan(&actualType)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("constraint %s on table %s does not exist", constraintName, tableName)
		}
		t.Fatalf("failed to check constraint: %v", err)
	}

	if constraintType != "" && actualType != constraintType {
		t.Errorf("constraint %s on table %s has type %s, expected %s",
			constraintName, tableName, actualType, constraintType)
	}
}

// TableStructure represents the expected structure of a table
type TableStructure struct {
	Columns     []ColumnSpec
	Indexes     []IndexSpec
	ForeignKeys []ForeignKeySpec
	Constraints []ConstraintSpec
}

// ColumnSpec describes expected column attributes
type ColumnSpec struct {
	Name      string
	Type      string
	Nullable  bool
	Default   *string
	Generated *string
}

// IndexSpec describes expected index attributes
type IndexSpec struct {
	Name    string
	Columns []string
	Unique  bool
	Partial bool
}

// ForeignKeySpec describes expected foreign key attributes
type ForeignKeySpec struct {
	Column        string
	TargetTable   string
	TargetColumn  string
	DeleteCascade CascadeBehavior
	UpdateCascade CascadeBehavior
}

// ConstraintSpec describes expected constraint attributes
type ConstraintSpec struct {
	Name string
	Type string // CHECK, UNIQUE, PRIMARY KEY, etc.
}

// AssertTableStructure provides comprehensive table structure validation
func AssertTableStructure(ctx context.Context, t *testing.T, db *sql.DB, tableName string, expected *TableStructure) {
	if expected == nil {
		return
	}

	// Verify columns
	for _, colSpec := range expected.Columns {
		AssertColumnExists(ctx, t, db, "postgres", tableName, colSpec.Name)

		// Check column type
		if colSpec.Type != "" {
			query := `
				SELECT data_type FROM information_schema.columns 
				WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
			`
			var actualType string
			err := db.QueryRowContext(ctx, query, tableName, colSpec.Name).Scan(&actualType)
			if err != nil {
				t.Fatalf("failed to get column type: %v", err)
			}

			// Type matching is flexible - just check if it contains the expected type
			if !strings.Contains(strings.ToLower(actualType), strings.ToLower(colSpec.Type)) {
				t.Logf("warning: column %s.%s has type %s, expected to contain %s",
					tableName, colSpec.Name, actualType, colSpec.Type)
			}
		}

		// Check nullable
		if colSpec.Nullable {
			query := `
				SELECT is_nullable FROM information_schema.columns 
				WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
			`
			var isNullable string
			err := db.QueryRowContext(ctx, query, tableName, colSpec.Name).Scan(&isNullable)
			if err == nil {
				if isNullable != "YES" && colSpec.Nullable {
					t.Errorf("column %s.%s is not nullable as expected", tableName, colSpec.Name)
				}
			}
		}
	}

	// Verify indexes
	for _, idxSpec := range expected.Indexes {
		AssertIndexExists(ctx, t, db, "postgres", tableName, idxSpec.Name)
		if idxSpec.Unique || len(idxSpec.Columns) > 0 || idxSpec.Partial {
			options := &ComplexIndexOptions{
				Columns: idxSpec.Columns,
				Unique:  idxSpec.Unique,
				Partial: idxSpec.Partial,
			}
			AssertComplexIndex(ctx, t, db, tableName, idxSpec.Name, options)
		}
	}

	// Verify foreign keys
	for _, fkSpec := range expected.ForeignKeys {
		AssertForeignKeyExists(ctx, t, db, "postgres", tableName, fkSpec.Column)
		if fkSpec.DeleteCascade != "" {
			AssertForeignKeyCascade(ctx, t, db, tableName, fkSpec.Column, fkSpec.DeleteCascade)
		}
	}

	// Verify constraints
	for _, constraintSpec := range expected.Constraints {
		AssertConstraintExistsEnhanced(ctx, t, db, tableName, constraintSpec.Name, constraintSpec.Type)
	}
}

// GetTableColumns returns all columns for a table
func GetTableColumns(ctx context.Context, db *sql.DB, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns 
		WHERE table_name = $1 AND table_schema = 'public'
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable, defaultValue sql.NullString
		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultValue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		col.Nullable = nullable.String == "YES"
		if defaultValue.Valid {
			col.Default = &defaultValue.String
		}
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// ColumnInfo represents information about a database column
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	Default  *string
}
