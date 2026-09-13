package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestBeginTx_NilDB(t *testing.T) {
	var database *DB
	_, err := database.BeginTx(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil DB, got nil")
	}
	if err.Error() != "database connection is nil" {
		t.Fatalf("expected 'database connection is nil', got %v", err)
	}

	database = &DB{} // Uninitialized, db.DB is nil
	_, err = database.BeginTx(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for uninitialized DB, got nil")
	}
	if err.Error() != "database connection is nil" {
		t.Fatalf("expected 'database connection is nil', got %v", err)
	}
}

func TestWithTx_NilDB(t *testing.T) {
	var database *DB
	err := database.WithTx(context.Background(), func(tx *Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nil DB, got nil")
	}
	if err.Error() != "database connection is nil" {
		t.Fatalf("expected 'database connection is nil', got %v", err)
	}
}

func TestValidateSavepointName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		// Valid cases
		{
			name:        "simple alphanumeric",
			input:       "savepoint1",
			expectedErr: nil,
		},
		{
			name:        "with underscore",
			input:       "save_point_1",
			expectedErr: nil,
		},
		{
			name:        "only letters",
			input:       "mysavepoint",
			expectedErr: nil,
		},
		{
			name:        "only underscores",
			input:       "___",
			expectedErr: nil,
		},
		{
			name:        "single character",
			input:       "a",
			expectedErr: nil,
		},
		{
			name:        "single digit",
			input:       "1",
			expectedErr: nil,
		},
		{
			name:        "mixed case letters",
			input:       "MySavePoint",
			expectedErr: nil,
		},
		{
			name:        "max length",
			input:       strings.Repeat("a", 128),
			expectedErr: nil,
		},

		// Invalid cases - empty
		{
			name:        "empty string",
			input:       "",
			expectedErr: ErrEmptySavepointName,
		},

		// Invalid cases - too long
		{
			name:        "too long",
			input:       strings.Repeat("a", 129),
			expectedErr: ErrSavepointNameTooLong,
		},

		// Invalid cases - special characters (SQL injection attempts)
		{
			name:        "with semicolon (SQL injection)",
			input:       "save;point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with space",
			input:       "save point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with hyphen",
			input:       "save-point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with dot",
			input:       "save.point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with quote (SQL injection)",
			input:       "save'point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with double quote (SQL injection)",
			input:       `save"point`,
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with backtick (SQL injection)",
			input:       "save`point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with parenthesis (SQL injection)",
			input:       "save(point)",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with equals sign (SQL injection)",
			input:       "save=point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "SQL injection attempt DROP",
			input:       "sp; DROP TABLE users;--",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with newline",
			input:       "save\npoint",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with tab",
			input:       "save\tpoint",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with special unicode",
			input:       "save©point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with at sign",
			input:       "save@point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with hash",
			input:       "save#point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with dollar sign",
			input:       "save$point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with percent (SQL injection)",
			input:       "save%point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with asterisk",
			input:       "save*point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with plus sign",
			input:       "save+point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with slash",
			input:       "save/point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with backslash",
			input:       `save\point`,
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with exclamation mark",
			input:       "save!point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with question mark",
			input:       "save?point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with less than",
			input:       "save<point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with greater than",
			input:       "save>point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with pipe",
			input:       "save|point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with ampersand",
			input:       "save&point",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with square brackets",
			input:       "save[point]",
			expectedErr: ErrInvalidSavepointName,
		},
		{
			name:        "with curly braces",
			input:       "save{point}",
			expectedErr: ErrInvalidSavepointName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSavepointName(tt.input)
			if err != tt.expectedErr {
				t.Errorf("validateSavepointName(%q) = %v, want %v", tt.input, err, tt.expectedErr)
			}
		})
	}
}

