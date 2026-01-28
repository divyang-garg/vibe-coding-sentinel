// Test Handlers Helper
// Provides test helpers for calling handlers from integration tests
// NOTE: These handlers are now in the handlers package. This file is kept for backward compatibility.

package services

import (
	"context"
	"database/sql"
	"net/http"

	"sentinel-hub-api/models"
)

// TestHandlerCaller provides a way to call handlers from tests
// This is used by integration tests to call handlers with proper context
type TestHandlerCaller struct {
	ProjectID string
	APIKey    string
	DB        *sql.DB
}

// NewTestHandlerCaller creates a new test handler caller
func NewTestHandlerCaller(projectID, apiKey string, testDB *sql.DB) *TestHandlerCaller {
	return &TestHandlerCaller{
		ProjectID: projectID,
		APIKey:    apiKey,
		DB:        testDB,
	}
}

// getProjectFromDB retrieves project from database
func (thc *TestHandlerCaller) getProjectFromDB() (*models.Project, error) {
	var project models.Project
	query := "SELECT id, org_id, name, api_key, created_at FROM projects WHERE id = $1"
	err := thc.DB.QueryRow(query, thc.ProjectID).Scan(&project.ID, &project.OrgID, &project.Name, &project.APIKey, &project.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// addProjectContext adds project context to request
func (thc *TestHandlerCaller) addProjectContext(r *http.Request) (*http.Request, error) {
	project, err := thc.getProjectFromDB()
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "project", project)
	return r.WithContext(ctx), nil
}

// CallValidateCodeHandler calls validateCodeHandler with test context
// NOTE: These handlers are now in the handlers package. Use handlers.CodeAnalysisHandler.ValidateCode instead.
func (thc *TestHandlerCaller) CallValidateCodeHandler(w http.ResponseWriter, r *http.Request) error {
	// Handlers are now in handlers package - this method is deprecated
	// Use the actual handlers from handlers package in tests
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.CodeAnalysisHandler.ValidateCode instead"}`))
	return nil
}

// CallApplyFixHandler calls applyFixHandler with test context
// NOTE: These handlers are now in the handlers package. Use handlers.FixHandler.ApplyFix instead.
func (thc *TestHandlerCaller) CallApplyFixHandler(w http.ResponseWriter, r *http.Request) error {
	// Handlers are now in handlers package - this method is deprecated
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.FixHandler.ApplyFix instead"}`))
	return nil
}

// CallValidateLLMConfigHandler calls validateLLMConfigHandler with test context
// NOTE: These handlers are now in the handlers package. Use handlers.LLMHandler.ValidateLLMConfig instead.
func (thc *TestHandlerCaller) CallValidateLLMConfigHandler(w http.ResponseWriter, r *http.Request) error {
	// Handlers are now in handlers package - this method is deprecated
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.LLMHandler.ValidateLLMConfig instead"}`))
	return nil
}

// CallGetCacheMetricsHandler calls getCacheMetricsHandler with test context
// NOTE: These handlers are now in the handlers package. Use handlers.MetricsHandler.GetCacheMetrics instead.
func (thc *TestHandlerCaller) CallGetCacheMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	// Handlers are now in handlers package - this method is deprecated
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.MetricsHandler.GetCacheMetrics instead"}`))
	return nil
}

// CallGetCostMetricsHandler calls getCostMetricsHandler with test context
// NOTE: These handlers are now in the handlers package. Use handlers.MetricsHandler.GetCostMetrics instead.
func (thc *TestHandlerCaller) CallGetCostMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	// Handlers are now in handlers package - this method is deprecated
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.MetricsHandler.GetCostMetrics instead"}`))
	return nil
}

// CallHandler is a generic handler caller
func (thc *TestHandlerCaller) CallHandler(handler http.HandlerFunc, w http.ResponseWriter, r *http.Request) error {
	req, err := thc.addProjectContext(r)
	if err != nil {
		return err
	}
	handler(w, req)
	return nil
}
