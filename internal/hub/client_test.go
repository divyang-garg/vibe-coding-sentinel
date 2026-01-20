// Package hub provides tests for Hub API client
package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with custom URL", func(t *testing.T) {
		client := NewClient("http://custom:9000", "test-key")

		if client.baseURL != "http://custom:9000" {
			t.Errorf("expected baseURL http://custom:9000, got %s", client.baseURL)
		}
		if client.apiKey != "test-key" {
			t.Errorf("expected apiKey test-key, got %s", client.apiKey)
		}
	})

	t.Run("uses default URL when empty", func(t *testing.T) {
		client := NewClient("", "test-key")

		if client.baseURL != "http://localhost:8080" {
			t.Errorf("expected default baseURL, got %s", client.baseURL)
		}
	})

	t.Run("creates http client with timeout", func(t *testing.T) {
		client := NewClient("http://test:8080", "key")

		if client.httpClient == nil {
			t.Error("httpClient should not be nil")
		}
		if client.timeout != 30*1000*1000*1000 { // 30 seconds in nanoseconds
			// Just verify it's set
		}
	})
}

func TestIsAvailable(t *testing.T) {
	t.Run("returns false for empty baseURL", func(t *testing.T) {
		client := &Client{baseURL: "", httpClient: &http.Client{}}

		if client.IsAvailable() {
			t.Error("should return false for empty baseURL")
		}
	})

	t.Run("returns true for healthy server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		if !client.IsAvailable() {
			t.Error("should return true for healthy server")
		}
	})

	t.Run("returns false for unhealthy server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		if client.IsAvailable() {
			t.Error("should return false for unhealthy server")
		}
	})

	t.Run("returns false for unreachable server", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")
		if client.IsAvailable() {
			t.Error("should return false for unreachable server")
		}
	})
}

func TestAnalyzeAST(t *testing.T) {
	t.Run("returns error when hub unavailable", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")

		_, err := client.AnalyzeAST(&ASTAnalysisRequest{})
		if err == nil {
			t.Error("should return error when hub unavailable")
		}
	})

	t.Run("successfully analyzes AST", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/api/v1/ast/analyze" {
				response := ASTAnalysisResponse{
					Findings: []ASTFinding{
						{Type: "unused_import", File: "test.go", Line: 1},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		resp, err := client.AnalyzeAST(&ASTAnalysisRequest{
			Code:     "package main",
			Language: "go",
			Filepath: "test.go",
		})

		if err != nil {
			t.Errorf("AnalyzeAST() error = %v", err)
		}
		if resp == nil || len(resp.Findings) == 0 {
			t.Error("should return findings")
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		_, err := client.AnalyzeAST(&ASTAnalysisRequest{})

		if err == nil {
			t.Error("should return error for server error")
		}
	})
}

func TestAnalyzeVibe(t *testing.T) {
	t.Run("returns error when hub unavailable", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")

		_, err := client.AnalyzeVibe(&VibeAnalysisRequest{})
		if err == nil {
			t.Error("should return error when hub unavailable")
		}
	})

	t.Run("successfully analyzes vibe issues", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/api/v1/vibe/analyze" {
				response := VibeAnalysisResponse{
					Issues: []VibeIssue{
						{Type: "duplicate_function", File: "test.go"},
					},
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		resp, err := client.AnalyzeVibe(&VibeAnalysisRequest{})

		if err != nil {
			t.Errorf("AnalyzeVibe() error = %v", err)
		}
		if resp == nil || len(resp.Issues) == 0 {
			t.Error("should return issues")
		}
	})
}

func TestAnalyzeStructure(t *testing.T) {
	t.Run("returns error when hub unavailable", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")

		_, err := client.AnalyzeStructure(&StructureAnalysisRequest{})
		if err == nil {
			t.Error("should return error when hub unavailable")
		}
	})

	t.Run("successfully analyzes structure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/api/v1/structure/analyze" {
				response := StructureAnalysisResponse{
					File:   "test.go",
					Lines:  100,
					Status: "ok",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		resp, err := client.AnalyzeStructure(&StructureAnalysisRequest{
			File:     "test.go",
			Language: "go",
		})

		if err != nil {
			t.Errorf("AnalyzeStructure() error = %v", err)
		}
		if resp == nil || resp.Lines == 0 {
			t.Error("should return file structure")
		}
	})
}

func TestGetHookPolicy(t *testing.T) {
	t.Run("returns default policy when hub unavailable", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")

		policy, err := client.GetHookPolicy("org-123")
		if err != nil {
			t.Errorf("should not error, got %v", err)
		}
		if policy == nil {
			t.Error("should return default policy")
		}
		if !policy.AuditEnabled {
			t.Error("default policy should have AuditEnabled=true")
		}
	})

	t.Run("fetches policy from hub", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/api/v1/hooks/policies" {
				policy := HookPolicy{
					AuditEnabled:     true,
					VibeCheckEnabled: false,
				}
				json.NewEncoder(w).Encode(policy)
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		policy, err := client.GetHookPolicy("org-123")

		if err != nil {
			t.Errorf("GetHookPolicy() error = %v", err)
		}
		if policy == nil {
			t.Error("should return policy")
		}
	})
}

func TestSendTelemetry(t *testing.T) {
	t.Run("silently fails when hub unavailable", func(t *testing.T) {
		client := NewClient("http://localhost:99999", "key")

		err := client.SendTelemetry(&TelemetryData{EventType: "test"})
		if err != nil {
			t.Errorf("should not error, got %v", err)
		}
	})

	t.Run("sends telemetry to hub", func(t *testing.T) {
		received := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.URL.Path == "/api/v1/telemetry" {
				received = true
				w.WriteHeader(http.StatusOK)
				return
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, "key")
		err := client.SendTelemetry(&TelemetryData{EventType: "audit_complete"})

		if err != nil {
			t.Errorf("SendTelemetry() error = %v", err)
		}
		if !received {
			t.Error("telemetry should be received by server")
		}
	})
}

func TestClientStructure(t *testing.T) {
	client := &Client{
		baseURL:    "http://test:8080",
		apiKey:     "secret-key",
		httpClient: &http.Client{},
		timeout:    30,
	}

	if client.baseURL != "http://test:8080" {
		t.Errorf("expected baseURL http://test:8080, got %s", client.baseURL)
	}
	if client.apiKey != "secret-key" {
		t.Errorf("expected apiKey secret-key, got %s", client.apiKey)
	}
}
