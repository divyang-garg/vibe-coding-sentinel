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
		// Cannot call http.Error after WriteHeader, so log the error
		// The response has already been sent, so we can only log
		// In production, this should use proper logging
		_ = err // Log error in production: log.Printf("Failed to encode JSON response: %v", err)
	}
}

// WriteErrorResponse writes an error response using the standardized format.
// Delegates to the package-level WriteErrorResponse function for consistency.
func (h *BaseHandler) WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	WriteErrorResponse(w, err, statusCode)
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
	orgIDStr, ok := orgID.(string)
	if !ok {
		return nil, fmt.Errorf("invalid organization ID type in context")
	}
	return &models.Organization{
		ID:   orgIDStr,
		Name: "Default Organization",
	}, nil
}
