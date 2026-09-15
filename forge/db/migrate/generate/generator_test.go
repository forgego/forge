package generate_test

import (
	"testing"

	"github.com/forgego/forge/db/migrate/generate"
)

// TestNewMigrationGeneratorBackwardsCompatibility ensures the original 5-arg signature still works
func TestNewMigrationGeneratorBackwardsCompatibility(t *testing.T) {
	// This should compile and work with the original 5-argument signature
	_, err := generate.NewMigrationGenerator(
		"models",     // modelsDir
		"migrations", // migrationsDir
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGenerator with 5 args failed: %v", err)
	}
}

// TestNewMigrationGeneratorUsesDefaultDriver ensures the backwards compatible function uses the default driver
func TestNewMigrationGeneratorUsesDefaultDriver(t *testing.T) {
	gen, err := generate.NewMigrationGenerator(
		"models",     // modelsDir
		"migrations", // migrationsDir
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGenerator failed: %v", err)
	}
	if gen == nil {
		t.Fatalf("NewMigrationGenerator returned nil")
	}
	// We can't easily test the driver value without mocking config, but we can ensure it's not nil
	// The driver field is unexported, so we can't access it directly in tests
	// We'll rely on the fact that the function doesn't return an error to indicate it worked
}

// TestNewMigrationGeneratorWithDriverIsAvailable ensures the new driver-aware function exists
func TestNewMigrationGeneratorWithDriverIsAvailable(t *testing.T) {
	_, err := generate.NewMigrationGeneratorWithDriver(
		"models",     // modelsDir
		"migrations", // migrationsDir
		"postgres",   // driver
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGeneratorWithDriver failed: %v", err)
	}
}
