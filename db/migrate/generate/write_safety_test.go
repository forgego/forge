package generate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMigrationPair_Success(t *testing.T) {
	dir := t.TempDir()
	upPath := filepath.Join(dir, "000001_init.up.sql")
	downPath := filepath.Join(dir, "000001_init.down.sql")
	upContent := []byte("CREATE TABLE users (id int);")
	downContent := []byte("DROP TABLE users;")

	err := writeMigrationPair(upPath, upContent, downPath, downContent)
	require.NoError(t, err)

	// Verify exact contents
	gotUp, err := os.ReadFile(upPath)
	require.NoError(t, err)
	assert.Equal(t, upContent, gotUp)

	gotDown, err := os.ReadFile(downPath)
	require.NoError(t, err)
	assert.Equal(t, downContent, gotDown)

	// Verify file mode (0644)
	upInfo, err := os.Stat(upPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), upInfo.Mode().Perm())

	downInfo, err := os.Stat(downPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), downInfo.Mode().Perm())

	// Verify no .migration-*.tmp files remain
	tmpMatches, err := filepath.Glob(filepath.Join(dir, ".migration-*.tmp"))
	require.NoError(t, err)
	assert.Empty(t, tmpMatches)
}

func TestWriteMigrationPair_UpExists(t *testing.T) {
	dir := t.TempDir()
	upPath := filepath.Join(dir, "000001_init.up.sql")
	downPath := filepath.Join(dir, "000001_init.down.sql")
	existingUpContent := []byte("EXISTING UP")
	require.NoError(t, os.WriteFile(upPath, existingUpContent, 0644))

	err := writeMigrationPair(upPath, []byte("NEW UP"), downPath, []byte("NEW DOWN"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrExist), "error should wrap os.ErrExist, got: %v", err)

	// Existing up content unchanged
	gotUp, err := os.ReadFile(upPath)
	require.NoError(t, err)
	assert.Equal(t, existingUpContent, gotUp)

	// Down file NOT created
	_, err = os.Stat(downPath)
	assert.True(t, os.IsNotExist(err), "down file should not be created")

	// No temp files left
	tmpMatches, err := filepath.Glob(filepath.Join(dir, ".migration-*.tmp"))
	require.NoError(t, err)
	assert.Empty(t, tmpMatches)
}

func TestWriteMigrationPair_DownExists(t *testing.T) {
	dir := t.TempDir()
	upPath := filepath.Join(dir, "000001_init.up.sql")
	downPath := filepath.Join(dir, "000001_init.down.sql")
	existingDownContent := []byte("EXISTING DOWN")
	require.NoError(t, os.WriteFile(downPath, existingDownContent, 0644))

	err := writeMigrationPair(upPath, []byte("NEW UP"), downPath, []byte("NEW DOWN"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrExist), "error should wrap os.ErrExist, got: %v", err)

	// Up file NOT left behind
	_, err = os.Stat(upPath)
	assert.True(t, os.IsNotExist(err), "up file should not be left behind")

	// Existing down content unchanged
	gotDown, err := os.ReadFile(downPath)
	require.NoError(t, err)
	assert.Equal(t, existingDownContent, gotDown)

	// No temp files left
	tmpMatches, err := filepath.Glob(filepath.Join(dir, ".migration-*.tmp"))
	require.NoError(t, err)
	assert.Empty(t, tmpMatches)
}

func TestGetNextVersion_Overflow(t *testing.T) {
	t.Run("generator getNextVersion overflow", func(t *testing.T) {
		dir := t.TempDir()
		filename := filepath.Join(dir, "18446744073709551615_overflow.up.sql")
		require.NoError(t, os.WriteFile(filename, []byte("-- sql"), 0644))

		_, err := getNextVersion(dir)
		require.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "overflow")
	})

	t.Run("squasher getNextVersion overflow", func(t *testing.T) {
		dir := t.TempDir()
		filename := filepath.Join(dir, "18446744073709551615_overflow.up.sql")
		require.NoError(t, os.WriteFile(filename, []byte("-- sql"), 0644))

		s := NewSquasher(dir)
		_, err := s.getNextVersion()
		require.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "overflow")
	})
}

func TestGetNextVersion_SkipNonNumeric(t *testing.T) {
	t.Run("generator getNextVersion skips non-numeric", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "abc_x.up.sql"), []byte("-- sql"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "000007_y.up.sql"), []byte("-- sql"), 0644))

		ver, err := getNextVersion(dir)
		require.NoError(t, err)
		assert.Equal(t, "000008", ver)
	})

	t.Run("squasher getNextVersion skips non-numeric", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "abc_x.up.sql"), []byte("-- sql"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "000007_y.up.sql"), []byte("-- sql"), 0644))

		s := NewSquasher(dir)
		ver, err := s.getNextVersion()
		require.NoError(t, err)
		assert.Equal(t, "000008", ver)
	})
}
