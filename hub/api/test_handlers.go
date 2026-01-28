// Test Handlers Helper
// Provides test helpers for calling handlers from integration tests
// This file is in package main so handlers can be called directly

package main

import (
	"context"
	"database/sql"
	"net/http"
)

// Handler functions for testing - these delegate to actual handlers in handlers package
// NOTE: These handlers are now fully implemented in the handlers package:
// - validateCodeHandler -> handlers.CodeAnalysisHandler.ValidateCode (already implemented)
// - applyFixHandler -> handlers.FixHandler.ApplyFix (implemented in handlers/fix_handler.go)
// - validateLLMConfigHandler -> handlers.LLMHandler.ValidateLLMConfig (implemented in handlers/llm_handler.go)
// - getCacheMetricsHandler -> handlers.MetricsHandler.GetCacheMetrics (implemented in handlers/metrics_handler.go)
// - getCostMetricsHandler -> handlers.MetricsHandler.GetCostMetrics (implemented in handlers/metrics_handler.go)
//
// For integration tests, use the actual handlers from the handlers package via TestHandlerCaller methods.

func validateCodeHandler(w http.ResponseWriter, _ *http.Request) {
	// This is a test helper - in production, use handlers.CodeAnalysisHandler.ValidateCode
	// This stub is kept for backward compatibility with existing tests
	// Request parameter unused - required for http.HandlerFunc interface compliance
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.CodeAnalysisHandler.ValidateCode instead"}`))
}

func applyFixHandler(w http.ResponseWriter, _ *http.Request) {
	// This is a test helper - in production, use handlers.FixHandler.ApplyFix
	// Request parameter unused - required for http.HandlerFunc interface compliance
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.FixHandler.ApplyFix instead"}`))
}

func validateLLMConfigHandler(w http.ResponseWriter, _ *http.Request) {
	// This is a test helper - in production, use handlers.LLMHandler.ValidateLLMConfig
	// Request parameter unused - required for http.HandlerFunc interface compliance
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.LLMHandler.ValidateLLMConfig instead"}`))
}

func getCacheMetricsHandler(w http.ResponseWriter, _ *http.Request) {
	// This is a test helper - in production, use handlers.MetricsHandler.GetCacheMetrics
	// Request parameter unused - required for http.HandlerFunc interface compliance
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.MetricsHandler.GetCacheMetrics instead"}`))
}

func getCostMetricsHandler(w http.ResponseWriter, _ *http.Request) {
	// This is a test helper - in production, use handlers.MetricsHandler.GetCostMetrics
	// Request parameter unused - required for http.HandlerFunc interface compliance
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Use handlers.MetricsHandler.GetCostMetrics instead"}`))
}

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
func (thc *TestHandlerCaller) getProjectFromDB() (*Project, error) {
	var project Project
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
	ctx = context.WithValue(ctx, projectKey, project)
	return r.WithContext(ctx), nil
}

// CallValidateCodeHandler calls validateCodeHandler with test context
func (thc *TestHandlerCaller) CallValidateCodeHandler(w http.ResponseWriter, r *http.Request) error {
	// Get project from database
	var project Project
	query := "SELECT id, org_id, name, api_key, created_at FROM projects WHERE id = $1"
	err := thc.DB.QueryRow(query, thc.ProjectID).Scan(&project.ID, &project.OrgID, &project.Name, &project.APIKey, &project.CreatedAt)
	if err != nil {
		return err
	}

	// Add project context to request
	ctx := r.Context()
	ctx = context.WithValue(ctx, projectKey, &project)
	r = r.WithContext(ctx)

	validateCodeHandler(w, r)
	return nil
}

// CallApplyFixHandler calls applyFixHandler with test context
func (thc *TestHandlerCaller) CallApplyFixHandler(w http.ResponseWriter, r *http.Request) error {
	req, err := thc.addProjectContext(r)
	if err != nil {
		return err
	}
	applyFixHandler(w, req)
	return nil
}

// CallValidateLLMConfigHandler calls validateLLMConfigHandler with test context
func (thc *TestHandlerCaller) CallValidateLLMConfigHandler(w http.ResponseWriter, r *http.Request) error {
	req, err := thc.addProjectContext(r)
	if err != nil {
		return err
	}
	validateLLMConfigHandler(w, req)
	return nil
}

// CallGetCacheMetricsHandler calls getCacheMetricsHandler with test context
func (thc *TestHandlerCaller) CallGetCacheMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	req, err := thc.addProjectContext(r)
	if err != nil {
		return err
	}
	getCacheMetricsHandler(w, req)
	return nil
}

// CallGetCostMetricsHandler calls getCostMetricsHandler with test context
func (thc *TestHandlerCaller) CallGetCostMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	req, err := thc.addProjectContext(r)
	if err != nil {
		return err
	}
	getCostMetricsHandler(w, req)
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
