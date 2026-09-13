package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectory ensures a directory exists, creating it if necessary
func EnsureDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetAbsolutePath returns the absolute path for a given path
func GetAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}
