// Package services_test provides unit tests for business logic layer
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
	"github.com/divyang-garg/sentinel-hub-api/internal/services"
)

// Mock types for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Verify(password, hash string) error {
	args := m.Called(password, hash)
	return args.Error(0)
}

type UserServiceTestSuite struct {
	suite.Suite
	mockRepo   *MockUserRepository
	mockHasher *MockPasswordHasher
	service    *services.PostgresUserService
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockUserRepository)
	suite.mockHasher = new(MockPasswordHasher)
	suite.service = services.NewPostgresUserService(suite.mockRepo, suite.mockHasher)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	req := &services.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	// Mock repository calls
	suite.mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	suite.mockHasher.On("Hash", "password123").Return("hashedpassword", nil)
	suite.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == "test@example.com" && user.Name == "Test User"
	})).Return(&models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}, nil)

	user, err := suite.service.CreateUser(context.Background(), req)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("test@example.com", user.Email)
	suite.Equal("Test User", user.Name)
	suite.Equal("", user.Password) // Password should be cleared in response
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockHasher.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestCreateUser_ValidationError_EmailExists() {
	req := &services.CreateUserRequest{
		Email:    "existing@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	existingUser := &models.User{ID: 1, Email: "existing@example.com"}
	suite.mockRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	user, err := suite.service.CreateUser(context.Background(), req)

	suite.Error(err)
	suite.Nil(user)
	var validationErr *models.ValidationError
	suite.ErrorAs(err, &validationErr)
	suite.Equal("email", validationErr.Field)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestCreateUser_ValidationError_InvalidRequest() {
	req := &services.CreateUserRequest{
		Email:    "", // Invalid: empty email
		Name:     "Test User",
		Password: "password123",
	}

	user, err := suite.service.CreateUser(context.Background(), req)

	suite.Error(err)
	suite.Nil(user)
	var validationErr *models.ValidationError
	suite.ErrorAs(err, &validationErr)
	suite.Equal("email", validationErr.Field)
}

func (suite *UserServiceTestSuite) TestCreateUser_PasswordTooShort() {
	req := &services.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "123", // Too short
	}

	user, err := suite.service.CreateUser(context.Background(), req)

	suite.Error(err)
	suite.Nil(user)
	var validationErr *models.ValidationError
	suite.ErrorAs(err, &validationErr)
	suite.Equal("password", validationErr.Field)
}

func (suite *UserServiceTestSuite) TestGetUser_Success() {
	expectedUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
		IsActive: true,
	}

	suite.mockRepo.On("GetByID", mock.Anything, 1).Return(expectedUser, nil)

	user, err := suite.service.GetUser(context.Background(), 1)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(expectedUser.Email, user.Email)
	suite.Equal(expectedUser.Name, user.Name)
	suite.Equal("", user.Password) // Password should be cleared
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUser_NotFound() {
	suite.mockRepo.On("GetByID", mock.Anything, 999).Return(nil, &models.NotFoundError{Resource: "user", ID: 999})

	user, err := suite.service.GetUser(context.Background(), 999)

	suite.Error(err)
	suite.Nil(user)
	var notFoundErr *models.NotFoundError
	suite.ErrorAs(err, &notFoundErr)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestUpdateUser_Success() {
	req := &services.UpdateUserRequest{
		Name:  stringPtr("Updated Name"),
		Email: stringPtr("updated@example.com"),
	}

	existingUser := &models.User{
		ID:    1,
		Email: "old@example.com",
		Name:  "Old Name",
	}

	suite.mockRepo.On("GetByID", mock.Anything, 1).Return(existingUser, nil)
	suite.mockRepo.On("GetByEmail", mock.Anything, "updated@example.com").Return(nil, nil)
	suite.mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.ID == 1 && user.Name == "Updated Name" && user.Email == "updated@example.com"
	})).Return(nil)

	user, err := suite.service.UpdateUser(context.Background(), 1, req)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal("Updated Name", user.Name)
	suite.Equal("updated@example.com", user.Email)
	suite.Equal("", user.Password) // Password should be cleared
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestUpdateUser_EmailAlreadyExists() {
	req := &services.UpdateUserRequest{
		Email: stringPtr("existing@example.com"),
	}

	existingUser := &models.User{
		ID:    1,
		Email: "old@example.com",
		Name:  "Test User",
	}

	conflictingUser := &models.User{
		ID:    2,
		Email: "existing@example.com",
	}

	suite.mockRepo.On("GetByID", mock.Anything, 1).Return(existingUser, nil)
	suite.mockRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(conflictingUser, nil)

	user, err := suite.service.UpdateUser(context.Background(), 1, req)

	suite.Error(err)
	suite.Nil(user)
	var validationErr *models.ValidationError
	suite.ErrorAs(err, &validationErr)
	suite.Equal("email", validationErr.Field)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestDeleteUser_Success() {
	suite.mockRepo.On("Delete", mock.Anything, 1).Return(nil)

	err := suite.service.DeleteUser(context.Background(), 1)

	suite.NoError(err)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAuthenticateUser_Success() {
	email := "test@example.com"
	password := "password123"
	hashedPassword := "hashedpassword"

	user := &models.User{
		ID:       1,
		Email:    email,
		Name:     "Test User",
		Password: hashedPassword,
		Role:     models.RoleUser,
		IsActive: true,
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)
	suite.mockHasher.On("Verify", password, hashedPassword).Return(nil)

	authenticatedUser, err := suite.service.AuthenticateUser(context.Background(), email, password)

	suite.NoError(err)
	suite.NotNil(authenticatedUser)
	suite.Equal(user.Email, authenticatedUser.Email)
	suite.Equal(user.Name, authenticatedUser.Name)
	suite.Equal("", authenticatedUser.Password) // Password should be cleared
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockHasher.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAuthenticateUser_InvalidCredentials() {
	email := "test@example.com"
	password := "wrongpassword"

	user := &models.User{
		ID:       1,
		Email:    email,
		Name:     "Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
		IsActive: true,
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)
	suite.mockHasher.On("Verify", password, user.Password).Return(errors.New("invalid password"))

	authenticatedUser, err := suite.service.AuthenticateUser(context.Background(), email, password)

	suite.Error(err)
	suite.Nil(authenticatedUser)
	var authErr *models.AuthenticationError
	suite.ErrorAs(err, &authErr)
	suite.Equal("invalid credentials", authErr.Message)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockHasher.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAuthenticateUser_InactiveAccount() {
	email := "test@example.com"
	password := "password123"

	user := &models.User{
		ID:       1,
		Email:    email,
		Name:     "Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
		IsActive: false, // Account is inactive
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, email).Return(user, nil)

	authenticatedUser, err := suite.service.AuthenticateUser(context.Background(), email, password)

	suite.Error(err)
	suite.Nil(authenticatedUser)
	var authErr *models.AuthenticationError
	suite.ErrorAs(err, &authErr)
	suite.Equal("account is disabled", authErr.Message)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
