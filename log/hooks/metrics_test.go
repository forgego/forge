package hooks

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestMetricsHook_ProcessCountsKnownLevels(t *testing.T) {
	hook := NewMetricsHook()

	levels := []zapcore.Level{
		zapcore.DebugLevel - 1,
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.DPanicLevel,
		zapcore.PanicLevel,
		zapcore.FatalLevel,
	}

	for _, level := range levels {
		_, _, ok := hook.Process(zapcore.Entry{Level: level}, nil)
		if !ok {
			t.Fatalf("Process returned shouldLog=false for level %v", level)
		}
	}

	snapshot := hook.Snapshot()
	if snapshot.Total != uint64(len(levels)) {
		t.Fatalf("expected total=%d, got %d", len(levels), snapshot.Total)
	}
	if snapshot.Unknown != 0 {
		t.Fatalf("expected unknown=0, got %d", snapshot.Unknown)
	}
	for _, level := range levels {
		if snapshot.ByLevel[level] != 1 {
			t.Fatalf("expected level %v count=1, got %d", level, snapshot.ByLevel[level])
		}
	}
}

func TestMetricsHook_ProcessCountsUnknownAndDropped(t *testing.T) {
	hook := NewMetricsHook()

	hook.Process(zapcore.Entry{Level: zapcore.Level(99)}, nil)
	hook.MarkDropped()
	hook.MarkDropped()

	snapshot := hook.Snapshot()
	if snapshot.Total != 1 {
		t.Fatalf("expected total=1, got %d", snapshot.Total)
	}
	if snapshot.Unknown != 1 {
		t.Fatalf("expected unknown=1, got %d", snapshot.Unknown)
	}
	if snapshot.Dropped != 2 {
		t.Fatalf("expected dropped=2, got %d", snapshot.Dropped)
	}
}
