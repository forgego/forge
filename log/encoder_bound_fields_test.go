package log

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Finding 4: One-line mode must include fields bound via logger.With in order, before per-call fields.
func TestConsoleEncoder_OneLine_IncludesBoundFields(t *testing.T) {
	enc := NewConsoleEncoder(DevelopmentConfig{
		Colored: false,
		OneLine: true,
	})

	buf := &bytes.Buffer{}
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)

	logger.With(zap.String("request_id", "abc")).Info("msg")

	output := buf.String()
	assert.Contains(t, output, "request_id=abc")
}

func TestConsoleEncoder_OneLine_FieldOrder(t *testing.T) {
	enc := NewConsoleEncoder(DevelopmentConfig{
		Colored: false,
		OneLine: true,
	})

	buf := &bytes.Buffer{}
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)

	// Bound fields must appear in order added, before per-call fields
	logger.With(zap.String("first", "1"), zap.String("second", "2")).Info("msg", zap.String("call", "3"))

	output := buf.String()
	assert.Contains(t, output, "first=1 second=2 call=3")
}

func TestConsoleEncoder_OneLine_FieldFormatting(t *testing.T) {
	enc := NewConsoleEncoder(DevelopmentConfig{
		Colored: false,
		OneLine: true,
	})

	buf := &bytes.Buffer{}
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)

	fixed := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	logger.With(
		zap.Duration("d", 2*time.Second),
		zap.Float64("f", 1.5),
		zap.Time("t", fixed),
	).Info("msg")

	output := buf.String()
	assert.Contains(t, output, "d=2s")
	assert.Contains(t, output, "f=1.5")
	assert.Contains(t, output, "t=2025-01-01T12:00:00Z")
	assert.NotContains(t, output, "<nil>")
	assert.NotContains(t, output, "Local")
}
