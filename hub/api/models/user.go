package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id" validate:"required,uuid"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Role      UserRole  `json:"role" db:"role" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents user roles in the system
type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleManager   UserRole = "manager"
	UserRoleDeveloper UserRole = "developer"
	UserRoleViewer    UserRole = "viewer"
)

// String returns the string representation of UserRole
func (r UserRole) String() string {
	return string(r)
}

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleManager, UserRoleDeveloper, UserRoleViewer:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (r UserRole) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (r *UserRole) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) > 0 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	*r = UserRole(s)
	if !r.IsValid() {
		*r = UserRoleViewer // Default to viewer for invalid roles
	}
	return nil
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email string   `json:"email" validate:"required,email"`
	Name  string   `json:"name" validate:"required,min=1,max=255"`
	Role  UserRole `json:"role" validate:"required"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email *string   `json:"email,omitempty" validate:"omitempty,email"`
	Name  *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Role  *UserRole `json:"role,omitempty" validate:"omitempty"`
}
