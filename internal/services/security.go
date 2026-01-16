// Package services provides security-related business logic
// Complies with CODING_STANDARDS.md: Security implementations
package services

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements secure password hashing
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new bcrypt hasher
func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{cost: cost}
}

// Hash creates a bcrypt hash of the password
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

// Verify checks if password matches hash
func (h *BcryptPasswordHasher) Verify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
