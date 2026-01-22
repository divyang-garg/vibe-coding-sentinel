// Package extraction provides tests for LLM client
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package extraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOllamaConfig(t *testing.T) {
	t.Run("returns default configuration", func(t *testing.T) {
		// When
		cfg := DefaultOllamaConfig()

		// Then
		assert.Equal(t, "http://localhost:11434", cfg.BaseURL)
		assert.Equal(t, "llama3.2", cfg.Model)
		assert.Equal(t, 120*time.Second, cfg.Timeout)
	})

	t.Run("uses environment variables when set", func(t *testing.T) {
		// Given
		os.Setenv("OLLAMA_HOST", "http://custom:11434")
		os.Setenv("OLLAMA_MODEL", "custom-model")
		defer func() {
			os.Unsetenv("OLLAMA_HOST")
			os.Unsetenv("OLLAMA_MODEL")
		}()

		// When
		cfg := DefaultOllamaConfig()

		// Then
		assert.Equal(t, "http://custom:11434", cfg.BaseURL)
		assert.Equal(t, "custom-model", cfg.Model)
	})
}

func TestNewOllamaClient(t *testing.T) {
	t.Run("creates client with configuration", func(t *testing.T) {
		// Given
		cfg := OllamaConfig{
			BaseURL: "http://localhost:11434",
			Model:   "test-model",
			Timeout: 30 * time.Second,
		}

		// When
		client := NewOllamaClient(cfg)

		// Then
		assert.NotNil(t, client)
		assert.Equal(t, "http://localhost:11434", client.baseURL)
		assert.Equal(t, "test-model", client.model)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	})
}

func TestOllamaClient_Call(t *testing.T) {
	t.Run("successful call returns response", func(t *testing.T) {
		// Given
		expectedResponse := "Test response from LLM"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/generate", r.URL.Path)

			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, "test-model", reqBody["model"])

			response := map[string]interface{}{
				"response": expectedResponse,
				"context":  []int{1, 2, 3},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := OllamaConfig{
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		}
		client := NewOllamaClient(cfg)

		// When
		result, tokens, err := client.Call(context.Background(), "test prompt", "test_task")

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		assert.Greater(t, tokens, 0)
	})

	t.Run("handles HTTP errors", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := OllamaConfig{
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		}
		client := NewOllamaClient(cfg)

		// When
		result, tokens, err := client.Call(context.Background(), "test prompt", "test_task")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
		assert.Empty(t, result)
		assert.Equal(t, 0, tokens)
	})

	t.Run("handles invalid JSON response", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		cfg := OllamaConfig{
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		}
		client := NewOllamaClient(cfg)

		// When
		result, tokens, err := client.Call(context.Background(), "test prompt", "test_task")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
		assert.Empty(t, result)
		assert.Equal(t, 0, tokens)
	})

	t.Run("handles network errors", func(t *testing.T) {
		// Given
		cfg := OllamaConfig{
			BaseURL: "http://nonexistent:11434",
			Model:   "test-model",
			Timeout: 1 * time.Second,
		}
		client := NewOllamaClient(cfg)

		// When
		result, tokens, err := client.Call(context.Background(), "test prompt", "test_task")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
		assert.Empty(t, result)
		assert.Equal(t, 0, tokens)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		// Given
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer server.Close()

		cfg := OllamaConfig{
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		}
		client := NewOllamaClient(cfg)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// When
		result, tokens, err := client.Call(ctx, "test prompt", "test_task")

		// Then
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 0, tokens)
	})

	t.Run("estimates tokens correctly", func(t *testing.T) {
		// Given
		longPrompt := "This is a very long prompt " + string(make([]byte, 1000))
		longResponse := "This is a very long response " + string(make([]byte, 2000))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"response": longResponse,
				"context":  []int{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := OllamaConfig{
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		}
		client := NewOllamaClient(cfg)

		// When
		result, tokens, err := client.Call(context.Background(), longPrompt, "test_task")

		// Then
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Greater(t, tokens, 0)
		// Token estimation: (prompt_length + response_length) / 4
		expectedTokens := (len(longPrompt) + len(longResponse)) / 4
		assert.InDelta(t, expectedTokens, tokens, float64(expectedTokens)*0.1) // 10% tolerance
	})
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("returns environment variable when set", func(t *testing.T) {
		// Given
		os.Setenv("TEST_ENV_VAR", "custom-value")
		defer os.Unsetenv("TEST_ENV_VAR")

		// When
		result := getEnvOrDefault("TEST_ENV_VAR", "default-value")

		// Then
		assert.Equal(t, "custom-value", result)
	})

	t.Run("returns default when not set", func(t *testing.T) {
		// Given
		os.Unsetenv("NONEXISTENT_VAR")

		// When
		result := getEnvOrDefault("NONEXISTENT_VAR", "default-value")

		// Then
		assert.Equal(t, "default-value", result)
	})

	t.Run("returns default when empty", func(t *testing.T) {
		// Given
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		// When
		result := getEnvOrDefault("EMPTY_VAR", "default-value")

		// Then
		assert.Equal(t, "default-value", result)
	})
}
