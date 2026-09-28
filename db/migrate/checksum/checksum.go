package checksum

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrUpFileNotFound identifies a version without an up migration.
var ErrUpFileNotFound = errors.New("up migration file not found")

// FilePair holds migration paths matched by numeric version.
type FilePair struct{ Up, Down string }

// FileIndex is a snapshot of migration paths, indexed once per run.
type FileIndex map[uint]FilePair

// IndexMigrationFiles matches up and down files by numeric version.
func IndexMigrationFiles(path string) (FileIndex, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	index := make(FileIndex)
	for _, entry := range entries {
		name := entry.Name()
		prefix, _, ok := strings.Cut(name, "_")
		n, err := strconv.ParseUint(prefix, 10, strconv.IntSize)
		if entry.IsDir() || !ok || err != nil {
			continue
		}
		pair := index[uint(n)]
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			pair.Up = filepath.Join(path, name)
		case strings.HasSuffix(name, ".down.sql"):
			pair.Down = filepath.Join(path, name)
		default:
			continue
		}
		index[uint(n)] = pair
	}
	return index, nil
}

// Checksums hashes the indexed files without rescanning the directory.
func (index FileIndex) Checksums(version uint) (string, *string, error) {
	pair := index[version]
	if pair.Up == "" {
		return "", nil, fmt.Errorf("up migration file for version %d not found: %w", version, ErrUpFileNotFound)
	}
	up, err := os.ReadFile(pair.Up)
	if err != nil {
		return "", nil, fmt.Errorf("read up migration version %d: %w", version, err)
	}
	upHash := fmt.Sprintf("%x", sha256.Sum256(up))
	if pair.Down == "" {
		return upHash, nil, nil
	}
	down, err := os.ReadFile(pair.Down)
	if err != nil {
		return "", nil, fmt.Errorf("read down migration version %d: %w", version, err)
	}
	downHash := fmt.Sprintf("%x", sha256.Sum256(down))
	return upHash, &downHash, nil
}

// MigrationFileChecksums computes hashes for one migration version.
func MigrationFileChecksums(path string, version uint) (string, *string, error) {
	index, err := IndexMigrationFiles(path)
	if err != nil {
		return "", nil, err
	}
	return index.Checksums(version)
}
