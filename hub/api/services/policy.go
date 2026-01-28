// Fixed import structure
package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// =============================================================================
// MODELS
// =============================================================================

type HookPolicy struct {
	AuditConfig struct {
		Security      bool `json:"security"`
		Vibe          bool `json:"vibe"`
		BusinessRules bool `json:"business_rules"`
		FileSize      bool `json:"file_size"`
	} `json:"audit_config"`
	OverridePolicy struct {
		CriticalRequiresApproval      bool `json:"critical_requires_approval"`
		MaxOverridesPerDay            int  `json:"max_overrides_per_day"`
		OverrideRequiresJustification bool `json:"override_requires_justification"`
	} `json:"override_policy"`
	BaselinePolicy struct {
		RequiresReview       bool `json:"requires_review"`
		AutoApproveAfterDays int  `json:"auto_approve_after_days"`
		MaxBaselinesPerWeek  int  `json:"max_baselines_per_week"`
	} `json:"baseline_policy"`
	ExceptionPolicy struct {
		RequiresApproval       bool `json:"requires_approval"`
		TemporaryExceptionDays int  `json:"temporary_exception_days"`
	} `json:"exception_policy"`
}

// =============================================================================
// VALIDATION
// =============================================================================

func validateHookPolicy(policy *HookPolicy) error {
	// Validate override policy
	if policy.OverridePolicy.MaxOverridesPerDay < 0 {
		return fmt.Errorf("max_overrides_per_day must be >= 0")
	}
	if policy.OverridePolicy.MaxOverridesPerDay > 100 {
		return fmt.Errorf("max_overrides_per_day cannot exceed 100")
	}

	// Validate baseline policy
	if policy.BaselinePolicy.AutoApproveAfterDays < 0 {
		return fmt.Errorf("auto_approve_after_days must be >= 0")
	}
	if policy.BaselinePolicy.AutoApproveAfterDays > 365 {
		return fmt.Errorf("auto_approve_after_days cannot exceed 365")
	}
	if policy.BaselinePolicy.MaxBaselinesPerWeek < 0 {
		return fmt.Errorf("max_baselines_per_week must be >= 0")
	}
	if policy.BaselinePolicy.MaxBaselinesPerWeek > 1000 {
		return fmt.Errorf("max_baselines_per_week cannot exceed 1000")
	}

	// Validate exception policy
	if policy.ExceptionPolicy.TemporaryExceptionDays < 0 {
		return fmt.Errorf("temporary_exception_days must be >= 0")
	}
	if policy.ExceptionPolicy.TemporaryExceptionDays > 365 {
		return fmt.Errorf("temporary_exception_days cannot exceed 365")
	}

	return nil
}

// =============================================================================
// HANDLERS
// =============================================================================

// createOrUpdateHookPolicyHandler handles POST /api/v1/hooks/policies
func createOrUpdateHookPolicyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get org_id from query first
	orgID := r.URL.Query().Get("org_id")

	// Read body once into bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// If org_id not in query, try to get from body
	if orgID == "" {
		var bodyStruct struct {
			OrgID string `json:"org_id"`
		}
		if err := json.Unmarshal(bodyBytes, &bodyStruct); err == nil {
			orgID = bodyStruct.OrgID
		}
	}

	if orgID == "" {
		http.Error(w, "org_id required", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(orgID); err != nil {
		http.Error(w, "org_id must be a valid UUID", http.StatusBadRequest)
		return
	}

	// Decode policy from body bytes
	var policy HookPolicy
	if err := json.Unmarshal(bodyBytes, &policy); err != nil {
		log.Printf("Error decoding hook policy: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate policy
	if err := validateHookPolicy(&policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert to JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		log.Printf("Error marshaling hook policy: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Check if policy exists
	var existingID string
	err = db.QueryRow("SELECT id FROM hook_policies WHERE org_id = $1", orgID).Scan(&existingID)

	if errors.Is(err, sql.ErrNoRows) {
		// Create new policy
		policyID := uuid.New().String()
		query := `INSERT INTO hook_policies (id, org_id, policy_config, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())`
		_, err = database.ExecWithTimeout(r.Context(), db, query, policyID, orgID, policyJSON)
		if err != nil {
			log.Printf("Error creating hook policy: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         policyID,
			"status":     "created",
			"message":    "Hook policy created",
			"updated_at": time.Now().Format(time.RFC3339),
		})
		return
	}

	if err != nil {
		log.Printf("Error checking existing hook policy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Update existing policy
	query := `UPDATE hook_policies 
		SET policy_config = $1, updated_at = NOW()
		WHERE org_id = $2`
	_, err = database.ExecWithTimeout(r.Context(), db, query, policyJSON, orgID)
	if err != nil {
		log.Printf("Error updating hook policy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         existingID,
		"status":     "updated",
		"message":    "Hook policy updated",
		"updated_at": time.Now().Format(time.RFC3339),
	})
}
