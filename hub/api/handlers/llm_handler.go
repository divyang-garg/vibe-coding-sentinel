// Package handlers - LLM configuration HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// LLMHandler handles LLM configuration requests
type LLMHandler struct {
	BaseHandler
	LLMService services.LLMService
}

// NewLLMHandler creates a new LLM handler with dependencies
func NewLLMHandler(llmService services.LLMService) *LLMHandler {
	return &LLMHandler{
		LLMService: llmService,
	}
}

// ValidateLLMConfig handles POST /api/v1/llm/validate-config
// Validates LLM configuration including API keys, models, and cost optimization settings
func (h *LLMHandler) ValidateLLMConfig(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateLLMConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Config.Provider == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "config.provider",
			Message: "Provider is required",
		}, http.StatusBadRequest)
		return
	}

	if req.Config.APIKey == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "config.api_key",
			Message: "API key is required",
		}, http.StatusBadRequest)
		return
	}

	if req.Config.Model == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "config.model",
			Message: "Model is required",
		}, http.StatusBadRequest)
		return
	}

	// Validate configuration using service
	result, err := h.LLMService.ValidateConfig(r.Context(), req.Config)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
