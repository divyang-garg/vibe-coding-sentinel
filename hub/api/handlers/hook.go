// Package handlers - Hook and telemetry HTTP handlers
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package handlers

import (
	"database/sql"
	"net/http"
)

// HookHandler handles hook and telemetry HTTP requests
type HookHandler struct {
	BaseHandler
	db *sql.DB
}

// NewHookHandler creates a new hook handler with dependencies
func NewHookHandler(db *sql.DB) *HookHandler {
	return &HookHandler{
		db: db,
	}
}

// ReportHookTelemetry handles POST /api/v1/telemetry/hook
func (h *HookHandler) ReportHookTelemetry(w http.ResponseWriter, r *http.Request) {
	// Set DB for hook handlers
	SetDB(h.db)
	hookTelemetryHandler(w, r)
}

// GetHookMetrics handles GET /api/v1/hooks/metrics
func (h *HookHandler) GetHookMetrics(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	hookMetricsHandler(w, r)
}

// GetHookMetricsTeam handles GET /api/v1/hooks/metrics/team
func (h *HookHandler) GetHookMetricsTeam(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	// Use same handler but with team filter
	hookMetricsHandler(w, r)
}

// GetHookPolicies handles GET /api/v1/hooks/policies
func (h *HookHandler) GetHookPolicies(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	hookPoliciesHandler(w, r)
}

// UpdateHookPolicies handles POST /api/v1/hooks/policies
func (h *HookHandler) UpdateHookPolicies(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	// For POST, we need to create/update policy
	hookPoliciesHandler(w, r)
}

// GetHookLimits handles GET /api/v1/hooks/limits
func (h *HookHandler) GetHookLimits(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	hookLimitsHandler(w, r)
}

// CreateHookBaseline handles POST /api/v1/hooks/baselines
func (h *HookHandler) CreateHookBaseline(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	hookBaselineHandler(w, r)
}

// ReviewHookBaseline handles POST /api/v1/hooks/baselines/{id}/review
func (h *HookHandler) ReviewHookBaseline(w http.ResponseWriter, r *http.Request) {
	SetDB(h.db)
	reviewHookBaselineHandler(w, r)
}
