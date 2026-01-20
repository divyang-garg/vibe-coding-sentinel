// Package llm provides LLM configuration management
// Complies with CODING_STANDARDS.md: Config modules max 300 lines
package llm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"
)

var db *sql.DB // Will be set during initialization

// SetDB sets the database connection for LLM operations
func SetDB(database *sql.DB) {
	db = database
}

// saveLLMConfig saves or updates LLM configuration for a project
func saveLLMConfig(ctx context.Context, projectID string, config *LLMConfig) (string, error) {
	// Encrypt API key
	encryptedKey, err := encryptAPIKey(config.APIKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Marshal cost optimization config
	costOptJSON, err := json.Marshal(config.CostOptimization)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cost optimization: %w", err)
	}

	// Check if config exists for project
	checkQuery := `SELECT id FROM llm_configurations WHERE project_id = $1 LIMIT 1`
	var existingID string
	checkErr := database.QueryRowWithTimeout(ctx, db, checkQuery, projectID).Scan(&existingID)

	if checkErr == nil && existingID != "" {
		// Update existing config
		updateErr := updateLLMConfig(ctx, existingID, projectID, config)
		if updateErr != nil {
			return "", updateErr
		}
		return existingID, nil
	}

	// Insert new config
	query := `
		INSERT INTO llm_configurations (project_id, provider, api_key_encrypted, model, key_type, cost_optimization)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id
	`

	var configID string
	err = database.QueryRowWithTimeout(ctx, db, query,
		projectID,
		config.Provider,
		encryptedKey,
		config.Model,
		config.KeyType,
		string(costOptJSON),
	).Scan(&configID)

	if err != nil {
		return "", fmt.Errorf("failed to save LLM config: %w", err)
	}

	return configID, nil
}

// updateLLMConfig updates existing LLM configuration
func updateLLMConfig(ctx context.Context, configID string, projectID string, config *LLMConfig) error {
	// Encrypt API key if provided
	var encryptedKey []byte
	var err error
	if config.APIKey != "" {
		encryptedKey, err = encryptAPIKey(config.APIKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt API key: %w", err)
		}
	}

	// Marshal cost optimization config
	costOptJSON, err := json.Marshal(config.CostOptimization)
	if err != nil {
		return fmt.Errorf("failed to marshal cost optimization: %w", err)
	}

	// Update config
	query := `
		UPDATE llm_configurations
		SET provider = $1, model = $2, key_type = $3, cost_optimization = $4::jsonb, updated_at = NOW()
		WHERE id = $5 AND project_id = $6
	`

	// Add API key to update if provided
	if config.APIKey != "" {
		query = `
			UPDATE llm_configurations
			SET provider = $1, api_key_encrypted = $2, model = $3, key_type = $4, cost_optimization = $5::jsonb, updated_at = NOW()
			WHERE id = $6 AND project_id = $7
		`
		_, err = database.ExecWithTimeout(ctx, db, query,
			config.Provider,
			encryptedKey,
			config.Model,
			config.KeyType,
			string(costOptJSON),
			configID,
			projectID,
		)
	} else {
		_, err = database.ExecWithTimeout(ctx, db, query,
			config.Provider,
			config.Model,
			config.KeyType,
			string(costOptJSON),
			configID,
			projectID,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to update LLM config: %w", err)
	}

	return nil
}

// deleteLLMConfig deletes LLM configuration
func deleteLLMConfig(ctx context.Context, configID string, projectID string) error {
	// Implementation extracted from main llm_integration.go
	return fmt.Errorf("not implemented")
}

// ListLLMConfigs lists all configurations for a project
func ListLLMConfigs(ctx context.Context, projectID string) ([]*LLMConfig, error) {
	query := `
		SELECT id, provider, api_key_encrypted, model, key_type, cost_optimization, created_at, updated_at
		FROM llm_configurations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM configs: %w", err)
	}
	defer rows.Close()

	var configs []*LLMConfig
	for rows.Next() {
		var id, provider, model, keyType string
		var apiKeyEncrypted []byte
		var costOptJSON sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &provider, &apiKeyEncrypted, &model, &keyType, &costOptJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}

		// Decrypt API key (but mask it in response)
		apiKey, err := decryptAPIKey(apiKeyEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt API key: %w", err)
		}

		// Mask API key (show last 4 chars only)
		maskedKey := maskAPIKey(apiKey)

		// Parse cost optimization config
		var costOpt *CostOptimizationConfig
		if costOptJSON.Valid && costOptJSON.String != "" {
			var costOptMap map[string]interface{}
			if err := json.Unmarshal([]byte(costOptJSON.String), &costOptMap); err == nil {
				costOpt = &CostOptimizationConfig{}
				if useCache, ok := costOptMap["use_cache"].(bool); ok {
					costOpt.UseCache = useCache
				}
				if ttl, ok := costOptMap["cache_ttl_hours"].(float64); ok {
					costOpt.CacheTTLHours = int(ttl)
				}
				if progressive, ok := costOptMap["progressive_depth"].(bool); ok {
					costOpt.ProgressiveDepth = progressive
				}
				if maxCost, ok := costOptMap["max_cost_per_request"].(float64); ok {
					costOpt.MaxCostPerRequest = maxCost
				}
			}
		}

		config := &LLMConfig{
			ID:               id,
			Provider:         provider,
			APIKey:           maskedKey, // Return masked key
			Model:            model,
			KeyType:          keyType,
			CostOptimization: costOpt,
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating configs: %w", err)
	}

	return configs, nil
}