func TestValidateSavepointNameBoundaryConditions(t *testing.T) {
	// Test exactly at boundary (128 chars)
	exactlyMaxLength := strings.Repeat("x", 128)
	if err := validateSavepointName(exactlyMaxLength); err != nil {
		t.Errorf("name with exactly 128 chars should be valid, got error: %v", err)
	}

	// Test just over boundary (129 chars)
	justOverMaxLength := strings.Repeat("x", 129)
	if err := validateSavepointName(justOverMaxLength); err != ErrSavepointNameTooLong {
		t.Errorf("name with 129 chars should return ErrSavepointNameTooLong, got: %v", err)
	}
}

func TestValidateSavepointNameUnicodeLetters(t *testing.T) {
	// Unicode letters should be accepted (unicode.IsLetter returns true)
	unicodeTests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "greek letters",
			input:    "αβγδ",
			expected: nil, // Greek letters are valid letters
		},
		{
			name:     "cyrillic letters",
			input:    "тест",
			expected: nil, // Cyrillic letters are valid letters
		},
		{
			name:     "chinese characters",
			input:    "测试",
			expected: nil, // Chinese characters are valid letters
		},
		{
			name:     "mixed unicode and ascii",
			input:    "test_测试_1",
			expected: nil,
		},
	}

	for _, tt := range unicodeTests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSavepointName(tt.input)
			if err != tt.expected {
				t.Errorf("validateSavepointName(%q) = %v, want %v", tt.input, err, tt.expected)
			}
		})
	}
}

func setupTransactionTestDB(t *testing.T) *DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db := &DB{
		DB:     sqlDB,
		Driver: "sqlite3",
	}

	_, err = db.Exec("CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	return db
}

func TestWithTx_CommitFailure(t *testing.T) {
	db := setupTransactionTestDB(t)

	err := db.WithTx(context.Background(), func(tx *Tx) error {
		_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (1, 'item1')")
		if execErr != nil {
			return execErr
		}
		// Manually commit inside fn so deferred Commit fails with sql.ErrTxDone
		return tx.Commit()
	})

	if err == nil {
		t.Fatal("expected non-nil error when tx commit fails in defer, got nil")
	}
	if !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("expected sql.ErrTxDone, got %v", err)
	}
}

func TestWithTx_ErrorRollback(t *testing.T) {
	db := setupTransactionTestDB(t)

	expectedErr := errors.New("business logic error")
	err := db.WithTx(context.Background(), func(tx *Tx) error {
		_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (2, 'item2')")
		if execErr != nil {
			return execErr
		}
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	var count int
	queryErr := db.QueryRow("SELECT COUNT(*) FROM test_items WHERE id = 2").Scan(&count)
	if queryErr != nil {
		t.Fatalf("failed to query test_items: %v", queryErr)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after rollback, got %d", count)
	}
}

func TestWithTx_PanicRollback(t *testing.T) {
	db := setupTransactionTestDB(t)

	var recovered interface{}
	func() {
		defer func() {
			recovered = recover()
		}()
		_ = db.WithTx(context.Background(), func(tx *Tx) error {
			_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (3, 'item3')")
			if execErr != nil {
				return execErr
			}
			panic("something went critically wrong")
		})
	}()

	if recovered != "something went critically wrong" {
		t.Fatalf("expected panic 'something went critically wrong', got %v", recovered)
	}

	var count int
	queryErr := db.QueryRow("SELECT COUNT(*) FROM test_items WHERE id = 3").Scan(&count)
	if queryErr != nil {
		t.Fatalf("failed to query test_items: %v", queryErr)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after panic rollback, got %d", count)
	}
}

func TestWithTx_Success(t *testing.T) {
	db := setupTransactionTestDB(t)

	err := db.WithTx(context.Background(), func(tx *Tx) error {
		_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (4, 'item4')")
		return execErr
	})

	if err != nil {
		t.Fatalf("expected nil error on success, got %v", err)
	}

	var name string
	queryErr := db.QueryRow("SELECT name FROM test_items WHERE id = 4").Scan(&name)
	if queryErr != nil {
		t.Fatalf("failed to query test_items: %v", queryErr)
	}
	if name != "item4" {
		t.Fatalf("expected 'item4', got %q", name)
	}
}
