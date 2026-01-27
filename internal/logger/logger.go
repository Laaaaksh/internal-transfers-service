// Package logger provides structured logging for the application.
package logger

import (
	"context"
	"os"

	"github.com/internal-transfers-service/internal/constants"
	"github.com/internal-transfers-service/internal/constants/contextkeys"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L is the global logger instance
var L *zap.SugaredLogger

// Initialize creates and configures the global logger
func Initialize(level string, format string) error {
	cfg := createLogConfig(format)
	cfg.Level = parseLogLevel(level)
	configureEncoder(&cfg)
	configureOutputPaths(&cfg)

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	L = logger.Sugar()
	return nil
}

// createLogConfig returns the appropriate zap config based on format
func createLogConfig(format string) zap.Config {
	if format == constants.LogFormatJSON {
		return zap.NewProductionConfig()
	}
	return zap.NewDevelopmentConfig()
}

// parseLogLevel converts a string log level to zap atomic level
func parseLogLevel(level string) zap.AtomicLevel {
	switch level {
	case constants.LogLevelDebug:
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case constants.LogLevelInfo:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case constants.LogLevelWarn:
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case constants.LogLevelError:
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}

// configureEncoder sets up the encoder configuration
func configureEncoder(cfg *zap.Config) {
	cfg.EncoderConfig.TimeKey = constants.LogEncoderTimeKey
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = constants.LogEncoderMessageKey
	cfg.EncoderConfig.LevelKey = constants.LogEncoderLevelKey
	cfg.EncoderConfig.CallerKey = constants.LogEncoderCallerKey
}

// configureOutputPaths sets up the output paths
func configureOutputPaths(cfg *zap.Config) {
	cfg.OutputPaths = []string{constants.LogOutputStdout}
	cfg.ErrorOutputPaths = []string{constants.LogOutputStderr}
}

// Ctx returns a logger with request context fields
func Ctx(ctx context.Context) *zap.SugaredLogger {
	if L == nil {
		// Fallback to a basic logger if not initialized
		l, _ := zap.NewProduction()
		L = l.Sugar()
	}

	logger := L

	// Add request ID if present
	if requestID, ok := ctx.Value(contextkeys.RequestID).(string); ok {
		logger = logger.With(constants.LogKeyRequestID, requestID)
	}

	return logger
}

// WithField returns a logger with an additional field
func WithField(key string, value interface{}) *zap.SugaredLogger {
	if L == nil {
		l, _ := zap.NewProduction()
		L = l.Sugar()
	}
	return L.With(key, value)
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	if L == nil {
		l, _ := zap.NewProduction()
		L = l.Sugar()
	}

	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return L.With(args...)
}

// WithError returns a logger with an error field
func WithError(err error) *zap.SugaredLogger {
	if L == nil {
		l, _ := zap.NewProduction()
		L = l.Sugar()
	}
	return L.With(constants.LogKeyError, err)
}

// Info logs an info message
func Info(msg string, keysAndValues ...interface{}) {
	if L == nil {
		return
	}
	L.Infow(msg, keysAndValues...)
}

// Debug logs a debug message
func Debug(msg string, keysAndValues ...interface{}) {
	if L == nil {
		return
	}
	L.Debugw(msg, keysAndValues...)
}

// Warn logs a warning message
func Warn(msg string, keysAndValues ...interface{}) {
	if L == nil {
		return
	}
	L.Warnw(msg, keysAndValues...)
}

// Error logs an error message
func Error(msg string, keysAndValues ...interface{}) {
	if L == nil {
		return
	}
	L.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, keysAndValues ...interface{}) {
	if L == nil {
		os.Exit(1)
	}
	L.Fatalw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
