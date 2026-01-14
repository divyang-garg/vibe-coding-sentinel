// Phase 14A: Hub LLM Integration
// Manages LLM API keys, model selection, and cost optimization

package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	ID               string                 `json:"id,omitempty"`
	Provider         string                 `json:"provider"`
	APIKey           string                 `json:"api_key"` // Decrypted for use
	Model            string                 `json:"model"`
	KeyType          string                 `json:"key_type"`
	CostOptimization CostOptimizationConfig `json:"cost_optimization,omitempty"`
}

// CostOptimizationConfig contains cost optimization settings
type CostOptimizationConfig struct {
	UseCache          bool    `json:"use_cache"`
	CacheTTLHours     int     `json:"cache_ttl_hours"`
	ProgressiveDepth  bool    `json:"progressive_depth"`
	MaxCostPerRequest float64 `json:"max_cost_per_request,omitempty"`
}

// LLMUsage tracks token usage and costs
type LLMUsage struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"project_id"`
	ValidationID  string  `json:"validation_id,omitempty"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	TokensUsed    int     `json:"tokens_used"`
	EstimatedCost float64 `json:"estimated_cost"`
	CreatedAt     string  `json:"created_at"`
}

// getEncryptionKey retrieves or generates the encryption key for API keys
func getEncryptionKey() ([]byte, error) {
	// In production, this should be stored securely (e.g., in a secrets manager)
	// For now, use an environment variable or generate a key
	keyStr := os.Getenv("SENTINEL_ENCRYPTION_KEY")
	if keyStr == "" {
		// Generate a key (32 bytes for AES-256)
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		// In production, this should be persisted securely
		return key, nil
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}

	// Ensure key is 32 bytes (AES-256)
	if len(key) != 32 {
		// Hash to 32 bytes
		hash := sha256.Sum256(key)
		key = hash[:]
	}

	return key, nil
}

// encryptAPIKey encrypts an API key using AES-256
func encryptAPIKey(apiKey string) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return ciphertext, nil
}

