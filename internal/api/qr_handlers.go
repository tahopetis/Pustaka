package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/pustaka/pustaka/internal/ci"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type QRHandlers struct {
	*Handler
	ciService *ci.Service
	logger    *pustakaLogger.Logger
	baseURL   string
}

func NewQRHandlers(handler *Handler, ciService *ci.Service, logger *pustakaLogger.Logger, baseURL string) *QRHandlers {
	return &QRHandlers{
		Handler:   handler,
		ciService: ciService,
		logger:    logger,
		baseURL:   baseURL,
	}
}

// GetCIQRCode generates and returns a QR code for a Configuration Item
// @Summary Generate QR code for a CI
// @Description Generate a QR code that contains the public URL for a CI
// @Tags qr
// @Produce json
// @Param id path string true "Configuration item ID"
// @Param size query int false "QR code size in pixels (default: 256)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/qr/ci/{id} [get]
func (h *QRHandlers) GetCIQRCode(w http.ResponseWriter, r *http.Request) {
	// Get CI ID from URL
	ciIDStr := chi.URLParam(r, "id")
	ciID, err := uuid.Parse(ciIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	// Verify CI exists
	_, err = h.ciService.GetCI(r.Context(), ciID)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_id", ciIDStr).Msg("CI not found")
		h.writeError(w, http.StatusNotFound, "Configuration item not found")
		return
	}

	// Get QR code size from query parameter (default 256)
	size := 256
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if parsedSize, err := parseIntFromString(sizeStr); err == nil && parsedSize >= 100 && parsedSize <= 1000 {
			size = parsedSize
		}
	}

	// Generate public URL for this CI
	publicURL := h.baseURL + "/public/ci/" + ciIDStr

	// Generate QR code
	qrCode, err := qrcode.Encode(publicURL, qrcode.Medium, size)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_id", ciIDStr).Msg("Failed to generate QR code")
		h.writeError(w, http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	// Convert to base64
	qrBase64 := base64.StdEncoding.EncodeToString(qrCode)

	// Return QR code data
	response := map[string]interface{}{
		"ci_id":     ciID,
		"public_url": publicURL,
		"qr_code":   "data:image/png;base64," + qrBase64,
		"size":      size,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCIQRCodeImage generates and returns a QR code as PNG image
// @Summary Get QR code as PNG image
// @Description Generate and return a QR code as a PNG image for a CI
// @Tags qr
// @Produce image/png
// @Param id path string true "Configuration item ID"
// @Param size query int false "QR code size in pixels (default: 256)"
// @Success 200 {file} binary "PNG image"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/qr/ci/{id}/image [get]
func (h *QRHandlers) GetCIQRCodeImage(w http.ResponseWriter, r *http.Request) {
	// Get CI ID from URL
	ciIDStr := chi.URLParam(r, "id")
	ciID, err := uuid.Parse(ciIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	// Verify CI exists
	_, err = h.ciService.GetCI(r.Context(), ciID)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_id", ciIDStr).Msg("CI not found")
		h.writeError(w, http.StatusNotFound, "Configuration item not found")
		return
	}

	// Get QR code size from query parameter (default 256)
	size := 256
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if parsedSize, err := parseIntFromString(sizeStr); err == nil && parsedSize >= 100 && parsedSize <= 1000 {
			size = parsedSize
		}
	}

	// Generate public URL for this CI
	publicURL := h.baseURL + "/public/ci/" + ciIDStr

	// Generate QR code
	qrCode, err := qrcode.Encode(publicURL, qrcode.Medium, size)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_id", ciIDStr).Msg("Failed to generate QR code")
		h.writeError(w, http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	// Set headers for PNG download
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=ci-"+ciIDStr+"-qr.png")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Write QR code image
	w.Write(qrCode)
}

// GetPublicCI retrieves CI details without authentication
// @Summary Get public CI details
// @Description Get a configuration item's details without authentication (for QR code scanning)
// @Tags public
// @Produce json
// @Param id path string true "Configuration item ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /public/ci/{id} [get]
func (h *QRHandlers) GetPublicCI(w http.ResponseWriter, r *http.Request) {
	// Get CI ID from URL
	ciIDStr := chi.URLParam(r, "id")
	ciID, err := uuid.Parse(ciIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid CI ID")
		return
	}

	// Get CI from service
	ciItem, err := h.ciService.GetCI(r.Context(), ciID)
	if err != nil {
		h.logger.Error().Err(err).Str("ci_id", ciIDStr).Msg("CI not found")
		h.writeError(w, http.StatusNotFound, "Configuration item not found")
		return
	}

	// Create public response (exclude sensitive information)
	publicCI := map[string]interface{}{
		"id":              ciItem.ID,
		"name":            ciItem.Name,
		"ci_type":         ciItem.CIType,
		"attributes":      ciItem.Attributes,
		"tags":            ciItem.Tags,
		"lifecycle_status": ciItem.LifecycleStatus,
		"created_at":      ciItem.CreatedAt,
		"updated_at":      ciItem.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, publicCI)
}

// Helper function to parse int from string
func parseIntFromString(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
