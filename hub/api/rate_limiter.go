// Phase 12: Rate Limiting Middleware
// Provides rate limiting protection for API endpoints

package main

import (
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

var (
	// Default rate limiter configuration
	defaultRateLimiter = &rateLimiter{
		limiter: rate.NewLimiter(rate.Limit(100), 200), // 100 req/s, burst 200
	}

	// Per-endpoint rate limiters
	endpointLimiters = make(map[string]*rateLimiter)
	limiterMutex     sync.RWMutex

	// Per-API-key rate limiters (Phase E: Security Hardening)
	apiKeyLimiters     = make(map[string]*rateLimiter)
	apiKeyLimiterMutex sync.RWMutex
)

type rateLimiter struct {
	limiter *rate.Limiter
}

// rateLimitMiddleware is defined in main.go

// getEndpointRateLimiter returns a rate limiter for a specific endpoint
func getEndpointRateLimiter(endpoint string) *rateLimiter {
	limiterMutex.RLock()
	defer limiterMutex.RUnlock()

	if limiter, ok := endpointLimiters[endpoint]; ok {
		return limiter
	}
	return defaultRateLimiter
}

// setEndpointRateLimiter sets a custom rate limiter for an endpoint
func setEndpointRateLimiter(endpoint string, rps float64, burst int) {
	limiterMutex.Lock()
	defer limiterMutex.Unlock()

	endpointLimiters[endpoint] = &rateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// rateLimitByEndpointMiddleware creates rate limiting middleware that applies different limits per endpoint
func rateLimitByEndpointMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := r.URL.Path
			limiter := getEndpointRateLimiter(endpoint)

			if !limiter.limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Rate limit exceeded for this endpoint. Please try again later."}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// checkAPIKeyRateLimit checks per-API-key rate limit (Phase E: Security Hardening)
// Returns true if request is allowed, false if rate limit exceeded
func checkAPIKeyRateLimit(apiKey string, endpoint string) bool {
	apiKeyLimiterMutex.Lock()
	defer apiKeyLimiterMutex.Unlock()

	// Get or create limiter for this API key
	limiter, exists := apiKeyLimiters[apiKey]
	if !exists {
		// Determine rate limit based on endpoint
		var rps float64 = 10 // Default: 10 req/s
		var burst int = 20   // Default burst: 20

		// Comprehensive analysis endpoints have stricter limits
		if strings.Contains(endpoint, "/analyze/comprehensive") {
			rps = 2 // 2 req/s for comprehensive analysis
			burst = 5
		} else if strings.Contains(endpoint, "/analyze/") {
			rps = 5 // 5 req/s for other analysis endpoints
			burst = 10
		}

		limiter = &rateLimiter{
			limiter: rate.NewLimiter(rate.Limit(rps), burst),
		}
		apiKeyLimiters[apiKey] = limiter
	}

	return limiter.limiter.Allow()
}
