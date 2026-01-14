// Package handlers - Organization and project HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
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
