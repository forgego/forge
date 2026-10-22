// Package safepath reads files beneath a trusted directory.
package safepath

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads path only when it is contained in root, including after
// following symlinks. Absolute and relative paths are both accepted.
func ReadFile(root, path string) ([]byte, error) {
	root, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil, err
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %q is outside %q", path, root)
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	file, err := dir.Open(rel)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
