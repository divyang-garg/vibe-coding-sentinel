// Hook handler types and helper functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// MODELS
// =============================================================================

type HookExecution struct {
	ID              string                 `json:"id"`
	AgentID         string                 `json:"agent_id"`
	HookType        string                 `json:"hook_type"` // "pre-commit" | "pre-push"
	Result          string                 `json:"result"`    // "blocked" | "allowed" | "overridden"
	OverrideReason  string                 `json:"override_reason,omitempty"`
	FindingsSummary map[string]interface{} `json:"findings_summary"`
	UserActions     []string               `json:"user_actions"`
	DurationMs      int64                  `json:"duration_ms"`
	CreatedAt       time.Time              `json:"created_at"`
}

type HookBaseline struct {
	ID            string          `json:"id"`
	AgentID       string          `json:"agent_id"`
	BaselineEntry json.RawMessage `json:"baseline_entry"`
	Source        string          `json:"source"` // "hook" | "manual"
	HookType      string          `json:"hook_type,omitempty"`
	Reviewed      bool            `json:"reviewed"`
	ReviewedBy    *string         `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type HookPolicyRecord struct {
	ID           string                 `json:"id"`
	OrgID        string                 `json:"org_id"`
	PolicyConfig map[string]interface{} `json:"policy_config"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type PatternCount struct {
	Pattern string `json:"pattern"`
	Count   int    `json:"count"`
}

type HookMetrics struct {
	TotalExecutions     int64             `json:"total_executions"`
	BlockedCount        int64             `json:"blocked_count"`
	AllowedCount        int64             `json:"allowed_count"`
	OverriddenCount     int64             `json:"overridden_count"`
	OverrideRate        float64           `json:"override_rate"`
	AvgDurationMs       float64           `json:"avg_duration_ms"`
	MostBlockedPatterns []PatternCount    `json:"most_blocked_patterns"`
	Errors              map[string]string `json:"errors,omitempty"` // Error indicators
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// db is the package-level database connection
var db *sql.DB

// SetDB sets the database connection for handlers
func SetDB(database *sql.DB) {
	db = database
}

// validateRequired validates that a field is not empty
func validateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// validateHookType validates hook type is valid
func validateHookType(hookType string) error {
	validTypes := map[string]bool{"pre-commit": true, "pre-push": true}
	if !validTypes[hookType] {
		return fmt.Errorf("invalid hook_type: must be 'pre-commit' or 'pre-push'")
	}
	return nil
}

// validateResult validates hook result is valid
func validateResult(result string) error {
	validResults := map[string]bool{"blocked": true, "allowed": true, "overridden": true}
	if !validResults[result] {
		return fmt.Errorf("invalid result: must be 'blocked', 'allowed', or 'overridden'")
	}
	return nil
}

// validateUUID validates a UUID string
func validateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", id)
	}
	return nil
}

// Note: validateDate and validateAction are defined in helpers.go

// buildPatternQuery builds a query for pattern extraction
func buildPatternQuery(orgID, startDate, endDate string) (string, []interface{}, error) {
	query := `SELECT findings_summary FROM hook_executions WHERE result = 'blocked'`
	args := []interface{}{}
	argIndex := 1

	if orgID != "" {
		query += fmt.Sprintf(" AND org_id = $%d", argIndex)
		args = append(args, orgID)
		argIndex++
	}
	if startDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	// Validate parameter count matches
	expectedParams := strings.Count(query, "$")
	if expectedParams != len(args) {
		return "", nil, fmt.Errorf("parameter count mismatch: query has %d params, args has %d", expectedParams, len(args))
	}

	return query, args, nil
}

// getMapKeys returns keys of a map for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
