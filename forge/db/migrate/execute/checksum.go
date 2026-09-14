package execute

import (
	"github.com/forgego/forge/db/migrate/verify"
)

// ChecksumValidator validates migration file checksums.
//
// Deprecated: use verify.ChecksumValidator.
type ChecksumValidator = verify.ChecksumValidator

// NewChecksumValidator creates a new checksum validator.
//
// Deprecated: use verify.NewChecksumValidator.
func NewChecksumValidator(migrationsDir string) *ChecksumValidator {
	return verify.NewChecksumValidator(migrationsDir)
}

// CalculateChecksum calculates SHA256 checksum of SQL content (standalone helper).
//
// Deprecated: use verify.CalculateChecksum.
func CalculateChecksum(sql string) string {
	return verify.CalculateChecksum(sql)
}

// ValidateChecksum validates that SQL content's checksum matches the expected value (standalone helper).
//
// Deprecated: use verify.ValidateChecksum.
func ValidateChecksum(sql, expectedChecksum string) error {
	return verify.ValidateChecksum(sql, expectedChecksum)
}
