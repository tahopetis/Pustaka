package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/ci"
)

type CIHandlers struct {
	*Handler
	ciService *ci.Service
}

func NewCIHandlers(handler *Handler, ciService *ci.Service) *CIHandlers {
	return &CIHandlers{
		Handler:   handler,
		ciService: ciService,
	}
}

// CreateCI godoc
// @Summary Create a configuration item
// @Description Create a new configuration item with validation against CI type schema
// @Tags ci
// @Accept json
// @Produce json
// @Param request body ci.CreateCIRequest true "Configuration item to create"
// @Success 201 {object} ci.ConfigurationItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci [post]
func (h *CIHandlers) CreateCI(w http.ResponseWriter, r *http.Request) {
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

	var req ci.CreateCIRequest
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

	ci, err := h.ciService.CreateCI(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "Attribute validation failed" {
			h.logger.ErrorService("ci", "CREATE_CI_VALIDATION", err, map[string]interface{}{
				"request": req,
				"user_id": userID,
				"error_type": "validation",
			})
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.ErrorService("ci", "CREATE_CI", err, map[string]interface{}{
			"request": req,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to create configuration item")
		return
	}

	h.writeJSON(w, http.StatusCreated, ci)
}

// GetCI godoc
// @Summary Get a configuration item
// @Description Get a configuration item by ID
// @Tags ci
// @Produce json
// @Param id path string true "Configuration item ID"
// @Success 200 {object} ci.ConfigurationItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id} [get]
func (h *CIHandlers) GetCI(w http.ResponseWriter, r *http.Request) {
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.logger.ErrorService("ci", "GET_CI_UUID_PARSE", err, map[string]interface{}{
			"id_param": mux.Vars(r)["id"],
		})
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	ci, err := h.ciService.GetCI(r.Context(), ciID)
	if err != nil {
		if err.Error() == "CI not found" {
			h.writeError(w, http.StatusNotFound, "Configuration item not found")
			return
		}
		h.logger.ErrorService("ci", "GET_CI", err, map[string]interface{}{
			"ci_id": ciID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get configuration item")
		return
	}

	h.writeJSON(w, http.StatusOK, ci)
}

// ListCIs godoc
// @Summary List configuration items
// @Description List configuration items with filtering and pagination
// @Tags ci
// @Produce json
// @Param ci_type query string false "Filter by CI type"
// @Param search query string false "Search in name and attributes"
// @Param tags query []string false "Filter by tags"
// @Param created_by query string false "Filter by creator ID"
// @Param sort query string false "Sort field (name, type, created_at, updated_at)"
// @Param order query string false "Sort order (asc, desc)" Enums(asc, desc)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} ci.CIListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci [get]
func (h *CIHandlers) ListCIs(w http.ResponseWriter, r *http.Request) {
	filters := ci.ListCIFilters{
		CIType:    h.getQueryString(r, "ci_type"),
		Search:    h.getQueryString(r, "search"),
		Tags:      h.getQueryStrings(r, "tags"),
		CreatedBy: h.getQueryString(r, "created_by"),
		Sort:      h.getQueryString(r, "sort"),
		Order:     h.getQueryString(r, "order"),
	}

	page := h.getQueryInt(r, "page", 1)
	limit := h.getQueryInt(r, "limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	response, err := h.ciService.ListCIs(r.Context(), filters, page, limit)
	if err != nil {
		h.logger.ErrorService("ci", "LIST_CIS", err, map[string]interface{}{
			"filters": filters,
			"page":    page,
			"limit":   limit,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to list configuration items")
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

// UpdateCI godoc
// @Summary Update a configuration item
// @Description Update a configuration item with validation against CI type schema
// @Tags ci
// @Accept json
// @Produce json
// @Param id path string true "Configuration item ID"
// @Param request body ci.UpdateCIRequest true "Configuration item updates"
// @Success 200 {object} ci.ConfigurationItem
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id} [put]
func (h *CIHandlers) UpdateCI(w http.ResponseWriter, r *http.Request) {
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
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	var req ci.UpdateCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ci, err := h.ciService.UpdateCI(r.Context(), ciID, &req, userID)
	if err != nil {
		if err.Error() == "CI not found" {
			h.writeError(w, http.StatusNotFound, "Configuration item not found")
			return
		}
		if err.Error() == "Attribute validation failed" {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.ErrorService("ci", "UPDATE_CI", err, map[string]interface{}{
			"ci_id":   ciID,
			"request": req,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to update configuration item")
		return
	}

	h.writeJSON(w, http.StatusOK, ci)
}

// DeleteCI godoc
// @Summary Delete a configuration item
// @Description Delete a configuration item (only if no relationships exist)
// @Tags ci
// @Param id path string true "Configuration item ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id} [delete]
func (h *CIHandlers) DeleteCI(w http.ResponseWriter, r *http.Request) {
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
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	err = h.ciService.DeleteCI(r.Context(), ciID, userID)
	if err != nil {
		if err.Error() == "CI not found" {
			h.writeError(w, http.StatusNotFound, "Configuration item not found")
			return
		}
		if err.Error() == "cannot delete CI with existing relationships" {
			h.writeError(w, http.StatusConflict, "Cannot delete configuration item with existing relationships")
			return
		}
		h.logger.ErrorService("ci", "DELETE_CI", err, map[string]interface{}{
			"ci_id": ciID,
			"user_id": userID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to delete configuration item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCIRelationships godoc
// @Summary Get CI relationships
// @Description Get all relationships for a configuration item
// @Tags ci
// @Produce json
// @Param id path string true "Configuration item ID"
// @Success 200 {array} ci.RelationshipGraph
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id}/relationships [get]
func (h *CIHandlers) GetCIRelationships(w http.ResponseWriter, r *http.Request) {
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	relationships, err := h.ciService.GetCIRelationships(r.Context(), ciID)
	if err != nil {
		h.logger.ErrorService("ci", "GET_CI_RELATIONSHIPS", err, map[string]interface{}{
			"ci_id": ciID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI relationships")
		return
	}

	h.writeJSON(w, http.StatusOK, relationships)
}

// GetGraphData godoc
// @Summary Get graph data
// @Description Get graph data for visualization with filtering
// @Tags graph
// @Produce json
// @Param ci_types query []string false "Filter by CI types"
// @Param search query string false "Search in CI names"
// @Param limit query int false "Maximum number of nodes" default(100)
// @Success 200 {object} ci.GraphData
// @Failure 500 {object} map[string]string
// @Router /api/v1/graph [get]
func (h *CIHandlers) GetGraphData(w http.ResponseWriter, r *http.Request) {
	filters := ci.GraphFilters{
		CITypes: h.getQueryStrings(r, "ci_types"),
		Search:  h.getQueryString(r, "search"),
		Limit:   h.getQueryInt(r, "limit", 100),
	}

	if filters.Limit < 1 || filters.Limit > 500 {
		filters.Limit = 100
	}

	graphData, err := h.ciService.GetGraphData(r.Context(), filters)
	if err != nil {
		h.logger.ErrorService("ci", "GET_GRAPH_DATA", err, map[string]interface{}{
			"filters": filters,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get graph data")
		return
	}

	h.writeJSON(w, http.StatusOK, graphData)
}

// GetCINetwork godoc
// @Summary Get CI network
// @Description Get the network of a CI with specified depth
// @Tags graph
// @Produce json
// @Param id path string true "Configuration item ID"
// @Param depth query int false "Network depth" default(2)
// @Success 200 {object} ci.CINetwork
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id}/network [get]
func (h *CIHandlers) GetCINetwork(w http.ResponseWriter, r *http.Request) {
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	depth := h.getQueryInt(r, "depth", 2)
	if depth < 1 || depth > 5 {
		depth = 2
	}

	network, err := h.ciService.GetCINetwork(r.Context(), ciID, depth)
	if err != nil {
		h.logger.ErrorService("ci", "GET_CI_NETWORK", err, map[string]interface{}{
			"ci_id": ciID,
			"depth": depth,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI network")
		return
	}

	h.writeJSON(w, http.StatusOK, network)
}

// GetImpactAnalysis godoc
// @Summary Get impact analysis
// @Description Get impact analysis for a configuration item
// @Tags analysis
// @Produce json
// @Param id path string true "Configuration item ID"
// @Success 200 {object} ci.ImpactAnalysis
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ci/{id}/impact [get]
func (h *CIHandlers) GetImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	ciID, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	analysis, err := h.ciService.GetImpactAnalysis(r.Context(), ciID)
	if err != nil {
		h.logger.ErrorService("ci", "GET_IMPACT_ANALYSIS", err, map[string]interface{}{
			"ci_id": ciID,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get impact analysis")
		return
	}

	h.writeJSON(w, http.StatusOK, analysis)
}

// ExploreGraph godoc
// @Summary Explore CI graph
// @Description Explore the CI graph with filters
// @Tags graph
// @Produce json
// @Param ci_types query []string false "Filter by CI types"
// @Param search query string false "Search term"
// @Param limit query int false "Maximum number of nodes" default(100)
// @Success 200 {object} ci.GraphData
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/graph/explore [get]
func (h *CIHandlers) ExploreGraph(w http.ResponseWriter, r *http.Request) {
	filters := ci.GraphFilters{
		CITypes: h.getQueryStrings(r, "ci_types"),
		Search:  h.getQueryString(r, "search"),
		Limit:   h.getQueryInt(r, "limit", 100),
	}
	if filters.Limit < 1 || filters.Limit > 500 {
		filters.Limit = 100
	}
	graphData, err := h.ciService.GetGraphData(r.Context(), filters)
	if err != nil {
		h.logger.ErrorService("ci", "EXPLORE_GRAPH", err, map[string]interface{}{
			"filters": filters,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to explore graph data")
		return
	}
	h.writeJSON(w, http.StatusOK, graphData)
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get statistics for the dashboard including counts of CIs, CI types, relationships, and users
// @Tags dashboard
// @Produce json
// @Success 200 {object} ci.DashboardStats
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/dashboard/stats [get]
func (h *CIHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.ciService.GetDashboardStats(r.Context())
	if err != nil {
		h.logger.ErrorService("dashboard", "GET_STATS", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get dashboard statistics")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// GetCIGrowth godoc
// @Summary Get CI and relationship growth over time
// @Description Get time-series data showing CI and relationship creation counts
// @Tags analytics
// @Produce json
// @Param from_date query string false "Start date (YYYY-MM-DD)"
// @Param to_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} ci.GrowthData
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/ci-growth [get]
func (h *CIHandlers) GetCIGrowth(w http.ResponseWriter, r *http.Request) {
	fromDate := h.getQueryString(r, "from_date")
	toDate := h.getQueryString(r, "to_date")

	growthData, err := h.ciService.GetCIGrowth(r.Context(), fromDate, toDate)
	if err != nil {
		h.logger.ErrorService("analytics", "GET_CI_GROWTH", err, map[string]interface{}{
			"from_date": fromDate,
			"to_date":   toDate,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get CI growth data")
		return
	}

	h.writeJSON(w, http.StatusOK, growthData)
}

// GetHealthScore godoc
// @Summary Get CMDB health score
// @Description Get the overall CMDB health score with sub-scores for completeness, correctness, and compliance
// @Tags dashboard
// @Produce json
// @Success 200 {object} ci.HealthScore
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/dashboard/health-score [get]
func (h *CIHandlers) GetHealthScore(w http.ResponseWriter, r *http.Request) {
	score, err := h.ciService.GetHealthScore(r.Context())
	if err != nil {
		h.logger.ErrorService("dashboard", "GET_HEALTH_SCORE", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get health score")
		return
	}

	h.writeJSON(w, http.StatusOK, score)
}

// GetDataQualityMetrics returns data quality metrics for the CMDB
// @Summary Get Data Quality Metrics
// @Description Returns data quality metrics including missing attributes, orphaned CIs, and other quality issues
// @Tags dashboard
// @Produce json
// @Param include_details query bool false "Include detailed CI lists (default: false)"
// @Param limit query int false "Max number of CIs to return per category (default: 10)"
// @Success 200 {object} ci.DataQualityMetrics
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/dashboard/data-quality [get]
func (h *CIHandlers) GetDataQualityMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	includeDetails := r.URL.Query().Get("include_details") == "true"
	limit := 10 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	metrics, err := h.ciService.GetDataQualityMetrics(r.Context(), includeDetails, limit)
	if err != nil {
		h.logger.ErrorService("dashboard", "GET_DATA_QUALITY_METRICS", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get data quality metrics")
		return
	}

	h.writeJSON(w, http.StatusOK, metrics)
}

// GetAssetAgingMetrics returns asset aging metrics for the dashboard
// @Summary Get Asset Aging Metrics
// @Description Returns asset age distribution, EOL alerts, and aging statistics
// @Tags dashboard
// @Produce json
// @Param eol_threshold_months query int false "EOL threshold in months (default: 6)"
// @Param limit query int false "Max number of EOL assets to return (default: 10)"
// @Success 200 {object} ci.AssetAgingMetrics
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/dashboard/asset-aging [get]
func (h *CIHandlers) GetAssetAgingMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	eolThresholdMonths := 6 // default
	if thresholdStr := r.URL.Query().Get("eol_threshold_months"); thresholdStr != "" {
		if parsedThreshold, err := strconv.Atoi(thresholdStr); err == nil && parsedThreshold > 0 {
			eolThresholdMonths = parsedThreshold
		}
	}

	limit := 10 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	metrics, err := h.ciService.GetAssetAgingMetrics(r.Context(), eolThresholdMonths, limit)
	if err != nil {
		h.logger.ErrorService("dashboard", "GET_ASSET_AGING_METRICS", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get asset aging metrics")
		return
	}

	h.writeJSON(w, http.StatusOK, metrics)
}

// GetRiskMetrics returns risk assessment metrics for the dashboard
// @Summary Get Risk Metrics
// @Description Returns risk assessment including SPOFs, critical assets, and compliance issues
// @Tags dashboard
// @Produce json
// @Param limit query int false "Max number of high-risk CIs to return (default: 10)"
// @Success 200 {object} ci.RiskMetrics
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/dashboard/risk-metrics [get]
func (h *CIHandlers) GetRiskMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 10 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	metrics, err := h.ciService.GetRiskMetrics(r.Context(), limit)
	if err != nil {
		h.logger.ErrorService("dashboard", "GET_RISK_METRICS", err, nil)
		h.writeError(w, http.StatusInternalServerError, "Failed to get risk metrics")
		return
	}

	h.writeJSON(w, http.StatusOK, metrics)
}