package db

import (
	"strings"
	"testing"
)

func TestNewDBFromConfig_NilConfig(t *testing.T) {
	database, err := NewDBFromConfig(nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
	if database != nil {
		t.Fatalf("expected nil database for nil config, got %#v", database)
	}
	if !strings.Contains(err.Error(), "config is nil") {
		t.Fatalf("expected nil-config error message, got %q", err.Error())
	}
}

