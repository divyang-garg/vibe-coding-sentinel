// Package handlers - Unit tests for LLM handler
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLLMService implements LLMService for testing
type mockLLMService struct {
	mock.Mock
}

func (m *mockLLMService) ValidateConfig(ctx context.Context, config models.LLMConfig) (*models.ValidateLLMConfigResponse, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ValidateLLMConfigResponse), args.Error(1)
}

func TestLLMHandler_ValidateLLMConfig_Success(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test1234567890",
			Model:    "gpt-4",
			KeyType:  "api_key",
		},
	}

	expectedResponse := &models.ValidateLLMConfigResponse{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	mockService.On("ValidateConfig", mock.Anything, reqBody.Config).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.ValidateLLMConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Valid)
	assert.Empty(t, response.Errors)
	mockService.AssertExpectations(t)
}

func TestLLMHandler_ValidateLLMConfig_InvalidJSON(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ValidateConfig")
}

func TestLLMHandler_ValidateLLMConfig_MissingProvider(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			APIKey: "sk-test1234567890",
			Model:  "gpt-4",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ValidateConfig")
}

func TestLLMHandler_ValidateLLMConfig_MissingAPIKey(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ValidateConfig")
}

func TestLLMHandler_ValidateLLMConfig_MissingModel(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test1234567890",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ValidateConfig")
}

func TestLLMHandler_ValidateLLMConfig_WithErrors(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test1234567890",
			Model:    "gpt-4",
		},
	}

	expectedResponse := &models.ValidateLLMConfigResponse{
		Valid:    false,
		Errors:   []string{"Invalid provider"},
		Warnings: []string{},
	}

	mockService.On("ValidateConfig", mock.Anything, reqBody.Config).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.ValidateLLMConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Valid)
	assert.NotEmpty(t, response.Errors)
	mockService.AssertExpectations(t)
}

func TestLLMHandler_ValidateLLMConfig_ServiceError(t *testing.T) {
	mockService := new(mockLLMService)
	handler := NewLLMHandler(mockService)

	reqBody := models.ValidateLLMConfigRequest{
		Config: models.LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test1234567890",
			Model:    "gpt-4",
		},
	}

	mockService.On("ValidateConfig", mock.Anything, reqBody.Config).Return(nil, assert.AnError)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/llm/validate-config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateLLMConfig(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
