// Package handlers - Metrics HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	BaseHandler
	MetricsService services.MetricsService
}

// NewMetricsHandler creates a new metrics handler with dependencies
func NewMetricsHandler(metricsService services.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		MetricsService: metricsService,
	}
}

// GetCacheMetrics handles GET /api/v1/metrics/cache
// Returns cache metrics for the current project
func (h *MetricsHandler) GetCacheMetrics(w http.ResponseWriter, r *http.Request) {
	// Get project from context
	project, err := h.GetProjectFromContext(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project",
			Message: "Project not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get cache metrics from service
	result, err := h.MetricsService.GetCacheMetrics(r.Context(), project.ID)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetCostMetrics handles GET /api/v1/metrics/cost
// Returns cost metrics for the current project
func (h *MetricsHandler) GetCostMetrics(w http.ResponseWriter, r *http.Request) {
	// Get project from context
	project, err := h.GetProjectFromContext(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project",
			Message: "Project not found in context",
		}, http.StatusUnauthorized)
		return
	}

	// Get cost metrics from service
	result, err := h.MetricsService.GetCostMetrics(r.Context(), project.ID)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
