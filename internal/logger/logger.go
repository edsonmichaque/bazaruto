package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level, format string) *Logger {
	// Set log level
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "trace":
		zapLevel = zapcore.DebugLevel // zap doesn't have trace, use debug
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	case "panic":
		zapLevel = zapcore.PanicLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if strings.ToLower(format) == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	// Create encoder
	var encoder zapcore.Encoder
	if strings.ToLower(format) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapLevel)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{logger}
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.With(zap.Any(key, value))
	return &Logger{logger}
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	logger := l.With(zapFields...)
	return &Logger{logger}
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *Logger {
	logger := l.With(zap.Error(err))
	return &Logger{logger}
}

// WithString adds a string field to the logger context
func (l *Logger) WithString(key, value string) *Logger {
	logger := l.With(zap.String(key, value))
	return &Logger{logger}
}

// WithInt adds an int field to the logger context
func (l *Logger) WithInt(key string, value int) *Logger {
	logger := l.With(zap.Int(key, value))
	return &Logger{logger}
}

// WithInt64 adds an int64 field to the logger context
func (l *Logger) WithInt64(key string, value int64) *Logger {
	logger := l.With(zap.Int64(key, value))
	return &Logger{logger}
}

// WithFloat64 adds a float64 field to the logger context
func (l *Logger) WithFloat64(key string, value float64) *Logger {
	logger := l.With(zap.Float64(key, value))
	return &Logger{logger}
}

// WithBool adds a bool field to the logger context
func (l *Logger) WithBool(key string, value bool) *Logger {
	logger := l.With(zap.Bool(key, value))
	return &Logger{logger}
}

// WithTime adds a time field to the logger context
func (l *Logger) WithTime(key string, value time.Time) *Logger {
	logger := l.With(zap.Time(key, value))
	return &Logger{logger}
}

// WithDuration adds a duration field to the logger context
func (l *Logger) WithDuration(key string, value time.Duration) *Logger {
	logger := l.With(zap.Duration(key, value))
	return &Logger{logger}
}

// WithUUID adds a UUID field to the logger context
func (l *Logger) WithUUID(key string, value string) *Logger {
	logger := l.With(zap.String(key, value))
	return &Logger{logger}
}

// WithRequest adds HTTP request fields to the logger context
func (l *Logger) WithRequest(method, path, userAgent string) *Logger {
	logger := l.With(
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
	)
	return &Logger{logger}
}

// WithResponse adds HTTP response fields to the logger context
func (l *Logger) WithResponse(status int, size int64, duration time.Duration) *Logger {
	logger := l.With(
		zap.Int("status", status),
		zap.Int64("size", size),
		zap.Duration("duration", duration),
	)
	return &Logger{logger}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	zap.ReplaceGlobals(logger.Logger)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger and flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Logger.Sync()
}
