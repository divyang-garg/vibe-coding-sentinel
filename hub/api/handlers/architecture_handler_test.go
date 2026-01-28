// Package handlers - Architecture handler tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/services"
)

func TestArchitectureHandler_AnalyzeArchitecture(t *testing.T) {
	handler := NewArchitectureHandler()

	t.Run("success", func(t *testing.T) {
		reqBody := services.ArchitectureAnalysisRequest{
			Files: []services.FileContent{
				{Path: "a.go", Content: "package p\nfunc F() {}", Language: "go"},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/analyze/architecture", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeArchitecture(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("empty_files", func(t *testing.T) {
		reqBody := services.ArchitectureAnalysisRequest{Files: []services.FileContent{}}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/analyze/architecture", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeArchitecture(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for empty files, got %d", w.Code)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/analyze/architecture", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()

		handler.AnalyzeArchitecture(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for invalid JSON, got %d", w.Code)
		}
	})

	t.Run("empty_file_path", func(t *testing.T) {
		reqBody := services.ArchitectureAnalysisRequest{
			Files: []services.FileContent{
				{Path: "", Content: "package p", Language: "go"},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/analyze/architecture", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.AnalyzeArchitecture(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for empty file path, got %d", w.Code)
		}
	})
}
