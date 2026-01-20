// Package handlers provides unit tests for HTTP handlers
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
	"github.com/divyang-garg/sentinel-hub-api/internal/services"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *services.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int, req *services.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

type UserHandlerTestSuite struct {
	suite.Suite
	mockService *MockUserService
	handler     *UserHandler
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.mockService = new(MockUserService)
	suite.handler = NewUserHandler(suite.mockService)
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (suite *UserHandlerTestSuite) TestCreateUser_Success() {
	req := services.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleUser,
	}

	suite.mockService.On("CreateUser", mock.Anything, &req).Return(expectedUser, nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.CreateUser(w, httpReq)

	suite.Equal(http.StatusCreated, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedUser.ID, response.ID)
	suite.Equal(expectedUser.Email, response.Email)
	suite.Equal(expectedUser.Name, response.Name)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestCreateUser_InvalidJSON() {
	httpReq := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.CreateUser(w, httpReq)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestCreateUser_ValidationError() {
	req := services.CreateUserRequest{
		Email:    "", // Invalid: empty email
		Name:     "Test User",
		Password: "password123",
	}

	// Mock service to return validation error
	validationErr := &models.ValidationError{
		Field:   "email",
		Message: "email is required",
	}
	suite.mockService.On("CreateUser", mock.Anything, &req).Return(nil, validationErr)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/users", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.CreateUser(w, httpReq)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("validation_failed", response["error"])
	suite.Equal("email", response["field"])
	suite.Equal("email is required", response["message"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGetUser_Success() {
	expectedUser := &models.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  models.RoleUser,
	}

	suite.mockService.On("GetUser", mock.Anything, 1).Return(expectedUser, nil)

	httpReq := httptest.NewRequest("GET", "/api/v1/users/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	suite.handler.GetUser(w, httpReq)

	suite.Equal(http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedUser.ID, response.ID)
	suite.Equal(expectedUser.Email, response.Email)
	suite.Equal(expectedUser.Name, response.Name)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGetUser_InvalidID() {
	httpReq := httptest.NewRequest("GET", "/api/v1/users/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	suite.handler.GetUser(w, httpReq)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid user ID", response["error"])
}

func (suite *UserHandlerTestSuite) TestGetUser_NotFound() {
	suite.mockService.On("GetUser", mock.Anything, 999).Return(nil, &models.NotFoundError{Resource: "user", ID: 999})

	httpReq := httptest.NewRequest("GET", "/api/v1/users/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	suite.handler.GetUser(w, httpReq)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user with id 999 not found", response["error"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestUpdateUser_Success() {
	req := services.UpdateUserRequest{
		Name:  stringPtr("Updated Name"),
		Email: stringPtr("updated@example.com"),
	}

	expectedUser := &models.User{
		ID:    1,
		Email: "updated@example.com",
		Name:  "Updated Name",
		Role:  models.RoleUser,
	}

	suite.mockService.On("UpdateUser", mock.Anything, 1, &req).Return(expectedUser, nil)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/api/v1/users/1", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	suite.handler.UpdateUser(w, httpReq)

	suite.Equal(http.StatusOK, w.Code)

	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedUser.Email, response.Email)
	suite.Equal(expectedUser.Name, response.Name)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestDeleteUser_Success() {
	suite.mockService.On("DeleteUser", mock.Anything, 1).Return(nil)

	httpReq := httptest.NewRequest("DELETE", "/api/v1/users/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	suite.handler.DeleteUser(w, httpReq)

	suite.Equal(http.StatusNoContent, w.Code)
	suite.Empty(w.Body.String())
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestHandleServiceError_ValidationError() {
	validationErr := &models.ValidationError{
		Field:   "email",
		Message: "invalid email format",
	}

	w := httptest.NewRecorder()
	suite.handler.handleServiceError(w, validationErr)

	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("validation_failed", response["error"])
	suite.Equal("email", response["field"])
	suite.Equal("invalid email format", response["message"])
}

func (suite *UserHandlerTestSuite) TestHandleServiceError_NotFoundError() {
	notFoundErr := &models.NotFoundError{
		Resource: "user",
		ID:       123,
	}

	w := httptest.NewRecorder()
	suite.handler.handleServiceError(w, notFoundErr)

	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user with id 123 not found", response["error"])
}

func (suite *UserHandlerTestSuite) TestHandleServiceError_AuthenticationError() {
	authErr := &models.AuthenticationError{
		Message: "invalid credentials",
	}

	w := httptest.NewRecorder()
	suite.handler.handleServiceError(w, authErr)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid credentials", response["error"])
}

func (suite *UserHandlerTestSuite) TestHandleServiceError_GenericError() {
	genericErr := assert.AnError

	w := httptest.NewRecorder()
	suite.handler.handleServiceError(w, genericErr)

	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("internal server error", response["error"])
}

func (suite *UserHandlerTestSuite) TestUpdateUser_InvalidJSON() {
	// Given: Invalid JSON in request body
	httpReq := httptest.NewRequest("PUT", "/api/v1/users/1", bytes.NewReader([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Updating user
	suite.handler.UpdateUser(w, httpReq)

	// Then: Should return bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestUpdateUser_InvalidID() {
	// Given: Non-numeric user ID in URL
	req := services.UpdateUserRequest{
		Name: stringPtr("Updated Name"),
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/api/v1/users/invalid", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Updating user
	suite.handler.UpdateUser(w, httpReq)

	// Then: Should return bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid user ID", response["error"])
}

func (suite *UserHandlerTestSuite) TestDeleteUser_InvalidID() {
	// Given: Non-numeric user ID in URL
	httpReq := httptest.NewRequest("DELETE", "/api/v1/users/invalid", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Deleting user
	suite.handler.DeleteUser(w, httpReq)

	// Then: Should return bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid user ID", response["error"])
}

func (suite *UserHandlerTestSuite) TestDeleteUser_ServiceError() {
	// Given: Service returns NotFoundError
	suite.mockService.On("DeleteUser", mock.Anything, 999).Return(&models.NotFoundError{Resource: "user", ID: 999})

	httpReq := httptest.NewRequest("DELETE", "/api/v1/users/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Deleting user
	suite.handler.DeleteUser(w, httpReq)

	// Then: Should return not found
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user with id 999 not found", response["error"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestHandleServiceError_AuthorizationError() {
	// Given: AuthorizationError
	authzErr := &models.AuthorizationError{
		Message: "insufficient permissions",
	}

	w := httptest.NewRecorder()

	// When: Handling service error
	suite.handler.handleServiceError(w, authzErr)

	// Then: Should return forbidden
	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("insufficient permissions", response["error"])
}

func (suite *UserHandlerTestSuite) TestUpdateUser_ValidationError() {
	// Given: Update request that causes validation error
	req := services.UpdateUserRequest{
		Email: stringPtr("invalid-email"),
	}

	validationErr := &models.ValidationError{
		Field:   "email",
		Message: "invalid email format",
	}
	suite.mockService.On("UpdateUser", mock.Anything, 1, &req).Return(nil, validationErr)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/api/v1/users/1", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Updating user
	suite.handler.UpdateUser(w, httpReq)

	// Then: Should return bad request
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("validation_failed", response["error"])
	suite.Equal("email", response["field"])
	suite.Equal("invalid email format", response["message"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestUpdateUser_NotFoundError() {
	// Given: Update request for non-existent user
	req := services.UpdateUserRequest{
		Name: stringPtr("Updated Name"),
	}

	notFoundErr := &models.NotFoundError{Resource: "user", ID: 999}
	suite.mockService.On("UpdateUser", mock.Anything, 999, &req).Return(nil, notFoundErr)

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("PUT", "/api/v1/users/999", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// When: Updating user
	suite.handler.UpdateUser(w, httpReq)

	// Then: Should return not found
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user with id 999 not found", response["error"])
	suite.mockService.AssertExpectations(suite.T())
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
