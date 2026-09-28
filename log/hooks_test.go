package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type secretFilterHook struct{}

func (h secretFilterHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	if strings.Contains(entry.Message, "secret") {
		return entry, fields, false
	}
	return entry, fields, true
}

func TestBuilder_Hooks(t *testing.T) {
	// 1. Builder with a filter hook that drops messages containing "secret"
	hookLogPath := filepath.Join(t.TempDir(), "hook.log")
	hookBuilder := NewBuilder().Production()
	hookBuilder.AddFileOutput(hookLogPath, LevelInfo, 10, 1, 1, false)
	hookBuilder.Hooks(secretFilterHook{})

	hookLogger, err := hookBuilder.Build()
	require.NoError(t, err)

	hookLogger.Info("this is a secret token")
	hookLogger.Info("this is a public event")
	_ = hookLogger.Sync()

	hookData, err := os.ReadFile(hookLogPath)
	require.NoError(t, err)
	assert.NotContains(t, string(hookData), "secret")
	assert.Contains(t, string(hookData), "public")

	// 2. Builder without hooks: output is unchanged (both entries written)
	noHookLogPath := filepath.Join(t.TempDir(), "nohook.log")
	noHookBuilder := NewBuilder().Production()
	noHookBuilder.AddFileOutput(noHookLogPath, LevelInfo, 10, 1, 1, false)

	noHookLogger, err := noHookBuilder.Build()
	require.NoError(t, err)

	noHookLogger.Info("this is a secret token")
	noHookLogger.Info("this is a public event")
	_ = noHookLogger.Sync()

	noHookData, err := os.ReadFile(noHookLogPath)
	require.NoError(t, err)
	assert.Contains(t, string(noHookData), "secret")
	assert.Contains(t, string(noHookData), "public")
}
