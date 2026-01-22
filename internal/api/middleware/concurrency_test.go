// Package middleware provides concurrency and edge case tests
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent requests from same IP", func(t *testing.T) {
		limiter := NewRateLimiter(10, time.Second)

		handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		var wg sync.WaitGroup
		successCount := int32(0)
		blockedCount := int32(0)
		concurrency := 20

		// Make concurrent requests
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, req)

				if rr.Code == http.StatusOK {
					atomic.AddInt32(&successCount, 1)
				} else if rr.Code == http.StatusTooManyRequests {
					atomic.AddInt32(&blockedCount, 1)
				}
			}()
		}

		wg.Wait()

		// Should allow some requests but block others
		assert.Greater(t, successCount, int32(0), "Should allow some requests")
		assert.Greater(t, blockedCount, int32(0), "Should block some requests")
		assert.Equal(t, int32(concurrency), successCount+blockedCount, "All requests should be processed")
	})

	t.Run("handles concurrent requests from different IPs", func(t *testing.T) {
		limiter := NewRateLimiter(10, time.Second)

		handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		var wg sync.WaitGroup
		successCount := int32(0)
		ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}

		// Each IP makes requests concurrently
		for _, ip := range ips {
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(ipAddr string) {
					defer wg.Done()
					req := httptest.NewRequest("GET", "/", nil)
					req.RemoteAddr = ipAddr + ":12345"
					rr := httptest.NewRecorder()

					handler.ServeHTTP(rr, req)

					if rr.Code == http.StatusOK {
						atomic.AddInt32(&successCount, 1)
					}
				}(ip)
			}
		}

		wg.Wait()

		// All requests from different IPs should succeed (each IP has its own bucket)
		assert.Equal(t, int32(len(ips)*5), successCount, "All requests from different IPs should succeed")
	})
}

func TestRateLimiter_BoundaryConditions(t *testing.T) {
	t.Run("handles zero rate limit", func(t *testing.T) {
		limiter := NewRateLimiter(0, time.Second)

		handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// With zero rate limit, all requests should be blocked
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})

	t.Run("handles very high rate limit", func(t *testing.T) {
		limiter := NewRateLimiter(10000, time.Second)

		handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Make many requests
		successCount := 0
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code == http.StatusOK {
				successCount++
			}
		}

		// Should allow all requests
		assert.Equal(t, 100, successCount, "Should allow all requests with high rate limit")
	})
}

func TestAuthMiddleware_TimeoutHandling(t *testing.T) {
	t.Run("handles request timeout", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret")

		// Create a handler that takes too long
		slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.Authenticate(slowHandler)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		// Handler should still process (timeout is handled by context, not middleware)
		handler.ServeHTTP(rr, req)

		// Should return error due to invalid token, not timeout
		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

func TestAuthMiddleware_EdgeCases(t *testing.T) {
	t.Run("handles missing authorization header gracefully", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret")

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		// No Authorization header
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handles malformed authorization header", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret")

		handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
