package exporters

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestFileExporter_WritesToConfiguredPath(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	exp := NewFileExporter(FileConfig{
		Path:       logPath,
		MaxSize:    0,
		MaxBackups: 0,
	}, zapcore.InfoLevel)
	defer func() { _ = exp.Close() }()

	payload := []byte("hello exporter\n")
	n, err := exp.GetWriter().Write(payload)
	require.NoError(t, err)
	require.Equal(t, len(payload), n)
	require.NoError(t, exp.GetWriter().Sync())

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	require.Equal(t, string(payload), string(data))
}

func TestFileExporter_RotatesWhenMaxSizeExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "rotate.log")

	exp := NewFileExporter(FileConfig{
		Path:       logPath,
		MaxSize:    1, // MB
		MaxBackups: 2,
	}, zapcore.InfoLevel)
	defer func() { _ = exp.Close() }()

	chunkA := bytes.Repeat([]byte("A"), 700*1024)
	chunkB := bytes.Repeat([]byte("B"), 700*1024)

	_, err := exp.GetWriter().Write(chunkA)
	require.NoError(t, err)
	_, err = exp.GetWriter().Write(chunkB)
	require.NoError(t, err)
	require.NoError(t, exp.GetWriter().Sync())

	backupPath := logPath + ".1"
	backupInfo, err := os.Stat(backupPath)
	require.NoError(t, err)
	require.Greater(t, backupInfo.Size(), int64(0))

	currentData, err := os.ReadFile(logPath)
	require.NoError(t, err)
	require.NotEmpty(t, currentData)
	require.Equal(t, byte('B'), currentData[0])
}

type fakeWriteSyncer struct {
	syncErr error
}

func (f *fakeWriteSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (f *fakeWriteSyncer) Sync() error                 { return f.syncErr }

func TestFileExporter_Close_CallsCloserOnSyncError(t *testing.T) {
	closed := false
	exp := &FileExporter{
		writer: &fakeWriteSyncer{syncErr: errors.New("sync failed")},
		level:  zapcore.InfoLevel,
		closer: func() error {
			closed = true
			return nil
		},
	}

	err := exp.Close()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sync failed")
	assert.True(t, closed, "closer must be called even when Sync fails")
}

func TestFileExporter_Close_IgnoresEINVAL(t *testing.T) {
	closed := false
	exp := &FileExporter{
		writer: &fakeWriteSyncer{syncErr: syscall.EINVAL},
		level:  zapcore.InfoLevel,
		closer: func() error {
			closed = true
			return nil
		},
	}

	err := exp.Close()
	assert.NoError(t, err)
	assert.True(t, closed)
}

func TestFileExporter_Close_JoinsBothErrors(t *testing.T) {
	closed := false
	exp := &FileExporter{
		writer: &fakeWriteSyncer{syncErr: errors.New("sync failed")},
		level:  zapcore.InfoLevel,
		closer: func() error {
			closed = true
			return errors.New("closer failed")
		},
	}

	err := exp.Close()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sync failed")
	assert.Contains(t, err.Error(), "closer failed")
	assert.True(t, closed)
}
