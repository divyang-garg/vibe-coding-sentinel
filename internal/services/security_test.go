// Package services provides unit tests for security implementations
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptPasswordHasher_Hash(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8) // Use lower cost for faster tests

	password := "testpassword123"

	hash, err := hasher.Hash(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash should not equal plain password

	// Verify the hash can be verified
	err = hasher.Verify(password, hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Hash_EmptyPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.DefaultCost)

	hash, err := hasher.Hash("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Empty password should still work
	err = hasher.Verify("", hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Verify_Success(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	password := "correctpassword"
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Verify(password, hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Verify_WrongPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	password := "correctpassword"
	wrongPassword := "wrongpassword"

	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	err = hasher.Verify(wrongPassword, hash)
	assert.Error(t, err)
	assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
}

func TestBcryptPasswordHasher_Verify_InvalidHash(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.DefaultCost)

	err := hasher.Verify("password", "invalidhash")
	assert.Error(t, err)
}

func TestBcryptPasswordHasher_CostConfiguration(t *testing.T) {
	// Test default cost
	hasher := NewBcryptPasswordHasher(0) // Should use default cost
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)

	// Test custom cost
	customCost := 10
	hasher = NewBcryptPasswordHasher(customCost)
	assert.Equal(t, customCost, hasher.cost)
}

func TestBcryptPasswordHasher_Hash_VerificationCompatibility(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	password := "testpassword123"

	// Hash with our hasher
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	// Verify with standard bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err)

	// Also verify that standard bcrypt hash works with our hasher
	standardHash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	assert.NoError(t, err)

	err = hasher.Verify(password, string(standardHash))
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Hash_Uniqueness(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	password := "samepassword"

	// Hash the same password multiple times
	hash1, err := hasher.Hash(password)
	assert.NoError(t, err)

	hash2, err := hasher.Hash(password)
	assert.NoError(t, err)

	// Hashes should be different (due to salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should verify correctly
	err = hasher.Verify(password, hash1)
	assert.NoError(t, err)

	err = hasher.Verify(password, hash2)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Hash_LongPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	// bcrypt has a maximum password length of 72 bytes
	longPassword := string(make([]byte, 100)) // 100 bytes, longer than bcrypt limit

	// Hashing should fail for passwords longer than 72 bytes
	hash, err := hasher.Hash(longPassword)
	assert.Error(t, err, "bcrypt should reject passwords longer than 72 bytes")
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "password length exceeds 72 bytes")

	// Test with a password exactly at the limit
	validLongPassword := string(make([]byte, 72)) // Exactly 72 bytes
	hash, err = hasher.Hash(validLongPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Should verify correctly
	err = hasher.Verify(validLongPassword, hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Hash_SpecialCharacters(t *testing.T) {
	hasher := NewBcryptPasswordHasher(8)

	password := "P@ssw0rd!#$%^&*()_+{}|:<>?[]\\;',./"

	hash, err := hasher.Hash(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = hasher.Verify(password, hash)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_Verify_EmptyInputs(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.DefaultCost)

	// Empty password with valid hash
	hash, err := hasher.Hash("")
	assert.NoError(t, err)

	err = hasher.Verify("", hash)
	assert.NoError(t, err)

	// Valid password with empty hash
	err = hasher.Verify("password", "")
	assert.Error(t, err)
}

func BenchmarkBcryptPasswordHasher_Hash(b *testing.B) {
	hasher := NewBcryptPasswordHasher(8) // Lower cost for benchmark
	password := "benchmarkpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Hash(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBcryptPasswordHasher_Verify(b *testing.B) {
	hasher := NewBcryptPasswordHasher(8)
	password := "benchmarkpassword123"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := hasher.Verify(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}
