package utils

import (
	"testing"
	"time"
)

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "-"},
		{"string", "hello", "hello"},
		{"int", int(42), "42"},
		{"int32", int32(42), "42"},
		{"int64", int64(42), "42"},
		{"float32", float32(3.14), "3.14"},
		{"float64", float64(3.14159), "3.14"},
		{"bool true", true, "Yes"},
		{"bool false", false, "No"},
		{"[]byte", []byte("test"), "<binary data: 4 bytes>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.value)
			if result != tt.expected {
				t.Errorf("FormatValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatValue_Time(t *testing.T) {
	// Test time formatting
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result := FormatValue(now)
	expected := "2024-01-15 10:30:00"
	if result != expected {
		t.Errorf("FormatValue(time) = %q, want %q", result, expected)
	}
}

func TestFormatValue_Default(t *testing.T) {
	// Test default case with struct
	type testStruct struct {
		Name string
	}
	result := FormatValue(testStruct{Name: "test"})
	if result == "" {
		t.Error("FormatValue(struct) should return non-empty string")
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "-"},
		{"time", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), "2024-01-15"},
		{"string date", "2024-01-15", "2024-01-15"},
		{"invalid string", "not-a-date", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.value)
			if result != tt.expected {
				t.Errorf("FormatDate(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatDate_Default(t *testing.T) {
	result := FormatDate(12345)
	if result == "" {
		t.Error("FormatDate(int) should return non-empty string")
	}
}

func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "-"},
		{"time", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), "2024-01-15 10:30:00"},
		{"string RFC3339", "2024-01-15T10:30:00Z", "2024-01-15 10:30:00"},
		{"invalid string", "not-a-datetime", "not-a-datetime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(tt.value)
			if result != tt.expected {
				t.Errorf("FormatDateTime(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestFormatDateTime_Default(t *testing.T) {
	result := FormatDateTime(12345)
	if result == "" {
		t.Error("FormatDateTime(int) should return non-empty string")
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		currency string
		expected string
	}{
		{"nil", nil, "$", "-"},
		{"float64", 123.45, "$", "$123.45"},
		{"float32", float32(99.99), "$", "$99.99"},
		{"int", 100, "$", "$100.00"},
		{"int64", int64(1000), "$", "$1000.00"},
		{"empty currency", 100, "", "$100.00"},
		{"euro", 100, "EUR ", "EUR 100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCurrency(tt.value, tt.currency)
			if result != tt.expected {
				t.Errorf("FormatCurrency(%v, %q) = %q, want %q", tt.value, tt.currency, result, tt.expected)
			}
		})
	}
}

func TestFormatCurrency_Default(t *testing.T) {
	result := FormatCurrency("not-a-number", "$")
	if result == "" {
		t.Error("FormatCurrency(string, $) should return non-empty string")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"kilobytes partial", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"megabytes partial", 1572864, "1.5 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 5, "hello..."},
		{"empty string", "", 5, ""},
		{"zero maxLen", "hello", 0, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestSafeHTML(t *testing.T) {
	input := "<script>alert('xss')</script>"
	result := SafeHTML(input)
	// SafeHTML should return the string as template.HTML
	if string(result) != input {
		t.Errorf("SafeHTML(%q) = %q, want %q", input, string(result), input)
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		singular  string
		plural    string
		expected  string
	}{
		{"zero", 0, "item", "items", "items"},
		{"one", 1, "item", "items", "item"},
		{"two", 2, "item", "items", "items"},
		{"many", 100, "item", "items", "items"},
		{"empty plural", 2, "item", "", "items"},
		{"custom plural", 2, "person", "people", "people"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Pluralize(tt.count, tt.singular, tt.plural)
			if result != tt.expected {
				t.Errorf("Pluralize(%d, %q, %q) = %q, want %q", tt.count, tt.singular, tt.plural, result, tt.expected)
			}
		})
	}
}
