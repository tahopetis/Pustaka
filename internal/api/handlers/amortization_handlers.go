package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/amortization"
	"github.com/pustaka/pustaka/internal/api/middleware"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// AmortizationHandler handles amortization HTTP requests
type AmortizationHandler struct {
	service amortization.Service
	logger  *pustakaLogger.Logger
}

// NewAmortizationHandler creates a new amortization handler
func NewAmortizationHandler(service amortization.Service, logger *pustakaLogger.Logger) *AmortizationHandler {
	return &AmortizationHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers amortization routes
func (h *AmortizationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/amortization", func(r chi.Router) {
		// Configuration Items with amortization
		r.Get("/configuration-items", h.ListAmortizableCIs)
		r.Get("/configuration-items/{ciId}", h.GetAmortizationDetails)
		r.Put("/configuration-items/{ciId}", h.UpdateAmortizationConfig)

		// Ledger management
		r.Get("/ledger", h.GetLedgerEntries)
		r.Get("/ledger/{entryId}", h.GetLedgerEntry)
		r.Post("/adjustments", h.CreateAdjustment)

		// Amortization runs
		r.Get("/runs", h.ListAmortizationRuns)
		r.Get("/runs/{runId}", h.GetAmortizationRun)
		r.Post("/runs", h.TriggerManualRun)

		// Reports and summaries
		r.Get("/summaries", h.GetAmortizationSummaries)
		r.Get("/reports/depreciation-schedule", h.GenerateDepreciationSchedule)

		// Restructuring (useful life changes with prospective recalculation)
		r.Post("/restructuring/preview", h.PreviewRestructuring)
		r.Post("/restructuring", h.ExecuteRestructuring)
	})
}

