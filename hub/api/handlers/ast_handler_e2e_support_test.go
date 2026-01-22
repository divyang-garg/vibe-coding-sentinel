// Package handlers - AST handler end-to-end tests for support endpoints and scenarios
// Tests the complete HTTP flow from request to response
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/models"
)

// TestASTEndToEnd_GetSupportedAnalyses tests the complete flow for getting supported analyses
// Uses setupTestRouter() from ast_handler_e2e_analyze_test.go
func TestASTEndToEnd_GetSupportedAnalyses(t *testing.T) {
	router := setupTestRouter()

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/ast/supported", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.SupportedAnalysesResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify languages are returned
		if len(response.Languages) == 0 {
			t.Error("Expected at least one supported language")
		}

		// Verify analyses are returned
		if len(response.Analyses) == 0 {
			t.Error("Expected at least one supported analysis")
		}

		// Verify common languages are present
		langMap := make(map[string]bool)
		for _, lang := range response.Languages {
			langMap[lang.Name] = true
		}

		expectedLangs := []string{"go", "javascript", "typescript", "python"}
		for _, expected := range expectedLangs {
			if !langMap[expected] {
				t.Errorf("Expected language '%s' to be supported", expected)
			}
		}

		// Verify common analyses are present
		analysisMap := make(map[string]bool)
		for _, analysis := range response.Analyses {
			analysisMap[analysis] = true
		}

		expectedAnalyses := []string{"duplicates", "unused", "unreachable"}
		for _, expected := range expectedAnalyses {
			if !analysisMap[expected] {
				t.Errorf("Expected analysis '%s' to be supported", expected)
			}
		}
	})
}

// TestASTEndToEnd_RealWorldScenarios tests realistic usage scenarios
func TestASTEndToEnd_RealWorldScenarios(t *testing.T) {
	router := setupTestRouter()

	t.Run("complete_go_project_analysis", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{
				{
					Path: "main.go",
					Content: `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}

func helper() {
    fmt.Println("helper")
}`,
					Language: "go",
				},
				{
					Path: "utils.go",
					Content: `package main

func unusedFunction() {
    // This function is never called
}

func usedFunction() {
    helper()
}`,
					Language: "go",
				},
			},
			Analyses: []string{"unused", "duplicates"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/multi", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.MultiFileASTResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify analysis completed successfully
		if len(response.Files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(response.Files))
		}
	})

	t.Run("security_audit_flow", func(t *testing.T) {
		reqBody := models.SecurityASTRequest{
			Code: `package main

import (
    "database/sql"
    "fmt"
    "net/http"
)

func vulnerableHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    db, _ := sql.Open("postgres", "...")
    
    // SQL Injection vulnerability
    query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
    db.Query(query)
    
    // XSS vulnerability
    w.Write([]byte("<div>" + userID + "</div>"))
}`,
			Language: "go",
			Severity: "high",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/security", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.SecurityASTResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify response structure
		if response.RiskScore < 0 || response.RiskScore > 100 {
			t.Errorf("Expected risk score between 0-100, got %f", response.RiskScore)
		}
	})
}