// decryptAPIKey decrypts an API key using AES-256
func decryptAPIKey(encrypted []byte) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// getLLMConfig retrieves LLM configuration for a project
func getLLMConfig(ctx context.Context, projectID string) (*LLMConfig, error) {
	query := `
		SELECT id, provider, api_key_encrypted, model, key_type, cost_optimization
		FROM llm_configurations
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var id, provider, model, keyType string
	var apiKeyEncrypted []byte
	var costOptJSON sql.NullString

	err := queryRowWithTimeout(ctx, query, projectID).Scan(
		&id, &provider, &apiKeyEncrypted, &model, &keyType, &costOptJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no LLM configuration found for project %s", projectID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM config: %w", err)
	}

	// Decrypt API key
	apiKey, err := decryptAPIKey(apiKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Parse cost optimization config
	var costOpt CostOptimizationConfig
	if costOptJSON.Valid && costOptJSON.String != "" {
		if err := json.Unmarshal([]byte(costOptJSON.String), &costOpt); err != nil {
			// Use defaults if parsing fails
			costOpt = CostOptimizationConfig{
				UseCache:         true,
				CacheTTLHours:    24,
				ProgressiveDepth: true,
			}
		}
	} else {
		// Defaults
		costOpt = CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		}
	}

	return &LLMConfig{
		ID:               id,
		Provider:         provider,
		APIKey:           apiKey,
		Model:            model,
		KeyType:          keyType,
		CostOptimization: costOpt,
	}, nil
}

// getLLMConfigByID retrieves LLM configuration by ID
func getLLMConfigByID(ctx context.Context, configID string, projectID string) (*LLMConfig, error) {
	query := `
		SELECT id, provider, api_key_encrypted, model, key_type, cost_optimization
		FROM llm_configurations
		WHERE id = $1 AND project_id = $2
	`

	var id, provider, model, keyType string
	var apiKeyEncrypted []byte
	var costOptJSON sql.NullString

	err := queryRowWithTimeout(ctx, query, configID, projectID).Scan(
		&id, &provider, &apiKeyEncrypted, &model, &keyType, &costOptJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no LLM configuration found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM config: %w", err)
	}

	// Decrypt API key
	apiKey, err := decryptAPIKey(apiKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	// Parse cost optimization config
	var costOpt CostOptimizationConfig
	if costOptJSON.Valid && costOptJSON.String != "" {
		if err := json.Unmarshal([]byte(costOptJSON.String), &costOpt); err != nil {
			costOpt = CostOptimizationConfig{
				UseCache:         true,
				CacheTTLHours:    24,
				ProgressiveDepth: true,
			}
		}
	} else {
		costOpt = CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		}
	}

	return &LLMConfig{
		ID:               id,
		Provider:         provider,
		APIKey:           apiKey,
		Model:            model,
		KeyType:          keyType,
		CostOptimization: costOpt,
	}, nil
}

// validateProvider validates that provider is in supported list
func validateProvider(provider string) error {
	supportedProviders := []string{"openai", "anthropic", "azure"}
	for _, p := range supportedProviders {
		if provider == p {
			return nil
		}
	}
	return fmt.Errorf("unsupported provider: %s. Supported providers: %v", provider, supportedProviders)
}

// validateModel validates that model exists for provider
func validateModel(provider, model string) error {
	models := getSupportedModels(provider)
	for _, m := range models {
		if m.Name == model {
			return nil
		}
	}
	return fmt.Errorf("unsupported model '%s' for provider '%s'", model, provider)
}

// validateAPIKeyFormat validates API key format based on provider
func validateAPIKeyFormat(provider, apiKey string) error {
	if len(apiKey) < 10 {
		return fmt.Errorf("API key is too short")
	}

	switch provider {
	case "openai":
		if !strings.HasPrefix(apiKey, "sk-") {
			return fmt.Errorf("OpenAI API key must start with 'sk-'")
		}
	case "anthropic":
		if len(apiKey) < 20 {
			return fmt.Errorf("anthropic API key appears to be invalid")
		}
	case "azure":
		// Azure API keys can vary, just check minimum length
		if len(apiKey) < 20 {
			return fmt.Errorf("azure API key appears to be invalid")
		}
	}

	return nil
}

// validateCostOptimization validates cost optimization settings
func validateCostOptimization(config CostOptimizationConfig) error {
	if config.CacheTTLHours < 0 {
		return fmt.Errorf("cache TTL hours must be non-negative")
	}
	if config.CacheTTLHours > 8760 { // 1 year
		return fmt.Errorf("cache TTL hours cannot exceed 8760 (1 year)")
	}
	if config.MaxCostPerRequest < 0 {
		return fmt.Errorf("max cost per request must be non-negative")
	}
	return nil
}

// testLLMConnection tests the LLM API connection with detailed error messages
func testLLMConnection(ctx context.Context, config *LLMConfig) error {
	if config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// Validate provider
	if err := validateProvider(config.Provider); err != nil {
		return err
	}

	// Validate model
	if err := validateModel(config.Provider, config.Model); err != nil {
		return err
	}

	// Validate API key format
	if err := validateAPIKeyFormat(config.Provider, config.APIKey); err != nil {
		return err
	}

	// Make a minimal test API call
	testPrompt := "test"
	_, _, err := callLLM(ctx, config, testPrompt, "test")
	if err != nil {
		// Provide more detailed error messages
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
			return fmt.Errorf("authentication failed: invalid API key")
		}
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("model '%s' not found for provider '%s'", config.Model, config.Provider)
		}
		if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("connection timeout: provider may be unavailable")
		}
		return fmt.Errorf("LLM connection test failed: %w", err)
	}

	return nil
}

// ProviderInfo contains information about a supported LLM provider
type ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// ModelInfo contains information about a supported model
type ModelInfo struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	PricePer1K  float64 `json:"price_per_1k"`
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
	checkErr := queryRowWithTimeout(ctx, checkQuery, projectID).Scan(&existingID)

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
	err = queryRowWithTimeout(ctx, query,
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
	} else {
		// If API key not provided, get existing encrypted key
		getKeyQuery := `SELECT api_key_encrypted FROM llm_configurations WHERE id = $1 AND project_id = $2`
		err = queryRowWithTimeout(ctx, getKeyQuery, configID, projectID).Scan(&encryptedKey)
		if err != nil {
			return fmt.Errorf("failed to get existing API key: %w", err)
		}
	}

	// Marshal cost optimization config
	costOptJSON, err := json.Marshal(config.CostOptimization)
	if err != nil {
		return fmt.Errorf("failed to marshal cost optimization: %w", err)
	}

	// Build update query dynamically based on what's provided
	var query string
	var args []interface{}
	argIndex := 1

	if config.APIKey != "" {
		query = `
			UPDATE llm_configurations
			SET provider = $` + strconv.Itoa(argIndex) + `,
				api_key_encrypted = $` + strconv.Itoa(argIndex+1) + `,
				model = $` + strconv.Itoa(argIndex+2) + `,
				key_type = $` + strconv.Itoa(argIndex+3) + `,
				cost_optimization = $` + strconv.Itoa(argIndex+4) + `::jsonb,
				updated_at = NOW()
			WHERE id = $` + strconv.Itoa(argIndex+5) + ` AND project_id = $` + strconv.Itoa(argIndex+6) + `
		`
		args = []interface{}{
			config.Provider,
			encryptedKey,
			config.Model,
			config.KeyType,
			string(costOptJSON),
			configID,
			projectID,
		}
	} else {
		query = `
			UPDATE llm_configurations
			SET provider = $` + strconv.Itoa(argIndex) + `,
				model = $` + strconv.Itoa(argIndex+1) + `,
				key_type = $` + strconv.Itoa(argIndex+2) + `,
				cost_optimization = $` + strconv.Itoa(argIndex+3) + `::jsonb,
				updated_at = NOW()
			WHERE id = $` + strconv.Itoa(argIndex+4) + ` AND project_id = $` + strconv.Itoa(argIndex+5) + `
		`
		args = []interface{}{
			config.Provider,
			config.Model,
			config.KeyType,
			string(costOptJSON),
			configID,
			projectID,
		}
	}

	result, err := execWithTimeout(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update LLM config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("config not found or access denied")
	}

	return nil
}

// deleteLLMConfig deletes LLM configuration
func deleteLLMConfig(ctx context.Context, configID string, projectID string) error {
	query := `
		DELETE FROM llm_configurations
		WHERE id = $1 AND project_id = $2
	`

	result, err := execWithTimeout(ctx, query, configID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete LLM config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("config not found or access denied")
	}

	return nil
}

// listLLMConfigs lists all configurations for a project
func listLLMConfigs(ctx context.Context, projectID string) ([]*LLMConfig, error) {
	query := `
		SELECT id, provider, api_key_encrypted, model, key_type, cost_optimization, created_at, updated_at
		FROM llm_configurations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := queryWithTimeout(ctx, query, projectID)
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
		var costOpt CostOptimizationConfig
		if costOptJSON.Valid && costOptJSON.String != "" {
			if err := json.Unmarshal([]byte(costOptJSON.String), &costOpt); err != nil {
				costOpt = CostOptimizationConfig{
					UseCache:         true,
					CacheTTLHours:    24,
					ProgressiveDepth: true,
				}
			}
		} else {
			costOpt = CostOptimizationConfig{
				UseCache:         true,
				CacheTTLHours:    24,
				ProgressiveDepth: true,
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

// maskAPIKey masks an API key showing only last 4 characters
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 4 {
		return "****"
	}
	return "****" + apiKey[len(apiKey)-4:]
}

// getSupportedProviders returns list of supported providers
func getSupportedProviders() []ProviderInfo {
	return []ProviderInfo{
		{
			Name:        "openai",
			DisplayName: "OpenAI",
			Description: "OpenAI GPT models (GPT-4, GPT-3.5)",
		},
		{
			Name:        "anthropic",
			DisplayName: "Anthropic",
			Description: "Anthropic Claude models (Claude 3 Opus, Sonnet, Haiku)",
		},
		{
			Name:        "azure",
			DisplayName: "Azure OpenAI",
			Description: "Azure OpenAI Service (GPT-4, GPT-3.5)",
		},
	}
}

// getSupportedModels returns list of models for a provider
func getSupportedModels(provider string) []ModelInfo {
	// Pricing per 1K tokens (from calculateEstimatedCost)
	pricing := map[string]map[string]float64{
		"openai": {
			"gpt-4":             0.03,
			"gpt-4-turbo":       0.01,
			"gpt-3.5-turbo":     0.0015,
			"gpt-3.5-turbo-16k": 0.003,
		},
		"anthropic": {
			"claude-3-opus":   0.015,
			"claude-3-sonnet": 0.003,
			"claude-3-haiku":  0.00025,
		},
		"azure": {
			"gpt-4":         0.03,
			"gpt-4-turbo":   0.01,
			"gpt-3.5-turbo": 0.0015,
		},
	}

	providerModels, ok := pricing[provider]
	if !ok {
		return []ModelInfo{}
	}

	var models []ModelInfo
	for modelName, price := range providerModels {
		displayName := modelName
		// Format display names
		switch modelName {
		case "gpt-4":
			displayName = "GPT-4"
		case "gpt-4-turbo":
			displayName = "GPT-4 Turbo"
		case "gpt-3.5-turbo":
			displayName = "GPT-3.5 Turbo"
		case "gpt-3.5-turbo-16k":
			displayName = "GPT-3.5 Turbo 16K"
		case "claude-3-opus":
			displayName = "Claude 3 Opus"
		case "claude-3-sonnet":
			displayName = "Claude 3 Sonnet"
		case "claude-3-haiku":
			displayName = "Claude 3 Haiku"
		}

		models = append(models, ModelInfo{
			Name:        modelName,
			DisplayName: displayName,
			PricePer1K:  price,
		})
	}

	return models
}

// updateLLMConfigByID is an alias for updateLLMConfig for clarity in handlers
func updateLLMConfigByID(ctx context.Context, configID string, projectID string, config *LLMConfig) error {
	return updateLLMConfig(ctx, configID, projectID, config)
}

// logConfigChange logs a configuration change to the audit log
func logConfigChange(ctx context.Context, projectID string, configID string, action string, changedBy string, oldValue interface{}, newValue interface{}, ipAddress string) error {
	oldValueJSON, err := json.Marshal(oldValue)
	if err != nil {
		oldValueJSON = []byte("null")
	}

	newValueJSON, err := json.Marshal(newValue)
	if err != nil {
		newValueJSON = []byte("null")
	}

	query := `
		INSERT INTO config_audit_log (project_id, config_id, action, changed_by, old_value, new_value, ip_address)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7)
	`

	_, err = execWithTimeout(ctx, query, projectID, configID, action, changedBy, string(oldValueJSON), string(newValueJSON), ipAddress)
	if err != nil {
		return fmt.Errorf("failed to log config change: %w", err)
	}

	return nil
}

// getIPAddress extracts IP address from request context or headers
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the chain
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// updateLLMUsageValidationID updates validation ID for LLM usage records
func updateLLMUsageValidationID(ctx context.Context, validationID string, projectID string) error {
	query := `
		UPDATE llm_usage 
		SET validation_id = $1 
		WHERE project_id = $2 AND (validation_id IS NULL OR validation_id = '')
	`
	_, err := execWithTimeout(ctx, query, validationID, projectID)
	if err != nil {
		return fmt.Errorf("failed to update LLM usage validation ID: %w", err)
	}
	return nil
}

// trackUsage stores LLM usage in the database
func trackUsage(ctx context.Context, usage *LLMUsage) error {
	query := `
		INSERT INTO llm_usage (project_id, validation_id, provider, model, tokens_used, estimated_cost)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	var id string
	var createdAt string

	validationID := sql.NullString{}
	if usage.ValidationID != "" {
		validationID.Valid = true
		validationID.String = usage.ValidationID
	}

	err := queryRowWithTimeout(ctx, query,
		usage.ProjectID,
		validationID,
		usage.Provider,
		usage.Model,
		usage.TokensUsed,
		usage.EstimatedCost,
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("failed to track LLM usage: %w", err)
	}

	usage.ID = id
	usage.CreatedAt = createdAt

	return nil
}

// UsageReport contains aggregated usage data for a project
type UsageReport struct {
	ProjectID       string                   `json:"project_id"`
	Period          string                   `json:"period"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	TotalTokens     int64                    `json:"total_tokens"`
	TotalCost       float64                  `json:"total_cost"`
	UsageByProvider map[string]ProviderUsage `json:"usage_by_provider"`
	UsageByModel    map[string]ModelUsage    `json:"usage_by_model"`
	DailyUsage      []DailyUsage             `json:"daily_usage"`
}

// ProviderUsage contains usage statistics for a provider
type ProviderUsage struct {
	Provider     string  `json:"provider"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int64   `json:"request_count"`
}

// ModelUsage contains usage statistics for a model
type ModelUsage struct {
	Model        string  `json:"model"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int64   `json:"request_count"`
}

// DailyUsage contains usage statistics for a day
type DailyUsage struct {
	Date         string  `json:"date"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int64   `json:"request_count"`
}

// UsageStats contains aggregated usage statistics
type UsageStats struct {
	TotalRequests int64           `json:"total_requests"`
	TotalTokens   int64           `json:"total_tokens"`
	TotalCost     float64         `json:"total_cost"`
	AverageCost   float64         `json:"average_cost"`
	TopModels     []ModelStats    `json:"top_models"`
	CostTrend     []CostDataPoint `json:"cost_trend"`
}

// ModelStats contains statistics for a model
type ModelStats struct {
	Model        string  `json:"model"`
	RequestCount int64   `json:"request_count"`
	TotalTokens  int64   `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
}

// CostDataPoint contains cost data for a time period
type CostDataPoint struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

// CostBreakdown contains cost breakdown by provider and model
type CostBreakdown struct {
	ProjectID           string             `json:"project_id"`
	Period              string             `json:"period"`
	TotalCost           float64            `json:"total_cost"`
	ByProvider          map[string]float64 `json:"by_provider"`
	ByModel             map[string]float64 `json:"by_model"`
	ProviderPercentages map[string]float64 `json:"provider_percentages"`
	ModelPercentages    map[string]float64 `json:"model_percentages"`
}

// getUsageReport generates usage report for a project
func getUsageReport(ctx context.Context, projectID string, startDate, endDate time.Time) (*UsageReport, error) {
	query := `
		SELECT provider, model, tokens_used, estimated_cost, DATE(created_at) as usage_date
		FROM llm_usage
		WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at ASC
	`

	rows, err := queryWithTimeout(ctx, query, projectID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage: %w", err)
	}
	defer rows.Close()

	report := &UsageReport{
		ProjectID:       projectID,
		StartDate:       startDate.Format("2006-01-02"),
		EndDate:         endDate.Format("2006-01-02"),
		UsageByProvider: make(map[string]ProviderUsage),
		UsageByModel:    make(map[string]ModelUsage),
		DailyUsage:      []DailyUsage{},
	}

	dailyMap := make(map[string]*DailyUsage)

	for rows.Next() {
		var provider, model string
		var tokensUsed int
		var estimatedCost float64
		var usageDate time.Time

		err := rows.Scan(&provider, &model, &tokensUsed, &estimatedCost, &usageDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage row: %w", err)
		}

		// Update totals
		report.TotalTokens += int64(tokensUsed)
		report.TotalCost += estimatedCost

		// Update provider usage
		if provUsage, exists := report.UsageByProvider[provider]; exists {
			provUsage.Tokens += int64(tokensUsed)
			provUsage.Cost += estimatedCost
			provUsage.RequestCount++
			report.UsageByProvider[provider] = provUsage
		} else {
			report.UsageByProvider[provider] = ProviderUsage{
				Provider:     provider,
				Tokens:       int64(tokensUsed),
				Cost:         estimatedCost,
				RequestCount: 1,
			}
		}

		// Update model usage
		modelKey := provider + ":" + model
		if modUsage, exists := report.UsageByModel[modelKey]; exists {
			modUsage.Tokens += int64(tokensUsed)
			modUsage.Cost += estimatedCost
			modUsage.RequestCount++
			report.UsageByModel[modelKey] = modUsage
		} else {
			report.UsageByModel[modelKey] = ModelUsage{
				Model:        model,
				Tokens:       int64(tokensUsed),
				Cost:         estimatedCost,
				RequestCount: 1,
			}
		}

		// Update daily usage
		dateStr := usageDate.Format("2006-01-02")
		if daily, exists := dailyMap[dateStr]; exists {
			daily.Tokens += int64(tokensUsed)
			daily.Cost += estimatedCost
			daily.RequestCount++
		} else {
			dailyMap[dateStr] = &DailyUsage{
				Date:         dateStr,
				Tokens:       int64(tokensUsed),
				Cost:         estimatedCost,
				RequestCount: 1,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage rows: %w", err)
	}

	// Convert daily map to slice
	for _, daily := range dailyMap {
		report.DailyUsage = append(report.DailyUsage, *daily)
	}

	return report, nil
}

// getUsageStats returns aggregated usage statistics
func getUsageStats(ctx context.Context, projectID string, period string) (*UsageStats, error) {
	// Calculate date range based on period
	var startDate time.Time
	endDate := time.Now()

	switch period {
	case "daily":
		startDate = endDate.AddDate(0, 0, -1)
	case "weekly":
		startDate = endDate.AddDate(0, 0, -7)
	case "monthly":
		startDate = endDate.AddDate(0, -1, 0)
	case "yearly":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30) // Default to 30 days
	}

	query := `
		SELECT provider, model, tokens_used, estimated_cost, DATE(created_at) as usage_date
		FROM llm_usage
		WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at ASC
	`

	rows, err := queryWithTimeout(ctx, query, projectID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage stats: %w", err)
	}
	defer rows.Close()

	stats := &UsageStats{
		TopModels: []ModelStats{},
		CostTrend: []CostDataPoint{},
	}

	modelMap := make(map[string]*ModelStats)
	dailyCostMap := make(map[string]float64)

	for rows.Next() {
		var provider, model string
		var tokensUsed int
		var estimatedCost float64
		var usageDate time.Time

		err := rows.Scan(&provider, &model, &tokensUsed, &estimatedCost, &usageDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage row: %w", err)
		}

		// Update totals
		stats.TotalRequests++
		stats.TotalTokens += int64(tokensUsed)
		stats.TotalCost += estimatedCost

		// Update model stats
		modelKey := provider + ":" + model
		if modelStat, exists := modelMap[modelKey]; exists {
			modelStat.RequestCount++
			modelStat.TotalTokens += int64(tokensUsed)
			modelStat.TotalCost += estimatedCost
		} else {
			modelMap[modelKey] = &ModelStats{
				Model:        model,
				RequestCount: 1,
				TotalTokens:  int64(tokensUsed),
				TotalCost:    estimatedCost,
			}
		}

		// Update daily cost
		dateStr := usageDate.Format("2006-01-02")
		dailyCostMap[dateStr] += estimatedCost
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating usage rows: %w", err)
	}

	// Calculate average cost
	if stats.TotalRequests > 0 {
		stats.AverageCost = stats.TotalCost / float64(stats.TotalRequests)
	}

	// Convert model map to slice and sort by total cost
	for _, modelStat := range modelMap {
		stats.TopModels = append(stats.TopModels, *modelStat)
	}

	// Sort top models by cost (descending)
	for i := 0; i < len(stats.TopModels)-1; i++ {
		for j := i + 1; j < len(stats.TopModels); j++ {
			if stats.TopModels[i].TotalCost < stats.TopModels[j].TotalCost {
				stats.TopModels[i], stats.TopModels[j] = stats.TopModels[j], stats.TopModels[i]
			}
		}
	}

	// Limit to top 10
	if len(stats.TopModels) > 10 {
		stats.TopModels = stats.TopModels[:10]
	}

	// Build cost trend
	for dateStr, cost := range dailyCostMap {
		stats.CostTrend = append(stats.CostTrend, CostDataPoint{
			Date: dateStr,
			Cost: cost,
		})
	}

	return stats, nil
}

// getCostBreakdown returns cost breakdown by provider and model
func getCostBreakdown(ctx context.Context, projectID string, period string) (*CostBreakdown, error) {
	// Calculate date range based on period
	var startDate time.Time
	endDate := time.Now()

	switch period {
	case "daily":
		startDate = endDate.AddDate(0, 0, -1)
	case "weekly":
		startDate = endDate.AddDate(0, 0, -7)
	case "monthly":
		startDate = endDate.AddDate(0, -1, 0)
	case "yearly":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30) // Default to 30 days
	}

	query := `
		SELECT provider, model, estimated_cost
		FROM llm_usage
		WHERE project_id = $1 AND created_at >= $2 AND created_at <= $3
	`

	rows, err := queryWithTimeout(ctx, query, projectID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query cost breakdown: %w", err)
	}
	defer rows.Close()

	breakdown := &CostBreakdown{
		ProjectID:           projectID,
		Period:              period,
		ByProvider:          make(map[string]float64),
		ByModel:             make(map[string]float64),
		ProviderPercentages: make(map[string]float64),
		ModelPercentages:    make(map[string]float64),
	}

	for rows.Next() {
		var provider, model string
		var estimatedCost float64

		err := rows.Scan(&provider, &model, &estimatedCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cost row: %w", err)
		}

		// Update totals
		breakdown.TotalCost += estimatedCost
		breakdown.ByProvider[provider] += estimatedCost
		breakdown.ByModel[model] += estimatedCost
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cost rows: %w", err)
	}

	// Calculate percentages
	if breakdown.TotalCost > 0 {
		for provider, cost := range breakdown.ByProvider {
			breakdown.ProviderPercentages[provider] = (cost / breakdown.TotalCost) * 100
		}
		for model, cost := range breakdown.ByModel {
			breakdown.ModelPercentages[model] = (cost / breakdown.TotalCost) * 100
		}
	}

	return breakdown, nil
}

// calculateEstimatedCost calculates estimated cost based on provider and model
func calculateEstimatedCost(provider string, model string, tokensUsed int) float64 {
	// Pricing per 1K tokens (approximate, as of 2024)
	// These should be configurable or fetched from provider APIs

	pricing := map[string]map[string]float64{
		"openai": {
			"gpt-4":             0.03, // $0.03 per 1K input tokens
			"gpt-4-turbo":       0.01,
			"gpt-3.5-turbo":     0.0015,
			"gpt-3.5-turbo-16k": 0.003,
		},
		"anthropic": {
			"claude-3-opus":   0.015,
			"claude-3-sonnet": 0.003,
			"claude-3-haiku":  0.00025,
		},
		"azure": {
			"gpt-4":         0.03,
			"gpt-4-turbo":   0.01,
			"gpt-3.5-turbo": 0.0015,
		},
	}

	providerPricing, ok := pricing[provider]
	if !ok {
		return 0.0 // Unknown provider
	}

	pricePer1K, ok := providerPricing[model]
	if !ok {
		// Use default for provider
		if provider == "openai" {
			pricePer1K = 0.002
		} else if provider == "anthropic" {
			pricePer1K = 0.001
		} else {
			pricePer1K = 0.001
		}
	}

	// Calculate cost
	cost := (float64(tokensUsed) / 1000.0) * pricePer1K
	return cost
}

// Phase 14D: Model cost database (cost per 1K tokens)
var modelCosts = map[string]map[string]float64{
	"openai": {
		"gpt-4":         0.03,   // $0.03 per 1K tokens
		"gpt-4-turbo":   0.01,   // $0.01 per 1K tokens
		"gpt-3.5-turbo": 0.0015, // $0.0015 per 1K tokens
	},
	"anthropic": {
		"claude-3-opus":   0.015,   // $0.015 per 1K tokens
		"claude-3-sonnet": 0.003,   // $0.003 per 1K tokens
		"claude-3-haiku":  0.00025, // $0.00025 per 1K tokens
	},
	"azure": {
		"gpt-4":         0.03,
		"gpt-4-turbo":   0.01,
		"gpt-3.5-turbo": 0.0015,
	},
}

// getModelCost returns the cost per 1K tokens for a given provider and model
func getModelCost(provider, model string) float64 {
	if costs, ok := modelCosts[provider]; ok {
		if cost, ok := costs[model]; ok {
			return cost
		}
	}
	// Default fallback cost
	return 0.01
}

// estimateCost estimates the cost for a given number of tokens
func estimateCost(provider, model string, tokens int) float64 {
	costPer1K := getModelCost(provider, model)
	return (float64(tokens) / 1000.0) * costPer1K
}

// selectCheaperModel selects a cheaper model for the given provider
func selectCheaperModel(provider string) (string, error) {
	switch provider {
	case "openai":
		return "gpt-3.5-turbo", nil
	case "anthropic":
		return "claude-3-haiku", nil
	case "azure":
		return "gpt-3.5-turbo", nil
	default:
		return "gpt-3.5-turbo", nil
	}
}

// selectExpensiveModel selects an expensive/high-accuracy model for the given provider
func selectExpensiveModel(provider string) (string, error) {
	switch provider {
	case "openai":
		return "gpt-4", nil
	case "anthropic":
		return "claude-3-opus", nil
	case "azure":
		return "gpt-4", nil
	default:
		return "gpt-4", nil
	}
}

// selectModel selects the appropriate LLM model based on task criticality
// Phase 14D: Enhanced to consider depth parameter and cost limits
func selectModel(ctx context.Context, taskType string, config *LLMConfig) (string, error) {
	return selectModelWithDepth(ctx, taskType, config, "medium", 0)
}

// selectModelWithDepth selects the appropriate LLM model considering depth and cost limits
// Phase 14D: New function that considers depth parameter and cost limits
// Returns selected model and whether a cheaper model was selected (for savings tracking)
// projectID is optional - if provided, tracks savings for that project
func selectModelWithDepth(ctx context.Context, taskType string, config *LLMConfig, depth string, estimatedTokens int, projectID ...string) (string, error) {
	// Classify task as critical or non-critical
	isCritical := isCriticalTask(taskType)

	// Phase 14D: Consider depth parameter
	// For medium depth and non-critical tasks, prefer cheaper models
	// For deep depth and critical tasks, prefer expensive models
	preferCheaper := depth == "medium" && !isCritical
	preferExpensive := depth == "deep" && isCritical

	var selectedModel string
	var cheaperSelected bool

	// Route to appropriate model based on criticality and depth
	// Always respect user-configured model if set (but check cost limit)
	if config.Model != "" {
		// Check cost limit if set
		if config.CostOptimization.MaxCostPerRequest > 0 && estimatedTokens > 0 {
			estimatedCost := estimateCost(config.Provider, config.Model, estimatedTokens)
			if estimatedCost > config.CostOptimization.MaxCostPerRequest {
				// Fallback to cheaper model
				cheaperModel, _ := selectCheaperModel(config.Provider)
				cheaperCost := estimateCost(config.Provider, cheaperModel, estimatedTokens)
				savings := estimatedCost - cheaperCost
				if savings > 0 && len(projectID) > 0 && projectID[0] != "" {
					trackModelSelectionSavings(projectID[0], savings, true)
				}
				LogWarn(ctx, "Estimated cost $%.4f exceeds limit $%.4f, using cheaper model", estimatedCost, config.CostOptimization.MaxCostPerRequest)
				return cheaperModel, nil
			}
		}
		return config.Model, nil
	}

	// If no model configured, select based on task criticality and depth
	if preferExpensive {
		// Deep depth + critical: use high-accuracy model
		selectedModel, _ = selectExpensiveModel(config.Provider)
	} else if preferCheaper {
		// Medium depth + non-critical: use cheaper model
		selectedModel, _ = selectCheaperModel(config.Provider)
		cheaperSelected = true
	} else if isCritical {
		// Critical tasks: use high-accuracy model
		selectedModel, _ = selectExpensiveModel(config.Provider)
	} else {
		// Non-critical tasks: use cheaper/faster model
		selectedModel, _ = selectCheaperModel(config.Provider)
		cheaperSelected = true
	}

	// Phase 14D: Check cost limit before returning
	if config.CostOptimization.MaxCostPerRequest > 0 && estimatedTokens > 0 {
		estimatedCost := estimateCost(config.Provider, selectedModel, estimatedTokens)
		if estimatedCost > config.CostOptimization.MaxCostPerRequest {
			// Fallback to cheaper model
			cheaperModel, _ := selectCheaperModel(config.Provider)
			cheaperCost := estimateCost(config.Provider, cheaperModel, estimatedTokens)
			savings := estimatedCost - cheaperCost
			if savings > 0 && len(projectID) > 0 && projectID[0] != "" {
				trackModelSelectionSavings(projectID[0], savings, true)
			}
			LogWarn(ctx, "Estimated cost $%.4f exceeds limit $%.4f, using cheaper model", estimatedCost, config.CostOptimization.MaxCostPerRequest)
			return cheaperModel, nil
		}
	}

	// Phase 14D: Track model selection savings if cheaper model was selected
	if cheaperSelected && estimatedTokens > 0 && len(projectID) > 0 && projectID[0] != "" {
		// Calculate savings vs expensive model
		expensiveModel, _ := selectExpensiveModel(config.Provider)
		expensiveCost := estimateCost(config.Provider, expensiveModel, estimatedTokens)
		cheaperCost := estimateCost(config.Provider, selectedModel, estimatedTokens)
		savings := expensiveCost - cheaperCost
		if savings > 0 {
			trackModelSelectionSavings(projectID[0], savings, true)
		}
	}

	return selectedModel, nil
}

// isCriticalTask determines if a task is critical
func isCriticalTask(taskType string) bool {
	criticalTasks := []string{
		"business_rule_validation",
		"security_analysis",
		"semantic_analysis",
		"requirement_compliance",
	}

	for _, critical := range criticalTasks {
		if taskType == critical {
			return true
		}
	}

	return false
}

// isHighAccuracyModel checks if a model is considered high-accuracy
func isHighAccuracyModel(model string) bool {
	highAccuracyModels := []string{
		"gpt-4",
		"gpt-4-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
	}

	modelLower := strings.ToLower(model)
	for _, highAcc := range highAccuracyModels {
		if strings.Contains(modelLower, strings.ToLower(highAcc)) {
			return true
		}
	}

	return false
}

// LLMResponse contains the response and token usage from an LLM call
type LLMResponse struct {
	Content          string
	TokensUsed       int
	PromptTokens     int
	CompletionTokens int
}

// callLLM makes an API call to the selected LLM model
// Phase 14D: Enhanced to consider depth and cost limits
func callLLM(ctx context.Context, config *LLMConfig, prompt string, taskType string) (string, int, error) {
	return callLLMWithDepth(ctx, config, prompt, taskType, "medium")
}

// callLLMWithDepth makes an API call with depth consideration
// Phase 14D: New function that considers depth parameter
// projectID is optional - if provided, tracks savings for that project
func callLLMWithDepth(ctx context.Context, config *LLMConfig, prompt string, taskType string, depth string, projectID ...string) (string, int, error) {
	// Phase 14D: Estimate tokens before selecting model (rough: 1 token ≈ 4 characters)
	estimatedTokens := len(prompt) / 4

	// Phase 14D: Enforce cost limit before LLM call
	if config.CostOptimization.MaxCostPerRequest > 0 {
		// Try with user model first if set
		testModel := config.Model
		if testModel == "" {
			// Estimate with default selection
			isCritical := isCriticalTask(taskType)
			if isCritical {
				testModel, _ = selectExpensiveModel(config.Provider)
			} else {
				testModel, _ = selectCheaperModel(config.Provider)
			}
		}

		estimatedCost := estimateCost(config.Provider, testModel, estimatedTokens)
		if estimatedCost > config.CostOptimization.MaxCostPerRequest {
			// Try cheaper model
			cheaperModel, _ := selectCheaperModel(config.Provider)
			cheaperCost := estimateCost(config.Provider, cheaperModel, estimatedTokens)
			if cheaperCost > config.CostOptimization.MaxCostPerRequest {
				return "", 0, fmt.Errorf("estimated cost $%.4f exceeds limit $%.4f (even with cheaper model: $%.4f)", estimatedCost, config.CostOptimization.MaxCostPerRequest, cheaperCost)
			}
			// Use cheaper model
			config.Model = cheaperModel
		}
	}

	// Select model based on task criticality and depth
	selectedModel, err := selectModelWithDepth(ctx, taskType, config, depth, estimatedTokens, projectID...)
	if err != nil {
		return "", 0, fmt.Errorf("failed to select model: %w", err)
	}

	// Use selected model
	config.Model = selectedModel

	// Make API call based on provider
	var response LLMResponse
	switch config.Provider {
	case "openai":
		response, err = callOpenAI(ctx, config, prompt)
	case "anthropic":
		response, err = callAnthropic(ctx, config, prompt)
	case "azure":
		response, err = callAzure(ctx, config, prompt)
	default:
		return "", 0, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}

	if err != nil {
		return "", 0, fmt.Errorf("LLM API call failed: %w", err)
	}

	// Phase 14D: Track actual cost vs estimated cost for savings calculation
	if len(projectID) > 0 && projectID[0] != "" && response.TokensUsed > 0 {
		actualCost := calculateEstimatedCost(config.Provider, config.Model, response.TokensUsed)
		if estimatedTokens > 0 {
			estimatedCost := estimateCost(config.Provider, config.Model, estimatedTokens)
			// If actual cost is less than estimated, track the savings
			if actualCost < estimatedCost {
				savings := estimatedCost - actualCost
				trackModelSelectionSavings(projectID[0], savings, false)
			}
		}
	}

	return response.Content, response.TokensUsed, nil
}

// callOpenAI makes an API call to OpenAI
func callOpenAI(ctx context.Context, config *LLMConfig, prompt string) (LLMResponse, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a code analysis assistant."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  4096,
		"temperature": 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < maxRetries-1 {
			delay := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(delay)
		}
	}

	if err != nil {
		return LLMResponse{}, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return LLMResponse{}, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return LLMResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return LLMResponse{}, fmt.Errorf("no choices in response")
	}

	return LLMResponse{
		Content:          apiResponse.Choices[0].Message.Content,
		TokensUsed:       apiResponse.Usage.TotalTokens,
		PromptTokens:     apiResponse.Usage.PromptTokens,
		CompletionTokens: apiResponse.Usage.CompletionTokens,
	}, nil
}

// callAnthropic makes an API call to Anthropic Claude
func callAnthropic(ctx context.Context, config *LLMConfig, prompt string) (LLMResponse, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := map[string]interface{}{
		"model":      config.Model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < maxRetries-1 {
			delay := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(delay)
		}
	}

	if err != nil {
		return LLMResponse{}, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return LLMResponse{}, fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return LLMResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Content) == 0 {
		return LLMResponse{}, fmt.Errorf("no content in response")
	}

	return LLMResponse{
		Content:          apiResponse.Content[0].Text,
		TokensUsed:       apiResponse.Usage.InputTokens + apiResponse.Usage.OutputTokens,
		PromptTokens:     apiResponse.Usage.InputTokens,
		CompletionTokens: apiResponse.Usage.OutputTokens,
	}, nil
}

// callAzure makes an API call to Azure OpenAI
func callAzure(ctx context.Context, config *LLMConfig, prompt string) (LLMResponse, error) {
	// Azure endpoint format: {endpoint}/openai/deployments/{deployment}/chat/completions?api-version={version}
	// For now, assume endpoint contains full path or use default format
	endpoint := config.Provider // This should be the full endpoint URL
	if !strings.Contains(endpoint, "/openai/deployments/") {
		// Construct URL if not full path
		endpoint = fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-01", endpoint, config.Model)
	}

	reqBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": "You are a code analysis assistant."},
			{"role": "user", "content": prompt},
		},
		"max_tokens":  4096,
		"temperature": 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request with retry logic
	client := &http.Client{Timeout: 60 * time.Second}
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < maxRetries-1 {
			delay := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(delay)
		}
	}

	if err != nil {
		return LLMResponse{}, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return LLMResponse{}, fmt.Errorf("azure API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return LLMResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return LLMResponse{}, fmt.Errorf("no choices in response")
	}

	return LLMResponse{
		Content:          apiResponse.Choices[0].Message.Content,
		TokensUsed:       apiResponse.Usage.TotalTokens,
		PromptTokens:     apiResponse.Usage.PromptTokens,
		CompletionTokens: apiResponse.Usage.CompletionTokens,
	}, nil
}
