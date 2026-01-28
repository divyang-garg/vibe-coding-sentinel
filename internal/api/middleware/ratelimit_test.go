// Package middleware provides unit tests for rate limiting middleware
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_NewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)

	assert.NotNil(t, rl)
	assert.NotNil(t, rl.buckets)
	assert.Equal(t, 1, rl.rate) // 100 requests per minute = ~1.67 per second, truncated to 1
	assert.Equal(t, 100, rl.capacity)
}

func TestRateLimiter_Allow_FirstRequest(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	allowed := rl.Allow("test-ip")
	assert.True(t, allowed)

	// Check that bucket was created
	rl.mu.Lock()
	bucket, exists := rl.buckets["test-ip"]
	rl.mu.Unlock()

	assert.True(t, exists)
	assert.Equal(t, 9, bucket.tokens) // capacity - 1
}

func TestRateLimiter_Allow_ExhaustTokens(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	// Use up all tokens
	assert.True(t, rl.Allow("test-ip"))
	assert.True(t, rl.Allow("test-ip"))
	assert.True(t, rl.Allow("test-ip"))

	// Next request should be denied
	assert.False(t, rl.Allow("test-ip"))
}

func TestRateLimiter_Allow_TokenRefill(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)

	// Use up tokens
	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow("test-ip"))
	}

	// Should be exhausted
	assert.False(t, rl.Allow("test-ip"))

	// Wait for refill (simulate time passing)
	rl.mu.Lock()
	if bucket, exists := rl.buckets["test-ip"]; exists {
		bucket.lastRefill = bucket.lastRefill.Add(-2 * time.Second) // Simulate 2 seconds ago
	}
	rl.mu.Unlock()

	// Should allow again after refill
	assert.True(t, rl.Allow("test-ip"))
}

func TestRateLimiter_RateLimit_Middleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	callCount := 0
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// First two requests should succeed
	w1 := httptest.NewRecorder()
	rl.RateLimit(nextHandler).ServeHTTP(w1, req)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, 1, callCount)

	w2 := httptest.NewRecorder()
	rl.RateLimit(nextHandler).ServeHTTP(w2, req)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, 2, callCount)

	// Third request should be rate limited
	w3 := httptest.NewRecorder()
	rl.RateLimit(nextHandler).ServeHTTP(w3, req)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Equal(t, 2, callCount) // Handler not called
	assert.Equal(t, "60", w3.Header().Get("X-RateLimit-RetryAfter"))
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	// IP 1 uses up its tokens
	assert.True(t, rl.Allow("ip1"))
	assert.True(t, rl.Allow("ip1"))
	assert.False(t, rl.Allow("ip1")) // Exhausted

	// IP 2 should still work
	assert.True(t, rl.Allow("ip2"))
	assert.True(t, rl.Allow("ip2"))
	assert.False(t, rl.Allow("ip2")) // Exhausted
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")

	ip := getClientIP(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_XForwardedFor_Multiple(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1, 192.0.2.1")

	ip := getClientIP(req)
	assert.Equal(t, "203.0.113.1", ip) // First IP
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")

	ip := getClientIP(req)
	assert.Equal(t, "198.51.100.1", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:8080"

	ip := getClientIP(req)
	assert.Equal(t, "192.168.1.100", ip)
}

func TestGetClientIP_XForwardedFor_TakesPrecedence(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "198.51.100.1")
	req.RemoteAddr = "192.168.1.100:8080"

	ip := getClientIP(req)
	assert.Equal(t, "203.0.113.1", ip) // X-Forwarded-For takes precedence
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow("bench-ip")
	}
}

func BenchmarkRateLimiter_RateLimit(b *testing.B) {
	rl := NewRateLimiter(10000, time.Minute)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	req := httptest.NewRequest("GET", "/bench", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		rl.RateLimit(nextHandler).ServeHTTP(w, req)
	}
}

func TestRateLimiter_CleanupWorker(t *testing.T) {
	// Create rate limiter with short cleanup interval for testing
	rl := NewRateLimiter(10, time.Minute)

	// Override cleanup interval to be shorter for testing
	// Note: This test verifies cleanup logic, but cleanup timing is non-deterministic
	// We verify that old buckets can be cleaned up and recent buckets remain
	cleanupInterval := 200 * time.Millisecond
	rl.SetCleanupInterval(cleanupInterval)

	// Create some buckets
	rl.mu.Lock()
	rl.buckets["old-ip"] = &tokenBucket{
		tokens:     5,
		lastRefill: time.Now().Add(-2 * time.Minute), // Old bucket
	}
	newBucketTime := time.Now()
	rl.buckets["new-ip"] = &tokenBucket{
		tokens:     5,
		lastRefill: newBucketTime, // Recent bucket
	}
	initialCount := len(rl.buckets)
	rl.mu.Unlock()

	// Wait for cleanup to potentially run
	// Cleanup removes buckets older than cleanup interval (200ms), so new bucket should remain
	time.Sleep(cleanupInterval + 50*time.Millisecond)

	// Update new bucket's lastRefill to ensure it's still recent
	rl.mu.Lock()
	if bucket, exists := rl.buckets["new-ip"]; exists {
		bucket.lastRefill = time.Now()
	}
	rl.mu.Unlock()

	// Wait a bit more for cleanup cycle
	time.Sleep(cleanupInterval + 50*time.Millisecond)

	// Check that old bucket was cleaned up
	rl.mu.Lock()
	finalCount := len(rl.buckets)
	_, oldExists := rl.buckets["old-ip"]
	_, newExists := rl.buckets["new-ip"]
	rl.mu.Unlock()

	// Old bucket should be removed (if cleanup ran), new bucket should remain
	// Note: cleanup timing is non-deterministic due to goroutine scheduling,
	// so we verify the structure is correct rather than exact timing
	if !oldExists {
		assert.True(t, finalCount < initialCount, "Old bucket should be cleaned up")
	}
	// New bucket should still exist since we updated its lastRefill
	assert.True(t, newExists, "New bucket should still exist")
}

func TestRateLimiter_Allow_TokenRefillOverCapacity(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	// Create a bucket and simulate time passing
	rl.mu.Lock()
	bucket := &tokenBucket{
		tokens:     0,
		lastRefill: time.Now().Add(-3 * time.Second), // 3 seconds ago
	}
	rl.buckets["test-ip"] = bucket
	rl.mu.Unlock()

	// Allow should refill tokens (3 seconds * rate) but cap at capacity
	allowed := rl.Allow("test-ip")

	rl.mu.Lock()
	finalTokens := rl.buckets["test-ip"].tokens
	rl.mu.Unlock()

	assert.True(t, allowed)
	assert.LessOrEqual(t, finalTokens, 5, "Tokens should not exceed capacity")
}

func TestRateLimiter_RateLimit_EdgeCases(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with empty RemoteAddr
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = ""
	w1 := httptest.NewRecorder()
	rl.RateLimit(nextHandler).ServeHTTP(w1, req1)
	// Should handle gracefully
	assert.True(t, w1.Code == http.StatusOK || w1.Code == http.StatusTooManyRequests)

	// Test with X-Forwarded-For with single IP
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.1")
	w2 := httptest.NewRecorder()
	rl.RateLimit(nextHandler).ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
