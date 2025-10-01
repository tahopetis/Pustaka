package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pustaka/pustaka/internal/auth"
	"github.com/pustaka/pustaka/internal/api/middleware"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type AuthHandler struct {
	jwtService     *auth.JWTService
	passwordService *auth.PasswordService
	rbacService    *auth.RBACService
	logger         *pustakaLogger.Logger
}

func NewAuthHandler(jwtService *auth.JWTService, passwordService *auth.PasswordService, rbacService *auth.RBACService, logger *pustakaLogger.Logger) *AuthHandler {
	return &AuthHandler{
		jwtService:     jwtService,
		passwordService: passwordService,
		rbacService:    rbacService,
		logger:         logger,
	}
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType   string     `json:"token_type"`
	ExpiresIn   int64      `json:"expires_in"`
	User        *auth.User `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.rbacService.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Error().Str("action", "login").Str("username", req.Username).Str("ip", getClientIP(r)).Err(err).Msg("Authentication failed")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.IsActive {
		h.logger.Info().Str("action", "login_failed").Str("user_id", user.ID.String()).Str("username", req.Username).Str("ip", getClientIP(r)).Interface("reason", map[string]interface{}{
			"reason": "user_inactive",
		}).Msg("Login failed")
		http.Error(w, "Account is inactive", http.StatusUnauthorized)
		return
	}

	// Verify password
	passwordHash, err := h.rbacService.GetUserPasswordHash(r.Context(), req.Username)
	if err != nil {
		h.logger.Error().Str("action", "login").Str("username", req.Username).Str("ip", getClientIP(r)).Err(err).Msg("Authentication failed - user not found")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if valid, err := h.passwordService.VerifyPassword(req.Password, passwordHash); err != nil || !valid {
		h.logger.Info().Str("action", "login_failed").Str("username", req.Username).Str("ip", getClientIP(r)).Interface("reason", map[string]interface{}{
			"reason": "invalid_password",
		}).Msg("Login failed - invalid password")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user.ID, user.Username, user.Email, getRoleNames(user.Roles), user.Permissions)
	if err != nil {
		h.logger.ErrorAuth("login", user.ID.String(), req.Username, getClientIP(r), err, nil)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		h.logger.ErrorAuth("login", user.ID.String(), req.Username, getClientIP(r), err, nil)
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Log successful login
	h.logger.Info().Str("action", "login").Str("user_id", user.ID.String()).Str("username", req.Username).Str("ip", getClientIP(r)).Msg("Login successful")

	// Return response
	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.jwtService.GetAccessTokenTTL().Seconds()),
		User:       user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from refresh token
	userID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Error().Str("action", "refresh_token").Str("ip", getClientIP(r)).Err(err).Msg("Refresh token failed")
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := h.rbacService.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Error().Str("action", "refresh_token").Str("user_id", userID.String()).Str("ip", getClientIP(r)).Err(err).Msg("Refresh token failed")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.IsActive {
		h.logger.Info().Str("action", "refresh_token_failed").Str("user_id", user.ID.String()).Str("username", user.Username).Str("ip", getClientIP(r)).Interface("reason", map[string]interface{}{
			"reason": "user_inactive",
		}).Msg("Refresh token failed")
		http.Error(w, "Account is inactive", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := h.jwtService.RefreshToken(
		req.RefreshToken,
		user.Username,
		user.Email,
		getRoleNames(user.Roles),
		user.Permissions,
	)
	if err != nil {
		h.logger.ErrorAuth("refresh_token", user.ID.String(), user.Username, getClientIP(r), err, nil)
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}

	// Log successful token refresh
	h.logger.Info().Str("action", "refresh_token").Str("user_id", user.ID.String()).Str("username", user.Username).Str("ip", getClientIP(r)).Msg("Refresh token successful")

	// Return response
	response := map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(h.jwtService.GetAccessTokenTTL().Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	// Get full user details from database
	fullUser, err := h.rbacService.GetUserByID(r.Context(), user.UserID)
	if err != nil {
		http.Error(w, "Failed to get user details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"user": fullUser,
	}
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getRoleNames(roles []auth.Role) []string {
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	return roleNames
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}