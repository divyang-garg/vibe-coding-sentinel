// Test Handlers Helper
// Provides test helpers for calling handlers from integration tests
// This file is in package main so handlers can be called directly

package services

import (
	"context"
	"database/sql"
	"net/http"
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
