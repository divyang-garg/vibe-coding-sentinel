// Package handlers - Knowledge management HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// KnowledgeHandler handles knowledge-related HTTP requests
type KnowledgeHandler struct {
	BaseHandler
	KnowledgeService services.KnowledgeService
}

// NewKnowledgeHandler creates a new knowledge handler with dependencies
func NewKnowledgeHandler(knowledgeService services.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		KnowledgeService: knowledgeService,
	}
}

// GapAnalysis handles POST /api/v1/knowledge/gap-analysis
func (h *KnowledgeHandler) GapAnalysis(w http.ResponseWriter, r *http.Request) {
	var req services.GapAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	if req.CodebasePath == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "codebase_path",
			Message: "codebase_path is required",
		}, http.StatusBadRequest)
		return
	}

	report, err := h.KnowledgeService.RunGapAnalysis(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, report)
}

// ListKnowledgeItems handles GET /api/v1/knowledge/items
func (h *KnowledgeHandler) ListKnowledgeItems(w http.ResponseWriter, r *http.Request) {
	var req services.ListKnowledgeItemsRequest

	// Parse query parameters
	req.ProjectID = r.URL.Query().Get("project_id")
	req.Type = r.URL.Query().Get("type")
	req.Status = r.URL.Query().Get("status")

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	items, err := h.KnowledgeService.ListKnowledgeItems(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, items)
}

// CreateKnowledgeItem handles POST /api/v1/knowledge/items
func (h *KnowledgeHandler) CreateKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	var item services.KnowledgeItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	created, err := h.KnowledgeService.CreateKnowledgeItem(r.Context(), item)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, created)
}

// GetKnowledgeItem handles GET /api/v1/knowledge/items/{id}
func (h *KnowledgeHandler) GetKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "id is required",
		}, http.StatusBadRequest)
		return
	}

	item, err := h.KnowledgeService.GetKnowledgeItem(r.Context(), id)
	if err != nil {
		if err.Error() == "knowledge item not found: "+id {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, item)
}

// UpdateKnowledgeItem handles PUT /api/v1/knowledge/items/{id}
func (h *KnowledgeHandler) UpdateKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "id is required",
		}, http.StatusBadRequest)
		return
	}

	var item services.KnowledgeItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	updated, err := h.KnowledgeService.UpdateKnowledgeItem(r.Context(), id, item)
	if err != nil {
		if err.Error() == "knowledge item not found: "+id {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, updated)
}

// DeleteKnowledgeItem handles DELETE /api/v1/knowledge/items/{id}
func (h *KnowledgeHandler) DeleteKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "id is required",
		}, http.StatusBadRequest)
		return
	}

	err := h.KnowledgeService.DeleteKnowledgeItem(r.Context(), id)
	if err != nil {
		if err.Error() == "knowledge item not found: "+id {
			h.WriteErrorResponse(w, err, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusNoContent, nil)
}

// GetBusinessContext handles GET /api/v1/knowledge/business
func (h *KnowledgeHandler) GetBusinessContext(w http.ResponseWriter, r *http.Request) {
	var req services.BusinessContextRequest

	// Parse query parameters
	req.ProjectID = r.URL.Query().Get("project_id")
	req.Feature = r.URL.Query().Get("feature")
	req.Entity = r.URL.Query().Get("entity")

	if keywordsStr := r.URL.Query().Get("keywords"); keywordsStr != "" {
		// Parse comma-separated keywords
		keywords := strings.Split(keywordsStr, ",")
		for i := range keywords {
			keywords[i] = strings.TrimSpace(keywords[i])
		}
		req.Keywords = keywords
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	context, err := h.KnowledgeService.GetBusinessContext(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, context)
}

// SyncKnowledge handles POST /api/v1/knowledge/sync
func (h *KnowledgeHandler) SyncKnowledge(w http.ResponseWriter, r *http.Request) {
	var req services.SyncKnowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "project_id",
			Message: "project_id is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.KnowledgeService.SyncKnowledge(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}
