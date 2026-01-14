// Phase 15: Intent & Simple Language Analyzer
// Handles unclear prompts gracefully through intent analysis, simple language templates,
// context gathering, decision recording, and pattern refinement

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SimpleLanguageTemplate defines a template for clarifying questions
type SimpleLanguageTemplate struct {
	Type        IntentType
	Template    string
	MultiChoice bool
}

// GetTemplates returns all available simple language templates
func GetTemplates() map[IntentType]SimpleLanguageTemplate {
	return map[IntentType]SimpleLanguageTemplate{
		IntentLocationUnclear: {
			Type:        IntentLocationUnclear,
			Template:    "Where should this go?\n1. %s\n2. %s",
			MultiChoice: true,
		},
		IntentEntityUnclear: {
			Type:        IntentEntityUnclear,
			Template:    "Which %s?\n1. %s\n2. %s",
			MultiChoice: true,
		},
		IntentActionConfirm: {
			Type:        IntentActionConfirm,
			Template:    "I will %s. Correct? [Y/n]",
			MultiChoice: false,
		},
	}
}

// FormatTemplate formats a template with provided values
func FormatTemplate(template SimpleLanguageTemplate, values ...string) string {
	// Convert []string to []interface{} for fmt.Sprintf
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return fmt.Sprintf(template.Template, args...)
}

// ContextData contains gathered context for intent analysis
type ContextData struct {
	RecentFiles      []string               `json:"recent_files"`
	RecentErrors     []string               `json:"recent_errors"`
	GitStatus        map[string]string      `json:"git_status"`
	ProjectStructure map[string]interface{} `json:"project_structure"`
	BusinessRules    []string               `json:"business_rules"`
	CodePatterns     []string               `json:"code_patterns"`
}

// GatherContext collects relevant context for intent analysis
func GatherContext(ctx context.Context, projectID string, codebasePath string) (*ContextData, error) {
	contextData := &ContextData{
		RecentFiles:      []string{},
		RecentErrors:     []string{},
		GitStatus:        make(map[string]string),
		ProjectStructure: make(map[string]interface{}),
		BusinessRules:    []string{},
		CodePatterns:     []string{},
	}

	// Gather recent files (from git or file system)
	if codebasePath != "" {
		recentFiles, err := getRecentFiles(ctx, codebasePath)
		if err == nil {
			contextData.RecentFiles = recentFiles
		} else {
			LogWarn(ctx, "Failed to gather recent files: %v", err)
		}
	}

	// Gather git status
	if codebasePath != "" {
		gitStatus, err := getGitStatus(ctx, codebasePath)
		if err == nil {
			contextData.GitStatus = gitStatus
		} else {
			LogWarn(ctx, "Failed to gather git status: %v", err)
		}
	}

	// Gather project structure
	if codebasePath != "" {
		projectStructure, err := getProjectStructure(ctx, codebasePath)
		if err == nil {
			contextData.ProjectStructure = projectStructure
		} else {
			LogWarn(ctx, "Failed to gather project structure: %v", err)
		}
	}

	// Gather business rules from knowledge_items
	if projectID != "" {
		rules, err := extractBusinessRules(ctx, projectID, nil, "", nil)
		if err == nil {
			businessRules := make([]string, 0, len(rules))
			for _, rule := range rules {
				businessRules = append(businessRules, rule.Title)
			}
			contextData.BusinessRules = businessRules
		} else {
			LogWarn(ctx, "Failed to gather business rules: %v", err)
		}
	}

	// Gather code patterns (from recent files)
	if len(contextData.RecentFiles) > 0 {
		patterns := extractCodePatterns(contextData.RecentFiles[:minInt(5, len(contextData.RecentFiles))])
		contextData.CodePatterns = patterns
	}

	return contextData, nil
}

// getRecentFiles gets recently modified files from git or file system
func getRecentFiles(ctx context.Context, codebasePath string) ([]string, error) {
	// Try git first
	cmd := exec.CommandContext(ctx, "git", "-C", codebasePath, "log", "--name-only", "--pretty=format:", "-10")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		files := make([]string, 0)
		seen := make(map[string]bool)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !seen[line] {
				files = append(files, line)
				seen[line] = true
			}
		}
		if len(files) > 0 {
			return files[:minInt(10, len(files))], nil
		}
	}

	// Fallback: get files from directory
	files := []string{}
	err = filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && isCodeFileForIntent(path) {
			files = append(files, path)
		}
		if len(files) >= 10 {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files[:minInt(10, len(files))], nil
}

