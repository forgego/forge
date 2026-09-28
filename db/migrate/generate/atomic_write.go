package generate

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeMigrationPair creates both files exclusively. If either file already exists it
// returns an error wrapping os.ErrExist and writes nothing. If writing the second file
// fails, the first is removed.
func writeMigrationPair(upPath string, upContent []byte, downPath string, downContent []byte) (err error) {
	// Check if either file already exists before doing any work
	if _, err := os.Lstat(upPath); err == nil {
		return fmt.Errorf("%w: up migration already exists: %s", os.ErrExist, upPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if _, err := os.Lstat(downPath); err == nil {
		return fmt.Errorf("%w: down migration already exists: %s", os.ErrExist, downPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	upDir := filepath.Dir(upPath)
	downDir := filepath.Dir(downPath)

	// Write up migration to temp file in the same directory
	upTmp, err := writeTempFile(upDir, upContent)
	if err != nil {
		return err
	}
	defer os.Remove(upTmp)

	// Write down migration to temp file in the same directory
	downTmp, err := writeTempFile(downDir, downContent)
	if err != nil {
		return err
	}
	defer os.Remove(downTmp)

	// Before publishing the up file, check the down path does not exist either
	if _, err := os.Lstat(downPath); err == nil {
		return fmt.Errorf("%w: down migration already exists: %s", os.ErrExist, downPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	// Publish up file
	if err := os.Link(upTmp, upPath); err != nil {
		return err
	}

	// Track whether up file is linked so we can clean it up on subsequent failures
	upLinked := true
	defer func() {
		if upLinked && err != nil {
			_ = os.Remove(upPath)
		}
	}()

	// Publish down file
	if err = os.Link(downTmp, downPath); err != nil {
		return err
	}

	upLinked = false
	return nil
}

// writeTempFile writes content to a temporary file in dir with mode 0644, fsyncs, and closes it.
func writeTempFile(dir string, content []byte) (string, error) {
	tmpFile, err := os.CreateTemp(dir, ".migration-*.tmp")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()

	if err := tmpFile.Chmod(0644); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}

	if _, err := tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}
