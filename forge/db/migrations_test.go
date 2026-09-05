package db

import "testing"

func TestNewMigrationRunner_NilDB(t *testing.T) {
	runner, err := NewMigrationRunner(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil database, got nil")
	}
	if runner != nil {
		t.Fatalf("expected nil runner for nil database, got %#v", runner)
	}
	if err.Error() != "database connection is nil" {
		t.Fatalf("expected nil-database error, got %q", err.Error())
	}
}

func TestFallbackDetailedMigrationStatus_Dirty(t *testing.T) {
	result := fallbackDetailedMigrationStatus(&MigrationStatus{
		Version: 7,
		Dirty:   true,
	})

	if result.Current != "7" {
		t.Fatalf("expected current version 7, got %q", result.Current)
	}
	if result.Status != "DIRTY" {
		t.Fatalf("expected status DIRTY, got %q", result.Status)
	}
	if !result.Dirty {
		t.Fatal("expected dirty=true")
	}
}

func TestFallbackDetailedMigrationStatus_Clean(t *testing.T) {
	result := fallbackDetailedMigrationStatus(&MigrationStatus{
		Version: 12,
		Dirty:   false,
	})

	if result.Current != "12" {
		t.Fatalf("expected current version 12, got %q", result.Current)
	}
	if result.Status != "OK" {
		t.Fatalf("expected status OK, got %q", result.Status)
	}
	if result.Dirty {
		t.Fatal("expected dirty=false")
	}
}