// getGitStatus gets git status information
func getGitStatus(ctx context.Context, codebasePath string) (map[string]string, error) {
	status := make(map[string]string)

	// Get current branch
	cmd := exec.CommandContext(ctx, "git", "-C", codebasePath, "branch", "--show-current")
	output, err := cmd.Output()
	if err == nil {
		status["branch"] = strings.TrimSpace(string(output))
	}

	// Get modified files count
	cmd = exec.CommandContext(ctx, "git", "-C", codebasePath, "status", "--porcelain")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		modifiedCount := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				modifiedCount++
			}
		}
		status["modified_files"] = fmt.Sprintf("%d", modifiedCount)
	}

	return status, nil
}

// getProjectStructure gets project structure information
func getProjectStructure(ctx context.Context, codebasePath string) (map[string]interface{}, error) {
	structure := make(map[string]interface{})
	dirs := []string{"src", "lib", "app", "components", "packages", "server", "client", "api", "routes"}

	foundDirs := []string{}
	for _, dir := range dirs {
		path := filepath.Join(codebasePath, dir)
		if _, err := os.Stat(path); err == nil {
			foundDirs = append(foundDirs, dir)
		}
	}
	structure["directories"] = foundDirs

	// Get file extensions
	extensions := make(map[string]int)
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != "" {
			extensions[ext]++
		}
		return nil
	})
	if err == nil {
		structure["file_extensions"] = extensions
	}

	return structure, nil
}

// extractCodePatterns extracts code patterns from file paths
func extractCodePatterns(files []string) []string {
	patterns := []string{}
	seen := make(map[string]bool)

	for _, file := range files {
		dir := filepath.Dir(file)
		if dir != "." && !seen[dir] {
			patterns = append(patterns, dir)
			seen[dir] = true
		}
	}

	return patterns
}

