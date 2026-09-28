package verify

import (
	"fmt"
	"strings"
)

// CheckDangerousOperations checks for operations that should be reviewed
func CheckDangerousOperations(sql string) error {
	upperSQL := strings.ToUpper(sql)

	// List of dangerous operations that require explicit confirmation
	dangerousOps := []struct {
		op      string
		message string
	}{
		{"DROP DATABASE", "DROP DATABASE operations are extremely destructive"},
		{"DROP SCHEMA", "DROP SCHEMA operations are destructive"},
		{"TRUNCATE", "TRUNCATE operations delete all data"},
		{"DELETE FROM", "DELETE FROM operations modify data"},
		{"DROP TABLE", "DROP TABLE operations are destructive"},
		{"DROP COLUMN", "DROP COLUMN operations are destructive"},
	}

	for _, op := range dangerousOps {
		if strings.Contains(upperSQL, op.op) {
			// In production, you might want to require a flag or confirmation
			return fmt.Errorf("%s: %s", op.message, op.op)
		}
	}

	return nil
}

// CheckDataLossOperations checks for operations that could cause data loss
func CheckDataLossOperations(sql string) []string {
	var warnings []string
	upperSQL := strings.ToUpper(sql)

	// Check for DROP COLUMN
	if strings.Contains(upperSQL, "DROP COLUMN") {
		warnings = append(warnings, "DROP COLUMN detected - ensure data is backed up or migrated first")
	}

	// Check for ALTER COLUMN with type change (could cause data loss)
	if strings.Contains(upperSQL, "ALTER COLUMN") && strings.Contains(upperSQL, "TYPE") {
		warnings = append(warnings, "Column type change detected - verify data compatibility to prevent data loss")
	}

	// Check for DROP TABLE
	if strings.Contains(upperSQL, "DROP TABLE") {
		warnings = append(warnings, "DROP TABLE detected - this will permanently delete data")
	}

	// Check for TRUNCATE
	if strings.Contains(upperSQL, "TRUNCATE") {
		warnings = append(warnings, "TRUNCATE detected - this will delete all data in the table")
	}

	return warnings
}

// CheckPerformanceImpact checks for operations that could impact performance
func CheckPerformanceImpact(sql string) []string {
	var warnings []string
	upperSQL := strings.ToUpper(sql)

	// Check for CREATE INDEX (could lock table)
	if strings.Contains(upperSQL, "CREATE INDEX") && !strings.Contains(upperSQL, "CONCURRENTLY") {
		warnings = append(warnings, "CREATE INDEX without CONCURRENTLY - will lock table during creation (PostgreSQL)")
	}

	// Check for ALTER TABLE ADD COLUMN with default (could rewrite table in PostgreSQL)
	if strings.Contains(upperSQL, "ALTER TABLE") && strings.Contains(upperSQL, "ADD COLUMN") && strings.Contains(upperSQL, "DEFAULT") {
		warnings = append(warnings, "Adding column with default value may rewrite table (PostgreSQL) - consider adding as NULL first")
	}

	return warnings
}
