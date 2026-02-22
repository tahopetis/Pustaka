package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/ea"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// EAHandlers handles EA entity HTTP requests
type EAHandlers struct {
	eaService *ea.Service
	logger    *pustakaLogger.Logger
}

// NewEAHandlers creates a new EA handlers instance
func NewEAHandlers(eaService *ea.Service, logger *pustakaLogger.Logger) *EAHandlers {
	return &EAHandlers{
		eaService: eaService,
		logger:    logger,
	}
}

// CreateEAEntity handles POST /api/v1/ea/entities
func (h *EAHandlers) CreateEAEntity(w http.ResponseWriter, r *http.Request) {
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

	var req ea.CreateEACIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.CIType == "" {
		h.writeError(w, http.StatusBadRequest, "CI type is required")
		return
	}
	if req.Owner == "" {
		h.writeError(w, http.StatusBadRequest, "Owner (EA team) is required")
		return
	}

	entity, err := h.eaService.CreateEntity(r.Context(), &req, userID)
	if err != nil {
		// Only block on critical errors (not validation failures)
		if err == ea.ErrValidationFailed {
			// Warn-but-allow: return 422 only for critical validation errors
			// The service should have already saved with data_quality_score
			h.logger.ErrorService("ea", "create_entity", err, map[string]interface{}{
				"request": req,
				"user_id": userID,
			})
			h.writeError(w, http.StatusUnprocessableEntity, "Critical validation error: "+err.Error())
			return
		}
		h.logger.ErrorService("ea", "create_entity", err, map[string]interface{}{
			"request": req,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to create EA entity")
		return
	}

	// Return entity with validation warnings if data quality score < 100
	response := map[string]interface{}{
		"id":                 entity.ID,
		"name":               entity.Name,
		"ci_type":            entity.CIType,
		"owner":              entity.Owner,
		"attributes":         entity.Attributes,
		"tags":               entity.Tags,
		"lifecycle_status":   entity.LifecycleStatus,
		"data_quality_score": entity.DataQualityScore,
		"created_at":         entity.CreatedAt,
		"updated_at":         entity.UpdatedAt,
	}

	// Add validation warnings if score is less than perfect
	if entity.DataQualityScore < 100 {
		// Extract validation warnings from service if available
		// For now, just indicate score is less than perfect
		response["validation_warnings"] = []string{
			fmt.Sprintf("Data quality score is %.1f%% (recommended: 100%%)", entity.DataQualityScore),
		}
	} else {
		response["validation_warnings"] = []string{}
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetEAEntity handles GET /api/v1/ea/entities/{id}
func (h *EAHandlers) GetEAEntity(w http.ResponseWriter, r *http.Request) {
	id := extractIDParam(r)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	entity, err := h.eaService.GetEntity(r.Context(), id)
	if err != nil {
		if err.Error() == "EA entity not found" {
			h.writeError(w, http.StatusNotFound, "EA entity not found")
			return
		}
		h.logger.ErrorService("ea", "get_entity", err, map[string]interface{}{
			"entity_id": id,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get EA entity")
		return
	}

	h.writeJSON(w, http.StatusOK, entity)
}

// UpdateEAEntity handles PUT /api/v1/ea/entities/{id}
func (h *EAHandlers) UpdateEAEntity(w http.ResponseWriter, r *http.Request) {
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

	id := extractIDParam(r)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	var req ea.UpdateEACIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entity, err := h.eaService.UpdateEntity(r.Context(), id, &req, userID)
	if err != nil {
		// Only block on critical errors (not validation failures)
		if err == ea.ErrValidationFailed {
			// Warn-but-allow: return 422 only for critical validation errors
			h.logger.ErrorService("ea", "update_entity", err, map[string]interface{}{
				"entity_id": id,
				"user_id":   userID,
			})
			h.writeError(w, http.StatusUnprocessableEntity, "Critical validation error: "+err.Error())
			return
		}
		if err.Error() == "EA entity not found" {
			h.writeError(w, http.StatusNotFound, "EA entity not found")
			return
		}
		h.logger.ErrorService("ea", "update_entity", err, map[string]interface{}{
			"entity_id": id,
			"user_id":   userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to update EA entity")
		return
	}

	// Return entity with validation warnings if data quality score < 100
	response := map[string]interface{}{
		"id":                 entity.ID,
		"name":               entity.Name,
		"ci_type":            entity.CIType,
		"owner":              entity.Owner,
		"attributes":         entity.Attributes,
		"tags":               entity.Tags,
		"lifecycle_status":   entity.LifecycleStatus,
		"data_quality_score": entity.DataQualityScore,
		"created_at":         entity.CreatedAt,
		"updated_at":         entity.UpdatedAt,
	}

	// Add validation warnings if score is less than perfect
	if entity.DataQualityScore < 100 {
		response["validation_warnings"] = []string{
			fmt.Sprintf("Data quality score is %.1f%% (recommended: 100%%)", entity.DataQualityScore),
		}
	} else {
		response["validation_warnings"] = []string{}
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteEAEntity handles DELETE /api/v1/ea/entities/{id}
func (h *EAHandlers) DeleteEAEntity(w http.ResponseWriter, r *http.Request) {
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

	id := extractIDParam(r)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Check for force flag (to bypass relationship check after confirmation)
	forceDelete := r.URL.Query().Get("force") == "true"

	if err := h.eaService.DeleteEntity(r.Context(), id, userID, forceDelete); err != nil {
		if err.Error() == "EA entity not found" {
			h.writeError(w, http.StatusNotFound, "EA entity not found")
			return
		}
		// Check for relationship dependency error - return count
		if relErr, ok := err.(*ea.ErrRelationshipsExist); ok {
			h.writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":            fmt.Sprintf("Cannot delete entity with %d relationships", relErr.Count),
				"relationship_count": relErr.Count,
			})
			return
		}
		h.logger.ErrorService("ea", "delete_entity", err, map[string]interface{}{
			"entity_id": id,
			"user_id":   userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to delete EA entity")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEAEntities handles GET /api/v1/ea/entities
func (h *EAHandlers) ListEAEntities(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := ea.EAFilter{
		Domain:   r.URL.Query().Get("domain"),
		CIType:   r.URL.Query().Get("ci_type"),
		Search:   r.URL.Query().Get("search"),
		Tags:     r.URL.Query()["tags"], // Multi-value query param
		Page:     1,
		PageSize: 25,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.PageSize = pageSize
		}
	}

	entities, total, err := h.eaService.ListEntities(r.Context(), filter)
	if err != nil {
		h.logger.ErrorService("ea", "list_entities", err, map[string]interface{}{
			"filter": filter,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to list EA entities")
		return
	}

	totalPages := int64(0)
	if filter.PageSize > 0 {
		totalInt64 := int64(total)
		pageSizeInt64 := int64(filter.PageSize)
		totalPages = (totalInt64 + pageSizeInt64 - 1) / pageSizeInt64
	}

	response := map[string]interface{}{
		"entities": entities,
		"page":     filter.Page,
		"page_size": filter.PageSize,
		"total":    total,
		"total_pages": totalPages,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ValidateEAEntity handles GET /api/v1/ea/entities/{id}/validate
func (h *EAHandlers) ValidateEAEntity(w http.ResponseWriter, r *http.Request) {
	id := extractIDParam(r)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	result, err := h.eaService.ValidateEntity(r.Context(), id)
	if err != nil {
		if err.Error() == "EA entity not found" {
			h.writeError(w, http.StatusNotFound, "EA entity not found")
			return
		}
		h.logger.ErrorService("ea", "validate_entity", err, map[string]interface{}{
			"entity_id": id,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to validate EA entity")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Helper functions

func (h *EAHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EAHandlers) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func extractIDParam(r *http.Request) string {
	return chi.URLParam(r, "id")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || indexOfString(s, substr) >= 0))
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetEAEntityAuditLogs handles GET /api/v1/ea/entities/{id}/audit
func (h *EAHandlers) GetEAEntityAuditLogs(w http.ResponseWriter, r *http.Request) {
	id := extractIDParam(r)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 50

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Parse entity ID as UUID
	entityID, err := uuid.Parse(id)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid entity ID format")
		return
	}

	// Get audit logs from service
	auditLogs, total, err := h.eaService.GetEntityAuditLogs(r.Context(), entityID, page, pageSize)
	if err != nil {
		h.logger.ErrorService("ea", "get_entity_audit_logs", err, map[string]interface{}{
			"entity_id": id,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit logs")
		return
	}

	totalPages := int64(0)
	if pageSize > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	response := map[string]interface{}{
		"audit_logs":  auditLogs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListEACITypes handles GET /api/v1/ea/ci-types
// @Summary List EA CI Types
// @Tags ea
// @Accept json
// @Produce json
// @Success 200 {array} ea.CITypeDefinition
// @Router /ea/ci-types [get]
func (h *EAHandlers) ListEACITypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ciTypes, err := h.eaService.ListEACITypes(ctx)
	if err != nil {
		h.logger.ErrorService("ea", "list_ea_ci_types", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to list EA CI types")
		return
	}

	h.writeJSON(w, http.StatusOK, ciTypes)
}

// ListEATeams handles GET /api/v1/ea/teams
// @Summary List EA Teams
// @Tags ea
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ea/teams [get]
func (h *EAHandlers) ListEATeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teams, err := h.eaService.ListTeams(ctx)
	if err != nil {
		h.logger.ErrorService("ea", "list_ea_teams", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to list EA teams")
		return
	}

	response := map[string]interface{}{
		"data":  teams,
		"total": len(teams),
	}

	h.writeJSON(w, http.StatusOK, response)
}