// isCodeFileForIntent checks if a file is a code file (local version to avoid conflict)
func isCodeFileForIntent(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := []string{".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".kt", ".swift", ".rs", ".php", ".rb"}
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

// minInt returns the minimum of two integers (local version to avoid conflict)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AnalyzeIntent analyzes a user prompt to determine intent and clarity
func AnalyzeIntent(ctx context.Context, prompt string, contextData *ContextData, projectID string) (*IntentAnalysisResponse, error) {
	// Quick rule-based check for obviously clear prompts
	if isClearPrompt(prompt) {
		return &IntentAnalysisResponse{
			Success:               true,
			RequiresClarification: false,
			IntentType:            IntentClear,
			Confidence:            1.0,
			SuggestedAction:       prompt,
			ResolvedPrompt:        prompt,
		}, nil
	}

	// Use LLM for intent analysis
	llmConfig, err := getLLMConfig(ctx, projectID)
	if err != nil {
		LogWarn(ctx, "Failed to get LLM config, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Build LLM prompt
	llmPrompt := buildIntentAnalysisPrompt(prompt, contextData)

	// Call LLM
	response, tokensUsed, err := callLLM(ctx, llmConfig, llmPrompt, "intent_analysis")
	if err != nil {
		LogWarn(ctx, "LLM call failed, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Parse LLM response
	result, err := parseLLMIntentResponse(response)
	if err != nil {
		LogWarn(ctx, "Failed to parse LLM response, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Track LLM usage
	if projectID != "" {
		usage := &LLMUsage{
			ProjectID:     projectID,
			Provider:      llmConfig.Provider,
			Model:         llmConfig.Model,
			TokensUsed:    tokensUsed,
			EstimatedCost: calculateEstimatedCost(llmConfig.Provider, llmConfig.Model, tokensUsed),
		}
		if err := trackUsage(ctx, usage); err != nil {
			LogWarn(ctx, "Failed to track LLM usage: %v", err)
		}
	}

	return result, nil
}

// isClearPrompt checks if a prompt is clearly specified
func isClearPrompt(prompt string) bool {
	// Check for specific file paths
	if strings.Contains(prompt, "/") || strings.Contains(prompt, "\\") {
		return true
	}

	// Check for specific function/class names
	if strings.Contains(prompt, "function ") || strings.Contains(prompt, "class ") {
		return true
	}

	// Check for specific commands
	clearKeywords := []string{"create", "add", "implement", "fix", "update", "delete", "remove"}
	promptLower := strings.ToLower(prompt)
	for _, keyword := range clearKeywords {
		if strings.Contains(promptLower, keyword) && len(prompt) > 20 {
			return true
		}
	}

	return false
}

// buildIntentAnalysisPrompt builds the LLM prompt for intent analysis
func buildIntentAnalysisPrompt(prompt string, contextData *ContextData) string {
	builder := strings.Builder{}
	builder.WriteString("Analyze the following user prompt and determine if it needs clarification.\n\n")
	builder.WriteString("User Prompt: " + prompt + "\n\n")

	if contextData != nil {
		if len(contextData.RecentFiles) > 0 {
			builder.WriteString("Recent files: " + strings.Join(contextData.RecentFiles[:minInt(5, len(contextData.RecentFiles))], ", ") + "\n")
		}
		if len(contextData.BusinessRules) > 0 {
			builder.WriteString("Business rules: " + strings.Join(contextData.BusinessRules[:minInt(5, len(contextData.BusinessRules))], ", ") + "\n")
		}
		if len(contextData.CodePatterns) > 0 {
			builder.WriteString("Code patterns: " + strings.Join(contextData.CodePatterns, ", ") + "\n")
		}
	}

	builder.WriteString("\nDetermine:\n")
	builder.WriteString("1. Is the prompt clear or vague?\n")
	builder.WriteString("2. What type of clarification is needed? (location_unclear, entity_unclear, action_confirm, ambiguous, or clear)\n")
	builder.WriteString("3. If clarification is needed, what clarifying question should be asked?\n")
	builder.WriteString("4. What are the likely options for the user to choose from?\n")
	builder.WriteString("\nRespond in JSON format:\n")
	builder.WriteString(`{"requires_clarification": true/false, "intent_type": "location_unclear|entity_unclear|action_confirm|ambiguous|clear", "confidence": 0.0-1.0, "clarifying_question": "...", "options": ["option1", "option2"], "suggested_action": "..."}`)

	return builder.String()
}

// parseLLMIntentResponse parses the LLM response into IntentAnalysisResponse
func parseLLMIntentResponse(response string) (*IntentAnalysisResponse, error) {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]
	var result struct {
		RequiresClarification bool     `json:"requires_clarification"`
		IntentType            string   `json:"intent_type"`
		Confidence            float64  `json:"confidence"`
		ClarifyingQuestion    string   `json:"clarifying_question"`
		Options               []string `json:"options"`
		SuggestedAction       string   `json:"suggested_action"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	intentType := IntentType(result.IntentType)
	if intentType == "" {
		intentType = IntentAmbiguous
	}

	responseObj := &IntentAnalysisResponse{
		Success:               true,
		RequiresClarification: result.RequiresClarification,
		IntentType:            intentType,
		Confidence:            result.Confidence,
		ClarifyingQuestion:    result.ClarifyingQuestion,
		Options:               result.Options,
		SuggestedAction:       result.SuggestedAction,
	}

	// Generate clarifying question from template if needed
	if responseObj.RequiresClarification && responseObj.ClarifyingQuestion == "" {
		templates := GetTemplates()
		template, ok := templates[responseObj.IntentType]
		if ok && len(responseObj.Options) >= 2 {
			responseObj.ClarifyingQuestion = FormatTemplate(template, responseObj.Options[0], responseObj.Options[1])
		}
	}

	return responseObj, nil
}

// analyzeIntentRuleBased performs rule-based intent analysis as fallback
func analyzeIntentRuleBased(prompt string, contextData *ContextData) (*IntentAnalysisResponse, error) {
	promptLower := strings.ToLower(prompt)

	// Check for location uncertainty
	locationKeywords := []string{"add", "create", "put", "place", "where"}
	hasLocationKeyword := false
	for _, keyword := range locationKeywords {
		if strings.Contains(promptLower, keyword) {
			hasLocationKeyword = true
			break
		}
	}

	if hasLocationKeyword && len(prompt) < 30 {
		// Suggest common locations based on context
		options := []string{"src/", "lib/", "app/"}
		if contextData != nil && len(contextData.CodePatterns) > 0 {
			options = contextData.CodePatterns[:minInt(2, len(contextData.CodePatterns))]
		}

		templates := GetTemplates()
		template := templates[IntentLocationUnclear]
		question := FormatTemplate(template, options[0], options[1])

		return &IntentAnalysisResponse{
			Success:               true,
			RequiresClarification: true,
			IntentType:            IntentLocationUnclear,
			Confidence:            0.6,
			ClarifyingQuestion:    question,
			Options:               options,
		}, nil
	}

	// Default: prompt seems clear
	return &IntentAnalysisResponse{
		Success:               true,
		RequiresClarification: false,
		IntentType:            IntentClear,
		Confidence:            0.7,
		SuggestedAction:       prompt,
		ResolvedPrompt:        prompt,
	}, nil
}

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
		err = queryRowWithTimeout(ctx, updateQuery,
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
		err = queryRowWithTimeout(ctx, query,
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
	result, err := execWithTimeout(ctx, updateQuery, projectID, string(intentType), patternDataJSON)
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
	_, err = execWithTimeout(ctx, insertQuery, projectID, string(intentType), patternDataJSON)
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

	rows, err := queryWithTimeout(ctx, query, projectID)
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
		_, err = execWithTimeout(ctx, updateQuery, projectID, intentType, patternDataJSON, frequency)
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

	rows, err := queryWithTimeout(ctx, query, projectID)
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
