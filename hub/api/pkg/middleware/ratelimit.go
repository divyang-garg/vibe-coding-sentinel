// Package middleware provides HTTP middleware utilities
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
	Burst    int           `json:"burst"`
}

// RateLimitInfo represents current rate limit status
type RateLimitInfo struct {
	Limit      int       `json:"limit"`
	Remaining  int       `json:"remaining"`
	ResetTime  time.Time `json:"reset_time"`
	WindowSize string    `json:"window_size"`
}

// RateLimitMiddleware implements rate limiting middleware
// Uses simple in-memory rate limiter (in production, use Redis or similar)
func RateLimitMiddleware() func(http.Handler) http.Handler {
	limits := make(map[string]*RateLimitInfo)
	var mu sync.RWMutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address for now)
			clientIP := getClientIP(r)

			// Get current rate limit info
			now := time.Now()
			limit := RateLimit{
				Requests: 100,         // 100 requests
				Window:   time.Minute, // per minute
				Burst:    10,          // burst allowance
			}

			key := clientIP + ":" + r.Method + ":" + r.URL.Path

			mu.Lock()
			info, exists := limits[key]
			if !exists || now.After(info.ResetTime) {
				// Reset or create new limit info
				info = &RateLimitInfo{
					Limit:      limit.Requests,
					Remaining:  limit.Requests - 1,
					ResetTime:  now.Add(limit.Window),
					WindowSize: "1m",
				}
				limits[key] = info
			} else {
				// Check if limit exceeded
				if info.Remaining <= 0 {
					mu.Unlock()
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))
					w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(info.ResetTime).Seconds())))

					rateLimitErr := &models.RateLimitError{
						Message:    "Rate limit exceeded",
						RetryAfter: int(time.Until(info.ResetTime).Seconds()),
						ResetTime:  info.ResetTime,
					}
					// Note: This assumes WriteErrorResponse is available or needs to be passed as parameter
					http.Error(w, rateLimitErr.Error(), http.StatusTooManyRequests)
					return
				}

				// Decrement remaining requests
				info.Remaining--
			}
			mu.Unlock()

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client's IP address from the request
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if idx := len(xff); idx > 0 {
			for i, char := range xff {
				if char == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
