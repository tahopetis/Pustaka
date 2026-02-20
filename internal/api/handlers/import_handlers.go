package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
	"github.com/pustaka/pustaka/internal/ea"
)

// ImportHandlers handles EA entity import operations
type ImportHandlers struct {
	importService *ea.ImportService
	logger        *pustakaLogger.Logger
}

// NewImportHandlers creates a new import handlers instance
func NewImportHandlers(importService *ea.ImportService, logger *pustakaLogger.Logger) *ImportHandlers {
	return &ImportHandlers{
		importService: importService,
		logger:        logger,
	}
}

// GenerateTemplateRequest represents the request to generate a CSV template
type GenerateTemplateRequest struct {
	CIType string `json:"ci_type"`
}

// ValidateImportRequest represents the request to validate an import file
type ValidateImportRequest struct {
	File   []byte `json:"file"`
	CIType string `json:"ci_type"`
}

// ExecuteImportRequest represents the request to execute an import
type ExecuteImportRequest struct {
	File   []byte `json:"file"`
	CIType string `json:"ci_type"`
}

// GenerateImportTemplate handles POST /api/v1/ea/import/template
func (h *ImportHandlers) GenerateImportTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	// Parse request body
	var req GenerateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate CI type
	if req.CIType == "" {
		http.Error(w, "ci_type is required", http.StatusBadRequest)
		return
	}

	// Validate it's an EA CI type
	if !strings.HasPrefix(req.CIType, "EA.") {
		http.Error(w, "CI type must be an EA type (starts with 'EA.')", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("ci_type", req.CIType).
		Msg("Generating import template")

	// Generate template
	csvBytes, err := h.importService.GenerateTemplate(ctx, req.CIType)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_type", req.CIType).Msg("Failed to generate template")
		http.Error(w, fmt.Sprintf("Failed to generate template: %v", err), http.StatusBadRequest)
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("ea-%s-template.csv", strings.ReplaceAll(req.CIType, ".", "-"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w.WriteHeader(http.StatusOK)
	w.Write(csvBytes)

	log.Info().
		Str("ci_type", req.CIType).
		Str("filename", filename).
		Msg("Import template generated successfully")
}

// ValidateImport handles POST /api/v1/ea/import/validate
func (h *ImportHandlers) ValidateImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get CI type from form
	ciType := r.FormValue("ci_type")
	if ciType == "" {
		http.Error(w, "ci_type is required", http.StatusBadRequest)
		return
	}

	// Validate it's an EA CI type
	if !strings.HasPrefix(ciType, "EA.") {
		http.Error(w, "CI type must be an EA type (starts with 'EA.')", http.StatusBadRequest)
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file")
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("ci_type", ciType).
		Str("filename", header.Filename).
		Int("file_size", len(fileBytes)).
		Msg("Validating import file")

	// Parse CSV
	rows, err := h.importService.ParseCSV(fileBytes, ciType)
	if err != nil {
		h.logger.Error().Err(err).Str("filename", header.Filename).Msg("Failed to parse CSV")
		http.Error(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusUnprocessableEntity)
		return
	}

	// Validate import
	result, err := h.importService.ValidateImport(ctx, rows, ciType)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_type", ciType).Msg("Validation failed")
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return validation result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	log.Info().
		Str("ci_type", ciType).
		Int("total_rows", result.TotalRows).
		Int("success_count", result.SuccessCount).
		Int("error_count", result.ErrorCount).
		Msg("Import validation completed")
}

// ExecuteImport handles POST /api/v1/ea/import/execute
func (h *ImportHandlers) ExecuteImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	// Get user ID from context (set by JWT middleware)
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok || userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := uuid.Parse(userIDStr); err != nil {
		h.logger.Error().Err(err).Msg("Invalid user ID in context")
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse multipart form")
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get file from form")
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get CI type from form
	ciType := r.FormValue("ci_type")
	if ciType == "" {
		http.Error(w, "ci_type is required", http.StatusBadRequest)
		return
	}

	// Validate it's an EA CI type
	if !strings.HasPrefix(ciType, "EA.") {
		http.Error(w, "CI type must be an EA type (starts with 'EA.')", http.StatusBadRequest)
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read file")
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("ci_type", ciType).
		Str("filename", header.Filename).
		Str("user_id", userIDStr).
		Int("file_size", len(fileBytes)).
		Msg("Executing import")

	// Parse CSV
	rows, err := h.importService.ParseCSV(fileBytes, ciType)
	if err != nil {
		h.logger.Error().Err(err).Str("filename", header.Filename).Msg("Failed to parse CSV")
		http.Error(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusUnprocessableEntity)
		return
	}

	// Execute import
	result, err := h.importService.BulkCreateEntities(ctx, rows, ciType, userIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_type", ciType).Msg("Import failed")
		http.Error(w, fmt.Sprintf("Import failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return import result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	log.Info().
		Str("ci_type", ciType).
		Str("user_id", userIDStr).
		Int("total_rows", result.TotalRows).
		Int("success_count", result.SuccessCount).
		Int("error_count", result.ErrorCount).
		Msg("Import executed successfully")
}

// DownloadErrorCSV handles GET /api/v1/ea/import/errors/download
func (h *ImportHandlers) DownloadErrorCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	// Parse error data from request body (sent as JSON)
	var errors []ea.ImportError
	if err := json.NewDecoder(r.Body).Decode(&errors); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode errors")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(errors) == 0 {
		http.Error(w, "No errors to export", http.StatusBadRequest)
		return
	}

	log.Info().
		Int("error_count", len(errors)).
		Msg("Generating error CSV")

	// Create CSV buffer
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"import-errors.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"RowNumber", "FieldName", "ErrorMessage", "ExpectedFormat", "ActualValue"}
	if err := writer.Write(header); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write CSV header")
		return
	}

	// Write error rows
	for _, errItem := range errors {
		row := []string{
			strconv.Itoa(errItem.RowNumber),
			errItem.FieldName,
			errItem.ErrorMessage,
			errItem.ExpectedFormat,
			errItem.ActualValue,
		}
		if err := writer.Write(row); err != nil {
			h.logger.Error().Err(err).Msg("Failed to write CSV row")
			return
		}
	}

	log.Info().
		Int("error_count", len(errors)).
		Msg("Error CSV generated successfully")
}

// GetImportStatus handles GET /api/v1/ea/import/status/{batch_id}
// This is a placeholder for future async import functionality
func (h *ImportHandlers) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zerolog.Ctx(ctx)

	batchID := chi.URLParam(r, "batch_id")
	if batchID == "" {
		http.Error(w, "batch_id is required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("batch_id", batchID).
		Msg("Getting import status")

	// Placeholder response - async import not yet implemented
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"batch_id":     batchID,
		"status":       "completed",
		"progress":     100,
		"total_rows":   0,
		"processed":    0,
		"success_count": 0,
		"error_count":  0,
		"message":      "Import completed",
	})
}
