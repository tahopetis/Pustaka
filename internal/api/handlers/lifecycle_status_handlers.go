package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/auth"
	"github.com/pustaka/pustaka/internal/ci"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// LifecycleStatusHandler handles HTTP requests for lifecycle statuses
type LifecycleStatusHandler struct {
	lifecycleStatusService *ci.LifecycleStatusService
	rbacService           *auth.RBACService
	logger                *pustakaLogger.Logger
}

// NewLifecycleStatusHandler creates a new lifecycle status handler
func NewLifecycleStatusHandler(
	lifecycleStatusService *ci.LifecycleStatusService,
	rbacService *auth.RBACService,
	logger *pustakaLogger.Logger,
) *LifecycleStatusHandler {
	return &LifecycleStatusHandler{
		lifecycleStatusService: lifecycleStatusService,
		rbacService:           rbacService,
		logger:                logger,
	}
}

// CreateLifecycleStatus creates a new lifecycle status
func (h *LifecycleStatusHandler) CreateLifecycleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var req ci.CreateLifecycleStatusRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.lifecycleStatusService.CreateLifecycleStatus(ctx, &req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create lifecycle status")
		h.writeError(w, http.StatusInternalServerError, "Failed to create lifecycle status")
		return
	}

	h.writeJSON(w, http.StatusCreated, result)
}

// GetLifecycleStatus retrieves a lifecycle status by ID
func (h *LifecycleStatusHandler) GetLifecycleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid lifecycle status ID")
		return
	}

	result, err := h.lifecycleStatusService.GetLifecycleStatus(ctx, id)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get lifecycle status")
		h.writeError(w, http.StatusNotFound, "Lifecycle status not found")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// ListLifecycleStatuses retrieves lifecycle statuses with pagination and filtering
func (h *LifecycleStatusHandler) ListLifecycleStatuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filters := &ci.ListLifecycleStatusFilters{}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}
	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		isActiveBool := isActive == "true"
		filters.IsActive = &isActiveBool
	}
	if isSystem := r.URL.Query().Get("is_system"); isSystem != "" {
		isSystemBool := isSystem == "true"
		filters.IsSystem = &isSystemBool
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		filters.Sort = sort
	}
	if order := r.URL.Query().Get("order"); order != "" {
		filters.Order = order
	}

	result, err := h.lifecycleStatusService.ListLifecycleStatuses(ctx, filters, page, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list lifecycle statuses")
		h.writeError(w, http.StatusInternalServerError, "Failed to list lifecycle statuses")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// UpdateLifecycleStatus updates an existing lifecycle status
func (h *LifecycleStatusHandler) UpdateLifecycleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid lifecycle status ID")
		return
	}

	var req ci.UpdateLifecycleStatusRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.lifecycleStatusService.UpdateLifecycleStatus(ctx, id, &req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to update lifecycle status")
		h.writeError(w, http.StatusInternalServerError, "Failed to update lifecycle status")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// DeleteLifecycleStatus deletes a lifecycle status
func (h *LifecycleStatusHandler) DeleteLifecycleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid lifecycle status ID")
		return
	}

	err = h.lifecycleStatusService.DeleteLifecycleStatus(ctx, id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete lifecycle status")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete lifecycle status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveLifecycleStatuses retrieves all active lifecycle statuses
func (h *LifecycleStatusHandler) GetActiveLifecycleStatuses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result, err := h.lifecycleStatusService.GetActiveLifecycleStatuses(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get active lifecycle statuses")
		h.writeError(w, http.StatusInternalServerError, "Failed to get active lifecycle statuses")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetLifecycleStatusUsage retrieves usage statistics for lifecycle statuses
func (h *LifecycleStatusHandler) GetLifecycleStatusUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result, err := h.lifecycleStatusService.GetLifecycleStatusUsage(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get lifecycle status usage")
		h.writeError(w, http.StatusInternalServerError, "Failed to get lifecycle status usage")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// GetCIStatusDistribution retrieves CI status distribution for dashboard
func (h *LifecycleStatusHandler) GetCIStatusDistribution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result, err := h.lifecycleStatusService.GetCIStatusDistribution(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get CI status distribution")
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI status distribution")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Helper methods

func (h *LifecycleStatusHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *LifecycleStatusHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *LifecycleStatusHandler) decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}