// Hook handler core operations - Telemetry, Metrics, Limits, Policies
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// hookTelemetryHandler handles POST /api/v1/telemetry/hook
func hookTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event struct {
		AgentID         string                 `json:"agent_id"`
		OrgID           string                 `json:"org_id"`
		TeamID          string                 `json:"team_id"`
		HookType        string                 `json:"hook_type"`
		Result          string                 `json:"result"`
		OverrideReason  string                 `json:"override_reason,omitempty"`
		FindingsSummary map[string]interface{} `json:"findings_summary"`
		UserActions     []string               `json:"user_actions"`
		DurationMs      int64                  `json:"duration_ms"`
		Timestamp       string                 `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("Error decoding hook telemetry: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateRequired("agent_id", event.AgentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("hook_type", event.HookType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("result", event.Result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate hook_type
	if err := validateHookType(event.HookType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate result
	if err := validateResult(event.Result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate UUIDs if provided
	if event.OrgID != "" {
		if err := validateUUID(event.OrgID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if event.TeamID != "" {
		if err := validateUUID(event.TeamID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Insert into hook_executions table
	execID := uuid.New().String()
	findingsJSON, err := json.Marshal(event.FindingsSummary)
	if err != nil {
		log.Printf("Error marshaling findings summary: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	actionsJSON, err := json.Marshal(event.UserActions)
	if err != nil {
		log.Printf("Error marshaling user actions: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Handle empty org_id/team_id (convert to NULL)
	var orgID interface{}
	var teamID interface{}
	if event.OrgID != "" {
		orgID = event.OrgID
	} else {
		orgID = nil
	}
	if event.TeamID != "" {
		teamID = event.TeamID
	} else {
		teamID = nil
	}

	query := `INSERT INTO hook_executions 
		(id, agent_id, org_id, team_id, hook_type, result, override_reason, findings_summary, user_actions, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`

	_, err = db.Exec(query, execID, event.AgentID, orgID, teamID, event.HookType, event.Result,
		event.OverrideReason, findingsJSON, actionsJSON, event.DurationMs)
	if err != nil {
		log.Printf("Error inserting hook execution: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":      execID,
		"status":  "ok",
		"message": "Hook execution recorded",
	})
}

// hookMetricsHandler handles GET /api/v1/hooks/metrics
// NOTE: This handler is large (>200 lines) due to pattern extraction logic
func hookMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	orgID := r.URL.Query().Get("org_id")
	teamID := r.URL.Query().Get("team_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Validate UUIDs if provided
	if orgID != "" {
		if err := validateUUID(orgID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if teamID != "" {
		if err := validateUUID(teamID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Validate date formats if provided
	if startDate != "" {
		if err := validateDate(startDate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if endDate != "" {
		if err := validateDate(endDate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Build query
	query := `SELECT 
		COUNT(*) as total_executions,
		COUNT(*) FILTER (WHERE result = 'blocked') as blocked_count,
		COUNT(*) FILTER (WHERE result = 'allowed') as allowed_count,
		COUNT(*) FILTER (WHERE result = 'overridden') as overridden_count,
		AVG(duration_ms) as avg_duration_ms
		FROM hook_executions
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if orgID != "" {
		query += ` AND org_id = $` + strconv.Itoa(argIndex)
		args = append(args, orgID)
		argIndex++
	}
	if teamID != "" {
		query += ` AND team_id = $` + strconv.Itoa(argIndex)
		args = append(args, teamID)
		argIndex++
	}
	if startDate != "" {
		query += ` AND created_at >= $` + strconv.Itoa(argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != "" {
		query += ` AND created_at <= $` + strconv.Itoa(argIndex)
		args = append(args, endDate)
		argIndex++
	}

	var metrics HookMetrics
	err := database.QueryRowWithTimeout(r.Context(), db, query, args...).Scan(
		&metrics.TotalExecutions,
		&metrics.BlockedCount,
		&metrics.AllowedCount,
		&metrics.OverriddenCount,
		&metrics.AvgDurationMs,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error querying hook metrics: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Calculate override rate
	if metrics.TotalExecutions > 0 {
		metrics.OverrideRate = float64(metrics.OverriddenCount) / float64(metrics.TotalExecutions) * 100
	}

	// Get most blocked patterns by extracting from findings_summary
	var patternCounts = make(map[string]int)
	var patternExtractionError string
	patternQuery := `SELECT findings_summary 
		FROM hook_executions 
		WHERE result = 'blocked'`
	patternArgs := []interface{}{}
	patternArgIndex := 1
	if orgID != "" {
		patternQuery += " AND org_id = $1"
		patternArgs = append(patternArgs, orgID)
		patternArgIndex++
	}
	if startDate != "" {
		patternQuery += " AND created_at >= $" + strconv.Itoa(patternArgIndex)
		patternArgs = append(patternArgs, startDate)
		patternArgIndex++
	}
	if endDate != "" {
		patternQuery += " AND created_at <= $" + strconv.Itoa(patternArgIndex)
		patternArgs = append(patternArgs, endDate)
	}

	if err == nil {
		rows, err := database.QueryWithTimeout(r.Context(), db, patternQuery, patternArgs...)
		if err != nil {
			log.Printf("Error querying blocked patterns: %v", err)
			patternExtractionError = "Database query failed: " + err.Error()
		} else {
			defer rows.Close()
			for rows.Next() {
				var findingsJSON []byte
				if err := rows.Scan(&findingsJSON); err != nil {
					log.Printf("Error scanning pattern row: %v", err)
					continue // Skip this row but continue processing
				}

				var findings map[string]interface{}
				if err := json.Unmarshal(findingsJSON, &findings); err != nil {
					log.Printf("Error unmarshaling findings JSON: %v", err)
					continue // Skip this row
				}

				// Try multiple possible structures
				patternsFound := false

				// Structure 1: findings["findings"] array
				if findingsList, ok := findings["findings"].([]interface{}); ok {
					for _, f := range findingsList {
						if finding, ok := f.(map[string]interface{}); ok {
							if pattern, ok := finding["pattern"].(string); ok && pattern != "" {
								patternCounts[pattern]++
								patternsFound = true
							}
						}
					}
				}

				// Structure 2: findings["patterns"] array (direct)
				if !patternsFound {
					if patternsList, ok := findings["patterns"].([]interface{}); ok {
						for _, p := range patternsList {
							if pattern, ok := p.(string); ok && pattern != "" {
								patternCounts[pattern]++
								patternsFound = true
							}
						}
					}
				}

				// Structure 3: findings as array directly (if findings_summary is array)
				if !patternsFound {
					// Try to detect if findings is actually an array
					if findingsArray, ok := findings[""].([]interface{}); ok {
						for _, f := range findingsArray {
							if finding, ok := f.(map[string]interface{}); ok {
								if pattern, ok := finding["pattern"].(string); ok && pattern != "" {
									patternCounts[pattern]++
									patternsFound = true
								}
							}
						}
					}
				}

				if !patternsFound {
					log.Printf("Warning: Could not extract patterns from findings_summary. Structure keys: %v", getMapKeys(findings))
				}
			}

			// Check for iteration errors
			if err := rows.Err(); err != nil {
				log.Printf("Error iterating pattern rows: %v", err)
				patternExtractionError = "Row iteration failed: " + err.Error()
				// Continue with partial results
			}
		}
	}

	// Get top 10 patterns
	var topPatterns []PatternCount
	for pattern, count := range patternCounts {
		topPatterns = append(topPatterns, PatternCount{Pattern: pattern, Count: count})
	}
	// Sort by count (simple bubble sort for small list)
	for i := 0; i < len(topPatterns)-1; i++ {
		for j := i + 1; j < len(topPatterns); j++ {
			if topPatterns[i].Count < topPatterns[j].Count {
				topPatterns[i], topPatterns[j] = topPatterns[j], topPatterns[i]
			}
		}
	}
	// Keep only top 10
	if len(topPatterns) > 10 {
		topPatterns = topPatterns[:10]
	}
	metrics.MostBlockedPatterns = topPatterns

	// Add error indicator if pattern extraction failed
	if patternExtractionError != "" {
		if metrics.Errors == nil {
			metrics.Errors = make(map[string]string)
		}
		metrics.Errors["pattern_extraction"] = patternExtractionError
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// hookLimitsHandler handles GET /api/v1/hooks/limits
func hookLimitsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	orgID := r.URL.Query().Get("org_id")

	if err := validateRequired("agent_id", agentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateRequired("org_id", orgID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateUUID(orgID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current day's override count
	today := time.Now().Format("2006-01-02")
	var overrideCount int64
	overrideQuery := `SELECT COUNT(*) 
		FROM hook_executions 
		WHERE agent_id = $1 
		AND org_id = $2 
		AND result = 'overridden'
		AND DATE(created_at) = $3`
	err := database.QueryRowWithTimeout(r.Context(), db, overrideQuery, agentID, orgID, today).Scan(&overrideCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error querying override count: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get current week's baseline count
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")
	var baselineCount int64
	baselineQuery := `SELECT COUNT(*) 
		FROM hook_baselines 
		WHERE agent_id = $1 
		AND org_id = $2 
		AND source = 'hook'
		AND DATE(created_at) >= $3`
	err = database.QueryRowWithTimeout(r.Context(), db, baselineQuery, agentID, orgID, weekStart).Scan(&baselineCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error querying baseline count: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"override_count_today": overrideCount,
		"baseline_count_week":  baselineCount,
	})
}

// hookPoliciesHandler handles GET /api/v1/hooks/policies
func hookPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "org_id required", http.StatusBadRequest)
		return
	}

	var policyRecord HookPolicyRecord
	var policyConfigJSON []byte

	query := `SELECT id, org_id, policy_config, created_at, updated_at
		FROM hook_policies
		WHERE org_id = $1
		ORDER BY updated_at DESC
		LIMIT 1`

	err := db.QueryRow(query, orgID).Scan(
		&policyRecord.ID,
		&policyRecord.OrgID,
		&policyConfigJSON,
		&policyRecord.CreatedAt,
		&policyRecord.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		// Return default policy
		defaultPolicy := map[string]interface{}{
			"audit_config": map[string]interface{}{
				"security":       true,
				"vibe":           true,
				"business_rules": true,
				"file_size":      true,
			},
			"override_policy": map[string]interface{}{
				"critical_requires_approval":      false,
				"max_overrides_per_day":           5,
				"override_requires_justification": false,
			},
			"baseline_policy": map[string]interface{}{
				"requires_review":         true,
				"auto_approve_after_days": 7,
				"max_baselines_per_week":  10,
			},
			"exception_policy": map[string]interface{}{
				"requires_approval":        true,
				"temporary_exception_days": 30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(defaultPolicy)
		return
	}

	if err != nil {
		log.Printf("Error querying hook policies: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var policyConfig map[string]interface{}
	if err := json.Unmarshal(policyConfigJSON, &policyConfig); err != nil {
		log.Printf("Error unmarshaling policy config: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policyConfig)
}
