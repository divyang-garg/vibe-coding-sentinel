// Hook handler baseline operations
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// hookBaselineHandler handles POST /api/v1/hooks/baselines
func hookBaselineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry struct {
		AgentID  string `json:"agent_id"`
		OrgID    string `json:"org_id"`
		TeamID   string `json:"team_id,omitempty"`
		HookType string `json:"hook_type"`
		File     string `json:"file"`
		Line     int    `json:"line"`
		Pattern  string `json:"pattern"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Source   string `json:"source"`
		Reviewed bool   `json:"reviewed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		log.Printf("Error decoding hook baseline: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateRequired("agent_id", entry.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("org_id", entry.OrgID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("file", entry.File); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("hook_type", entry.HookType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate UUIDs
	if err := validateUUID(entry.OrgID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if entry.TeamID != "" {
		if err := validateUUID(entry.TeamID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate hook_type
	if err := validateHookType(entry.HookType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into hook_baselines table
	baselineID := uuid.New().String()

	// Handle empty team_id
	var teamID interface{}
	if entry.TeamID != "" {
		teamID = entry.TeamID
	} else {
		teamID = nil
	}

	query := `INSERT INTO hook_baselines 
		(id, agent_id, org_id, team_id, hook_type, file, line, pattern, message, severity, source, reviewed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())`

	_, err := db.Exec(query, baselineID, entry.AgentID, entry.OrgID, teamID, entry.HookType,
		entry.File, entry.Line, entry.Pattern, entry.Message, entry.Severity, entry.Source, entry.Reviewed)
	if err != nil {
		log.Printf("Error inserting hook baseline: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      baselineID,
		"status":  "ok",
		"message": "Baseline entry recorded",
	})
}

// reviewHookBaselineHandler handles POST /api/v1/hooks/baselines/{id}/review
func reviewHookBaselineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract baseline ID from URL path
	baselineID := r.URL.Path[len("/api/v1/hooks/baselines/"):]
	if idx := strings.Index(baselineID, "/"); idx != -1 {
		baselineID = baselineID[:idx]
	}

	if err := validateRequired("baseline_id", baselineID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateUUID(baselineID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var review struct {
		Action     string `json:"action"` // "approve" or "reject"
		ReviewedBy string `json:"reviewed_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		log.Printf("Error decoding baseline review: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateRequired("action", review.Action); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateAction(review.Action); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("reviewed_by", review.ReviewedBy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateUUID(review.ReviewedBy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update baseline entry
	var query string
	if review.Action == "approve" {
		query = `UPDATE hook_baselines 
			SET reviewed = true, reviewed_by = $1, reviewed_at = NOW()
			WHERE id = $2`
	} else {
		query = `UPDATE hook_baselines 
			SET reviewed = false, reviewed_by = $1, reviewed_at = NOW()
			WHERE id = $2`
	}

	result, err := db.Exec(query, review.ReviewedBy, baselineID)
	if err != nil {
		log.Printf("Error updating baseline review: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Baseline not found", http.StatusNotFound)
		return
	}

	// Return updated baseline entry
	var baseline struct {
		ID         string     `json:"id"`
		Reviewed   bool       `json:"reviewed"`
		ReviewedBy *string    `json:"reviewed_by"`
		ReviewedAt *time.Time `json:"reviewed_at"`
	}

	err = database.QueryRowWithTimeout(r.Context(), db, "SELECT id, reviewed, reviewed_by, reviewed_at FROM hook_baselines WHERE id = $1", baselineID).
		Scan(&baseline.ID, &baseline.Reviewed, &baseline.ReviewedBy, &baseline.ReviewedAt)
	if err != nil {
		log.Printf("Error querying updated baseline: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(baseline)
}
