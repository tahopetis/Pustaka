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

type RelationshipTypeHandler struct {
	relationshipTypeService *ci.RelationshipTypeService
	rbacService            *auth.RBACService
	logger                 *pustakaLogger.Logger
}

func NewRelationshipTypeHandler(
	relationshipTypeService *ci.RelationshipTypeService,
	rbacService *auth.RBACService,
	logger *pustakaLogger.Logger,
) *RelationshipTypeHandler {
	return &RelationshipTypeHandler{
		relationshipTypeService: relationshipTypeService,
		rbacService:            rbacService,
		logger:                 logger,
	}
}

// CreateRelationshipType creates a new relationship type
func (h *RelationshipTypeHandler) CreateRelationshipType(w http.ResponseWriter, r *http.Request) {
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

	var req ci.CreateRelationshipTypeRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	relationshipType, err := h.relationshipTypeService.CreateRelationshipType(ctx, &req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create relationship type")
		h.writeError(w, http.StatusInternalServerError, "Failed to create relationship type")
		return
	}

	h.writeJSON(w, http.StatusCreated, relationshipType)
}

// GetRelationshipType retrieves a relationship type by ID
func (h *RelationshipTypeHandler) GetRelationshipType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid relationship type ID")
		return
	}

	relationshipType, err := h.relationshipTypeService.GetRelationshipType(ctx, id)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get relationship type")
		h.writeError(w, http.StatusNotFound, "Relationship type not found")
		return
	}

	h.writeJSON(w, http.StatusOK, relationshipType)
}

// ListRelationshipTypes retrieves paginated relationship types
func (h *RelationshipTypeHandler) ListRelationshipTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Build filters
	filters := ci.ListRelationshipTypeFilters{
		Search:   r.URL.Query().Get("search"),
		Category: r.URL.Query().Get("category"),
		Sort:     r.URL.Query().Get("sort"),
		Order:    r.URL.Query().Get("order"),
	}

	// Parse boolean filters
	if activeStr := r.URL.Query().Get("is_active"); activeStr != "" {
		if active := activeStr == "true"; active {
			filters.IsActive = &active
		}
	}

	if systemStr := r.URL.Query().Get("is_system"); systemStr != "" {
		if system := systemStr == "true"; system {
			filters.IsSystem = &system
		}
	}

	response, err := h.relationshipTypeService.ListRelationshipTypes(ctx, filters, page, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list relationship types")
		h.writeError(w, http.StatusInternalServerError, "Failed to list relationship types")
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetActiveRelationshipTypes returns all active relationship types
func (h *RelationshipTypeHandler) GetActiveRelationshipTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	relationshipTypes, err := h.relationshipTypeService.GetActiveRelationshipTypes(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get active relationship types")
		h.writeError(w, http.StatusInternalServerError, "Failed to get active relationship types")
		return
	}

	h.writeJSON(w, http.StatusOK, relationshipTypes)
}

// UpdateRelationshipType updates an existing relationship type
func (h *RelationshipTypeHandler) UpdateRelationshipType(w http.ResponseWriter, r *http.Request) {
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
		h.writeError(w, http.StatusBadRequest, "Invalid relationship type ID")
		return
	}

	var req ci.UpdateRelationshipTypeRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	relationshipType, err := h.relationshipTypeService.UpdateRelationshipType(ctx, id, &req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to update relationship type")
		h.writeError(w, http.StatusInternalServerError, "Failed to update relationship type")
		return
	}

	h.writeJSON(w, http.StatusOK, relationshipType)
}

// DeleteRelationshipType deletes a relationship type
func (h *RelationshipTypeHandler) DeleteRelationshipType(w http.ResponseWriter, r *http.Request) {
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
		h.writeError(w, http.StatusBadRequest, "Invalid relationship type ID")
		return
	}

	err = h.relationshipTypeService.DeleteRelationshipType(ctx, id, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to delete relationship type")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete relationship type")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Relationship type deleted successfully"})
}

// ValidateRelationship validates if a relationship can be created
func (h *RelationshipTypeHandler) ValidateRelationship(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ci.RelationshipCompatibilityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.relationshipTypeService.ValidateRelationship(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to validate relationship")
		h.writeError(w, http.StatusInternalServerError, "Failed to validate relationship")
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetRelationshipTypeUsage returns usage statistics
func (h *RelationshipTypeHandler) GetRelationshipTypeUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	usage, err := h.relationshipTypeService.GetRelationshipTypeUsage(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get relationship type usage")
		h.writeError(w, http.StatusInternalServerError, "Failed to get relationship type usage")
		return
	}

	h.writeJSON(w, http.StatusOK, usage)
}

// GetRelationshipTypeStatistics returns comprehensive statistics
func (h *RelationshipTypeHandler) GetRelationshipTypeStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.relationshipTypeService.GetRelationshipTypeStatistics(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get relationship type statistics")
		h.writeError(w, http.StatusInternalServerError, "Failed to get relationship type statistics")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// GetRelationshipTypeCategories retrieves all unique categories
func (h *RelationshipTypeHandler) GetRelationshipTypeCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.relationshipTypeService.GetRelationshipTypeCategories(ctx)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to retrieve relationship type categories")
		h.writeError(w, http.StatusInternalServerError, "Failed to retrieve categories")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// Helper methods

func (h *RelationshipTypeHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RelationshipTypeHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *RelationshipTypeHandler) decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}