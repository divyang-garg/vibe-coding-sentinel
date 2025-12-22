package unit

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Note: These tests test the admin auth middleware logic conceptually.
// Actual middleware testing requires integration tests since the middleware
// is in package main and uses database/config dependencies.

// TestAdminAuthMiddleware_ConstantTimeEqual tests constant-time comparison logic
func TestAdminAuthMiddleware_ConstantTimeEqual(t *testing.T) {
	constantTimeEqual := func(a, b string) bool {
		if len(a) != len(b) {
			return false
		}
		result := 0
		for i := 0; i < len(a); i++ {
			result |= int(a[i]) ^ int(b[i])
		}
		return result == 0
	}

	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"equal strings", "test-key", "test-key", true},
		{"different strings", "test-key", "wrong-key", false},
		{"different lengths", "test-key", "test", false},
		{"empty strings", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constantTimeEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("constantTimeEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestAdminAuthMiddleware_InvalidKey tests that middleware rejects requests with invalid admin key
func TestAdminAuthMiddleware_InvalidKey(t *testing.T) {
	// Set admin key
	os.Setenv("ADMIN_API_KEY", "test-admin-key-123")
	defer os.Unsetenv("ADMIN_API_KEY")

	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("X-Admin-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestAdminAuthMiddleware_ValidKey_XHeader tests that middleware accepts valid admin key via X-Admin-API-Key header
func TestAdminAuthMiddleware_ValidKey_XHeader(t *testing.T) {
	adminKey := "test-admin-key-123"
	os.Setenv("ADMIN_API_KEY", adminKey)
	defer os.Unsetenv("ADMIN_API_KEY")

	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("X-Admin-API-Key", adminKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAdminAuthMiddleware_ValidKey_Bearer tests that middleware accepts valid admin key via Authorization Bearer header
func TestAdminAuthMiddleware_ValidKey_Bearer(t *testing.T) {
	adminKey := "test-admin-key-123"
	os.Setenv("ADMIN_API_KEY", adminKey)
	defer os.Unsetenv("ADMIN_API_KEY")

	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+adminKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestAdminAuthMiddleware_NotConfigured tests that middleware returns 500 if admin key not configured
func TestAdminAuthMiddleware_NotConfigured(t *testing.T) {
	// Ensure admin key is not set
	os.Unsetenv("ADMIN_API_KEY")

	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(adminAuthMiddleware)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("X-Admin-API-Key", "any-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

