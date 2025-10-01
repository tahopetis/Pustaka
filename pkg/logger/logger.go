package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	*zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return file[strings.LastIndex(file, "/")+1:] + ":" + fmt.Sprintf("%d", line)
	}

	var output io.Writer = os.Stdout

	// Configure output format
	switch strings.ToLower(cfg.Format) {
	case "console":
		output = &zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	case "json":
		output = os.Stdout
	default:
		output = os.Stdout
	}

	// Create logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: &logger}
}

// Default returns the default global logger
func Default() *Logger {
	return &Logger{Logger: &log.Logger}
}

// InfoRequest logs an HTTP request with info level
func (l *Logger) InfoRequest(method, path, ip, userAgent, requestID string, fields map[string]interface{}) {
	event := l.Info().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("request_id", requestID)

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("HTTP request")
}

// ErrorRequest logs an HTTP request with error level
func (l *Logger) ErrorRequest(method, path, ip, userAgent, requestID string, err error, fields map[string]interface{}) {
	event := l.Error().
		Err(err).
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Str("user_agent", userAgent).
		Str("request_id", requestID)

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("HTTP request failed")
}

// InfoAuth logs authentication events
func (l *Logger) InfoAuth(action, userID, username, ip string, fields map[string]interface{}) {
	event := l.Info().
		Str("action", action).
		Str("user_id", userID).
		Str("username", username).
		Str("ip", ip).
		Str("event_type", "auth")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Authentication event")
}

// ErrorAuth logs authentication errors
func (l *Logger) ErrorAuth(action, userID, username, ip string, err error, fields map[string]interface{}) {
	event := l.Error().
		Err(err).
		Str("action", action).
		Str("user_id", userID).
		Str("username", username).
		Str("ip", ip).
		Str("event_type", "auth")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Authentication failed")
}

// InfoAudit logs audit events
func (l *Logger) InfoAudit(entityType, entityID, action, userID string, fields map[string]interface{}) {
	event := l.Info().
		Str("entity_type", entityType).
		Str("entity_id", entityID).
		Str("action", action).
		Str("user_id", userID).
		Str("event_type", "audit")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Audit event")
}

// InfoDatabase logs database operations
func (l *Logger) InfoDatabase(operation, table string, duration int64, fields map[string]interface{}) {
	event := l.Info().
		Str("operation", operation).
		Str("table", table).
		Int64("duration_ms", duration).
		Str("event_type", "database")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Database operation")
}

// ErrorDatabase logs database errors
func (l *Logger) ErrorDatabase(operation, table string, err error, fields map[string]interface{}) {
	event := l.Error().
		Err(err).
		Str("operation", operation).
		Str("table", table).
		Str("event_type", "database")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Database operation failed")
}

// InfoService logs service events
func (l *Logger) InfoService(service, operation string, fields map[string]interface{}) {
	event := l.Info().
		Str("service", service).
		Str("operation", operation).
		Str("event_type", "service")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Service event")
}

// ErrorService logs service errors
func (l *Logger) ErrorService(service, operation string, err error, fields map[string]interface{}) {
	event := l.Error().
		Err(err).
		Str("service", service).
		Str("operation", operation).
		Str("event_type", "service")

	for k, v := range fields {
		event = event.Interface(k, v)
	}

	event.Msg("Service operation failed")
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	ctx := l.With().Interface(key, value)
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

// WithUserID adds user ID to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	ctx := l.With().Str("user_id", userID)
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

// WithRequestID adds request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	ctx := l.With().Str("request_id", requestID)
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}

// WithCorrelationID adds correlation ID to the logger
func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	ctx := l.With().Str("correlation_id", correlationID)
	logger := ctx.Logger()
	return &Logger{Logger: &logger}
}