// ListAmortizableCIs handles GET /amortization/configuration-items
func (h *AmortizationHandler) ListAmortizableCIs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters, err := h.parseCIFilters(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Call service
	result, err := h.service.ListAmortizableCIs(r.Context(), filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// GetAmortizationDetails handles GET /amortization/configuration-items/{ciId}
func (h *AmortizationHandler) GetAmortizationDetails(w http.ResponseWriter, r *http.Request) {
	// Parse CI ID
	ciID, err := uuid.Parse(chi.URLParam(r, "ciId"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID", err)
		return
	}

	// Call service
	result, err := h.service.GetAmortizationDetails(r.Context(), ciID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// UpdateAmortizationConfig handles PUT /amortization/configuration-items/{ciId}
func (h *AmortizationHandler) UpdateAmortizationConfig(w http.ResponseWriter, r *http.Request) {
	// Parse CI ID
	ciID, err := uuid.Parse(chi.URLParam(r, "ciId"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID", err)
		return
	}

	// Parse request body
	var req amortization.UpdateAmortizationConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from context
	userID, err := h.getUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Call service
	result, err := h.service.UpdateAmortizationConfig(r.Context(), ciID, &req, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// GetLedgerEntries handles GET /amortization/ledger
func (h *AmortizationHandler) GetLedgerEntries(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters, err := h.parseLedgerFilters(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Call service
	result, err := h.service.GetLedgerEntries(r.Context(), filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// GetLedgerEntry handles GET /amortization/ledger/{entryId}
func (h *AmortizationHandler) GetLedgerEntry(w http.ResponseWriter, r *http.Request) {
	// Parse entry ID
	entryID, err := uuid.Parse(chi.URLParam(r, "entryId"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid entry ID", err)
		return
	}

	// Call service
	result, err := h.service.GetLedgerEntry(r.Context(), entryID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// CreateAdjustment handles POST /amortization/adjustments
func (h *AmortizationHandler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req amortization.CreateAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from context
	userID, err := h.getUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Call service
	result, err := h.service.CreateAdjustment(r.Context(), &req, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusCreated, result)
}

// ListAmortizationRuns handles GET /amortization/runs
func (h *AmortizationHandler) ListAmortizationRuns(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters, err := h.parseRunFilters(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Call service
	result, err := h.service.ListAmortizationRuns(r.Context(), filters)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// GetAmortizationRun handles GET /amortization/runs/{runId}
func (h *AmortizationHandler) GetAmortizationRun(w http.ResponseWriter, r *http.Request) {
	// Parse run ID
	runID, err := uuid.Parse(chi.URLParam(r, "runId"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid run ID", err)
		return
	}

	// Call service
	result, err := h.service.GetAmortizationRun(r.Context(), runID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// TriggerManualRun handles POST /amortization/runs
func (h *AmortizationHandler) TriggerManualRun(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req amortization.ManualRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get user ID from context
	userID, err := h.getUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Call service
	result, err := h.service.TriggerManualRun(r.Context(), &req, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"run_id": result.ID,
		"status": "started",
		"message": "Amortization run initiated",
	})
}

// GetAmortizationSummaries handles GET /amortization/summaries
func (h *AmortizationHandler) GetAmortizationSummaries(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req, err := h.parseSummaryRequest(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Call service
	result, err := h.service.GetAmortizationSummaries(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// GenerateDepreciationSchedule handles GET /amortization/reports/depreciation-schedule
func (h *AmortizationHandler) GenerateDepreciationSchedule(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	req, err := h.parseDepreciationScheduleRequest(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Get format parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	// Call service
	result, err := h.service.GenerateDepreciationSchedule(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Handle different response formats
	if format == "csv" {
		h.writeCSV(w, result)
	} else {
		h.writeJSON(w, http.StatusOK, result)
	}
}

// Helper methods

func (h *AmortizationHandler) parseCIFilters(r *http.Request) (*amortization.AmortizableCIFilters, error) {
	search := r.URL.Query().Get("search")
	page := h.parsePage(r.URL.Query().Get("page"))
	pageSize := h.parseLimit(r.URL.Query().Get("limit"))

	filters := &amortization.AmortizableCIFilters{
		Search:    &search,
		SortBy:    func() *string { s := r.URL.Query().Get("sort"); return &s }(),
		SortOrder: func() *string { s := r.URL.Query().Get("order"); return &s }(),
		Page:      &page,
		PageSize:  &pageSize,
	}

	return filters, nil
}

func (h *AmortizationHandler) parseLedgerFilters(r *http.Request) (*amortization.LedgerFilters, error) {
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	page := h.parsePage(r.URL.Query().Get("page"))
	pageSize := h.parseLimit(r.URL.Query().Get("limit"))

	filters := &amortization.LedgerFilters{
		SortBy:    &sort,
		SortOrder: &order,
		Page:      &page,
		PageSize:  &pageSize,
	}

	// Parse CI ID
	if ciIDStr := r.URL.Query().Get("ci_id"); ciIDStr != "" {
		ciID, err := uuid.Parse(ciIDStr)
		if err != nil {
			return nil, err
		}
		filters.CIID = &ciID
	}

	// Parse entry types
	if entryTypesStr := r.URL.Query().Get("entry_types"); entryTypesStr != "" {
		entryTypes := strings.Split(entryTypesStr, ",")
		filters.EntryTypes = entryTypes
	}

	// Parse date range
	if dateFromStr := r.URL.Query().Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return nil, err
		}
		filters.DateFrom = &dateFrom
	}

	if dateToStr := r.URL.Query().Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return nil, err
		}
		filters.DateTo = &dateTo
	}

	return filters, nil
}

func (h *AmortizationHandler) parseRunFilters(r *http.Request) (*amortization.AmortizationRunFilters, error) {
	status := r.URL.Query().Get("status")
	statuses := strings.Split(status, ",")
	page := h.parsePage(r.URL.Query().Get("page"))
	pageSize := h.parseLimit(r.URL.Query().Get("limit"))

	filters := &amortization.AmortizationRunFilters{
		Status:   statuses,
		Page:     &page,
		PageSize: &pageSize,
	}

	// Parse date range
	if dateFromStr := r.URL.Query().Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return nil, err
		}
		filters.DateFrom = &dateFrom
	}

	if dateToStr := r.URL.Query().Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return nil, err
		}
		filters.DateTo = &dateTo
	}

	return filters, nil
}

func (h *AmortizationHandler) parseSummaryRequest(r *http.Request) (*amortization.SummaryRequest, error) {
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "ci_type"
	}

	req := &amortization.SummaryRequest{
		GroupBy: groupBy,
	}

	// Parse date range
	if dateFromStr := r.URL.Query().Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return nil, err
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := r.URL.Query().Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return nil, err
		}
		req.DateTo = &dateTo
	}

	return req, nil
}

func (h *AmortizationHandler) parseDepreciationScheduleRequest(r *http.Request) (*amortization.DepreciationScheduleRequest, error) {
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" || dateToStr == "" {
		return nil, fmt.Errorf("date_from and date_to are required")
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		return nil, err
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		return nil, err
	}

	req := &amortization.DepreciationScheduleRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	// Parse CI IDs
	if ciIDsStr := r.URL.Query().Get("ci_ids"); ciIDsStr != "" {
		// This would need to handle comma-separated UUIDs
		// For simplicity, we'll skip this implementation
	}

	return req, nil
}

func (h *AmortizationHandler) parsePage(pageStr string) int {
	if pageStr == "" {
		return 1
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func (h *AmortizationHandler) parseLimit(limitStr string) int {
	if limitStr == "" {
		return 20
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 20
	}
	return limit
}

func (h *AmortizationHandler) getUserID(r *http.Request) (uuid.UUID, error) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return uuid.Nil, fmt.Errorf("user not found in context")
	}
	return user.UserID, nil
}

// Response writing methods

func (h *AmortizationHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AmortizationHandler) writeCSV(w http.ResponseWriter, schedule *amortization.DepreciationSchedule) {
	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=depreciation_schedule_%s.csv", schedule.ReportID))

	// Write CSV header
	csvWriter := csv.NewWriter(w)
	csvWriter.Write([]string{
		"CI ID", "CI Name", "CI Type", "Period Start", "Period End",
		"Opening Book Value", "Depreciation Amount", "Closing Book Value",
		"Accumulated Depreciation",
	})

	// Write data rows
	for _, entry := range schedule.Schedule {
		csvWriter.Write([]string{
			entry.CIID.String(),
			entry.CIName,
			entry.CIType,
			entry.PeriodStart.Format("2006-01-02"),
			entry.PeriodEnd.Format("2006-01-02"),
			fmt.Sprintf("%.2f", entry.OpeningBookValue),
			fmt.Sprintf("%.2f", entry.DepreciationAmount),
			fmt.Sprintf("%.2f", entry.ClosingBookValue),
			fmt.Sprintf("%.2f", entry.AccumulatedDepreciation),
		})
	}

	csvWriter.Flush()
}

func (h *AmortizationHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error().Err(err).Msg(message)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"code":    http.StatusText(status),
		},
	}

	if err != nil {
		response["error"].(map[string]interface{})["details"] = err.Error()
	}

	h.writeJSON(w, status, response)
}

func (h *AmortizationHandler) handleError(w http.ResponseWriter, err error) {
	// For now, handle as a generic error
	// In a full implementation, you'd handle specific amortization error types
	h.writeError(w, http.StatusInternalServerError, "Internal server error", err)
}

// PreviewRestructuring handles POST /amortization/restructuring/preview
func (h *AmortizationHandler) PreviewRestructuring(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		CIID                uuid.UUID `json:"ci_id"`
		NewUsefulLifeMonths int        `json:"new_useful_life_months"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if req.CIID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "CI ID is required", nil)
		return
	}
	if req.NewUsefulLifeMonths <= 0 {
		h.writeError(w, http.StatusBadRequest, "New useful life months must be positive", nil)
		return
	}

	// Call service
	result, err := h.service.PreviewRestructuring(r.Context(), req.CIID, req.NewUsefulLifeMonths)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}

// ExecuteRestructuring handles POST /amortization/restructuring
func (h *AmortizationHandler) ExecuteRestructuring(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req amortization.RestructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if req.CIID == uuid.Nil {
		h.writeError(w, http.StatusBadRequest, "CI ID is required", nil)
		return
	}
	if req.NewUsefulLifeMonths <= 0 {
		h.writeError(w, http.StatusBadRequest, "New useful life months must be positive", nil)
		return
	}
	if req.Reason == "" {
		h.writeError(w, http.StatusBadRequest, "Reason is required", nil)
		return
	}

	// Get user ID from context
	userID, err := h.getUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	// Call service
	result, err := h.service.RestructureAmortization(r.Context(), &req, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Write response
	h.writeJSON(w, http.StatusOK, result)
}