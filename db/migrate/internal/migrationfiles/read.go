package migrationfiles

import (
	"fmt"
	"os"
)

// Read returns migration contents or an error that identifies the file.
func Read(file string) ([]byte, error) {
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration %s: %w", file, err)
	}
	return contents, nil
}
