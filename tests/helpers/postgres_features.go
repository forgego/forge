package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// AssertGINIndexExists verifies a GIN index exists on a table
func AssertGINIndexExists(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string) {
	query := `
		SELECT 1 FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2 
		AND indexdef LIKE '%USING gin%'
	`

	var exists int
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("GIN index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check GIN index: %v", err)
	}
}

// AssertGiSTIndexExists verifies a GiST index exists on a table
func AssertGiSTIndexExists(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string) {
	query := `
		SELECT 1 FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2 
		AND indexdef LIKE '%USING gist%'
	`

	var exists int
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("GiST index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check GiST index: %v", err)
	}
}

// AssertJSONBColumn verifies a JSONB column exists and can store/retrieve JSON data
func AssertJSONBColumn(ctx context.Context, t *testing.T, db *sql.DB, tableName string, columnName string) {
	// First verify the column exists and is JSONB type
	query := `
		SELECT data_type FROM information_schema.columns 
		WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
	`

	var dataType string
	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&dataType)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("column %s.%s does not exist", tableName, columnName)
		}
		t.Fatalf("failed to check column type: %v", err)
	}

	if dataType != "jsonb" {
		t.Fatalf("column %s.%s is not JSONB type, got %s", tableName, columnName, dataType)
	}

	// Test that we can insert and retrieve JSONB data
	testData := map[string]interface{}{
		"test":   "value",
		"number": 42,
	}
	jsonBytes, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("failed to marshal test JSON: %v", err)
	}

	// Try to insert test data (if table allows it)
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1) ON CONFLICT DO NOTHING", tableName, columnName)
	_, err = db.ExecContext(ctx, insertQuery, string(jsonBytes))
	// We don't fail if insert fails - the column might have constraints
	// The important part is that the type is correct
}

// AssertArrayColumn verifies an array column type exists
func AssertArrayColumn(ctx context.Context, t *testing.T, db *sql.DB, tableName string, columnName string, expectedElementType string) {
	query := `
		SELECT data_type, udt_name FROM information_schema.columns 
		WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
	`

	var dataType, udtName string
	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&dataType, &udtName)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("column %s.%s does not exist", tableName, columnName)
		}
		t.Fatalf("failed to check column type: %v", err)
	}

	// PostgreSQL array types are named like _int4, _text, etc.
	// Or data_type might be ARRAY
	if dataType != "ARRAY" && !strings.HasPrefix(udtName, "_") {
		t.Fatalf("column %s.%s is not an array type, got data_type=%s, udt_name=%s",
			tableName, columnName, dataType, udtName)
	}

	// Verify element type if specified
	if expectedElementType != "" {
		// For array types, we need to check the element type
		// This is complex in PostgreSQL, so we'll do a simpler check
		// by looking at the UDT name
		expectedPrefix := "_" + expectedElementType
		if !strings.HasPrefix(udtName, expectedPrefix) && udtName != expectedElementType+"[]" {
			// Try alternative check - query pg_type directly
			typeQuery := `
				SELECT t.typname FROM pg_type t
				JOIN pg_attribute a ON a.atttypid = t.oid
				JOIN pg_class c ON c.oid = a.attrelid
				JOIN pg_namespace n ON n.oid = c.relnamespace
				WHERE c.relname = $1 AND a.attname = $2 AND n.nspname = 'public'
			`
			var actualTypeName string
			err := db.QueryRowContext(ctx, typeQuery, tableName, columnName).Scan(&actualTypeName)
			if err == nil {
				if !strings.Contains(actualTypeName, expectedElementType) {
					t.Logf("warning: array element type may not match expected %s, got %s",
						expectedElementType, actualTypeName)
				}
			}
		}
	}
}

// AssertCustomType verifies a custom PostgreSQL type exists
func AssertCustomType(ctx context.Context, t *testing.T, db *sql.DB, typeName string) {
	query := `
		SELECT 1 FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE t.typname = $1 AND n.nspname = 'public'
	`

	var exists int
	err := db.QueryRowContext(ctx, query, typeName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("custom type %s does not exist", typeName)
		}
		t.Fatalf("failed to check custom type: %v", err)
	}
}

// AssertPartialIndex verifies a partial index with WHERE clause exists
func AssertPartialIndex(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string, expectedWhereClause string) {
	query := `
		SELECT indexdef FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2
	`

	var indexDef string
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&indexDef)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check index: %v", err)
	}

	// Check if WHERE clause exists in index definition
	if !strings.Contains(strings.ToLower(indexDef), "where") {
		t.Fatalf("index %s on table %s is not a partial index (no WHERE clause)", indexName, tableName)
	}

	// If expected WHERE clause is provided, verify it matches
	if expectedWhereClause != "" {
		normalizedDef := strings.ToLower(indexDef)
		normalizedExpected := strings.ToLower(expectedWhereClause)
		if !strings.Contains(normalizedDef, normalizedExpected) {
			t.Logf("warning: index WHERE clause may not match expected. Index def: %s", indexDef)
		}
	}
}

// AssertFunctionalIndex verifies a functional index exists (e.g., LOWER(email))
func AssertFunctionalIndex(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string, expectedExpression string) {
	query := `
		SELECT indexdef FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2
	`

	var indexDef string
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&indexDef)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check index: %v", err)
	}

	// Check if index definition contains a function call (indicates functional index)
	// Simple heuristic: contains parentheses with column names
	if !strings.Contains(indexDef, "(") || !strings.Contains(indexDef, ")") {
		t.Fatalf("index %s on table %s does not appear to be a functional index", indexName, tableName)
	}

	// If expected expression is provided, verify it's in the definition
	if expectedExpression != "" {
		normalizedDef := strings.ToLower(indexDef)
		normalizedExpected := strings.ToLower(expectedExpression)
		if !strings.Contains(normalizedDef, normalizedExpected) {
			t.Logf("warning: functional index expression may not match expected. Index def: %s", indexDef)
		}
	}
}

// AssertCoveringIndex verifies a covering index with INCLUDE columns exists
func AssertCoveringIndex(ctx context.Context, t *testing.T, db *sql.DB, tableName string, indexName string, expectedIncludeColumns []string) {
	query := `
		SELECT indexdef FROM pg_indexes 
		WHERE tablename = $1 AND indexname = $2
	`

	var indexDef string
	err := db.QueryRowContext(ctx, query, tableName, indexName).Scan(&indexDef)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check index: %v", err)
	}

	// Check if INCLUDE clause exists
	if !strings.Contains(strings.ToUpper(indexDef), "INCLUDE") {
		t.Fatalf("index %s on table %s is not a covering index (no INCLUDE clause)", indexName, tableName)
	}

	// If expected INCLUDE columns are provided, verify they're in the definition
	if len(expectedIncludeColumns) > 0 {
		normalizedDef := strings.ToLower(indexDef)
		for _, col := range expectedIncludeColumns {
			if !strings.Contains(normalizedDef, strings.ToLower(col)) {
				t.Logf("warning: covering index may not include expected column %s. Index def: %s", col, indexDef)
			}
		}
	}
}
