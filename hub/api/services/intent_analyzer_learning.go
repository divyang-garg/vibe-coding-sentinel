// Intent Analysis Learning Functions
// Records decisions and learns patterns from user interactions
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"
)

// RecordDecision records a user's decision for learning
// If decision.ID is set, it updates the existing decision; otherwise, it inserts a new one
func RecordDecision(ctx context.Context, projectID string, decision *IntentDecision) error {
	contextDataJSON, err := marshalJSONB(decision.ContextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	if decision.ID != "" {
		// UPDATE existing decision
		updateQuery := `
			UPDATE intent_decisions
			SET user_choice = $3, resolved_prompt = $4, context_data = $5::jsonb
			WHERE id = $1 AND project_id = $2
			RETURNING created_at
		`
		var createdAt time.Time
		err = database.QueryRowWithTimeout(ctx, db, updateQuery,
			decision.ID,
			projectID,
			decision.UserChoice,
			decision.ResolvedPrompt,
			contextDataJSON,
		).Scan(&createdAt)

		if err != nil {
			return fmt.Errorf("failed to update decision: %w", err)
		}

		decision.CreatedAt = createdAt.Format(time.RFC3339)
	} else {
		// INSERT new decision
		query := `
			INSERT INTO intent_decisions (project_id, original_prompt, intent_type, clarifying_question, user_choice, resolved_prompt, context_data)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
			RETURNING id, created_at
		`

		var id string
		var createdAt time.Time
		err = database.QueryRowWithTimeout(ctx, db, query,
			projectID,
			decision.OriginalPrompt,
			string(decision.IntentType),
			decision.ClarifyingQuestion,
			decision.UserChoice,
			decision.ResolvedPrompt,
			contextDataJSON,
		).Scan(&id, &createdAt)

		if err != nil {
			return fmt.Errorf("failed to record decision: %w", err)
		}

		decision.ID = id
		decision.CreatedAt = createdAt.Format(time.RFC3339)
	}

	// Update pattern frequency
	err = updatePatternFrequency(ctx, projectID, decision.IntentType, decision.UserChoice)
	if err != nil {
		LogWarn(ctx, "Failed to update pattern frequency: %v", err)
	}

	return nil
}

// updatePatternFrequency updates the frequency of a pattern
func updatePatternFrequency(ctx context.Context, projectID string, intentType IntentType, userChoice string) error {
	patternDataJSON, err := marshalJSONB(map[string]interface{}{
		"intent_type": string(intentType),
		"user_choice": userChoice,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal pattern data: %w", err)
	}

	// Try to update existing pattern
	updateQuery := `
		UPDATE intent_patterns
		SET frequency = frequency + 1, last_used = NOW()
		WHERE project_id = $1 AND pattern_type = $2 AND pattern_data = $3::jsonb
	`
	result, err := database.ExecWithTimeout(ctx, db, updateQuery, projectID, string(intentType), patternDataJSON)
	if err == nil {
		// Check if row was actually updated
		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr == nil && rowsAffected > 0 {
			return nil // Successfully updated
		}
		// No rows updated, continue to INSERT
	}

	// If no row updated, insert new pattern
	insertQuery := `
		INSERT INTO intent_patterns (project_id, pattern_type, pattern_data, frequency, last_used)
		VALUES ($1, $2, $3::jsonb, 1, NOW())
		ON CONFLICT (project_id, pattern_type, pattern_data) DO NOTHING
	`
	_, err = database.ExecWithTimeout(ctx, db, insertQuery, projectID, string(intentType), patternDataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert pattern: %w", err)
	}

	return nil
}

// RefinePatterns refines intent patterns based on learned decisions
func RefinePatterns(ctx context.Context, projectID string) error {
	// Query recent decisions
	query := `
		SELECT intent_type, user_choice, COUNT(*) as frequency
		FROM intent_decisions
		WHERE project_id = $1
		GROUP BY intent_type, user_choice
		ORDER BY frequency DESC
		LIMIT 20
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to query decisions: %w", err)
	}
	defer rows.Close()

	// Update patterns based on frequency
	for rows.Next() {
		var intentType string
		var userChoice string
		var frequency int

		if err := rows.Scan(&intentType, &userChoice, &frequency); err != nil {
			LogWarn(ctx, "Failed to scan decision row: %v", err)
			continue
		}

		patternDataJSON, err := marshalJSONB(map[string]interface{}{
			"intent_type": intentType,
			"user_choice": userChoice,
		})
		if err != nil {
			continue
		}

		// Update or insert pattern
		updateQuery := `
			INSERT INTO intent_patterns (project_id, pattern_type, pattern_data, frequency, last_used)
			VALUES ($1, $2, $3::jsonb, $4, NOW())
			ON CONFLICT (project_id, pattern_type, pattern_data) 
			DO UPDATE SET frequency = EXCLUDED.frequency, last_used = NOW()
		`
		_, err = database.ExecWithTimeout(ctx, db, updateQuery, projectID, intentType, patternDataJSON, frequency)
		if err != nil {
			LogWarn(ctx, "Failed to update pattern: %v", err)
		}
	}

	return nil
}

// GetLearnedPatterns retrieves learned patterns for a project
func GetLearnedPatterns(ctx context.Context, projectID string) ([]IntentPattern, error) {
	query := `
		SELECT id, project_id, pattern_type, pattern_data, frequency, last_used, created_at
		FROM intent_patterns
		WHERE project_id = $1
		ORDER BY frequency DESC, last_used DESC
		LIMIT 50
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query patterns: %w", err)
	}
	defer rows.Close()

	patterns := []IntentPattern{}
	for rows.Next() {
		var pattern IntentPattern
		var patternDataJSON sql.NullString
		var lastUsed, createdAt sql.NullTime

		err := rows.Scan(
			&pattern.ID,
			&pattern.ProjectID,
			&pattern.PatternType,
			&patternDataJSON,
			&pattern.Frequency,
			&lastUsed,
			&createdAt,
		)
		if err != nil {
			LogWarn(ctx, "Failed to scan pattern row: %v", err)
			continue
		}

		// Unmarshal pattern data
		if patternDataJSON.Valid {
			if err := unmarshalJSONB(patternDataJSON.String, &pattern.PatternData); err != nil {
				LogWarn(ctx, "Failed to unmarshal pattern data: %v", err)
				pattern.PatternData = make(map[string]interface{})
			}
		} else {
			pattern.PatternData = make(map[string]interface{})
		}

		if lastUsed.Valid {
			pattern.LastUsed = lastUsed.Time.Format(time.RFC3339)
		}
		if createdAt.Valid {
			pattern.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}
