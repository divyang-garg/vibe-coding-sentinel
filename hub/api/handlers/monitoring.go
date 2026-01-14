// Package handlers - Monitoring and error handling HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// MonitoringHandler handles monitoring and error-related HTTP requests
type MonitoringHandler struct {
	BaseHandler
	MonitoringService services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler with dependencies
func NewMonitoringHandler(monitoringService services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		MonitoringService: monitoringService,
	}
}

// GetErrorDashboard handles GET /api/v1/monitoring/errors/dashboard
func (h *MonitoringHandler) GetErrorDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.MonitoringService.GetErrorDashboard(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, dashboard)
}

// GetErrorAnalysis handles GET /api/v1/monitoring/errors/analysis
func (h *MonitoringHandler) GetErrorAnalysis(w http.ResponseWriter, r *http.Request) {
	analysis, err := h.MonitoringService.GetErrorAnalysis(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, analysis)
}

// ClassifyError handles POST /api/v1/monitoring/errors/classify
func (h *MonitoringHandler) ClassifyError(w http.ResponseWriter, r *http.Request) {
	var req models.ErrorClassification
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	classification, err := h.MonitoringService.ClassifyError(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, classification)
}

// GetErrorStats handles GET /api/v1/monitoring/errors/stats
func (h *MonitoringHandler) GetErrorStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.MonitoringService.GetErrorStats(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, stats)
}

// ReportError handles POST /api/v1/monitoring/errors/report
func (h *MonitoringHandler) ReportError(w http.ResponseWriter, r *http.Request) {
	var req models.ErrorReport
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	err := h.MonitoringService.ReportError(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusAccepted, map[string]string{
		"status": "error_reported",
	})
}

// GetHealthMetrics handles GET /api/v1/monitoring/health
func (h *MonitoringHandler) GetHealthMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.MonitoringService.GetHealthMetrics(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, metrics)
}

// GetPerformanceMetrics handles GET /api/v1/monitoring/performance
func (h *MonitoringHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.MonitoringService.GetPerformanceMetrics(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, metrics)
}
