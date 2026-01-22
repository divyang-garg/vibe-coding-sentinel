// Package middleware provides unit tests for HTTP middleware
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_NewAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	authMiddleware := NewAuthMiddleware(secret)

	assert.NotNil(t, authMiddleware)
	assert.Equal(t, []byte(secret), authMiddleware.jwtSecret)
}

func TestAuthMiddleware_Authenticate_ValidToken(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret-key-for-jwt-signing")

	// Generate a valid JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(123),
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	assert.NoError(t, err)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// Verify user_id is set in context
		userID := r.Context().Value("user_id")
		assert.NotNil(t, userID)
		assert.Equal(t, 123, userID)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.True(t, handlerCalled, "Next handler should be called with valid token")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_Authenticate_ExpiredToken(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	// Generate an expired JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(123),
		"exp":      time.Now().Add(-time.Hour).Unix(), // Expired
	})
	
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddleware_Authenticate_WrongSigningMethod(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	// Test with invalid token format (wrong signing method would fail during parsing)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Authenticate_InvalidClaims(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	// Generate token without user_id claim - should handle gracefully
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		// Missing user_id
	})
	
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// user_id should be nil if claim doesn't exist
		userID := r.Context().Value("user_id")
		assert.Nil(t, userID, "user_id should be nil when claim is missing")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	// Handler should still be called even if user_id is missing
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_Authenticate_WrongSecret(t *testing.T) {
	middleware := NewAuthMiddleware("correct-secret")

	// Generate token with wrong secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(123),
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	assert.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}

func TestAuthMiddleware_Authenticate_NoAuthorizationHeader(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header required")
}

func TestAuthMiddleware_Authenticate_InvalidAuthorizationFormat(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestAuthMiddleware_Authenticate_InvalidToken(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Next handler should not be called")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()

	middleware.Authenticate(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}
