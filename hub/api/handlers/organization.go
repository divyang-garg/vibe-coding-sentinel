// Package handlers - Organization and project HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// OrganizationHandler handles organization and project-related HTTP requests
type OrganizationHandler struct {
	BaseHandler
	OrganizationService services.OrganizationService
}

// NewOrganizationHandler creates a new organization handler with dependencies
func NewOrganizationHandler(orgService services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		OrganizationService: orgService,
	}
}

// CreateOrganization handles POST /api/v1/organizations
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	org, err := h.OrganizationService.CreateOrganization(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, org)
}

// GetOrganization handles GET /api/v1/organizations/{id}
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Organization ID is required",
		}, http.StatusBadRequest)
		return
	}

	org, err := h.OrganizationService.GetOrganization(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if org == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "organization",
			Message:  "Organization not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, org)
}

// CreateProject handles POST /api/v1/projects
func (h *OrganizationHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	// For now, use a default org ID - in production this would come from auth context
	orgID := "default-org"
	project, err := h.OrganizationService.CreateProject(r.Context(), orgID, req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, project)
}

// GetProject handles GET /api/v1/projects/{id}
func (h *OrganizationHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	project, err := h.OrganizationService.GetProject(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if project == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "project",
			Message:  "Project not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, project)
}

// ListProjects handles GET /api/v1/projects
func (h *OrganizationHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "org_id",
			Message: "Organization ID is required",
		}, http.StatusBadRequest)
		return
	}

	projects, err := h.OrganizationService.ListProjects(r.Context(), orgID)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
		"total":    len(projects),
	})
}

// GenerateAPIKey handles POST /api/v1/projects/{id}/api-key
// Generates a new API key for a project
func (h *OrganizationHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	apiKey, err := h.OrganizationService.GenerateAPIKey(r.Context(), projectID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "project not found" || err.Error() == "failed to find project: project not found" {
			h.WriteErrorResponse(w, &models.NotFoundError{
				Resource: "project",
				Message:  "Project not found",
			}, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	// Extract prefix for response
	prefix := ""
	if len(apiKey) >= 8 {
		prefix = apiKey[:8]
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"api_key":        apiKey,
		"api_key_prefix": prefix,
		"message":        "API key generated successfully. Save this key - it will not be shown again.",
		"warning":        "This is the only time you will see this key. Store it securely.",
	})
}

// RevokeAPIKey handles DELETE /api/v1/projects/{id}/api-key
// Revokes a project's API key
func (h *OrganizationHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	err := h.OrganizationService.RevokeAPIKey(r.Context(), projectID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "project not found" {
			h.WriteErrorResponse(w, &models.NotFoundError{
				Resource: "project",
				Message:  "Project not found",
			}, http.StatusNotFound)
			return
		}
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "API key revoked successfully",
	})
}

// GetAPIKeyInfo handles GET /api/v1/projects/{id}/api-key
// Returns API key information (prefix only, for security)
func (h *OrganizationHandler) GetAPIKeyInfo(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if projectID == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Project ID is required",
		}, http.StatusBadRequest)
		return
	}

	project, err := h.OrganizationService.GetProject(r.Context(), projectID)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if project == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "project",
			Message:  "Project not found",
		}, http.StatusNotFound)
		return
	}

	hasKey := project.APIKeyHash != "" || project.APIKeyPrefix != ""

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"has_api_key":    hasKey,
		"api_key_prefix": project.APIKeyPrefix,
		"message":        "Full API key is never returned for security reasons. Use POST to generate a new key.",
	})
}
