package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

type AuditHandlers struct {
	*Handler
	auditService *ci.AuditService
}

// NewAuditHandlers creates new audit handlers
func NewAuditHandlers(handler *Handler, auditService *ci.AuditService) *AuditHandlers {
	return &AuditHandlers{
		Handler:      handler,
		auditService: auditService,
	}
}

// ListAuditLogs handles GET /audit/logs
func (h *AuditHandlers) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	filters := h.buildAuditLogFilters(r)

	auditLogs, err := h.auditService.ListAuditLogs(r.Context(), filters)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list audit logs")
		return
	}

	h.writeJSON(w, http.StatusOK, auditLogs)
}

// GetAuditLog handles GET /audit/logs/{id}
func (h *AuditHandlers) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid audit log ID")
		return
	}

	auditLog, err := h.auditService.GetAuditLog(r.Context(), id)
	if err != nil {
		if err.Error() == "audit log not found" {
			h.writeError(w, http.StatusNotFound, "Audit log not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit log")
		return
	}

	h.writeJSON(w, http.StatusOK, auditLog)
}

// GetAuditStats handles GET /audit/stats
func (h *AuditHandlers) GetAuditStats(w http.ResponseWriter, r *http.Request) {
	filters := h.buildAuditLogFilters(r)

	stats, err := h.auditService.GetAuditStats(r.Context(), filters)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit statistics")
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// DeleteAuditLog handles DELETE /audit/logs/{id}
func (h *AuditHandlers) DeleteAuditLog(w http.ResponseWriter, r *http.Request) {
	id, err := h.getUUIDParam(r, "id")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid audit log ID")
		return
	}

	err = h.auditService.DeleteAuditLog(r.Context(), id)
	if err != nil {
		if err.Error() == "audit log not found" {
			h.writeError(w, http.StatusNotFound, "Audit log not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to delete audit log")
		return
	}

	h.writeJSON(w, http.StatusNoContent, nil)
}

// CleanupOldAuditLogs handles DELETE /audit/cleanup
func (h *AuditHandlers) CleanupOldAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse retention period (optional query parameter)
	retentionDays := h.getQueryInt(r, "retention_days", 365)

	var cutoffDate time.Time
	if retentionDays > 0 {
		cutoffDate = time.Now().AddDate(0, 0, -retentionDays)
	} else {
		// Use default retention period from service
		retentionPeriod := h.auditService.GetRetentionPeriod()
		cutoffDate = time.Now().Add(-retentionPeriod)
	}

	deletedCount, err := h.auditService.DeleteOldAuditLogs(r.Context(), cutoffDate)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to cleanup old audit logs")
		return
	}

	response := map[string]interface{}{
		"deleted_count": deletedCount,
		"cutoff_date":   cutoffDate,
		"retention_days": retentionDays,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ExportAuditLogs handles GET /audit/export
func (h *AuditHandlers) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	filters := h.buildAuditLogFilters(r)

	// Set a higher limit for export
	filters.Limit = 10000

	auditLogs, err := h.auditService.ListAuditLogs(r.Context(), filters)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to export audit logs")
		return
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=audit_logs.csv")

	// Write CSV header
	w.Write([]byte("ID,Entity Type,Entity ID,Action,Performed By,Timestamp,IP Address,User Agent\n"))

	// Write audit log rows
	for _, log := range auditLogs.AuditLogs {
		entityIDStr := ""
		if log.EntityID != nil {
			entityIDStr = log.EntityID.String()
		}

		row := []string{
			log.ID.String(),
			log.EntityType,
			entityIDStr,
			log.Action,
			log.PerformedBy.String(),
			log.Timestamp.Format(time.RFC3339),
			log.IPAddress,
			log.UserAgent,
		}

		// Write row as CSV (simple implementation)
		w.Write([]byte("\"" + row[0] + "\",\"" + row[1] + "\",\"" + row[2] + "\",\"" + row[3] + "\",\"" + row[4] + "\",\"" + row[5] + "\",\"" + row[6] + "\",\"" + row[7] + "\"\n"))
	}
}

// buildAuditLogFilters builds filters from query parameters
func (h *AuditHandlers) buildAuditLogFilters(r *http.Request) ci.AuditLogFilters {
	filters := ci.AuditLogFilters{
		EntityType: h.getQueryString(r, "entity_type"),
		Action:     h.getQueryString(r, "action"),
		Search:     h.getQueryString(r, "search"),
		Sort:       h.getQueryString(r, "sort"),
		Order:      h.getQueryString(r, "order"),
		Page:       h.getQueryInt(r, "page", 1),
		Limit:      h.getQueryInt(r, "limit", 50),
	}

	// Parse entity_id if provided
	if entityIDStr := h.getQueryString(r, "entity_id"); entityIDStr != "" {
		if entityID, err := uuid.Parse(entityIDStr); err == nil {
			filters.EntityID = &entityID
		}
	}

	// Parse performed_by if provided
	if performedByStr := h.getQueryString(r, "performed_by"); performedByStr != "" {
		if performedBy, err := uuid.Parse(performedByStr); err == nil {
			filters.PerformedBy = &performedBy
		}
	}

	// Parse date ranges
	if startDateStr := h.getQueryString(r, "start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := h.getQueryString(r, "end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.EndDate = &endDate
		}
	}

	return filters
}