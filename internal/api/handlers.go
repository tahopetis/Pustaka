package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type Handler struct {
	logger *pustakaLogger.Logger
}

func NewHandler(logger *pustakaLogger.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// Helper functions

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) getUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	vars := mux.Vars(r)
	idStr := vars[param]

	// Fallback: try to extract ID from path manually
	if idStr == "" && param == "id" {
		path := r.URL.Path
		// Extract UUID from path patterns like /api/v1/ci/{id} or /api/v1/ci-types/{id}
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			// Check if the last part looks like a UUID
			lastPart := parts[len(parts)-1]
			if len(lastPart) == 36 && strings.Count(lastPart, "-") == 4 {
				idStr = lastPart
			}
		}
	}

	return uuid.Parse(idStr)
}

func (h *Handler) getIntParam(r *http.Request, param string, defaultValue int) int {
	vars := mux.Vars(r)
	if val, ok := vars[param]; ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (h *Handler) getQueryInt(r *http.Request, param string, defaultValue int) int {
	if val := r.URL.Query().Get(param); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func (h *Handler) getQueryString(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

func (h *Handler) getQueryStrings(r *http.Request, param string) []string {
	return r.URL.Query()[param]
}