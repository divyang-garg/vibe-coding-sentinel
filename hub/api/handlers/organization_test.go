// Package handlers - Unit tests for organization handlers
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/models"

	"github.com/go-chi/chi/v5"
)

// mockOrganizationService implements OrganizationService for testing
type mockOrganizationService struct {
	projects          map[string]*models.Project
	organizations     map[string]*models.Organization
	generateAPIKeyErr error
	revokeAPIKeyErr   error
	generatedAPIKey   string
}

func newMockOrganizationService() *mockOrganizationService {
	return &mockOrganizationService{
		projects:      make(map[string]*models.Project),
		organizations: make(map[string]*models.Organization),
	}
}

func (m *mockOrganizationService) CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest) (*models.Organization, error) {
	org := &models.Organization{
		ID:   "org_123",
		Name: req.Name,
	}
	m.organizations[org.ID] = org
	return org, nil
}

func (m *mockOrganizationService) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	return m.organizations[id], nil
}

func (m *mockOrganizationService) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
	orgs := make([]models.Organization, 0, len(m.organizations))
	for _, org := range m.organizations {
		orgs = append(orgs, *org)
	}
	return orgs, nil
}

func (m *mockOrganizationService) UpdateOrganization(ctx context.Context, id string, req models.UpdateOrganizationRequest) (*models.Organization, error) {
	org := m.organizations[id]
	if org == nil {
		return nil, nil
	}
	if req.Name != "" {
		org.Name = req.Name
	}
	return org, nil
}

func (m *mockOrganizationService) DeleteOrganization(ctx context.Context, id string) error {
	delete(m.organizations, id)
	return nil
}

func (m *mockOrganizationService) CreateProject(ctx context.Context, orgID string, req models.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		ID:           "proj_123",
		OrgID:        orgID,
		Name:         req.Name,
		APIKey:       "test-api-key-12345",
		APIKeyHash:   "hash123",
		APIKeyPrefix: "test-api",
	}
	m.projects[project.ID] = project
	return project, nil
}

func (m *mockOrganizationService) GetProject(ctx context.Context, id string) (*models.Project, error) {
	return m.projects[id], nil
}

func (m *mockOrganizationService) ListProjects(ctx context.Context, orgID string) ([]models.Project, error) {
	projects := make([]models.Project, 0)
	for _, p := range m.projects {
		if p.OrgID == orgID {
			projects = append(projects, *p)
		}
	}
	return projects, nil
}

func (m *mockOrganizationService) UpdateProject(ctx context.Context, id string, req models.UpdateProjectRequest) (*models.Project, error) {
	project := m.projects[id]
	if project == nil {
		return nil, nil
	}
	if req.Name != "" {
		project.Name = req.Name
	}
	return project, nil
}

func (m *mockOrganizationService) DeleteProject(ctx context.Context, id string) error {
	delete(m.projects, id)
	return nil
}

func (m *mockOrganizationService) GenerateAPIKey(ctx context.Context, projectID string) (string, error) {
	if m.generateAPIKeyErr != nil {
		return "", m.generateAPIKeyErr
	}
	project := m.projects[projectID]
	if project == nil {
		return "", fmt.Errorf("project not found")
	}
	m.generatedAPIKey = "new-api-key-" + projectID
	return m.generatedAPIKey, nil
}

