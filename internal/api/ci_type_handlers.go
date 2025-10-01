package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

type CITypeHandlers struct {
	*Handler
	ciService *ci.Service
}

func NewCITypeHandlers(handler *Handler, ciService *ci.Service) *CITypeHandlers {
	return &CITypeHandlers{
		Handler:   handler,
		ciService: ciService,
	}
}

// CreateCIType godoc
// @Summary Create a CI type
// @Description Create a new configuration item type definition with schema
// @Tags ci-types
// @Accept json
// @Produce json
// @Param request body ci.CreateCITypeRequest true "CI type definition to create"
// @Success 201 {object} ci.CITypeDefinition
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci-types [post]
func (h *CITypeHandlers) CreateCIType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req ci.CreateCITypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "Name is required")
		return
	}

	ciType, err := h.ciService.CreateCIType(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "CI type already exists" {
			h.writeError(w, http.StatusConflict, "CI type with this name already exists")
			return
		}
		h.logger.ErrorService("ci_type", "CREATE_CI_TYPE", err, map[string]interface{}{
			"request": req,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to create CI type")
		return
	}

	h.writeJSON(w, http.StatusCreated, ciType)
}

// GetCIType godoc
// @Summary Get a CI type
// @Description Get a configuration item type definition by ID
// @Tags ci-types
// @Produce json
// @Param id path string true "CI type ID"
// @Success 200 {object} ci.CITypeDefinition
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci-types/{id} [get]
func (h *CITypeHandlers) GetCIType(w http.ResponseWriter, r *http.Request) {
	ciTypeID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI type ID")
		return
	}

	ciType, err := h.ciService.GetCIType(r.Context(), ciTypeID)
	if err != nil {
		if err.Error() == "CI type not found" {
			h.writeError(w, http.StatusNotFound, "CI type not found")
			return
		}
		h.logger.ErrorService("ci_type", "GET_CI_TYPE", err, map[string]interface{}{
			"ci_type_id": ciTypeID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI type")
		return
	}

	h.writeJSON(w, http.StatusOK, ciType)
}

// ListCITypes godoc
// @Summary List CI types
// @Description List configuration item type definitions with pagination and search
// @Tags ci-types
// @Produce json
// @Param search query string false "Search in name and description"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} ci.CITypeListResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci-types [get]
func (h *CITypeHandlers) ListCITypes(w http.ResponseWriter, r *http.Request) {
	search := h.getQueryString(r, "search")
	page := h.getQueryInt(r, "page", 1)
	limit := h.getQueryInt(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := h.ciService.ListCITypes(r.Context(), page, limit, search)
	if err != nil {
		h.logger.ErrorService("ci_type", "LIST_CI_TYPES", err, map[string]interface{}{
			"search": search,
			"page":   page,
			"limit":  limit,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to list CI types")
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// UpdateCIType godoc
// @Summary Update a CI type
// @Description Update a configuration item type definition
// @Tags ci-types
// @Accept json
// @Produce json
// @Param id path string true "CI type ID"
// @Param request body ci.UpdateCITypeRequest true "CI type updates"
// @Success 200 {object} ci.CITypeDefinition
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci-types/{id} [put]
func (h *CITypeHandlers) UpdateCIType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	ciTypeID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI type ID")
		return
	}

	var req ci.UpdateCITypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ciType, err := h.ciService.UpdateCIType(r.Context(), ciTypeID, &req, userID)
	if err != nil {
		if err.Error() == "CI type not found" {
			h.writeError(w, http.StatusNotFound, "CI type not found")
			return
		}
		h.logger.ErrorService("ci_type", "UPDATE_CI_TYPE", err, map[string]interface{}{
			"ci_type_id": ciTypeID,
			"request":    req,
			"user_id":    userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to update CI type")
		return
	}

	h.writeJSON(w, http.StatusOK, ciType)
}

// DeleteCIType godoc
// @Summary Delete a CI type
// @Description Delete a configuration item type definition (only if no CIs of this type exist)
// @Tags ci-types
// @Param id path string true "CI type ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci-types/{id} [delete]
func (h *CITypeHandlers) DeleteCIType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	ciTypeID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI type ID")
		return
	}

	err = h.ciService.DeleteCIType(r.Context(), ciTypeID, userID)
	if err != nil {
		if err.Error() == "CI type not found" {
			h.writeError(w, http.StatusNotFound, "CI type not found")
			return
		}
		if err.Error() == "cannot delete CI type with existing CIs" {
			h.writeError(w, http.StatusConflict, "Cannot delete CI type with existing configuration items")
			return
		}
		h.logger.ErrorService("ci_type", "DELETE_CI_TYPE", err, map[string]interface{}{
			"ci_type_id": ciTypeID,
			"user_id":    userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to delete CI type")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCITypesByUsage godoc
// @Summary Get CI types by usage
// @Description Get CI types sorted by usage frequency
// @Tags analytics
// @Produce json
// @Success 200 {array} ci.CITypeUsage
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/ci-types/usage [get]
func (h *CITypeHandlers) GetCITypesByUsage(w http.ResponseWriter, r *http.Request) {
	usage, err := h.ciService.GetCITypesByUsage(r.Context())
	if err != nil {
		h.logger.ErrorService("ci_type", "GET_CI_TYPES_BY_USAGE", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI types by usage")
		return
	}

	h.writeJSON(w, http.StatusOK, usage)
}