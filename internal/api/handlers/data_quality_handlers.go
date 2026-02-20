package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pustaka/pustaka/internal/repository"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// DataQualityHandlers handles data quality HTTP requests
type DataQualityHandlers struct {
	qualityRepo *repository.EADataQualityRepository
	logger      *pustakaLogger.Logger
}

// NewDataQualityHandlers creates a new data quality handlers instance
func NewDataQualityHandlers(qualityRepo *repository.EADataQualityRepository, logger *pustakaLogger.Logger) *DataQualityHandlers {
	return &DataQualityHandlers{
		qualityRepo: qualityRepo,
		logger:      logger,
	}
}

// DataQualityResponse represents the API response for data quality metrics
type DataQualityResponse struct {
	TotalEntities          int64              `json:"total_entities"`
	CompletenessPct        float64            `json:"completeness_pct"`
	StaleEntitiesCount     int                `json:"stale_entities_count"`
	EntitiesWithErrorsCount int               `json:"entities_with_errors_count"`
	LifecycleBreakdown     map[string]int     `json:"lifecycle_breakdown"`
	ErrorBreakdownByDomain map[string]int     `json:"error_breakdown_by_domain"`
	GeneratedAt            string             `json:"generated_at"`
}

// StaleEntitiesResponse represents the API response for stale entities list
type StaleEntitiesResponse struct {
	Entities []repository.EAEntitySummary `json:"entities"`
	Total    int                         `json:"total"`
	Query    repository.StaleEntityCriteria `json:"query"`
}

// EntitiesWithErrorsResponse represents the API response for entities with errors list
type EntitiesWithErrorsResponse struct {
	Entities []repository.EAEntitySummary `json:"entities"`
	Total    int                         `json:"total"`
	Domain   string                      `json:"domain,omitempty"`
}

