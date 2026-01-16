// Package services defines business logic interfaces
// Complies with CODING_STANDARDS.md: Interface-based design
package services

import (
	"context"

	"github.com/divyang-garg/sentinel-hub-api/internal/models"
)

// UserService defines user business logic methods
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, id int, req *UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id int) error
	AuthenticateUser(ctx context.Context, email, password string) (*models.User, error)
}

// PasswordHasher defines password hashing interface
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// CreateUserRequest represents user creation request
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// UserResponse represents user response (without sensitive data)
type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
