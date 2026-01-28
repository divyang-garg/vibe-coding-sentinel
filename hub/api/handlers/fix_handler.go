// Package handlers - Fix application HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// FixHandler handles fix application requests
type FixHandler struct {
	BaseHandler
	FixService services.FixService
}

// NewFixHandler creates a new fix handler with dependencies
func NewFixHandler(fixService services.FixService) *FixHandler {
	return &FixHandler{
		FixService: fixService,
	}
}

// ApplyFix handles POST /api/v1/fix/apply
// Applies automated fixes to code based on fix type (security, style, performance)
func (h *FixHandler) ApplyFix(w http.ResponseWriter, r *http.Request) {
	var req models.ApplyFixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Code == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "code",
			Message: "Code is required",
		}, http.StatusBadRequest)
		return
	}

	if req.Language == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "language",
			Message: "Language is required",
		}, http.StatusBadRequest)
		return
	}

	if req.FixType == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "fix_type",
			Message: "Fix type is required (security, style, performance)",
		}, http.StatusBadRequest)
		return
	}

	// Validate fix type
	validFixTypes := map[string]bool{
		"security":    true,
		"style":       true,
		"performance": true,
	}
	if !validFixTypes[req.FixType] {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "fix_type",
			Message: "Invalid fix type. Must be one of: security, style, performance",
		}, http.StatusBadRequest)
		return
	}

	// Apply fixes using the fix service
	result, err := h.FixService.ApplyFix(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
