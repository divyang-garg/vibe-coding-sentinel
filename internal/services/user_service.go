// Package services provides business logic implementations
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
	"github.com/divyang-garg/sentinel-hub-api/internal/repository"
)

// PostgresUserService implements UserService
type PostgresUserService struct {
	userRepo repository.UserRepository
	hasher   PasswordHasher
}

// NewPostgresUserService creates a new user service
func NewPostgresUserService(
	userRepo repository.UserRepository,
	hasher PasswordHasher,
) *PostgresUserService {
	return &PostgresUserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// CreateUser creates a new user with business logic validation
func (s *PostgresUserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	// Business logic validation
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !isNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, &models.ValidationError{Field: "email", Message: "user with this email already exists"}
	}

	// Hash password
	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  hashedPassword,
		Role:      models.RoleUser,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to repository
	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Don't return password in response
	created.Password = ""

	return created, nil
}

// GetUser retrieves a user by ID
func (s *PostgresUserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Don't return password
	user.Password = ""

	return user, nil
}

// UpdateUser updates an existing user
func (s *PostgresUserService) UpdateUser(ctx context.Context, id int, req *UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Check if email is already taken by another user
		existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil && !isNotFoundError(err) {
			return nil, fmt.Errorf("failed to check email availability: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, &models.ValidationError{Field: "email", Message: "email already in use"}
		}
		user.Email = *req.Email
	}

	user.UpdatedAt = time.Now()

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Don't return password
	user.Password = ""

	return user, nil
}

// DeleteUser deletes a user
func (s *PostgresUserService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepo.Delete(ctx, id)
}

// AuthenticateUser authenticates a user with email and password
func (s *PostgresUserService) AuthenticateUser(ctx context.Context, email, password string) (*models.User, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if isNotFoundError(err) {
			return nil, &models.AuthenticationError{Message: "invalid credentials"}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, &models.AuthenticationError{Message: "account is disabled"}
	}

	// Verify password
	if err := s.hasher.Verify(password, user.Password); err != nil {
		return nil, &models.AuthenticationError{Message: "invalid credentials"}
	}

	// Don't return password
	user.Password = ""

	return user, nil
}

// validateCreateUserRequest validates user creation request
func (s *PostgresUserService) validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Email == "" {
		return &models.ValidationError{Field: "email", Message: "email is required"}
	}
	if req.Name == "" {
		return &models.ValidationError{Field: "name", Message: "name is required"}
	}
	if req.Password == "" {
		return &models.ValidationError{Field: "password", Message: "password is required"}
	}
	if len(req.Password) < 8 {
		return &models.ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	return nil
}

// isNotFoundError checks if an error is a NotFoundError
func isNotFoundError(err error) bool {
	_, ok := err.(*models.NotFoundError)
	return ok
}
