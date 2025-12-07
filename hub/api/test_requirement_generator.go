// Phase 10A: Test Requirement Generator
// Generates test requirements from approved business rules in knowledge base

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TestRequirement represents a test requirement generated from a business rule
type TestRequirement struct {
	ID              string    `json:"id"`
	KnowledgeItemID string    `json:"knowledge_item_id"`
	RuleTitle       string    `json:"rule_title"`
	RequirementType string    `json:"requirement_type"` // happy_path, edge_case, error_case
	Description     string    `json:"description"`
	CodeFunction    string    `json:"code_function,omitempty"`
	Priority        string    `json:"priority"` // critical, high, medium, low
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GenerateTestRequirementsRequest represents the request to generate test requirements
type GenerateTestRequirementsRequest struct {
	ProjectID      string   `json:"projectId"`
	KnowledgeItemIDs []string `json:"knowledgeItemIds,omitempty"` // Optional: specific items, empty = all approved
}

// GenerateTestRequirementsResponse represents the response
type GenerateTestRequirementsResponse struct {
	Success       bool                `json:"success"`
	Requirements  []TestRequirement   `json:"requirements"`
	Count         int                 `json:"count"`
	Message       string              `json:"message,omitempty"`
}

// extractBusinessRules extracts approved business rules from knowledge_items table
func extractBusinessRules(ctx context.Context, projectID string, knowledgeItemIDs []string) ([]KnowledgeItem, error) {
	query := `
		SELECT ki.id, ki.document_id, ki.type, ki.title, ki.content, ki.confidence, 
		       ki.source_page, ki.status, ki.approved_by, ki.approved_at, ki.created_at
		FROM knowledge_items ki
		INNER JOIN documents d ON ki.document_id = d.id
		WHERE d.project_id = $1
		  AND ki.type = 'business_rule'
		  AND ki.status = 'approved'
	`
	
	args := []interface{}{projectID}
	argIndex := 2
	
	// If specific knowledge item IDs provided, filter by them
	if len(knowledgeItemIDs) > 0 {
		query += fmt.Sprintf(" AND ki.id = ANY($%d)", argIndex)
		args = append(args, knowledgeItemIDs)
		argIndex++
	}
	
	query += " ORDER BY ki.created_at DESC"
	
	rows, err := queryWithTimeout(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query business rules: %w", err)
	}
	defer rows.Close()
	
	var rules []KnowledgeItem
	for rows.Next() {
		var rule KnowledgeItem
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		
		err := rows.Scan(
			&rule.ID, &rule.DocumentID, &rule.Type, &rule.Title, &rule.Content,
			&rule.Confidence, &rule.SourcePage, &rule.Status,
			&approvedBy, &approvedAt, &rule.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning business rule: %v", err)
			continue
		}
		
		if approvedBy.Valid {
			rule.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			rule.ApprovedAt = &approvedAt.Time
		}
		
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// mapRuleToCode attempts to map a business rule to code functions using AST analysis
// This is a simplified version - in production, this would use the AST analyzer
func mapRuleToCode(rule KnowledgeItem, projectCode []string) string {
	// Simple heuristic: look for function names that match rule title keywords
	ruleTitleLower := strings.ToLower(rule.Title)
	keywords := extractKeywords(ruleTitleLower)
	
	// Search for functions containing keywords
	for _, code := range projectCode {
		codeLower := strings.ToLower(code)
		for _, keyword := range keywords {
			if strings.Contains(codeLower, keyword) {
				// Try to extract function name (simplified)
				if fn := extractFunctionNameFromCode(code, keyword); fn != "" {
					return fn
				}
			}
		}
	}
	
	return "" // No mapping found
}

// extractKeywords extracts meaningful keywords from a title
func extractKeywords(title string) []string {
	// Remove common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
	}
	
	words := strings.Fields(title)
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[strings.ToLower(word)] {
			keywords = append(keywords, strings.ToLower(word))
		}
	}
	
	return keywords
}

// extractFunctionNameFromCode attempts to extract a function name from code (simplified)
func extractFunctionNameFromCode(code, keyword string) string {
	// This is a placeholder - in production, use AST analysis
	// Look for function definitions containing the keyword
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "func ") && strings.Contains(lineLower, keyword) {
			// Try to extract function name
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "func" && i+1 < len(parts) {
					funcName := parts[i+1]
					// Remove receiver if present
					if strings.Contains(funcName, "(") {
						continue
					}
					return funcName
				}
			}
		}
	}
	return ""
}

// generateTestRequirements generates test requirements for a business rule
func generateTestRequirements(rule KnowledgeItem, codeFunction string) []TestRequirement {
	var requirements []TestRequirement
	now := time.Now()
	
	// Generate happy path requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "happy_path",
		Description:     fmt.Sprintf("Test normal operation: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "high",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	
	// Generate edge case requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "edge_case",
		Description:     fmt.Sprintf("Test boundary conditions and edge cases: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "medium",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	
	// Generate error case requirement
	requirements = append(requirements, TestRequirement{
		ID:              uuid.New().String(),
		KnowledgeItemID: rule.ID,
		RuleTitle:       rule.Title,
		RequirementType: "error_case",
		Description:     fmt.Sprintf("Test error handling and invalid inputs: %s", rule.Content),
		CodeFunction:    codeFunction,
		Priority:        "high",
		CreatedAt:       now,
		UpdatedAt:       now,
	})
	
	return requirements
}

// saveTestRequirements saves test requirements to database
func saveTestRequirements(ctx context.Context, requirements []TestRequirement) error {
	query := `
		INSERT INTO test_requirements 
		(id, knowledge_item_id, rule_title, requirement_type, description, code_function, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			description = EXCLUDED.description,
			code_function = EXCLUDED.code_function,
			priority = EXCLUDED.priority,
			updated_at = EXCLUDED.updated_at
	`
	
	for _, req := range requirements {
		_, err := execWithTimeout(ctx, query,
			req.ID, req.KnowledgeItemID, req.RuleTitle, req.RequirementType,
			req.Description, req.CodeFunction, req.Priority, req.CreatedAt, req.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save test requirement %s: %w", req.ID, err)
		}
	}
	
	return nil
}

// generateTestRequirementsHandler handles the API request to generate test requirements
func generateTestRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req GenerateTestRequirementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate project ID
	if req.ProjectID == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// Extract business rules
	rules, err := extractBusinessRules(ctx, req.ProjectID, req.KnowledgeItemIDs)
	if err != nil {
		log.Printf("Error extracting business rules: %v", err)
		http.Error(w, fmt.Sprintf("Failed to extract business rules: %v", err), http.StatusInternalServerError)
		return
	}
	
	if len(rules) == 0 {
		response := GenerateTestRequirementsResponse{
			Success:      true,
			Requirements: []TestRequirement{},
			Count:        0,
			Message:      "No approved business rules found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Generate test requirements for each rule
	var allRequirements []TestRequirement
	for _, rule := range rules {
		// TODO: In production, use AST analysis to map rules to code functions
		// For now, use empty code function (manual mapping can be done later)
		codeFunction := ""
		
		requirements := generateTestRequirements(rule, codeFunction)
		allRequirements = append(allRequirements, requirements...)
	}
	
	// Save to database
	if err := saveTestRequirements(ctx, allRequirements); err != nil {
		log.Printf("Error saving test requirements: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save test requirements: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := GenerateTestRequirementsResponse{
		Success:      true,
		Requirements: allRequirements,
		Count:        len(allRequirements),
		Message:      fmt.Sprintf("Generated %d test requirements from %d business rules", len(allRequirements), len(rules)),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

