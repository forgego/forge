package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TempWorkdir creates a temporary working directory for tests
func TempWorkdir(t *testing.T, prefix string) (string, func()) {
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}

	return dir, cleanup
}

// TempDirInTests creates a temporary directory under tests/tmp
// Returns absolute path to avoid working directory issues with migration runner
func TempDirInTests(t *testing.T, prefix string) (absPath string, cleanup func()) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Determine tests directory - look for testhelpers package
	var testsDir string
	// Check if we're already in tests directory
	if _, err := os.Stat(filepath.Join(wd, "testhelpers")); err == nil {
		testsDir = wd
	} else if _, err := os.Stat(filepath.Join(wd, "..", "tests", "testhelpers")); err == nil {
		// We're in a subdirectory of tests (e.g., tests/pkg_migrations)
		testsDir = filepath.Join(wd, "..", "tests")
		// Resolve to absolute path
		testsDir, _ = filepath.Abs(testsDir)
	} else if _, err := os.Stat(filepath.Join(wd, "tests", "testhelpers")); err == nil {
		// We're in project root
		testsDir = filepath.Join(wd, "tests")
	} else {
		// Fallback: try to find tests directory by going up
		current := wd
		for i := 0; i < 5; i++ {
			testPath := filepath.Join(current, "tests", "testhelpers")
			if _, err := os.Stat(testPath); err == nil {
				testsDir = filepath.Join(current, "tests")
				break
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
		// Last resort: use current directory if it's named "tests"
		if testsDir == "" && filepath.Base(wd) == "tests" {
			testsDir = wd
		}
	}

	// Create tests/tmp if it doesn't exist
	tmpBase := filepath.Join(testsDir, "tmp")
	if err := os.MkdirAll(tmpBase, 0755); err != nil {
		t.Fatalf("failed to create tmp base dir: %v", err)
	}

	// Create temp directory
	absDir, err := ioutil.TempDir(tmpBase, prefix)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Always return absolute path to avoid working directory issues
	// The migration runner converts paths to absolute anyway, so this is safer
	absPath, err = filepath.Abs(absDir)
	if err != nil {
		// Fallback to the path we got from TempDir
		absPath = absDir
	}

	cleanup = func() {
		if err := os.RemoveAll(absDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}

	return absPath, cleanup
}

// CaptureGeneratedMigrations scans a directory for migration files
func CaptureGeneratedMigrations(dir string) ([]string, error) {
	var migrations []string
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Look for migration files (typically NNNN_name.up.sql or similar)
		name := entry.Name()
		if len(name) >= 5 && name[0] >= '0' && name[0] <= '9' {
			migrations = append(migrations, filepath.Join(dir, name))
		}
	}

	return migrations, nil
}

// ReadFileString reads a file as a string for testing
func ReadFileString(t *testing.T, path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}
	return string(content)
}

// WriteFileString writes a string to a file for testing
func WriteFileString(t *testing.T, path string, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
