// Package handlers contains HTTP request handlers for the Sentinel Hub API.
// Handlers are thin wrappers around service calls, following the single responsibility principle.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"sentinel-hub-api/models"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	// Common dependencies can be added here
}

// WriteJSONResponse writes a JSON response with the given status code
func (h *BaseHandler) WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteErrorResponse writes an error response
func (h *BaseHandler) WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error": err.Error(),
	}

	if mcpErr, ok := err.(*models.MCPError); ok {
		errorResponse["code"] = mcpErr.Code
		errorResponse["message"] = mcpErr.Message
		if mcpErr.Data != nil {
			errorResponse["data"] = mcpErr.Data
		}
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// GetProjectFromContext extracts project from context
func (h *BaseHandler) GetProjectFromContext(ctx context.Context) (*models.Project, error) {
	value := ctx.Value("project")
	if value == nil {
		return nil, fmt.Errorf("project not found in context")
	}
	project, ok := value.(*models.Project)
	if !ok || project == nil {
		return nil, fmt.Errorf("invalid project type in context")
	}
	return project, nil
}

// GetOrgFromContext extracts organization from context
func (h *BaseHandler) GetOrgFromContext(ctx context.Context) (*models.Organization, error) {
	orgID := ctx.Value("org_id")
	if orgID == nil {
		return nil, fmt.Errorf("organization not found in context")
	}
	return &models.Organization{
		ID:   orgID.(string),
		Name: "Default Organization",
	}, nil
}
