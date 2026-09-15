package server

import (
	"errors"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

// Finding 5: ignoreSyncError on multierr aggregate must not hide real sync errors.
func TestIgnoreSyncError(t *testing.T) {
	// A lone EINVAL returns nil
	assert.NoError(t, ignoreSyncError(syscall.EINVAL))

	// multierr containing EINVAL and real error: non-nil error containing real error
	combined := multierr.Append(syscall.EINVAL, errors.New("disk full"))
	err := ignoreSyncError(combined)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")

	// Nil error returns nil
	assert.NoError(t, ignoreSyncError(nil))
}
