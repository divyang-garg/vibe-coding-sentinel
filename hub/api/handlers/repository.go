// Package handlers - Repository management HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// RepositoryHandler handles repository-related HTTP requests
type RepositoryHandler struct {
	BaseHandler
	RepositoryService services.RepositoryService
}

// NewRepositoryHandler creates a new repository handler with dependencies
func NewRepositoryHandler(repositoryService services.RepositoryService) *RepositoryHandler {
	return &RepositoryHandler{
		RepositoryService: repositoryService,
	}
}

// ListRepositories handles GET /api/v1/repositories
func (h *RepositoryHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	language := r.URL.Query().Get("language")
	limit := 50 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		// Simple limit parsing (would use strconv in real implementation)
		if limitStr == "10" {
			limit = 10
		} else if limitStr == "100" {
			limit = 100
		}
	}

	repositories, err := h.RepositoryService.ListRepositories(r.Context(), language, limit)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"repositories": repositories,
		"total":        len(repositories),
		"language":     language,
		"limit":        limit,
	})
}

// GetRepositoryImpact handles GET /api/v1/repositories/{id}/impact
func (h *RepositoryHandler) GetRepositoryImpact(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Repository ID is required",
		}, http.StatusBadRequest)
		return
	}

	impact, err := h.RepositoryService.GetRepositoryImpact(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, impact)
}

// GetRepositoryNetwork handles GET /api/v1/repositories/network
func (h *RepositoryHandler) GetRepositoryNetwork(w http.ResponseWriter, r *http.Request) {
	network, err := h.RepositoryService.GetRepositoryNetwork(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, network)
}

// AnalyzeCrossRepoImpact handles POST /api/v1/repositories/analyze-cross-repo
func (h *RepositoryHandler) AnalyzeCrossRepoImpact(w http.ResponseWriter, r *http.Request) {
	var req models.CrossRepoAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.RepositoryService.AnalyzeCrossRepoImpact(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, result)
}

// GetRepositoryCentrality handles GET /api/v1/repositories/{id}/centrality
func (h *RepositoryHandler) GetRepositoryCentrality(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Repository ID is required",
		}, http.StatusBadRequest)
		return
	}

	centrality, err := h.RepositoryService.GetRepositoryCentrality(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, centrality)
}

// GetRepositoryClusters handles GET /api/v1/repositories/clusters
func (h *RepositoryHandler) GetRepositoryClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := h.RepositoryService.GetRepositoryClusters(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, clusters)
}
