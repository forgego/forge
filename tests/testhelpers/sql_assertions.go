package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

// AssertTableExists checks if a table exists in the database
func AssertTableExists(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tableName string) {
	var query string
	switch {
	case strings.Contains(dialect, "postgres"):
		query = `SELECT 1 FROM information_schema.tables WHERE table_name = $1 AND table_schema = 'public'`
	case strings.Contains(dialect, "sqlite"):
		query = `SELECT 1 FROM sqlite_master WHERE type='table' AND name=?`
	default:
		t.Fatalf("unsupported dialect: %s", dialect)
	}

	var exists int
	err := db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("table %s does not exist", tableName)
		}
		t.Fatalf("failed to check table existence: %v", err)
	}
}

// AssertColumnExists checks if a column exists in a table
func AssertColumnExists(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tableName string, columnName string) {
	var query string
	switch {
	case strings.Contains(dialect, "postgres"):
		query = `SELECT 1 FROM information_schema.columns WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'`
	case strings.Contains(dialect, "sqlite"):
		query = `PRAGMA table_info(` + tableName + `)`
	default:
		t.Fatalf("unsupported dialect: %s", dialect)
	}

	if strings.Contains(dialect, "sqlite") {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			t.Fatalf("failed to check column: %v", err)
		}
		defer rows.Close()

		found := false
		for rows.Next() {
			var cid, notnull, pk int
			var name, type_ string
			var dflt_value sql.NullString
			if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
				t.Fatalf("failed to scan pragma result: %v", err)
			}
			if name == columnName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("column %s.%s does not exist", tableName, columnName)
		}
	} else {
		var exists int
		err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Fatalf("column %s.%s does not exist", tableName, columnName)
			}
			t.Fatalf("failed to check column existence: %v", err)
		}
	}
}

// AssertIndexExists checks if an index exists in the database
func AssertIndexExists(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tableName string, indexName string) {
	var query string
	switch {
	case strings.Contains(dialect, "postgres"):
		query = `SELECT 1 FROM pg_indexes WHERE tablename = $1 AND indexname = $2`
	case strings.Contains(dialect, "sqlite"):
		query = `SELECT 1 FROM sqlite_master WHERE type='index' AND name=? AND tbl_name=?`
	default:
		t.Fatalf("unsupported dialect: %s", dialect)
	}

	var exists int
	// For Postgres query uses (tablename, indexname), for SQLite query uses (name, tbl_name)
	var qErr error
	if strings.Contains(dialect, "sqlite") {
		qErr = db.QueryRowContext(ctx, query, indexName, tableName).Scan(&exists)
	} else {
		qErr = db.QueryRowContext(ctx, query, tableName, indexName).Scan(&exists)
	}
	if qErr != nil {
		if qErr == sql.ErrNoRows {
			t.Fatalf("index %s on table %s does not exist", indexName, tableName)
		}
		t.Fatalf("failed to check index existence: %v", qErr)
	}
}

// AssertConstraintExists checks if a constraint exists
func AssertConstraintExists(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tableName string, constraintName string) {
	var query string
	switch {
	case strings.Contains(dialect, "postgres"):
		query = `SELECT 1 FROM information_schema.table_constraints WHERE table_name = $1 AND constraint_name = $2 AND table_schema = 'public'`
	default:
		t.Fatalf("constraint checking not implemented for dialect: %s", dialect)
	}

	var exists int
	err := db.QueryRowContext(ctx, query, tableName, constraintName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatalf("constraint %s on table %s does not exist", constraintName, tableName)
		}
		t.Fatalf("failed to check constraint: %v", err)
	}
}

// AssertForeignKeyExists checks if a foreign key constraint exists
func AssertForeignKeyExists(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tableName string, columnName string) {
	switch {
	case strings.Contains(dialect, "postgres"):
		query := `
			SELECT 1 FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			WHERE tc.table_name = $1 AND kcu.column_name = $2 AND tc.constraint_type = 'FOREIGN KEY'
		`
		var exists int
		err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Fatalf("foreign key on %s.%s does not exist", tableName, columnName)
			}
			t.Fatalf("failed to check FK: %v", err)
		}
	default:
		t.Fatalf("FK checking not implemented for dialect: %s", dialect)
	}
}

// AssertRowCount checks the number of rows in a table
func AssertRowCount(ctx context.Context, t *testing.T, db *sql.DB, tableName string, expectedCount int64) {
	var count int64
	err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != expectedCount {
		t.Fatalf("expected %d rows in %s, got %d", expectedCount, tableName, count)
	}
}

// RunSQLExpectError executes SQL and expects an error
func RunSQLExpectError(ctx context.Context, db *sql.DB, sql string) error {
	_, err := db.ExecContext(ctx, sql)
	return err
}

// RunSQLExpectSuccess executes SQL and expects no error
func RunSQLExpectSuccess(ctx context.Context, t *testing.T, db *sql.DB, sql string) {
	_, err := db.ExecContext(ctx, sql)
	if err != nil {
		t.Fatalf("SQL execution failed: %v\nSQL: %s", err, sql)
	}
}

// CleanupTables drops the specified tables if they exist
// This is useful for test cleanup to avoid "table already exists" errors
func CleanupTables(ctx context.Context, t *testing.T, db *sql.DB, dialect string, tables []string) {
	if len(tables) == 0 {
		return
	}

	var dropSQL string
	switch {
	case strings.Contains(dialect, "postgres"):
		// Drop tables with CASCADE to handle foreign key dependencies
		for _, table := range tables {
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE;", table)
			_, err := db.ExecContext(ctx, dropSQL)
			if err != nil {
				t.Logf("Warning: failed to drop table %s: %v", table, err)
			}
		}
	case strings.Contains(dialect, "sqlite"):
		// SQLite doesn't support CASCADE, but we can drop in reverse order
		// For simplicity, just drop each table
		for _, table := range tables {
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
			_, err := db.ExecContext(ctx, dropSQL)
			if err != nil {
				t.Logf("Warning: failed to drop table %s: %v", table, err)
			}
		}
	default:
		t.Fatalf("unsupported dialect for cleanup: %s", dialect)
	}
}
