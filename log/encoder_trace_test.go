package log

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestTraceLevelEncoder_EncodesTrace(t *testing.T) {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:    "level",
		TimeKey:     "time",
		MessageKey:  "msg",
		EncodeLevel: TraceLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})

	buf, err := encoder.EncodeEntry(zapcore.Entry{
		Level:   traceZapLevel,
		Time:    time.Unix(0, 0),
		Message: "trace-test",
	}, nil)
	if err != nil {
		t.Fatalf("EncodeEntry failed: %v", err)
	}
	defer buf.Free()

	if !strings.Contains(buf.String(), `"level":"TRACE"`) {
		t.Fatalf("expected TRACE level, got %q", buf.String())
	}
}

func TestTraceLowercaseLevelEncoder_EncodesTrace(t *testing.T) {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		LevelKey:    "level",
		TimeKey:     "time",
		MessageKey:  "msg",
		EncodeLevel: TraceLowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})

	buf, err := encoder.EncodeEntry(zapcore.Entry{
		Level:   traceZapLevel,
		Time:    time.Unix(0, 0),
		Message: "trace-test",
	}, nil)
	if err != nil {
		t.Fatalf("EncodeEntry failed: %v", err)
	}
	defer buf.Free()

	if !strings.Contains(buf.String(), `"level":"trace"`) {
		t.Fatalf("expected lowercase trace level, got %q", buf.String())
	}
}

func TestConsoleEncoder_OneLineTraceLevel(t *testing.T) {
	encoder := NewConsoleEncoder(DevelopmentConfig{
		Colored: false,
		OneLine: true,
		Caller:  false,
	})

	buf, err := encoder.EncodeEntry(zapcore.Entry{
		Level:   traceZapLevel,
		Time:    time.Unix(0, 0),
		Message: "trace-test",
	}, nil)
	if err != nil {
		t.Fatalf("EncodeEntry failed: %v", err)
	}
	defer buf.Free()

	if !strings.Contains(buf.String(), " TRACE trace-test") {
		t.Fatalf("expected one-line TRACE level, got %q", buf.String())
	}
}
