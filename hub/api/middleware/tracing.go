// Package middleware provides HTTP middleware
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
	"context"
	"net/http"
	
	"github.com/google/uuid"
	"sentinel-hub-api/pkg"
)

// TracingMiddleware adds correlation IDs to requests
func TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate trace ID
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.New().String()
			}
			
			// Extract or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			// Generate span ID for this request
			spanID := uuid.New().String()[:8]
			
			// Add to context
			ctx := context.WithValue(r.Context(), pkg.RequestIDKey, requestID)
			ctx = context.WithValue(ctx, pkg.TraceIDKey, traceID)
			ctx = context.WithValue(ctx, pkg.SpanIDKey, spanID)
			
			// Add to response headers for client correlation
			w.Header().Set("X-Trace-ID", traceID)
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Span-ID", spanID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(pkg.TraceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(pkg.RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpanID extracts span ID from context
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(pkg.SpanIDKey).(string); ok {
		return id
	}
	return ""
}
