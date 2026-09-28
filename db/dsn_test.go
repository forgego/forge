package db

import (
	"testing"
)

func TestDetectDriverFromDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected string
	}{
		{
			dsn:      "postgres://u:p@h:5432/d?sslmode=disable",
			expected: "postgres",
		},
		{
			dsn:      "postgresql://h/d",
			expected: "postgres",
		},
		{
			dsn:      "host=localhost port=5432 user=u dbname=d sslmode=disable",
			expected: "postgres",
		},
		{
			dsn:      "dbname=app.sqlite host=localhost sslmode=disable",
			expected: "postgres",
		},
		{
			dsn:      "file:app.sqlite?sslmode=disable",
			expected: "sqlite3",
		},
		{
			dsn:      ":memory:",
			expected: "sqlite3",
		},
		{
			dsn:      "file::memory:?cache=shared",
			expected: "sqlite3",
		},
		{
			dsn:      "/tmp/host=data.sqlite",
			expected: "sqlite3",
		},
		{
			dsn:      "./data/app.db",
			expected: "sqlite3",
		},
		{
			dsn:      "app.SQLITE3",
			expected: "sqlite3",
		},
		{
			dsn:      "app.db?_foreign_keys=on",
			expected: "sqlite3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			actual := DetectDriverFromDSN(tt.dsn)
			if actual != tt.expected {
				t.Errorf("DetectDriverFromDSN(%q) = %q; want %q", tt.dsn, actual, tt.expected)
			}
		})
	}
}
