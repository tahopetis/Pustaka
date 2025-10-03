package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/ci"
)

type RelationshipHandlers struct {
	*Handler
	ciService *ci.Service
}

func NewRelationshipHandlers(handler *Handler, ciService *ci.Service) *RelationshipHandlers {
	return &RelationshipHandlers{
		Handler:   handler,
		ciService: ciService,
	}
}

// CreateRelationship godoc
// @Summary Create a relationship
// @Description Create a new relationship between two configuration items
// @Tags relationships
// @Accept json
// @Produce json
// @Param request body ci.CreateRelationshipRequest true "Relationship to create"
// @Success 201 {object} ci.Relationship
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/relationships [post]
func (h *RelationshipHandlers) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "User not found in context")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var req ci.CreateRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.SourceID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "Source CI ID is required")
		return
	}
	if req.TargetID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "Target CI ID is required")
		return
	}
	if req.RelationshipType == "" {
		h.writeError(w, http.StatusBadRequest, "Relationship type is required")
		return
	}

	relationship, err := h.ciService.CreateRelationship(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "source CI not found" || err.Error() == "target CI not found" {
			h.writeError(w, http.StatusNotFound, "Configuration item not found")
			return
		}
		if err.Error() == "cannot create self-referencing relationship" {
			h.writeError(w, http.StatusBadRequest, "Cannot create self-referencing relationship")
			return
		}
		h.logger.ErrorService("relationship", "CREATE_RELATIONSHIP", err, map[string]interface{}{
			"request": req,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to create relationship")
		return
	}

	h.writeJSON(w, http.StatusCreated, relationship)
}

// GetRelationship godoc
// @Summary Get a relationship
// @Description Get a relationship by ID
// @Tags relationships
// @Produce json
// @Param id path string true "Relationship ID"
// @Success 200 {object} ci.Relationship
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/relationships/{id} [get]
func (h *RelationshipHandlers) GetRelationship(w http.ResponseWriter, r *http.Request) {
	relationshipID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	relationship, err := h.ciService.GetRelationship(r.Context(), relationshipID)
	if err != nil {
		if err.Error() == "relationship not found" {
			h.writeError(w, http.StatusNotFound, "Relationship not found")
			return
		}
		h.logger.ErrorService("relationship", "GET_RELATIONSHIP", err, map[string]interface{}{
			"relationship_id": relationshipID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get relationship")
		return
	}

	h.writeJSON(w, http.StatusOK, relationship)
}

// ListRelationships godoc
// @Summary List relationships
// @Description List relationships with filtering and pagination
// @Tags relationships
// @Produce json
// @Param source_id query string false "Filter by source CI ID"
// @Param target_id query string false "Filter by target CI ID"
// @Param relationship_type query string false "Filter by relationship type"
// @Param search query string false "Search term for relationships"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} ci.RelationshipListResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/relationships [get]
func (h *RelationshipHandlers) ListRelationships(w http.ResponseWriter, r *http.Request) {
	filters := ci.ListRelationshipFilters{}

	// Parse optional UUID parameters
	if sourceIDStr := h.getQueryString(r, "source_id"); sourceIDStr != "" {
		if sourceID, err := uuid.Parse(sourceIDStr); err == nil {
			filters.SourceID = &sourceID
		}
	}

	if targetIDStr := h.getQueryString(r, "target_id"); targetIDStr != "" {
		if targetID, err := uuid.Parse(targetIDStr); err == nil {
			filters.TargetID = &targetID
		}
	}

	filters.RelationshipType = h.getQueryString(r, "relationship_type")
	filters.Search = h.getQueryString(r, "search")

	page := h.getQueryInt(r, "page", 1)
	limit := h.getQueryInt(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := h.ciService.ListRelationships(r.Context(), filters, page, limit)
	if err != nil {
		h.logger.ErrorService("relationship", "LIST_RELATIONSHIPS", err, map[string]interface{}{
			"filters": filters,
			"page":    page,
			"limit":   limit,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to list relationships")
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// UpdateRelationship godoc
// @Summary Update a relationship
// @Description Update a relationship between configuration items
// @Tags relationships
// @Accept json
// @Produce json
// @Param id path string true "Relationship ID"
// @Param request body ci.UpdateRelationshipRequest true "Relationship updates"
// @Success 200 {object} ci.Relationship
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/relationships/{id} [put]
func (h *RelationshipHandlers) UpdateRelationship(w http.ResponseWriter, r *http.Request) {
	// FIXED: Properly extract and parse user ID from context
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "User not found in context")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	relationshipID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	var req ci.UpdateRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	relationship, err := h.ciService.UpdateRelationship(r.Context(), relationshipID, &req, userID)
	if err != nil {
		if err.Error() == "relationship not found" {
			h.writeError(w, http.StatusNotFound, "Relationship not found")
			return
		}
		h.logger.ErrorService("relationship", "UPDATE_RELATIONSHIP", err, map[string]interface{}{
			"relationship_id": relationshipID,
			"request":         req,
			"user_id":         userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to update relationship")
		return
	}

	h.writeJSON(w, http.StatusOK, relationship)
}

// DeleteRelationship godoc
// @Summary Delete a relationship
// @Description Delete a relationship between configuration items
// @Tags relationships
// @Param id path string true "Relationship ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/relationships/{id} [delete]
func (h *RelationshipHandlers) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	// FIXED: Properly extract and parse user ID from context
	userIDStr, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "User not found in context")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	relationshipID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	err = h.ciService.DeleteRelationship(r.Context(), relationshipID, userID)
	if err != nil {
		h.logger.ErrorService("relationship", "DELETE_RELATIONSHIP", err, map[string]interface{}{
			"relationship_id": relationshipID,
			"user_id":         userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to delete relationship")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// FindCycles godoc
// @Summary Find cycles in relationships
// @Description Find circular dependencies in the CI graph
// @Tags analytics
// @Produce json
// @Success 200 {array} []uuid.UUID
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/cycles [get]
func (h *RelationshipHandlers) FindCycles(w http.ResponseWriter, r *http.Request) {
	cycles, err := h.ciService.FindCycles(r.Context())
	if err != nil {
		h.logger.ErrorService("relationship", "FIND_CYCLES", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to find cycles")
		return
	}

	h.writeJSON(w, http.StatusOK, cycles)
}

// GetMostConnectedCIs godoc
// @Summary Get most connected CIs
// @Description Get configuration items with the most relationships
// @Tags analytics
// @Produce json
// @Param limit query int false "Maximum number of results" default(10)
// @Success 200 {array} ci.CIConnectivity
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/most-connected [get]
func (h *RelationshipHandlers) GetMostConnectedCIs(w http.ResponseWriter, r *http.Request) {
	limit := h.getQueryInt(r, "limit", 10)
	if limit < 1 || limit > 50 {
		limit = 10
	}

	connectivity, err := h.ciService.GetMostConnectedCIs(r.Context(), limit)
	if err != nil {
		h.logger.ErrorService("relationship", "GET_MOST_CONNECTED_CIS", err, map[string]interface{}{
			"limit": limit,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get most connected CIs")
		return
	}

	h.writeJSON(w, http.StatusOK, connectivity)
}