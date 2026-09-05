package exporters

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
