// Package handlers - Architecture analysis HTTP handler
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// ArchitectureHandler handles architecture analysis HTTP requests
type ArchitectureHandler struct {
	BaseHandler
}

// NewArchitectureHandler creates a new architecture handler
func NewArchitectureHandler() *ArchitectureHandler {
	return &ArchitectureHandler{}
}

// AnalyzeArchitecture handles POST /api/v1/analyze/architecture
func (h *ArchitectureHandler) AnalyzeArchitecture(w http.ResponseWriter, r *http.Request) {
	var req services.ArchitectureAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "files",
			Message: "At least one file is required",
		}, http.StatusBadRequest)
		return
	}

	for _, f := range req.Files {
		if strings.TrimSpace(f.Path) == "" {
			h.WriteErrorResponse(w, &models.ValidationError{
				Field:   "files",
				Message: "file path is required for each file",
			}, http.StatusBadRequest)
			return
		}
	}

	result := services.AnalyzeArchitecture(req)
	h.WriteJSONResponse(w, http.StatusOK, result)
}
