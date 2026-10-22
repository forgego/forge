package migrationfiles

import (
	"fmt"

	"github.com/forgego/forge/internal/safepath"
)

// Read returns migration contents or an error that identifies the file.
func Read(root, file string) ([]byte, error) {
	contents, err := safepath.ReadFile(root, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration %s: %w", file, err)
	}
	return contents, nil
}
