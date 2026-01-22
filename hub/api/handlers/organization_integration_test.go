// Package handlers - Integration tests for API key management endpoints
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/middleware"
	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg/security"
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// Integration test helper - creates a router with all middleware
func createTestRouter(orgService services.OrganizationService) *chi.Mux {
	r := chi.NewRouter()

	// Add authentication middleware (but allow all for testing)
	mockLogger := &mockLoggerForAudit{}
	auditLogger := security.NewAuditLogger(mockLogger)

	r.Use(middleware.AuthMiddleware(middleware.AuthMiddlewareConfig{
		OrganizationService: orgService,
		SkipPaths:           []string{"/api/v1/projects"},
		Logger:              nil,
		AuditLogger:         auditLogger,
	}))

	orgHandler := NewOrganizationHandler(orgService)
	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Post("/", orgHandler.CreateProject)
		r.Get("/{id}", orgHandler.GetProject)
		r.Post("/{id}/api-key", orgHandler.GenerateAPIKey)
		r.Get("/{id}/api-key", orgHandler.GetAPIKeyInfo)
		r.Delete("/{id}/api-key", orgHandler.RevokeAPIKey)
	})

	return r
}

type mockLoggerForAudit struct{}

func (m *mockLoggerForAudit) Info(msg string, fields ...interface{})  {}
func (m *mockLoggerForAudit) Warn(msg string, fields ...interface{})  {}
func (m *mockLoggerForAudit) Error(msg string, fields ...interface{}) {}

func TestAPIKeyManagement_Integration(t *testing.T) {
	// Create mock service
	mockService := newMockOrganizationService()

	// Create project first
	createReq := models.CreateProjectRequest{Name: "Test Project"}
	project, err := mockService.CreateProject(nil, "org_123", createReq)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create router
	router := createTestRouter(mockService)

	t.Run("Generate API Key", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/projects/"+project.ID+"/api-key", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["api_key"] == nil {
			t.Error("Expected api_key in response")
		}

		apiKey, ok := response["api_key"].(string)
		if !ok || apiKey == "" {
			t.Error("api_key should be a non-empty string")
		}
	})

	t.Run("Get API Key Info", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/"+project.ID+"/api-key", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["has_api_key"] == nil {
			t.Error("Expected has_api_key in response")
		}
	})

	t.Run("Revoke API Key", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/projects/"+project.ID+"/api-key", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] == nil {
			t.Error("Expected message in response")
		}
	})

	t.Run("Full Workflow: Generate -> Get Info -> Revoke", func(t *testing.T) {
		// Create new project for this test
		project2, err := mockService.CreateProject(nil, "org_123", models.CreateProjectRequest{Name: "Test Project 2"})
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		// 1. Generate API key
		req1 := httptest.NewRequest("POST", "/api/v1/projects/"+project2.ID+"/api-key", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("Generate failed: expected 200, got %d", w1.Code)
		}

		var genResponse map[string]interface{}
		json.NewDecoder(w1.Body).Decode(&genResponse)
		generatedKey := genResponse["api_key"].(string)

		// 2. Get API key info
		req2 := httptest.NewRequest("GET", "/api/v1/projects/"+project2.ID+"/api-key", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("Get info failed: expected 200, got %d", w2.Code)
		}

		var infoResponse map[string]interface{}
		json.NewDecoder(w2.Body).Decode(&infoResponse)

		if infoResponse["has_api_key"] != true {
			t.Error("Expected has_api_key to be true")
		}

		// 3. Revoke API key
		req3 := httptest.NewRequest("DELETE", "/api/v1/projects/"+project2.ID+"/api-key", nil)
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, req3)

		if w3.Code != http.StatusOK {
			t.Errorf("Revoke failed: expected 200, got %d", w3.Code)
		}

		// 4. Verify key is revoked (get info should show no key)
		req4 := httptest.NewRequest("GET", "/api/v1/projects/"+project2.ID+"/api-key", nil)
		w4 := httptest.NewRecorder()
		router.ServeHTTP(w4, req4)

		var infoResponse2 map[string]interface{}
		json.NewDecoder(w4.Body).Decode(&infoResponse2)

		if infoResponse2["has_api_key"] != false {
			t.Error("Expected has_api_key to be false after revocation")
		}

		_ = generatedKey // Use generated key to avoid unused variable
	})
}
