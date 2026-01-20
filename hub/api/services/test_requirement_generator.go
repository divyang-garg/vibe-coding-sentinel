// Test Requirement Generator - Main Handler and Types
// Generates test requirements from approved business rules in knowledge base
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// GenerateTestRequirementsRequest represents the request to generate test requirements
type GenerateTestRequirementsRequest struct {
	ProjectID        string   `json:"project_id"`
	KnowledgeItemIDs []string `json:"knowledgeItemIds,omitempty"` // Optional: specific items, empty = all approved
}

// GenerateTestRequirementsResponse represents the response
type GenerateTestRequirementsResponse struct {
	Success      bool              `json:"success"`
	Requirements []TestRequirement `json:"requirements"`
	Count        int               `json:"count"`
	Message      string            `json:"message,omitempty"`
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
	rules, err := extractBusinessRules(ctx, req.ProjectID, req.KnowledgeItemIDs, "", nil)
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
		// CURRENT IMPLEMENTATION: Uses empty code function (manual mapping can be done later)
		// FUTURE ENHANCEMENT: Use AST analysis (Phase 6) to automatically map rules to code functions
		// This would enable automatic detection of which code implements which business rule
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
