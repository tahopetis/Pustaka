package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		pustakaLogger.Default().Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Msg("HTTP request")
	})
}

type requestLogger struct {
	logger *pustakaLogger.Logger
}

func (l *requestLogger) Print(v interface{}) {
	// This method is required by the chi middleware interface
	// but we implement our own logging in NewLogEntry
}

func (l *requestLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &logEntry{
		request: r,
		logger:  l.logger,
	}

	// Add request ID to context if not present
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	ctx := context.WithValue(r.Context(), "request_id", requestID)
	*r = *r.WithContext(ctx)

	return entry
}

type logEntry struct {
	request *http.Request
	logger  *pustakaLogger.Logger
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	fields := map[string]interface{}{
		"status":     status,
		"bytes":      bytes,
		"duration":   elapsed.String(),
		"duration_ms": elapsed.Milliseconds(),
	}

	// Get request ID from context
	if requestID, ok := e.request.Context().Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}

	// Get user ID from context if available
	if userID, ok := e.request.Context().Value("user_id").(string); ok {
		fields["user_id"] = userID
	}

	// Log based on status code
	if status >= 500 {
		e.logger.ErrorRequest(
			e.request.Method,
			e.request.URL.Path,
			e.request.RemoteAddr,
			e.request.UserAgent(),
			getRequestID(e.request),
			nil, // No specific error for logging
			fields,
		)
	} else if status >= 400 {
		e.logger.ErrorRequest(
			e.request.Method,
			e.request.URL.Path,
			e.request.RemoteAddr,
			e.request.UserAgent(),
			getRequestID(e.request),
			nil, // No specific error for logging
			fields,
		)
	} else {
		e.logger.InfoRequest(
			e.request.Method,
			e.request.URL.Path,
			e.request.RemoteAddr,
			e.request.UserAgent(),
			getRequestID(e.request),
			fields,
		)
	}
}

func (e *logEntry) Panic(v interface{}, stack []byte) {
	e.logger.Error().
		Interface("panic", v).
		Bytes("stack", stack).
		Str("method", e.request.Method).
		Str("path", e.request.URL.Path).
		Msg("Request panic")
}

func getRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value("request_id").(string); ok {
		return requestID
	}
	return r.Header.Get("X-Request-ID")
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}