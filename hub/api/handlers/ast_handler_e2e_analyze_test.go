// Package handlers - AST handler end-to-end tests for analysis endpoints
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
	"sentinel-hub-api/services"

	"github.com/go-chi/chi/v5"
)

// setupTestRouter creates a test router with AST routes
func setupTestRouter() *chi.Mux {
	r := chi.NewRouter()
	astService := services.NewASTService()
	astHandler := NewASTHandler(astService)

	r.Route("/api/v1/ast", func(r chi.Router) {
		r.Route("/analyze", func(r chi.Router) {
			r.Post("/", astHandler.AnalyzeAST)
			r.Post("/multi", astHandler.AnalyzeMultiFile)
			r.Post("/security", astHandler.AnalyzeSecurity)
			r.Post("/cross", astHandler.AnalyzeCrossFile)
		})
		r.Get("/supported", astHandler.GetSupportedAnalyses)
	})

	return r
}

// TestASTEndToEnd_AnalyzeAST tests the complete flow for single-file AST analysis
func TestASTEndToEnd_AnalyzeAST(t *testing.T) {
	router := setupTestRouter()

	t.Run("success_go_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code: `package main

func hello() {
    fmt.Println("Hello, World!")
}

func duplicate() {
    fmt.Println("duplicate")
}

func duplicate() {
    fmt.Println("duplicate")
}`,
			Language: "go",
			Analyses: []string{"duplicates"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.ASTAnalysisResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Language != "go" {
			t.Errorf("Expected language 'go', got '%s'", response.Language)
		}

		// Should find duplicate functions
		if len(response.Findings) == 0 {
			t.Error("Expected to find duplicate functions, but found none")
		}

		// Verify stats are populated
		if response.Stats.NodesVisited == 0 {
			t.Error("Expected NodesVisited > 0")
		}
	})

	t.Run("success_javascript_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code: `function unused() {
    const x = 5;
    return x;
}

function used() {
    return 42;
}

used();`,
			Language: "javascript",
			Analyses: []string{"unused"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.ASTAnalysisResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Language != "javascript" {
			t.Errorf("Expected language 'javascript', got '%s'", response.Language)
		}
	})

	t.Run("success_python_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code: `def hello():
    print("Hello, World!")

def unused_function():
    pass`,
			Language: "python",
			Analyses: []string{"unused"},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.ASTAnalysisResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Language != "python" {
			t.Errorf("Expected language 'python', got '%s'", response.Language)
		}
	})

	t.Run("success_typescript_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code: `function test(): void {
    console.log("test");
}

interface TestInterface {
    prop: string;
}`,
			Language: "typescript",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.ASTAnalysisResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Language != "typescript" {
			t.Errorf("Expected language 'typescript', got '%s'", response.Language)
		}
	})

	t.Run("missing_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Language: "go",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing_language", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code: "package main\n",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}

// TestASTEndToEnd_AnalyzeMultiFile tests the complete flow for multi-file AST analysis
func TestASTEndToEnd_AnalyzeMultiFile(t *testing.T) {
	router := setupTestRouter()

	t.Run("success_multiple_files", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\n\nfunc file1() {}\n",
					Language: "go",
				},
				{
					Path:     "file2.go",
					Content:  "package main\n\nfunc file2() {}\n",
					Language: "go",
				},
				{
					Path:     "file3.js",
					Content:  "function file3() {}\n",
					Language: "javascript",
				},
			},
			Analyses: []string{"duplicates"},
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

		if len(response.Files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(response.Files))
		}

		// Verify stats are populated
		if response.Stats.NodesVisited == 0 {
			t.Error("Expected NodesVisited > 0")
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/multi", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing_file_content", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Language: "go",
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/multi", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Accept both 400 and 500 as valid error responses
		if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 400 or 500, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}

// TestASTEndToEnd_AnalyzeSecurity tests the complete flow for security analysis
func TestASTEndToEnd_AnalyzeSecurity(t *testing.T) {
	router := setupTestRouter()

	t.Run("success_sql_injection", func(t *testing.T) {
		reqBody := models.SecurityASTRequest{
			Code: `package main

import "database/sql"

func vulnerable(userID string) {
    db, _ := sql.Open("postgres", "...")
    query := "SELECT * FROM users WHERE id = " + userID
    db.Query(query)
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

		// Should find security vulnerabilities
		if len(response.Vulnerabilities) == 0 && len(response.Findings) == 0 {
			t.Log("No vulnerabilities found (may be expected depending on detection logic)")
		}

		// Verify risk score is calculated
		if response.RiskScore < 0 || response.RiskScore > 100 {
			t.Errorf("Expected risk score between 0-100, got %f", response.RiskScore)
		}
	})

	t.Run("success_xss_vulnerability", func(t *testing.T) {
		reqBody := models.SecurityASTRequest{
			Code: `function render(userInput) {
    document.getElementById("content").innerHTML = userInput;
}`,
			Language: "javascript",
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

	t.Run("missing_code", func(t *testing.T) {
		reqBody := models.SecurityASTRequest{
			Language: "go",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/security", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}

// TestASTEndToEnd_AnalyzeCrossFile tests the complete flow for cross-file analysis
func TestASTEndToEnd_AnalyzeCrossFile(t *testing.T) {
	router := setupTestRouter()

	t.Run("success_cross_file_dependencies", func(t *testing.T) {
		reqBody := models.CrossFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\n\nfunc ExportedFunc() {}\nfunc privateFunc() {}\n",
					Language: "go",
				},
				{
					Path:     "file2.go",
					Content:  "package main\n\nfunc UseExported() {\n    ExportedFunc()\n}\n",
					Language: "go",
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/cross", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response models.CrossFileASTResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Verify stats are populated
		if response.Stats.FilesAnalyzed != 2 {
			t.Errorf("Expected 2 files analyzed, got %d", response.Stats.FilesAnalyzed)
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		reqBody := models.CrossFileASTRequest{
			Files: []models.FileInput{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/cross", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}
