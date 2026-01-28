// Package middleware provides rate limiting middleware
// Complies with CODING_STANDARDS.md: Middleware files max 200 lines
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     int           // requests per second
	capacity int           // maximum burst
	cleanup  time.Duration // cleanup interval
}

// tokenBucket represents a token bucket
type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	// Calculate rate per second
	rate := int(float64(requests) / window.Seconds())

	rl := &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: requests,
		cleanup:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupWorker()

	return rl
}

// SetCleanupInterval sets the cleanup interval (for testing)
// This should be called before the cleanup worker starts, or the worker
// should be restarted. For production use, set cleanup in NewRateLimiter.
func (rl *RateLimiter) SetCleanupInterval(interval time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.cleanup = interval
}

// Allow checks if request is allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// If capacity is 0, no requests are allowed
	if rl.capacity <= 0 {
		return false
	}

	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.capacity - 1, // Allow this request
			lastRefill: time.Now(),
		}
		rl.buckets[identifier] = bucket
		return true
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	refillTokens := int(elapsed.Seconds()) * rl.rate

	bucket.tokens += refillTokens
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}
	bucket.lastRefill = now

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// RateLimit middleware
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as identifier (in production, consider user ID)
		identifier := getClientIP(r)

		if !rl.Allow(identifier) {
			w.Header().Set("X-RateLimit-RetryAfter", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// cleanupWorker periodically removes old buckets
func (rl *RateLimiter) cleanupWorker() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for identifier, bucket := range rl.buckets {
			if time.Since(bucket.lastRefill) > rl.cleanup {
				delete(rl.buckets, identifier)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Take the first IP if multiple are present
		if idx := strings.Index(xForwardedFor, ","); idx > 0 {
			return strings.TrimSpace(xForwardedFor[:idx])
		}
		return strings.TrimSpace(xForwardedFor)
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
