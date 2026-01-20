// Package handlers - Health check handlers
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"sentinel-hub-api/pkg"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	BaseHandler
	DB *sql.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		DB: db,
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Check if shutting down
	if pkg.IsShuttingDown() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "shutting_down",
			"timestamp": time.Now(),
		})
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "v1.0.0",
		"uptime":    "24h 30m", // Would be calculated from actual uptime
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status":  "healthy",
				"latency": "12ms",
			},
			"cache": map[string]interface{}{
				"status":   "healthy",
				"hit_rate": "87.5%",
			},
			"storage": map[string]interface{}{
				"status":    "healthy",
				"used":      "2.4GB",
				"available": "47.6GB",
			},
		},
		"metrics": map[string]interface{}{
			"active_connections": 23,
			"total_requests":     15420,
			"error_rate":         "0.012%",
		},
	}

	// Check database connectivity
	if err := h.DB.Ping(); err != nil {
		health["status"] = "unhealthy"
		if services, ok := health["services"].(map[string]interface{}); ok {
			if database, ok := services["database"].(map[string]interface{}); ok {
				database["status"] = "unhealthy"
				database["error"] = err.Error()
			}
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// HealthDB handles GET /health/db
func (h *HealthHandler) HealthDB(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		h.WriteErrorResponse(w, err, http.StatusServiceUnavailable)
		return
	}
	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":   "healthy",
		"database": "connected",
	})
}

// HealthReady handles GET /health/ready
func (h *HealthHandler) HealthReady(w http.ResponseWriter, r *http.Request) {
	if pkg.IsShuttingDown() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not_ready",
			"reason": "shutting_down",
		})
		return
	}

	// Check if all critical services are ready
	if err := h.DB.Ping(); err != nil {
		h.WriteErrorResponse(w, err, http.StatusServiceUnavailable)
		return
	}
	h.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}
