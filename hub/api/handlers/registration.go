// Package handlers provides HTTP request handlers for the Sentinel API
// Complies with CODING_STANDARDS.md: Handler files max 300 lines
package handlers

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

// RegisterAll registers all HTTP handlers with the router
func RegisterAll(r *chi.Mux, db *sql.DB) {
	// Set database connection for handlers that need it
	SetDB(db)

	// Health check endpoints
	r.Get("/health", healthHandler)
	r.Get("/health/db", healthDBHandler)
	r.Get("/health/ready", healthReadyHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Register all API handlers here
		// This will be populated as we extract handlers from main.go
	})
}
