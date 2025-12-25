package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type contextKey struct{}

var (
	logger zerolog.Logger
)

// Init initializes the global logger
func Init() {
	// Set log level from environment or default to info
	level := zerolog.InfoLevel
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = zerolog.DebugLevel
	}

	// Create console writer for development
	// For production, this should be JSON format
	output := zerolog.ConsoleWriter{Out: os.Stdout}

	// Initialize logger
	l := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	logger = l
}

// WithRequestID adds request ID to logger context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(contextKey{}).(string); ok {
		return requestID
	}
	return ""
}

// Info logs info message
func Info(msg string) {
	logger.Info().Msg(msg)
}

// Infof logs formatted info message
func Infof(format string, v ...interface{}) {
	logger.Info().Msg(fmt.Sprintf(format, v...))
}

// Error logs error message
func Error(msg string) {
	logger.Error().Msg(msg)
}

// Errorf logs formatted error message
func Errorf(format string, v ...interface{}) {
	logger.Error().Msg(fmt.Sprintf(format, v...))
}

// ErrorWithErr logs error with error object
func ErrorWithErr(err error, msg string) {
	if err != nil {
		logger.Error().Err(err).Msg(msg)
	} else {
		logger.Error().Msg(msg)
	}
}

// Debug logs debug message
func Debug(msg string) {
	logger.Debug().Msg(msg)
}

// Debugf logs formatted debug message
func Debugf(format string, v ...interface{}) {
	logger.Debug().Msg(fmt.Sprintf(format, v...))
}

// Warn logs warning message
func Warn(msg string) {
	logger.Warn().Msg(msg)
}

// Warnf logs formatted warning message
func Warnf(format string, v ...interface{}) {
	logger.Warn().Msg(fmt.Sprintf(format, v...))
}

// Fatal logs fatal message and exits
func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

// FatalWithErr logs fatal error and exits
func FatalWithErr(err error, msg string) {
	if err != nil {
		logger.Fatal().Err(err).Msg(msg)
	} else {
		logger.Fatal().Msg(msg)
	}
}

// WithCtx returns logger with context fields
func WithCtx(ctx context.Context) zerolog.Logger {
	l := logger

	if requestID := GetRequestID(ctx); requestID != "" {
		l = l.With().Str("request_id", requestID).Logger()
	}

	return l
}
