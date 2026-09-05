package db

import (
	"strings"
	"testing"
)

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
