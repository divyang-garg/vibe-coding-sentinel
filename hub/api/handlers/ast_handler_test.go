// Package handlers - AST handler tests
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
)

func TestASTHandler_AnalyzeAST(t *testing.T) {
	astService := services.NewASTService()
	handler := NewASTHandler(astService)

	t.Run("success", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Code:     "package main\nfunc test() {}\n",
			Language: "go",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeAST(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("missing_code", func(t *testing.T) {
		reqBody := models.ASTAnalysisRequest{
			Language: "go",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeAST(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestASTHandler_AnalyzeMultiFile(t *testing.T) {
	astService := services.NewASTService()
	handler := NewASTHandler(astService)

	t.Run("success", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\n",
					Language: "go",
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/multi", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeMultiFile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		reqBody := models.MultiFileASTRequest{
			Files: []models.FileInput{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/multi", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeMultiFile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

func TestASTHandler_AnalyzeSecurity(t *testing.T) {
	astService := services.NewASTService()
	handler := NewASTHandler(astService)

	t.Run("success", func(t *testing.T) {
		reqBody := models.SecurityASTRequest{
			Code:     "package main\nfunc test() {}\n",
			Language: "go",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/security", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeSecurity(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestASTHandler_AnalyzeCrossFile(t *testing.T) {
	astService := services.NewASTService()
	handler := NewASTHandler(astService)

	t.Run("success", func(t *testing.T) {
		reqBody := models.CrossFileASTRequest{
			Files: []models.FileInput{
				{
					Path:     "file1.go",
					Content:  "package main\n",
					Language: "go",
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/ast/analyze/cross", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeCrossFile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestASTHandler_GetSupportedAnalyses(t *testing.T) {
	astService := services.NewASTService()
	handler := NewASTHandler(astService)

	req := httptest.NewRequest("GET", "/api/v1/ast/supported", nil)
	w := httptest.NewRecorder()

	handler.GetSupportedAnalyses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
