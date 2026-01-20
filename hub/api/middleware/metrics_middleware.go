// Package middleware provides HTTP middleware
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"sentinel-hub-api/pkg/metrics"
)

// MetricsMiddleware records HTTP metrics
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status
			wrapper := &responseWrapper{ResponseWriter: w, status: http.StatusOK}

			m.ActiveConnections.Inc()
			defer m.ActiveConnections.Dec()

			// Calculate request size
			requestSize := r.ContentLength
			if requestSize < 0 {
				requestSize = 0
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrapper.status)
			path := normalizePath(r.URL.Path)

			m.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
			m.HTTPRequestSize.WithLabelValues(r.Method, path).Observe(float64(requestSize))
			m.HTTPResponseSize.WithLabelValues(r.Method, path, status).Observe(float64(wrapper.size))
		})
	}
}

type responseWrapper struct {
	http.ResponseWriter
	status int
	size   int64
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	w.size += int64(len(b))
	return w.ResponseWriter.Write(b)
}

// normalizePath reduces cardinality by replacing IDs with placeholders
func normalizePath(path string) string {
	// Replace UUIDs and numeric IDs with placeholders
	parts := strings.Split(path, "/")
	for i, part := range parts {
		// Check if part looks like an ID (UUID or numeric)
		if len(part) > 0 {
			// UUID pattern: 8-4-4-4-12 hex digits
			if len(part) == 36 && strings.Count(part, "-") == 4 {
				parts[i] = ":id"
			} else if isNumeric(part) && len(part) > 3 {
				parts[i] = ":id"
			}
		}
	}
	return strings.Join(parts, "/")
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
