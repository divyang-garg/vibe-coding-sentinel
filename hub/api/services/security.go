// Phase 12: Security Middleware
// Provides security enhancements including CSRF protection, request size limits, and security headers

package services

import (
	"net/http"
	"strconv"
)

const (
	// DefaultMaxRequestSize is the default maximum request body size (10MB)
	DefaultMaxRequestSize = 10 * 1024 * 1024

	// DefaultMaxUploadSize is the default maximum upload size (50MB)
	DefaultMaxUploadSize = 50 * 1024 * 1024
)

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict Transport Security (HSTS) - only set if HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// requestSizeLimitMiddleware limits the size of request bodies
func requestSizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			// Set Content-Length header limit
			if r.ContentLength > maxSize {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				w.Write([]byte(`{"error": "Request body too large. Maximum size: ` + formatBytes(maxSize) + `"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// csrfProtectionMiddleware provides basic CSRF protection
// Note: For production, consider using a more robust CSRF library
func csrfProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for CSRF token in header
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			// For API endpoints, we can be more lenient - check Origin header instead
			origin := r.Header.Get("Origin")
			referer := r.Header.Get("Referer")

			// If no Origin/Referer, allow if it's an API request with proper auth
			if origin == "" && referer == "" {
				// Check if request has authorization header (API key or token)
				if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" {
					next.ServeHTTP(w, r)
					return
				}
			}

			// For browser requests, require Origin header
			if origin == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "CSRF protection: Missing Origin header"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "B"
}

// ipWhitelistMiddleware allows only requests from whitelisted IPs (optional)
func ipWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[ip] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no IPs specified, allow all
			if len(ipMap) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Check if IP is whitelisted
			if !ipMap[clientIP] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "Access denied: IP not whitelisted"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
