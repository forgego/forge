package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Linter lints migration files for common issues
type Linter struct {
	rules []LintRule
}

// NewLinter creates a new migration linter
func NewLinter() *Linter {
	return &Linter{
		rules: getDefaultRules(),
	}
}

// LintResult represents a linting result
type LintResult struct {
	File    string
	Level   string // "error", "warning", "info"
	Message string
	Line    int
}

// LintMigration lints a migration file
func (l *Linter) LintMigration(filePath string) ([]LintResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	var results []LintResult
	lines := strings.Split(string(content), "\n")

	for _, rule := range l.rules {
		ruleResults := rule.Check(filePath, string(content), lines)
		results = append(results, ruleResults...)
	}

	return results, nil
}

// LintMigrationsDir lints all migrations in a directory
func (l *Linter) LintMigrationsDir(migrationsDir string) ([]LintResult, error) {
	pattern := filepath.Join(migrationsDir, "*_*.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	var allResults []LintResult
	for _, match := range matches {
		results, err := l.LintMigration(match)
		if err != nil {
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// LintRule defines a linting rule
type LintRule interface {
	Check(filePath, content string, lines []string) []LintResult
	Name() string
}

// getDefaultRules returns default linting rules
func getDefaultRules() []LintRule {
	return []LintRule{
		&NoTransactionRule{},
		&MissingDownMigrationRule{},
		&LargeMigrationRule{},
		&HardcodedValuesRule{},
		&MissingCommentsRule{},
		&InconsistentNamingRule{},
		&DestructiveOperationRule{},
		&DataLossWarningRule{},
		&PerformanceImpactRule{},
	}
}

// NoTransactionRule checks for missing transaction wrappers
type NoTransactionRule struct{}

func (r *NoTransactionRule) Name() string {
	return "no-transaction"
}

func (r *NoTransactionRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	if !strings.Contains(strings.ToUpper(content), "BEGIN") && !strings.Contains(strings.ToUpper(content), "START TRANSACTION") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "Migration does not use explicit transaction",
			Line:    0,
		})
	}
	return results
}

// MissingDownMigrationRule checks for missing down migrations
type MissingDownMigrationRule struct{}

func (r *MissingDownMigrationRule) Name() string {
	return "missing-down-migration"
}

func (r *MissingDownMigrationRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	if strings.HasSuffix(filePath, ".up.sql") {
		downPath := strings.Replace(filePath, ".up.sql", ".down.sql", 1)
		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			results = append(results, LintResult{
				File:    filePath,
				Level:   "warning",
				Message: "Missing corresponding down migration file",
				Line:    0,
			})
		}
	}
	return results
}

// LargeMigrationRule checks for migrations that are too large
type LargeMigrationRule struct{}

func (r *LargeMigrationRule) Name() string {
	return "large-migration"
}

func (r *LargeMigrationRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	const maxLines = 500
	if len(lines) > maxLines {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: fmt.Sprintf("Migration is very large (%d lines), consider splitting", len(lines)),
			Line:    0,
		})
	}
	return results
}

// HardcodedValuesRule checks for hardcoded values that should be parameterized
type HardcodedValuesRule struct{}

func (r *HardcodedValuesRule) Name() string {
	return "hardcoded-values"
}

func (r *HardcodedValuesRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	// Check for hardcoded IDs (common anti-pattern)
	hardcodedIDPattern := regexp.MustCompile(`\b(id|ID)\s*=\s*\d+`)
	if hardcodedIDPattern.MatchString(content) {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "info",
			Message: "Migration contains hardcoded ID values",
			Line:    0,
		})
	}
	return results
}

// MissingCommentsRule checks for migrations without comments
type MissingCommentsRule struct{}

func (r *MissingCommentsRule) Name() string {
	return "missing-comments"
}

func (r *MissingCommentsRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	if !strings.Contains(content, "--") && !strings.Contains(content, "/*") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "info",
			Message: "Migration has no comments explaining its purpose",
			Line:    0,
		})
	}
	return results
}

// InconsistentNamingRule checks for inconsistent naming conventions
type InconsistentNamingRule struct{}

func (r *InconsistentNamingRule) Name() string {
	return "inconsistent-naming"
}

