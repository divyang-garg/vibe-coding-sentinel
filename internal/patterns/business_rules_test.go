// Package patterns provides tests for business rules functionality
package patterns

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchBusinessRulesFromHub(t *testing.T) {
	t.Run("empty hub URL", func(t *testing.T) {
		rules, err := fetchBusinessRulesFromHub("", "key", "project")
		if err == nil {
			t.Error("Expected error for empty hub URL")
		}
		if rules != nil && len(rules) > 0 {
			t.Error("Expected empty rules on error")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		rules, err := fetchBusinessRulesFromHub(":://invalid", "key", "project")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if rules != nil && len(rules) > 0 {
			t.Error("Expected empty rules on error")
		}
	})

	t.Run("network error returns empty slice", func(t *testing.T) {
		// Use unreachable URL to simulate network error
		rules, err := fetchBusinessRulesFromHub("http://127.0.0.1:99999/api", "key", "project")
		if err != nil {
			t.Errorf("Expected no error on network failure, got: %v", err)
		}
		if rules == nil {
			t.Error("Expected empty slice, not nil")
		}
		if len(rules) > 0 {
			t.Error("Expected empty rules on network error")
		}
	})

	t.Run("HTTP error status returns empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		rules, err := fetchBusinessRulesFromHub(server.URL, "key", "project")
		if err != nil {
			t.Errorf("Expected no error on HTTP error, got: %v", err)
		}
		if rules == nil {
			t.Error("Expected empty slice, not nil")
		}
		if len(rules) > 0 {
			t.Error("Expected empty rules on HTTP error")
		}
	})

	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rules": [
					{
						"id": "BR1",
						"type": "business_rule",
						"title": "Test Rule",
						"content": "Test content",
						"confidence": 0.9,
						"source_page": 1,
						"status": "approved"
					}
				],
				"entities": [],
				"user_journeys": []
			}`))
		}))
		defer server.Close()

		rules, err := fetchBusinessRulesFromHub(server.URL, "key", "project")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Errorf("Expected 1 rule, got %d", len(rules))
		}
		if rules[0].ID != "BR1" {
			t.Errorf("Expected rule ID 'BR1', got '%s'", rules[0].ID)
		}
		if rules[0].Title != "Test Rule" {
			t.Errorf("Expected title 'Test Rule', got '%s'", rules[0].Title)
		}
	})

	t.Run("filters non-approved rules", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rules": [
					{
						"id": "BR1",
						"type": "business_rule",
						"title": "Approved Rule",
						"content": "Content",
						"status": "approved"
					},
					{
						"id": "BR2",
						"type": "business_rule",
						"title": "Draft Rule",
						"content": "Content",
						"status": "draft"
					}
				],
				"entities": [],
				"user_journeys": []
			}`))
		}))
		defer server.Close()

		rules, err := fetchBusinessRulesFromHub(server.URL, "key", "project")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Errorf("Expected 1 approved rule, got %d", len(rules))
		}
		if rules[0].ID != "BR1" {
			t.Errorf("Expected approved rule BR1, got %s", rules[0].ID)
		}
	})

	t.Run("filters non-rule types", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rules": [
					{
						"id": "BR1",
						"type": "entity",
						"title": "Not a Rule",
						"content": "Content",
						"status": "approved"
					}
				],
				"entities": [],
				"user_journeys": []
			}`))
		}))
		defer server.Close()

		rules, err := fetchBusinessRulesFromHub(server.URL, "key", "project")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(rules) != 0 {
			t.Errorf("Expected 0 rules (entity type filtered), got %d", len(rules))
		}
	})

	t.Run("invalid JSON returns empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		rules, err := fetchBusinessRulesFromHub(server.URL, "key", "project")
		if err != nil {
			t.Errorf("Expected no error on invalid JSON, got: %v", err)
		}
		if rules == nil {
			t.Error("Expected empty slice, not nil")
		}
		if len(rules) > 0 {
			t.Error("Expected empty rules on invalid JSON")
		}
	})

	t.Run("project ID in query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("project_id") != "test-project" {
				t.Errorf("Expected project_id=test-project, got %s", r.URL.Query().Get("project_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"rules": [], "entities": [], "user_journeys": []}`))
		}))
		defer server.Close()

		_, err := fetchBusinessRulesFromHub(server.URL, "key", "test-project")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("API key in headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			apiKey := r.Header.Get("X-API-Key")
			if auth != "Bearer test-key" && apiKey != "test-key" {
				t.Errorf("Expected API key in headers, got Authorization=%s, X-API-Key=%s", auth, apiKey)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"rules": [], "entities": [], "user_journeys": []}`))
		}))
		defer server.Close()

		_, err := fetchBusinessRulesFromHub(server.URL, "test-key", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}
