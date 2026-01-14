// Package handlers - API versioning HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"encoding/json"
	"net/http"

	"sentinel-hub-api/models"
	"sentinel-hub-api/services"
)

// APIVersionHandler handles API versioning-related HTTP requests
type APIVersionHandler struct {
	BaseHandler
	APIVersionService services.APIVersionService
}

// NewAPIVersionHandler creates a new API version handler with dependencies
func NewAPIVersionHandler(apiVersionService services.APIVersionService) *APIVersionHandler {
	return &APIVersionHandler{
		APIVersionService: apiVersionService,
	}
}

// CreateAPIVersion handles POST /api/v1/versions
func (h *APIVersionHandler) CreateAPIVersion(w http.ResponseWriter, r *http.Request) {
	var req models.APIVersion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	version, err := h.APIVersionService.CreateAPIVersion(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, version)
}

// GetAPIVersion handles GET /api/v1/versions/{id}
func (h *APIVersionHandler) GetAPIVersion(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "id",
			Message: "Version ID is required",
		}, http.StatusBadRequest)
		return
	}

	version, err := h.APIVersionService.GetAPIVersion(r.Context(), id)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if version == nil {
		h.WriteErrorResponse(w, &models.NotFoundError{
			Resource: "api_version",
			Message:  "API version not found",
		}, http.StatusNotFound)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, version)
}

// ListAPIVersions handles GET /api/v1/versions
func (h *APIVersionHandler) ListAPIVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.APIVersionService.ListAPIVersions(r.Context())
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"versions": versions,
		"total":    len(versions),
	})
}

// GetVersionCompatibility handles GET /api/v1/versions/compatibility
func (h *APIVersionHandler) GetVersionCompatibility(w http.ResponseWriter, r *http.Request) {
	fromVersion := r.URL.Query().Get("from")
	toVersion := r.URL.Query().Get("to")

	if fromVersion == "" || toVersion == "" {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "versions",
			Message: "Both from_version and to_version are required",
		}, http.StatusBadRequest)
		return
	}

	report, err := h.APIVersionService.GetVersionCompatibility(r.Context(), fromVersion, toVersion)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusOK, report)
}

// CreateAPIVersionMigration handles POST /api/v1/versions/migrations
func (h *APIVersionHandler) CreateAPIVersionMigration(w http.ResponseWriter, r *http.Request) {
	var req models.VersionMigration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteErrorResponse(w, &models.ValidationError{
			Field:   "body",
			Message: "Invalid request format",
		}, http.StatusBadRequest)
		return
	}

	migration, err := h.APIVersionService.CreateVersionMigration(r.Context(), req)
	if err != nil {
		h.WriteErrorResponse(w, err, http.StatusInternalServerError)
		return
	}

	h.WriteJSONResponse(w, http.StatusCreated, migration)
}
