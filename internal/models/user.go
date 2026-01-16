// Package models contains data models and types
// Complies with CODING_STANDARDS.md: Data models max 200 lines
package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password"` // Never expose in JSON
	Role      UserRole  `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents user role types
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// Validate validates user data
func (u *User) Validate() error {
	if u.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if u.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	return nil
}
