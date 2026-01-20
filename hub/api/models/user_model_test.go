// Package models - User model tests
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUserRole tests UserRole enum functionality
func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role UserRole
		want bool
	}{
		{UserRoleAdmin, true},
		{UserRoleManager, true},
		{UserRoleDeveloper, true},
		{UserRoleViewer, true},
		{UserRole("invalid"), false},
	}

	for _, tt := range tests {
		if got := tt.role.IsValid(); got != tt.want {
			t.Errorf("UserRole.IsValid() = %v, want %v", got, tt.want)
		}
	}
}

func TestUserRole_JSONSerialization(t *testing.T) {
	role := UserRoleDeveloper

	// Test MarshalJSON
	data, err := role.MarshalJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"developer"`)

	// Test UnmarshalJSON
	var unmarshaled UserRole
	err = unmarshaled.UnmarshalJSON(data)
	assert.NoError(t, err)
	assert.Equal(t, UserRoleDeveloper, unmarshaled)
}

// TestUser model tests
func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				ID:        "user-123",
				Email:     "test@example.com",
				Name:      "Test User",
				Role:      UserRoleDeveloper,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			user: User{
				ID:        "user-123",
				Email:     "invalid-email",
				Name:      "Test User",
				Role:      UserRoleDeveloper,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			user: User{
				ID:        "user-123",
				Email:     "test@example.com",
				Name:      "Test User",
				Role:      UserRole("invalid-role"),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUser(&tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreateUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
				Role:  UserRoleDeveloper,
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			req: CreateUserRequest{
				Email: "invalid-email",
				Name:  "Test User",
				Role:  UserRoleDeveloper,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateUserRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: UpdateUserRequest{
				Name: stringPtr("Updated Name"),
				Role: func() *UserRole { r := UserRoleDeveloper; return &r }(),
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			req: UpdateUserRequest{
				Email: stringPtr("invalid-email"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateUserRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