func (r *InconsistentNamingRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	// Check for mixed case table names
	mixedCasePattern := regexp.MustCompile(`CREATE TABLE\s+([A-Z][a-z]+[A-Z])`)
	if mixedCasePattern.MatchString(content) {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "Migration uses mixed case table names (should use snake_case)",
			Line:    0,
		})
	}
	return results
}

// DestructiveOperationRule checks for destructive operations that could cause data loss
type DestructiveOperationRule struct{}

func (r *DestructiveOperationRule) Name() string {
	return "destructive-operation"
}

func (r *DestructiveOperationRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	upperContent := strings.ToUpper(content)

	// Check for DROP TABLE
	if strings.Contains(upperContent, "DROP TABLE") && !strings.Contains(upperContent, "IF EXISTS") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "error",
			Message: "DANGER: DROP TABLE without IF EXISTS - could cause data loss if table doesn't exist",
			Line:    0,
		})
	}

	// Check for DROP TABLE (even with IF EXISTS, it's still destructive)
	if strings.Contains(upperContent, "DROP TABLE") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "WARNING: DROP TABLE operation detected - this will permanently delete data",
			Line:    0,
		})
	}

	// Check for TRUNCATE
	if strings.Contains(upperContent, "TRUNCATE") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "WARNING: TRUNCATE operation detected - this will delete all data in the table",
			Line:    0,
		})
	}

	// Check for DELETE without WHERE
	deletePattern := regexp.MustCompile(`(?i)DELETE\s+FROM\s+\w+\s*(?:;|$|\n)`)
	if deletePattern.MatchString(content) {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "error",
			Message: "DANGER: DELETE without WHERE clause - will delete all rows in table",
			Line:    0,
		})
	}

	return results
}

// DataLossWarningRule checks for operations that could cause data loss
type DataLossWarningRule struct{}

func (r *DataLossWarningRule) Name() string {
	return "data-loss-warning"
}

func (r *DataLossWarningRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	upperContent := strings.ToUpper(content)

	// Check for DROP COLUMN
	if strings.Contains(upperContent, "DROP COLUMN") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "WARNING: DROP COLUMN detected - ensure data is backed up or migrated first",
			Line:    0,
		})
	}

	// Check for ALTER COLUMN with type change (could cause data loss)
	if strings.Contains(upperContent, "ALTER COLUMN") && strings.Contains(upperContent, "TYPE") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "warning",
			Message: "WARNING: Column type change detected - verify data compatibility to prevent data loss",
			Line:    0,
		})
	}

	return results
}

// PerformanceImpactRule checks for operations that could impact performance
type PerformanceImpactRule struct{}

func (r *PerformanceImpactRule) Name() string {
	return "performance-impact"
}

func (r *PerformanceImpactRule) Check(filePath, content string, lines []string) []LintResult {
	var results []LintResult
	upperContent := strings.ToUpper(content)

	// Check for CREATE INDEX (could lock table)
	if strings.Contains(upperContent, "CREATE INDEX") && !strings.Contains(upperContent, "CONCURRENTLY") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "info",
			Message: "INFO: CREATE INDEX without CONCURRENTLY - will lock table during creation (PostgreSQL)",
			Line:    0,
		})
	}

	// Check for ALTER TABLE ADD COLUMN with default (could rewrite table in PostgreSQL)
	if strings.Contains(upperContent, "ALTER TABLE") && strings.Contains(upperContent, "ADD COLUMN") && strings.Contains(upperContent, "DEFAULT") {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "info",
			Message: "INFO: Adding column with default value may rewrite table (PostgreSQL) - consider adding as NULL first",
			Line:    0,
		})
	}

	// Check for large UPDATE statements
	updatePattern := regexp.MustCompile(`(?i)UPDATE\s+\w+\s+SET`)
	updateCount := len(updatePattern.FindAllString(content, -1))
	if updateCount > 5 {
		results = append(results, LintResult{
			File:    filePath,
			Level:   "info",
			Message: fmt.Sprintf("INFO: Multiple UPDATE statements (%d) - consider batching for large tables", updateCount),
			Line:    0,
		})
	}

	return results
}
