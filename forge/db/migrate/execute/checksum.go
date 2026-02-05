package execute

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/forgego/forge/db/migrate/core"
)

// ChecksumValidator validates migration file checksums
type ChecksumValidator struct {
	migrationsDir string
}

// NewChecksumValidator creates a new checksum validator
func NewChecksumValidator(migrationsDir string) *ChecksumValidator {
	return &ChecksumValidator{
		migrationsDir: migrationsDir,
	}
}

// CalculateChecksum calculates SHA256 checksum of a migration file
func (v *ChecksumValidator) CalculateChecksum(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read migration file: %w", err)
	}

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), nil
}

// ValidateChecksum validates that a migration file's checksum matches the expected value
func (v *ChecksumValidator) ValidateChecksum(filePath, expectedChecksum string) error {
	actualChecksum, err := v.CalculateChecksum(filePath)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return core.NewMigrationError(
			core.ErrChecksumMismatch,
			fmt.Sprintf("checksum mismatch for %s: expected %s, got %s", filePath, expectedChecksum, actualChecksum),
			nil,
		)
	}

	return nil
}

// GetMigrationChecksum gets the checksum for a migration by name
func (v *ChecksumValidator) GetMigrationChecksum(migrationName string) (string, error) {
	upPath := filepath.Join(v.migrationsDir, fmt.Sprintf("%s.up.sql", migrationName))
	return v.CalculateChecksum(upPath)
}

// ValidateMigration validates a migration's checksum against stored value
// This would typically query the schema_migrations table for the stored checksum
func (v *ChecksumValidator) ValidateMigration(migrationName, storedChecksum string) error {
	if storedChecksum == "" {
		// No stored checksum, skip validation
		return nil
	}

	return v.ValidateChecksum(
		filepath.Join(v.migrationsDir, fmt.Sprintf("%s.up.sql", migrationName)),
		storedChecksum,
	)
}

// CalculateChecksum calculates SHA256 checksum of SQL content (standalone helper)
func CalculateChecksum(sql string) string {
	hash := sha256.Sum256([]byte(sql))
	return hex.EncodeToString(hash[:])
}

// ValidateChecksum validates that SQL content's checksum matches the expected value (standalone helper)
func ValidateChecksum(sql, expectedChecksum string) error {
	actualChecksum := CalculateChecksum(sql)
	if actualChecksum != expectedChecksum {
		return core.NewMigrationError(
			core.ErrChecksumMismatch,
			fmt.Sprintf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum),
			nil,
		)
	}
	return nil
}
