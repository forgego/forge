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
	colored     bool
	oneLine     bool
	caller      bool
	stacktrace  bool
	addedFields []zapcore.Field
	namespaces  []string
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

// Clone creates a copy of the console encoder
func (e *ConsoleEncoder) Clone() zapcore.Encoder {
	addedFields := make([]zapcore.Field, len(e.addedFields))
	copy(addedFields, e.addedFields)
	return &ConsoleEncoder{
		Encoder:     e.Encoder.Clone(),
		colored:     e.colored,
		oneLine:     e.oneLine,
		caller:      e.caller,
		stacktrace:  e.stacktrace,
		addedFields: addedFields,
		namespaces:  append([]string(nil), e.namespaces...),
	}
}

// addField snapshots the rendered value while binding, before mutable values
// can change. The embedded encoder maintains its own namespace state.
func (e *ConsoleEncoder) addField(field zapcore.Field) {
	if !e.oneLine {
		return
	}
	key := field.Key
	if len(e.namespaces) > 0 {
		key = strings.Join(e.namespaces, ".") + "." + key
	}
	e.addedFields = append(e.addedFields, zap.String(key, e.formatFieldValue(field)))
}

func (e *ConsoleEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	e.addField(zap.Array(key, marshaler))
	return e.Encoder.AddArray(key, marshaler)
}

func (e *ConsoleEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	e.addField(zap.Object(key, marshaler))
	return e.Encoder.AddObject(key, marshaler)
}

func (e *ConsoleEncoder) AddBinary(key string, value []byte) {
	e.addField(zap.Binary(key, value))
	e.Encoder.AddBinary(key, value)
}

func (e *ConsoleEncoder) AddByteString(key string, value []byte) {
	e.addField(zap.ByteString(key, value))
	e.Encoder.AddByteString(key, value)
}

func (e *ConsoleEncoder) AddBool(key string, value bool) {
	e.addField(zap.Bool(key, value))
	e.Encoder.AddBool(key, value)
}

func (e *ConsoleEncoder) AddComplex128(key string, value complex128) {
	e.addField(zap.Complex128(key, value))
	e.Encoder.AddComplex128(key, value)
}

func (e *ConsoleEncoder) AddComplex64(key string, value complex64) {
	e.addField(zap.Complex64(key, value))
	e.Encoder.AddComplex64(key, value)
}

func (e *ConsoleEncoder) AddDuration(key string, value time.Duration) {
	e.addField(zap.Duration(key, value))
	e.Encoder.AddDuration(key, value)
}

func (e *ConsoleEncoder) AddFloat64(key string, value float64) {
	e.addField(zap.Float64(key, value))
	e.Encoder.AddFloat64(key, value)
}

func (e *ConsoleEncoder) AddFloat32(key string, value float32) {
	e.addField(zap.Float32(key, value))
	e.Encoder.AddFloat32(key, value)
}

func (e *ConsoleEncoder) AddInt(key string, value int) {
	e.addField(zap.Int(key, value))
	e.Encoder.AddInt(key, value)
}

func (e *ConsoleEncoder) AddInt64(key string, value int64) {
	e.addField(zap.Int64(key, value))
	e.Encoder.AddInt64(key, value)
}

func (e *ConsoleEncoder) AddInt32(key string, value int32) {
	e.addField(zap.Int32(key, value))
	e.Encoder.AddInt32(key, value)
}

func (e *ConsoleEncoder) AddInt16(key string, value int16) {
	e.addField(zap.Int16(key, value))
	e.Encoder.AddInt16(key, value)
}

func (e *ConsoleEncoder) AddInt8(key string, value int8) {
	e.addField(zap.Int8(key, value))
	e.Encoder.AddInt8(key, value)
}

func (e *ConsoleEncoder) AddString(key, value string) {
	e.addField(zap.String(key, value))
	e.Encoder.AddString(key, value)
}

func (e *ConsoleEncoder) AddTime(key string, value time.Time) {
	e.addField(zap.Time(key, value))
	e.Encoder.AddTime(key, value)
}

func (e *ConsoleEncoder) AddUint(key string, value uint) {
	e.addField(zap.Uint(key, value))
	e.Encoder.AddUint(key, value)
}

func (e *ConsoleEncoder) AddUint64(key string, value uint64) {
	e.addField(zap.Uint64(key, value))
	e.Encoder.AddUint64(key, value)
}

func (e *ConsoleEncoder) AddUint32(key string, value uint32) {
	e.addField(zap.Uint32(key, value))
	e.Encoder.AddUint32(key, value)
}

func (e *ConsoleEncoder) AddUint16(key string, value uint16) {
	e.addField(zap.Uint16(key, value))
	e.Encoder.AddUint16(key, value)
}

func (e *ConsoleEncoder) AddUint8(key string, value uint8) {
	e.addField(zap.Uint8(key, value))
	e.Encoder.AddUint8(key, value)
}

func (e *ConsoleEncoder) AddUintptr(key string, value uintptr) {
	e.addField(zap.Uintptr(key, value))
	e.Encoder.AddUintptr(key, value)
}

func (e *ConsoleEncoder) AddReflected(key string, value interface{}) error {
	e.addField(zap.Reflect(key, value))
	return e.Encoder.AddReflected(key, value)
}

func (e *ConsoleEncoder) OpenNamespace(key string) {
	e.namespaces = append(e.namespaces, key)
	e.Encoder.OpenNamespace(key)
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

	// Bound fields are already rendered and qualified. Call fields inherit the
	// bound namespace; namespaces opened by this entry must not affect siblings.
	allFields := make([]zapcore.Field, 0, len(e.addedFields)+len(fields))
	allFields = append(allFields, e.addedFields...)
	namespaces := append([]string(nil), e.namespaces...)
	for _, field := range fields {
		if field.Type == zapcore.NamespaceType {
			namespaces = append(namespaces, field.Key)
			continue
		}
		key := field.Key
		if len(namespaces) > 0 {
			key = strings.Join(namespaces, ".") + "." + key
		}
		allFields = append(allFields, zap.String(key, e.formatFieldValue(field)))
	}

	if len(allFields) > 0 {
		buf.AppendString(" | ")
		for i, field := range allFields {
			if i > 0 {
				buf.AppendString(" ")
			}
			buf.AppendString(field.Key)
			buf.AppendString("=")
			buf.AppendString(field.String)
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
	if field.Type == zapcore.ErrorType {
		if err, ok := field.Interface.(error); ok && err != nil {
			return err.Error()
		}
		return "nil"
	}
	enc := zapcore.NewMapObjectEncoder()
	field.AddTo(enc)
	val, ok := enc.Fields[field.Key]
	if !ok {
		if field.Interface == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%v", field.Interface)
	}
	switch v := val.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case time.Duration:
		return v.String()
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case nil:
		return "<nil>"
	default:
		return fmt.Sprintf("%v", v)
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

// Clone creates a copy of the production encoder
func (e *ProductionEncoder) Clone() zapcore.Encoder {
	return &ProductionEncoder{
		Encoder:    e.Encoder.Clone(),
		caller:     e.caller,
		stacktrace: e.stacktrace,
	}
}

// EncodeEntry encodes a log entry
func (e *ProductionEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
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
