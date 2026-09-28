package generate

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	forgeerrors "github.com/forgego/forge/errors"
)

// Squasher squashes multiple migrations into a single migration
type Squasher struct {
	migrationsDir string
}

// NewSquasher creates a new migration squasher
func NewSquasher(migrationsDir string) *Squasher {
	return &Squasher{
		migrationsDir: migrationsDir,
	}
}

// SquashMigrations squashes migrations from startVersion to endVersion into a single migration
func (s *Squasher) SquashMigrations(startVersion, endVersion, newName string) error {
	return forgeerrors.NewNotImplementedError("migrations squash")
}

// getNextVersion gets the next migration version
func (s *Squasher) getNextVersion() (string, error) {
	pattern := filepath.Join(s.migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	maxVersion := uint64(0)
	hasVersions := false
	for _, match := range matches {
		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) < 1 {
			continue
		}

		versionStr := parts[0]
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			continue
		}
		if !hasVersions || version > maxVersion {
			maxVersion = version
			hasVersions = true
		}
	}

	if !hasVersions {
		return "000001", nil
	}

	if maxVersion == math.MaxUint64 {
		return "", fmt.Errorf("migration version overflow")
	}

	nextVersion := maxVersion + 1
	return fmt.Sprintf("%06d", nextVersion), nil
}
