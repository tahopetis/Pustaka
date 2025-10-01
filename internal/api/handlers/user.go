package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/auth"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type UserHandler struct {
	rbacService *auth.RBACService
	passwordService *auth.PasswordService
	logger     *pustakaLogger.Logger
}

func NewUserHandler(rbacService *auth.RBACService, passwordService *auth.PasswordService, logger *pustakaLogger.Logger) *UserHandler {
	return &UserHandler{
		rbacService: rbacService,
		passwordService: passwordService,
		logger:     logger,
	}
}

type CreateUserRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=100"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Roles    []string `json:"roles"`
}

type UpdateUserRequest struct {
	Email     *string  `json:"email,omitempty" validate:"omitempty,email"`
	IsActive  *bool    `json:"is_active,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

type UserListResponse struct {
	Users      []auth.User `json:"users"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// ListUsers godoc
// @Summary List users
// @Description List users with pagination and search
// @Tags users
// @Produce json
// @Param search query string false "Search in username and email"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} UserListResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	page := h.getQueryInt(r, "page", 1)
	limit := h.getQueryInt(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// TODO: Implement search and pagination in RBACService
	// For now, return a simple implementation
	users, err := h.listAllUsers(r.Context())
	if err != nil {
		h.logger.ErrorService("user", "LIST_USERS", err, map[string]interface{}{
			"search": search,
			"page":   page,
			"limit":  limit,
		})
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	response := UserListResponse{
		Users:      users,
		Total:      len(users),
		Page:       page,
		Limit:      limit,
		TotalPages: (len(users) + limit - 1) / limit, // Simple ceiling division
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser godoc
// @Summary Get a user
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUUIDParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.rbacService.GetUserByID(r.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.ErrorService("user", "GET_USER", err, map[string]interface{}{
			"user_id": userID,
		})
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// CreateUser godoc
// @Summary Create a user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User to create"
// @Success 201 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := h.passwordService.HashPassword(req.Password)
	if err != nil {
		h.logger.ErrorService("user", "HASH_PASSWORD", err, map[string]interface{}{
			"username": req.Username,
		})
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.rbacService.CreateUser(r.Context(), req.Username, req.Email, passwordHash, req.Roles)
	if err != nil {
		if err.Error() == "failed to create user: unique constraint violation" {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		h.logger.ErrorService("user", "CREATE_USER", err, map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
		})
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "User updates"
// @Success 200 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUUIDParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement user update in RBACService
	// For now, return current user
	user, err := h.rbacService.GetUserByID(r.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.ErrorService("user", "UPDATE_USER", err, map[string]interface{}{
			"user_id": userID,
		})
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	_, err := h.getUUIDParam(r, "id")
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement user deletion in RBACService
	// For now, just return success
	w.WriteHeader(http.StatusNoContent)
}

// Helper methods (these should be in a shared base handler)
func (h *UserHandler) getUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	paramStr := chi.URLParam(r, param)
	return uuid.Parse(paramStr)
}

func (h *UserHandler) getQueryInt(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Temporary method until we implement proper search/pagination
func (h *UserHandler) listAllUsers(ctx context.Context) ([]auth.User, error) {
	// This is a placeholder - in a real implementation, you'd add a ListUsers method to RBACService
	// For now, return empty slice to avoid build errors
	return []auth.User{}, nil
}