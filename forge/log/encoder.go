package log

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// ConsoleEncoder is a custom encoder for development with colored, one-line output
type ConsoleEncoder struct {
	zapcore.Encoder
	colored    bool
	oneLine    bool
	caller     bool
	stacktrace bool
}

const traceZapLevel = zapcore.DebugLevel - 1

// NewConsoleEncoder creates a new console encoder
func NewConsoleEncoder(config DevelopmentConfig) *ConsoleEncoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	encoderConfig.EncodeLevel = TraceColorLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	baseEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	return &ConsoleEncoder{
		Encoder:    baseEncoder,
		colored:    config.Colored,
		oneLine:    config.OneLine,
		caller:     config.Caller,
		stacktrace: config.Stacktrace,
	}
}

// EncodeEntry encodes a log entry
func (e *ConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	if e.oneLine {
		return e.encodeOneLine(entry, fields)
	}
	return e.Encoder.EncodeEntry(entry, fields)
}

// encodeOneLine encodes a log entry as a single line
func (e *ConsoleEncoder) encodeOneLine(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := buffer.NewPool().Get()

	// Timestamp
	timestamp := entry.Time.Format("15:04:05.000")
	buf.AppendString(timestamp)
	buf.AppendString(" ")

	// Level with color
	levelStr := e.formatLevel(entry.Level)
	buf.AppendString(levelStr)
	buf.AppendString(" ")

	// Caller (file:line)
	if e.caller && entry.Caller.Defined {
		caller := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		// Shorten path
		if idx := strings.LastIndex(caller, "/"); idx >= 0 {
			caller = caller[idx+1:]
		}
		buf.AppendString(caller)
		buf.AppendString(" ")
	}

	// Message
	buf.AppendString(entry.Message)

	// Fields
	if len(fields) > 0 {
		buf.AppendString(" | ")
		for i, field := range fields {
			if i > 0 {
				buf.AppendString(" ")
			}
			buf.AppendString(field.Key)
			buf.AppendString("=")
			buf.AppendString(e.formatFieldValue(field))
		}
	}

	// Stack trace
	if e.stacktrace && entry.Stack != "" {
		buf.AppendString(" | stack=")
		// Truncate stack trace for one-line format
		stack := entry.Stack
		if len(stack) > 200 {
			stack = stack[:200] + "..."
		}
		buf.AppendString(strings.ReplaceAll(stack, "\n", " "))
	}

	buf.AppendString("\n")
	return buf, nil
}

// formatLevel formats the log level with optional color
func (e *ConsoleEncoder) formatLevel(level zapcore.Level) string {
	levelStr := traceLevelString(level)
	if e.colored {
		switch level {
		case traceZapLevel:
			return "\033[35m" + levelStr + "\033[0m" // Magenta
		case zapcore.DebugLevel:
			return "\033[36m" + levelStr + "\033[0m" // Cyan
		case zapcore.InfoLevel:
			return "\033[32m" + levelStr + "\033[0m" // Green
		case zapcore.WarnLevel:
			return "\033[33m" + levelStr + "\033[0m" // Yellow
		case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
			return "\033[31m" + levelStr + "\033[0m" // Red
		default:
			return levelStr
		}
	}
	return levelStr
}

// formatFieldValue formats a field value
func (e *ConsoleEncoder) formatFieldValue(field zapcore.Field) string {
	switch field.Type {
	case zapcore.StringType:
		return field.String
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return fmt.Sprintf("%d", field.Integer)
	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
		return fmt.Sprintf("%d", field.Integer)
	case zapcore.Float64Type:
		if field.Interface != nil {
			if f, ok := field.Interface.(float64); ok {
				return fmt.Sprintf("%f", f)
			}
		}
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.Float32Type:
		if field.Interface != nil {
			if f, ok := field.Interface.(float32); ok {
				return fmt.Sprintf("%f", f)
			}
		}
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.BoolType:
		return fmt.Sprintf("%t", field.Integer == 1)
	case zapcore.DurationType:
		if field.Interface != nil {
			if d, ok := field.Interface.(time.Duration); ok {
				return d.String()
			}
		}
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.TimeType:
		if field.Interface != nil {
			if t, ok := field.Interface.(time.Time); ok {
				return t.Format(time.RFC3339)
			}
		}
		return fmt.Sprintf("%v", field.Interface)
	case zapcore.ErrorType:
		if field.Interface != nil {
			if err, ok := field.Interface.(error); ok {
				return err.Error()
			}
		}
		return "nil"
	default:
		return fmt.Sprintf("%v", field.Interface)
	}
}

// ProductionEncoder is a custom encoder for production with structured JSON output
type ProductionEncoder struct {
	zapcore.Encoder
	caller     bool
	stacktrace bool
}

// NewProductionEncoder creates a new production encoder
func NewProductionEncoder(config ProductionConfig) *ProductionEncoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = TraceLowercaseLevelEncoder
	if config.Caller {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	baseEncoder := zapcore.NewJSONEncoder(encoderConfig)

	return &ProductionEncoder{
		Encoder:    baseEncoder,
		caller:     config.Caller,
		stacktrace: config.Stacktrace,
	}
}

// EncodeEntry encodes a log entry
func (e *ProductionEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Add caller if enabled
	if e.caller && entry.Caller.Defined {
		caller := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		fields = append(fields, zap.String("caller", caller))
	}

	// Add stack trace if enabled and present
	if e.stacktrace && entry.Stack != "" {
		fields = append(fields, zap.String("stack", entry.Stack))
	}

	return e.Encoder.EncodeEntry(entry, fields)
}

// TraceLevelEncoder adds TRACE level support
func TraceLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == traceZapLevel {
		enc.AppendString("TRACE")
		return
	}
	zapcore.CapitalLevelEncoder(level, enc)
}

// TraceLowercaseLevelEncoder adds trace support for lowercase level output.
func TraceLowercaseLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == traceZapLevel {
		enc.AppendString("trace")
		return
	}
	zapcore.LowercaseLevelEncoder(level, enc)
}

// TraceColorLevelEncoder adds trace support for colorized level output.
func TraceColorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if level == traceZapLevel {
		enc.AppendString("\033[35mTRACE\033[0m")
		return
	}
	zapcore.CapitalColorLevelEncoder(level, enc)
}

func traceLevelString(level zapcore.Level) string {
	if level == traceZapLevel {
		return "TRACE"
	}
	return level.CapitalString()
}

// getZapLevel converts our Level to zapcore.Level
func getZapLevel(level Level) zapcore.Level {
	switch level {
	case LevelTrace:
		return traceZapLevel // TRACE is below DEBUG
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getZapFormat converts our Format to zap encoder
func getZapFormat(format Format, config *LoggingConfig, development bool) zapcore.Encoder {
	if development {
		devConfig := config.Development
		if format == FormatConsole || format == "" {
			return NewConsoleEncoder(devConfig)
		}
		if format == FormatJSON {
			encoderConfig := zap.NewDevelopmentEncoderConfig()
			encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			encoderConfig.EncodeLevel = TraceLevelEncoder
			return zapcore.NewJSONEncoder(encoderConfig)
		}
		// Text format
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = TraceLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Production
	prodConfig := config.Production
	if format == FormatJSON || format == "" {
		return NewProductionEncoder(prodConfig)
	}
	// Text format in production
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = TraceLowercaseLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// addTraceLevel adds TRACE level support to zap
func addTraceLevel() {
	// TRACE level is already supported by using DebugLevel - 1
	// This is a no-op but documents the approach
	// TRACE = DebugLevel - 1
}