// GetDataQualityMetrics handles GET /api/v1/ea/data-quality
func (h *DataQualityHandlers) GetDataQualityMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse optional query parameters
	domain := r.URL.Query().Get("domain")

	// Get overall metrics
	metrics, err := h.qualityRepo.GetOverallMetrics(r.Context())
	if err != nil {
		h.logger.ErrorService("data_quality", "get_metrics", err, map[string]interface{}{
			"domain":  domain,
			"elapsed": time.Since(startTime).Milliseconds(),
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get data quality metrics")
		return
	}

	// Filter by domain if requested
	if domain != "" {
		// Get domain-specific completeness
		completeness, err := h.qualityRepo.GetCompletenessMetrics(r.Context(), domain)
		if err != nil {
			h.logger.ErrorService("data_quality", "get_completeness", err, map[string]interface{}{
				"domain": domain,
			})
		} else {
			metrics.CompletenessPct = completeness
		}

		// Get domain-specific lifecycle breakdown
		lifecycleBreakdown, err := h.qualityRepo.GetLifecycleStatusBreakdown(r.Context(), domain)
		if err != nil {
			h.logger.ErrorService("data_quality", "get_lifecycle_breakdown", err, map[string]interface{}{
				"domain": domain,
			})
		} else {
			metrics.LifecycleBreakdown = lifecycleBreakdown
		}

		// Get entities with errors for this domain
		_, errorCount, err := h.qualityRepo.GetEntitiesWithErrors(r.Context(), domain)
		if err != nil {
			h.logger.ErrorService("data_quality", "get_error_entities", err, map[string]interface{}{
				"domain": domain,
			})
		} else {
			metrics.EntitiesWithErrorsCount = errorCount
		}

		// Filter error breakdown to only show this domain
		domainErrorCount := metrics.ErrorBreakdownByDomain[domain]
		metrics.ErrorBreakdownByDomain = map[string]int{domain: domainErrorCount}
	}

	response := DataQualityResponse{
		TotalEntities:          metrics.TotalEntities,
		CompletenessPct:        metrics.CompletenessPct,
		StaleEntitiesCount:     metrics.StaleEntitiesCount,
		EntitiesWithErrorsCount: metrics.EntitiesWithErrorsCount,
		LifecycleBreakdown:     metrics.LifecycleBreakdown,
		ErrorBreakdownByDomain: metrics.ErrorBreakdownByDomain,
		GeneratedAt:            time.Now().Format(time.RFC3339),
	}

	h.logger.Info().Str("component", "data_quality").
		Str("action", "get_metrics").
		Str("domain", domain).
		Int64("total_entities", metrics.TotalEntities).
		Float64("completeness_pct", metrics.CompletenessPct).
		Int64("elapsed_ms", time.Since(startTime).Milliseconds()).
		Msg("Data quality metrics retrieved")

	h.writeJSON(w, http.StatusOK, response)
}

// GetStaleEntities handles GET /api/v1/ea/data-quality/stale
func (h *DataQualityHandlers) GetStaleEntities(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	criteria := repository.StaleEntityCriteria{
		DaysThreshold:     90, // Default
		IncludeIncomplete: true,
	}

	if daysStr := r.URL.Query().Get("days_threshold"); daysStr != "" {
		var days int
		if err := json.Unmarshal([]byte(daysStr), &days); err == nil && days > 0 {
			criteria.DaysThreshold = days
		}
	}

	if includeIncompleteStr := r.URL.Query().Get("include_incomplete"); includeIncompleteStr != "" {
		var includeIncomplete bool
		if err := json.Unmarshal([]byte(includeIncompleteStr), &includeIncomplete); err == nil {
			criteria.IncludeIncomplete = includeIncomplete
		}
	}

	entities, total, err := h.qualityRepo.GetStaleEntities(r.Context(), criteria)
	if err != nil {
		h.logger.ErrorService("data_quality", "get_stale_entities", err, map[string]interface{}{
			"criteria": criteria,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get stale entities")
		return
	}

	response := StaleEntitiesResponse{
		Entities: entities,
		Total:    total,
		Query:    criteria,
	}

	h.logger.Info().Str("component", "data_quality").
		Str("action", "get_stale_entities").
		Int("count", len(entities)).
		Int("total", total).
		Int64("elapsed_ms", time.Since(startTime).Milliseconds()).
		Msg("Stale entities retrieved")

	h.writeJSON(w, http.StatusOK, response)
}

// GetEntitiesWithErrors handles GET /api/v1/ea/data-quality/errors
func (h *DataQualityHandlers) GetEntitiesWithErrors(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse query parameters
	domain := r.URL.Query().Get("domain")

	entities, total, err := h.qualityRepo.GetEntitiesWithErrors(r.Context(), domain)
	if err != nil {
		h.logger.ErrorService("data_quality", "get_entities_with_errors", err, map[string]interface{}{
			"domain": domain,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get entities with errors")
		return
	}

	response := EntitiesWithErrorsResponse{
		Entities: entities,
		Total:    total,
		Domain:   domain,
	}

	h.logger.Info().Str("component", "data_quality").
		Str("action", "get_entities_with_errors").
		Str("domain", domain).
		Int("count", len(entities)).
		Int("total", total).
		Int64("elapsed_ms", time.Since(startTime).Milliseconds()).
		Msg("Entities with errors retrieved")

	h.writeJSON(w, http.StatusOK, response)
}

// GetLifecycleBreakdown handles GET /api/v1/ea/data-quality/lifecycle
func (h *DataQualityHandlers) GetLifecycleBreakdown(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	domain := r.URL.Query().Get("domain")

	breakdown, err := h.qualityRepo.GetLifecycleStatusBreakdown(r.Context(), domain)
	if err != nil {
		h.logger.ErrorService("data_quality", "get_lifecycle_breakdown", err, map[string]interface{}{
			"domain": domain,
		})
		h.writeError(w, http.StatusInternalServerError, "Failed to get lifecycle breakdown")
		return
	}

	response := map[string]interface{}{
		"breakdown":   breakdown,
		"domain":      domain,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	h.logger.Info().Str("component", "data_quality").
		Str("action", "get_lifecycle_breakdown").
		Str("domain", domain).
		Int64("elapsed_ms", time.Since(startTime).Milliseconds()).
		Msg("Lifecycle breakdown retrieved")

	h.writeJSON(w, http.StatusOK, response)
}

// Helper functions

func (h *DataQualityHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *DataQualityHandlers) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
