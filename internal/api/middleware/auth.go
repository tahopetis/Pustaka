package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pustaka/pustaka/internal/auth"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	UserContextKey contextKey = "user"
)

// JWTAuth creates a JWT authentication middleware
func JWTAuth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health check and OPTIONS requests
			if r.URL.Path == "/health" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			token, err := auth.ExtractTokenFromHeader(authHeader)
			if err != nil {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
			ctx = context.WithValue(ctx, UserContextKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RBAC creates a role-based access control middleware
func RBAC(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value(UserContextKey).(*auth.JWTClaims)
			if !ok {
				http.Error(w, "User not found in context", http.StatusUnauthorized)
				return
			}

			// Check permissions
			for _, requiredPermission := range permissions {
				if !hasPermission(user, requiredPermission) {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth is a middleware that optionally authenticates users
func OptionalAuth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				token, err := auth.ExtractTokenFromHeader(authHeader)
				if err == nil {
					claims, err := jwtService.ValidateToken(token)
					if err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID.String())
						ctx = context.WithValue(ctx, UserContextKey, claims)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuditLogging creates an audit logging middleware
func AuditLogging(rbacService *auth.RBACService, logger *pustakaLogger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip audit logging for GET requests and health checks
			if r.Method == "GET" || r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Get user from context
			user, ok := r.Context().Value(UserContextKey).(*auth.JWTClaims)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			// Create response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Log audit event
			logAuditEvent(r, user, wrapper.statusCode, logger)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func hasPermission(user *auth.JWTClaims, permission string) bool {
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func logAuditEvent(r *http.Request, user *auth.JWTClaims, statusCode int, logger *pustakaLogger.Logger) {
	// Only log successful operations (2xx status codes)
	if statusCode < 200 || statusCode >= 300 {
		return
	}

	// Determine action based on HTTP method and path
	action := getActionFromRequest(r)
	entityType := getEntityTypeFromPath(r.URL.Path)

	fields := map[string]interface{}{
		"method":       r.Method,
		"path":         r.URL.Path,
		"ip_address":   getRealIP(r),
		"user_agent":   r.UserAgent(),
		"status_code":  statusCode,
	}

	logger.InfoAudit(
		entityType,
		"", // entity_id - would need to extract from response body or path
		action,
		user.UserID.String(),
		fields,
	)
}

func getActionFromRequest(r *http.Request) string {
	switch r.Method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return r.Method
	}
}

func getEntityTypeFromPath(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/")

	// Extract first segment as entity type
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[:idx]
	}
	return path
}

func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(r *http.Request) (*auth.JWTClaims, bool) {
	user, ok := r.Context().Value(UserContextKey).(*auth.JWTClaims)
	return user, ok
}