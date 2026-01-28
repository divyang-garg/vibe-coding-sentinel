// Package middleware provides HTTP middleware for security and request handling.
// Complies with CODING_STANDARDS.md: HTTP middleware and routing logic
package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"encoding/json"
	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg/security"
	"sentinel-hub-api/services"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= 1.0 {
		rl.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(maxTokens, refillRate float64) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(maxTokens, refillRate)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For simplicity, using global limiter (in production, use per-IP limiters)
			if !limiter.Allow() {
				writeRateLimitError(w, &models.RateLimitError{
					Message: "Rate limit exceeded",
				})
				// Note: r is unused in this path but required by http.HandlerFunc signature
				_ = r
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddlewareConfig holds CORS configuration
type CORSMiddlewareConfig struct {
	AllowedOrigins []string
}

// CORSMiddleware creates CORS middleware with origin validation
func CORSMiddleware(config CORSMiddlewareConfig) func(http.Handler) http.Handler {
	originMap := make(map[string]bool)
	for _, origin := range config.AllowedOrigins {
		originMap[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine allowed origin based on environment
			var allowedOrigin string
			env := os.Getenv("ENV")

			if env == "development" || env == "dev" {
				// Development: allow all origins or specific ones
				if origin != "" {
					allowedOrigin = origin
				} else if len(config.AllowedOrigins) > 0 && originMap["*"] {
					allowedOrigin = "*"
				} else {
					allowedOrigin = "*"
				}
			} else {
				// Production: strict whitelist validation
				if origin == "" {
					// No origin header - reject CORS requests
					allowedOrigin = ""
				} else if originMap[origin] || originMap["*"] {
					allowedOrigin = origin
				} else {
					// Origin not in whitelist - reject
					http.Error(w, "CORS: Origin not allowed", http.StatusForbidden)
					return
				}
			}

			// Set CORS headers
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddlewareConfig holds configuration for authentication middleware
type AuthMiddlewareConfig struct {
	OrganizationService services.OrganizationService
	SkipPaths           []string // Paths to skip authentication
	Logger              Logger
	AuditLogger         security.AuditLogger // Security event logging
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// AuthMiddleware creates authentication middleware with service integration
func AuthMiddleware(config AuthMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for configured paths
			if shouldSkipAuth(r.URL.Path, config.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract API key from headers
			apiKey := extractAPIKey(r)
			if apiKey == "" {
				ipAddr := getClientIP(r)
				userAgent := r.UserAgent()

				if config.Logger != nil {
					config.Logger.Warn("Authentication failed: missing API key",
						"path", r.URL.Path,
						"ip", ipAddr,
					)
				}

				// Log security event
				if config.AuditLogger != nil {
					config.AuditLogger.LogAuthFailure(r.Context(), "missing API key", ipAddr, userAgent, r.URL.Path)
				}

				http.Error(w, "Unauthorized: API key required", http.StatusUnauthorized)
				return
			}

			// Validate API key via service layer
			if config.OrganizationService == nil {
				log.Printf("ERROR: OrganizationService not configured in AuthMiddleware")
				http.Error(w, "Internal server error: authentication service not configured", http.StatusInternalServerError)
				return
			}

			project, err := config.OrganizationService.ValidateAPIKey(r.Context(), apiKey)
			if err != nil || project == nil {
				ipAddr := getClientIP(r)
				userAgent := r.UserAgent()
				reason := "invalid API key"
				if err != nil {
					reason = err.Error()
				}

				if config.Logger != nil {
					config.Logger.Warn("Authentication failed: invalid API key",
						"path", r.URL.Path,
						"ip", ipAddr,
						"error", err,
					)
				}

				// Log security event
				if config.AuditLogger != nil {
					config.AuditLogger.LogAuthFailure(r.Context(), reason, ipAddr, userAgent, r.URL.Path)
				}

				http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
				return
			}

			// Add authenticated context with project information
			ctx := r.Context()
			ctx = context.WithValue(ctx, "project_id", project.ID)
			ctx = context.WithValue(ctx, "org_id", project.OrgID)
			ctx = context.WithValue(ctx, "api_key_prefix", project.APIKeyPrefix)
			r = r.WithContext(ctx)

			// Log successful authentication
			ipAddr := getClientIP(r)
			userAgent := r.UserAgent()

			if config.Logger != nil {
				config.Logger.Info("Authentication successful",
					"project_id", project.ID,
					"org_id", project.OrgID,
					"path", r.URL.Path,
				)
			}

			// Log security event
			if config.AuditLogger != nil {
				config.AuditLogger.LogAuthSuccess(r.Context(), project.ID, project.OrgID, ipAddr, userAgent)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractAPIKey extracts API key from request headers
func extractAPIKey(r *http.Request) string {
	// Check X-API-Key header first
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check Authorization header (Bearer token)
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// shouldSkipAuth checks if path should skip authentication
func shouldSkipAuth(path string, skipPaths []string) bool {
	// Always skip health endpoints
	if strings.HasPrefix(path, "/health") {
		return true
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLoggingMiddleware creates detailed request logging middleware
func RequestLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log request
			log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, getClientIP(r))

			next.ServeHTTP(w, r)

			// Log completion
			duration := time.Since(start)
			log.Printf("Completed: %s %s in %v", r.Method, r.URL.Path, duration)
		})
	}
}

// RecoveryMiddleware creates panic recovery middleware
func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("Panic recovered: %v", rec)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ValidateAPIKey validates an API key format
func ValidateAPIKey(apiKey string) bool {
	if len(apiKey) < 10 {
		return false
	}
	// Additional validation logic can be added here
	return true
}

// GetUserFromContext extracts user ID from request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetAPIKeyFromContext extracts API key from request context
func GetAPIKeyFromContext(ctx context.Context) (string, bool) {
	apiKey, ok := ctx.Value("api_key").(string)
	return apiKey, ok
}

// writeRateLimitError writes a standardized rate limit error response.
// This avoids importing handlers package to prevent import cycles.
func writeRateLimitError(w http.ResponseWriter, err *models.RateLimitError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"type":    "rate_limit_error",
			"message": err.Error(),
			"details": map[string]interface{}{
				"retry_after": err.RetryAfter,
				"reset_time":  err.ResetTime,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}