func (m *mockOrganizationService) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Project, error) {
	for _, p := range m.projects {
		if p.APIKey == apiKey || p.APIKeyHash != "" {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockOrganizationService) RevokeAPIKey(ctx context.Context, projectID string) error {
	if m.revokeAPIKeyErr != nil {
		return m.revokeAPIKeyErr
	}
	project := m.projects[projectID]
	if project == nil {
		return fmt.Errorf("project not found")
	}
	project.APIKey = ""
	project.APIKeyHash = ""
	project.APIKeyPrefix = ""
	return nil
}

func TestOrganizationHandler_GenerateAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockService    *mockOrganizationService
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "successful generation",
			projectID: "proj_123",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				m.projects["proj_123"] = &models.Project{
					ID:   "proj_123",
					Name: "Test Project",
				}
				return m
			}(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing project ID",
			projectID:      "",
			mockService:    newMockOrganizationService(),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "project not found",
			projectID: "proj_nonexistent",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				// Service returns error, handler checks for "project not found" string
				m.generateAPIKeyErr = fmt.Errorf("project not found")
				return m
			}(),
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewOrganizationHandler(tt.mockService)

			// Create router to properly set URL params
			r := chi.NewRouter()
			r.Post("/api/v1/projects/{id}/api-key", handler.GenerateAPIKey)

			req := httptest.NewRequest("POST", "/api/v1/projects/"+tt.projectID+"/api-key", nil)
			if tt.projectID == "" {
				req = httptest.NewRequest("POST", "/api/v1/projects//api-key", nil)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GenerateAPIKey() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if !tt.expectError {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["api_key"] == nil {
					t.Error("Expected api_key in response")
				}
			}
		})
	}
}

func TestOrganizationHandler_RevokeAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockService    *mockOrganizationService
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "successful revocation",
			projectID: "proj_123",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				m.projects["proj_123"] = &models.Project{
					ID:           "proj_123",
					Name:         "Test Project",
					APIKeyHash:   "hash123",
					APIKeyPrefix: "test-api",
				}
				return m
			}(),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing project ID",
			projectID:      "",
			mockService:    newMockOrganizationService(),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:      "project not found",
			projectID: "proj_nonexistent",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				m.revokeAPIKeyErr = fmt.Errorf("project not found")
				return m
			}(),
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewOrganizationHandler(tt.mockService)

			// Create router to properly set URL params
			r := chi.NewRouter()
			r.Delete("/api/v1/projects/{id}/api-key", handler.RevokeAPIKey)

			req := httptest.NewRequest("DELETE", "/api/v1/projects/"+tt.projectID+"/api-key", nil)
			if tt.projectID == "" {
				req = httptest.NewRequest("DELETE", "/api/v1/projects//api-key", nil)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("RevokeAPIKey() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if !tt.expectError {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["message"] == nil {
					t.Error("Expected message in response")
				}
			}
		})
	}
}

func TestOrganizationHandler_GetAPIKeyInfo(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockService    *mockOrganizationService
		expectedStatus int
		hasKey         bool
	}{
		{
			name:      "project with API key",
			projectID: "proj_123",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				m.projects["proj_123"] = &models.Project{
					ID:           "proj_123",
					Name:         "Test Project",
					APIKeyHash:   "hash123",
					APIKeyPrefix: "test-api",
				}
				return m
			}(),
			expectedStatus: http.StatusOK,
			hasKey:         true,
		},
		{
			name:      "project without API key",
			projectID: "proj_456",
			mockService: func() *mockOrganizationService {
				m := newMockOrganizationService()
				m.projects["proj_456"] = &models.Project{
					ID:   "proj_456",
					Name: "Test Project",
				}
				return m
			}(),
			expectedStatus: http.StatusOK,
			hasKey:         false,
		},
		{
			name:           "missing project ID",
			projectID:      "",
			mockService:    newMockOrganizationService(),
			expectedStatus: http.StatusBadRequest,
			hasKey:         false,
		},
		{
			name:           "project not found",
			projectID:      "proj_nonexistent",
			mockService:    newMockOrganizationService(),
			expectedStatus: http.StatusNotFound,
			hasKey:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewOrganizationHandler(tt.mockService)

			// Create router to properly set URL params
			r := chi.NewRouter()
			r.Get("/api/v1/projects/{id}/api-key", handler.GetAPIKeyInfo)

			req := httptest.NewRequest("GET", "/api/v1/projects/"+tt.projectID+"/api-key", nil)
			if tt.projectID == "" {
				req = httptest.NewRequest("GET", "/api/v1/projects//api-key", nil)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("GetAPIKeyInfo() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["has_api_key"] != tt.hasKey {
					t.Errorf("has_api_key = %v, want %v", response["has_api_key"], tt.hasKey)
				}
			}
		})
	}
}
