// Package handlers - Unit tests for fix handler
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

// mockFixService implements services.FixService for testing
type mockFixService struct {
	mock.Mock
}

func (m *mockFixService) ApplyFix(ctx context.Context, req models.ApplyFixRequest) (*models.ApplyFixResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApplyFixResponse), args.Error(1)
}

func TestFixHandler_ApplyFix_Success(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Code:     "const apiKey = 'secret123';",
		Language: "javascript",
		FixType:  "security",
	}

	expectedResponse := &models.ApplyFixResponse{
		FixedCode: "const apiKey = process.env.API_KEY;",
		Changes: []map[string]interface{}{
			{
				"type":        "security",
				"description": "Remove hardcoded apiKey",
				"line":        1,
			},
		},
		Summary: "Applied 1 security fixes",
	}

	mockService.On("ApplyFix", mock.Anything, reqBody).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.ApplyFixResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.FixedCode, response.FixedCode)
	assert.Equal(t, len(expectedResponse.Changes), len(response.Changes))
	mockService.AssertExpectations(t)
}

func TestFixHandler_ApplyFix_InvalidJSON(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ApplyFix")
}

func TestFixHandler_ApplyFix_MissingCode(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Language: "javascript",
		FixType:  "security",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ApplyFix")
}

func TestFixHandler_ApplyFix_MissingLanguage(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Code:    "const x = 1;",
		FixType: "security",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ApplyFix")
}

func TestFixHandler_ApplyFix_MissingFixType(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Code:     "const x = 1;",
		Language: "javascript",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ApplyFix")
}

func TestFixHandler_ApplyFix_InvalidFixType(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Code:     "const x = 1;",
		Language: "javascript",
		FixType:  "invalid",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "ApplyFix")
}

func TestFixHandler_ApplyFix_ServiceError(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	reqBody := models.ApplyFixRequest{
		Code:     "const x = 1;",
		Language: "javascript",
		FixType:  "security",
	}

	mockService.On("ApplyFix", mock.Anything, reqBody).Return(nil, assert.AnError)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ApplyFix(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestFixHandler_ApplyFix_AllFixTypes(t *testing.T) {
	mockService := new(mockFixService)
	handler := NewFixHandler(mockService)

	fixTypes := []string{"security", "style", "performance"}

	for _, fixType := range fixTypes {
		reqBody := models.ApplyFixRequest{
			Code:     "const x = 1;",
			Language: "javascript",
			FixType:  fixType,
		}

		expectedResponse := &models.ApplyFixResponse{
			FixedCode: "const x = 1;",
			Changes:   []map[string]interface{}{},
			Summary:   "Applied 0 " + fixType + " fixes",
		}

		mockService.On("ApplyFix", mock.Anything, reqBody).Return(expectedResponse, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/fix/apply", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ApplyFix(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Fix type: %s", fixType)
	}
	mockService.AssertExpectations(t)
}
