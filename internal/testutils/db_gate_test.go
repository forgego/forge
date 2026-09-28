package testutils

import (
	"testing"
)

func TestDBGateContract(t *testing.T) {
	// Test testDatabaseURL branches
	t.Run("testDatabaseURL", func(t *testing.T) {
		const defaultURL = "postgres://default:5432/db"
		const customURL = "postgres://custom:5432/db"

		t.Run("uses custom URL when set", func(t *testing.T) {
			t.Setenv("FORGE_TEST_DATABASE_URL", customURL)
			if got := testDatabaseURL(defaultURL); got != customURL {
				t.Errorf("expected %q, got %q", customURL, got)
			}
		})

		t.Run("falls back to default URL when unset", func(t *testing.T) {
			t.Setenv("FORGE_TEST_DATABASE_URL", "")
			if got := testDatabaseURL(defaultURL); got != defaultURL {
				t.Errorf("expected %q, got %q", defaultURL, got)
			}
		})
	})

	// Test env parsing and pure decision functions
	t.Run("requireDB and dbUnavailableAction", func(t *testing.T) {
		t.Run("required when FORGE_REQUIRE_DB=1", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "1")
			if !requireDB() {
				t.Errorf("expected requireDB() to be true")
			}
			if got := dbUnavailableAction(); got != "fatal" {
				t.Errorf("expected dbUnavailableAction() = 'fatal', got %q", got)
			}
		})

		t.Run("not required when FORGE_REQUIRE_DB is unset", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "")
			if requireDB() {
				t.Errorf("expected requireDB() to be false")
			}
			if got := dbUnavailableAction(); got != "skip" {
				t.Errorf("expected dbUnavailableAction() = 'skip', got %q", got)
			}
		})

		t.Run("not required when FORGE_REQUIRE_DB is other value", func(t *testing.T) {
			t.Setenv("FORGE_REQUIRE_DB", "0")
			if requireDB() {
				t.Errorf("expected requireDB() to be false")
			}
			if got := dbUnavailableAction(); got != "skip" {
				t.Errorf("expected dbUnavailableAction() = 'skip', got %q", got)
			}
		})
	})
}
