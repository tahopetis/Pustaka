package middleware

import (
	"net/http"
)

// RequireEARead creates middleware that checks for ea:read permission
func RequireEARead(next http.Handler) http.Handler {
	return RBAC("ea:read")(next)
}

// RequireEACreate creates middleware that checks for ea:create permission
func RequireEACreate(next http.Handler) http.Handler {
	return RBAC("ea:create")(next)
}

// RequireEAUpdate creates middleware that checks for ea:update permission
func RequireEAUpdate(next http.Handler) http.Handler {
	return RBAC("ea:update")(next)
}

// RequireEADelete creates middleware that checks for ea:delete permission
func RequireEADelete(next http.Handler) http.Handler {
	return RBAC("ea:delete")(next)
}
