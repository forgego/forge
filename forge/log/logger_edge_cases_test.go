package log

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestHookLevelChangeReroutes(t *testing.T) {
	for _, target := range []zapcore.Level{zapcore.ErrorLevel, zapcore.DebugLevel} {
		t.Run(target.String(), func(t *testing.T) {
			infoCore, infos := observer.New(zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l == zapcore.InfoLevel }))
			errorCore, errs := observer.New(zapcore.ErrorLevel)
			registry := NewHookRegistry()
			registry.AddHook(fnHook{fn: func(e zapcore.Entry, f []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
				e.Level = target
				return e, f, true
			}})
			zap.New(NewHookCore(zapcore.NewTee(infoCore, errorCore), registry)).Info("promote")
			assert.Zero(t, infos.Len())
			if target == zapcore.ErrorLevel {
				assert.Equal(t, 1, errs.Len())
			} else {
				assert.Zero(t, errs.Len())
			}
		})
	}
}

func TestHookRemovedFieldsStayRemoved(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	registry := NewHookRegistry()
	registry.AddHook(fnHook{fn: func(e zapcore.Entry, _ []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) { return e, nil, true }})
	zap.New(NewHookCore(core, registry)).With(zap.String("user", "alice")).Info("redact", zap.String("secret", "s3cr3t"))
	require.Equal(t, 1, logs.Len())
	assert.Equal(t, map[string]interface{}{"user": "alice"}, logs.All()[0].ContextMap())
}

func TestConsoleBoundMutableFieldsSnapshot(t *testing.T) {
	for _, kind := range []string{"object", "array", "reflect", "binary", "bytes"} {
		t.Run(kind, func(t *testing.T) {
			value := []byte("old")
			mapping := map[string]string{"value": "old"}
			var field zap.Field
			switch kind {
			case "object":
				field = zap.Object("data", zapcore.ObjectMarshalerFunc(func(e zapcore.ObjectEncoder) error { e.AddString("value", string(value)); return nil }))
			case "array":
				field = zap.Array("data", zapcore.ArrayMarshalerFunc(func(e zapcore.ArrayEncoder) error { e.AppendString(string(value)); return nil }))
			case "reflect":
				field = zap.Reflect("data", mapping)
			case "binary":
				field = zap.Binary("data", value)
			case "bytes":
				field = zap.ByteString("data", value)
			}
			var out bytes.Buffer
			logger := zap.New(zapcore.NewCore(NewConsoleEncoder(DevelopmentConfig{OneLine: true}), zapcore.AddSync(&out), zapcore.InfoLevel)).With(field)
			logger.Info("snapshot")
			before := strings.SplitN(out.String(), " | ", 2)[1]
			out.Reset()
			copy(value, "new")
			mapping["value"] = "new"
			logger.Info("snapshot")
			assert.Equal(t, before, strings.SplitN(out.String(), " | ", 2)[1])
		})
	}
}

func TestProductionCallerOccursOnce(t *testing.T) {
	enc := NewProductionEncoder(ProductionConfig{Caller: true})
	buf, err := enc.EncodeEntry(zapcore.Entry{Caller: zapcore.EntryCaller{Defined: true, File: "/tmp/example.go", Line: 42}}, nil)
	require.NoError(t, err)
	defer buf.Free()
	assert.Equal(t, 1, strings.Count(buf.String(), "\"caller\":"))
}

func TestConsoleNamespaceSurvivesClone(t *testing.T) {
	var out bytes.Buffer
	logger := zap.New(zapcore.NewCore(NewConsoleEncoder(DevelopmentConfig{OneLine: true}), zapcore.AddSync(&out), zapcore.InfoLevel))
	parent := logger.With(zap.Namespace("req"), zap.String("id", "1"))
	child := parent.With(zap.Namespace("nested"), zap.Int("id", 2))
	child.Info("child", zap.String("call", "3"))
	parent.Info("parent", zap.String("call", "4"))
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Contains(t, lines[0], "req.id=1")
	assert.Contains(t, lines[0], "req.nested.id=2")
	assert.Contains(t, lines[0], "req.nested.call=3")
	assert.Contains(t, lines[1], "req.call=4")
	assert.NotContains(t, lines[1], "nested")
}

type reviewCloser struct {
	err   error
	calls int
}

func (c *reviewCloser) Close() error { c.calls++; return c.err }

type reviewSyncCore struct {
	zapcore.Core
	err error
}

func (c reviewSyncCore) Sync() error { return c.err }

func TestCloseCombinesAllErrors(t *testing.T) {
	for _, syncErr := range []error{nil, errors.New("sync failed"), syscall.EINVAL} {
		t.Run(strings.ReplaceAll(fmtError(syncErr), " ", "_"), func(t *testing.T) {
			first := &reviewCloser{err: errors.New("first")}
			second := &reviewCloser{err: errors.New("second")}
			logger := &Logger{Logger: zap.New(reviewSyncCore{Core: zapcore.NewNopCore(), err: syncErr}), resources: &loggerResources{closers: []io.Closer{first, second}}}
			err := logger.Close()
			assert.ErrorIs(t, err, first.err)
			assert.ErrorIs(t, err, second.err)
			want := 2
			if syncErr != nil && !errors.Is(syncErr, syscall.EINVAL) {
				want++
				assert.ErrorIs(t, err, syncErr)
			}
			assert.Len(t, multierr.Errors(err), want)
			assert.Equal(t, err, logger.Close())
			assert.Equal(t, 1, first.calls)
			assert.Equal(t, 1, second.calls)
		})
	}
}
func fmtError(err error) string {
	if err == nil {
		return "success"
	}
	return err.Error()
}

func TestBuilderHooksConfigRoundTripOnce(t *testing.T) {
	calls := 0
	hook := fnHook{fn: func(e zapcore.Entry, f []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
		calls++
		return e, f, true
	}}
	logger, err := NewBuilder().Development().Hooks(hook).Build()
	require.NoError(t, err)
	defer logger.Close()
	logger.Info("first")
	assert.Equal(t, 1, calls)
	assert.Len(t, logger.GetConfig().Hooks, 1)
	rebuilt, err := NewLoggerFromConfig(logger.GetConfig())
	require.NoError(t, err)
	defer rebuilt.Close()
	rebuilt.Info("second")
	assert.Equal(t, 2, calls)
}
