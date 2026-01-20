// Package patterns provides business rules integration for Cursor rules
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package patterns

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HubBusinessContextResponse represents the response from Hub API
type HubBusinessContextResponse struct {
	Rules        []HubKnowledgeItem `json:"rules"`
	Entities     []HubKnowledgeItem `json:"entities"`
	UserJourneys []HubKnowledgeItem `json:"user_journeys"`
}

// HubKnowledgeItem represents a knowledge item from Hub API
type HubKnowledgeItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Confidence float64                `json:"confidence"`
	SourcePage int                    `json:"source_page"`
	Status     string                 `json:"status"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
}

// fetchBusinessRulesFromHub fetches business rules from Hub API
// Returns empty slice and logs warning if Hub is unavailable (non-critical failure)
func fetchBusinessRulesFromHub(hubURL, apiKey, projectID string) ([]BusinessRule, error) {
	if hubURL == "" {
		return nil, fmt.Errorf("hub URL is required")
	}

	// Build request URL
	baseURL := strings.TrimSuffix(hubURL, "/")
	apiURL := fmt.Sprintf("%s/api/v1/knowledge/business", baseURL)
	
	// Add project_id query parameter
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Hub URL: %w", err)
	}
	
	q := u.Query()
	if projectID != "" {
		q.Set("project_id", projectID)
	}
	u.RawQuery = q.Encode()
	apiURL = u.String()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key header if provided
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		// Network error - Hub unavailable, return empty slice (non-critical)
		return []BusinessRule{}, nil
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Non-200 status - log but don't fail (non-critical)
		return []BusinessRule{}, nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var hubResponse HubBusinessContextResponse
	if err := json.Unmarshal(body, &hubResponse); err != nil {
		// Invalid JSON - log but don't fail (non-critical)
		return []BusinessRule{}, nil
	}

	// Convert Hub knowledge items to BusinessRule format
	rules := make([]BusinessRule, 0, len(hubResponse.Rules))
	for _, item := range hubResponse.Rules {
		// Only include approved business rules
		if item.Type == "business_rule" && item.Status == "approved" {
			sourcePage := item.SourcePage
			rules = append(rules, BusinessRule{
				ID:         item.ID,
				Title:      item.Title,
				Content:    item.Content,
				Confidence: item.Confidence,
				SourcePage: &sourcePage,
			})
		}
	}

	return rules, nil
}

// generateBusinessRulesForCursor generates Cursor rules markdown from business rules
func generateBusinessRulesForCursor(rules []BusinessRule) string {
	var buf strings.Builder
	
	// Write YAML frontmatter
	buf.WriteString("---\n")
	buf.WriteString("description: Business Rules and Domain Logic\n")
	buf.WriteString("globs: [\"**/*\"]\n")
	buf.WriteString("alwaysApply: true\n")
	buf.WriteString("---\n\n")
	
	// Write header
	buf.WriteString("# Business Rules\n\n")
	buf.WriteString("This file contains business rules and domain logic extracted from project documentation.\n")
	buf.WriteString("These rules guide code generation and ensure business logic compliance.\n\n")
	
	// If no rules, add a note
	if len(rules) == 0 {
		buf.WriteString("_No business rules found. Run `sentinel knowledge extract` to extract rules from documents._\n")
		return buf.String()
	}
	
	// Write each rule
	for i, rule := range rules {
		// Rule heading
		buf.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, rule.Title))
		
		// Confidence metadata (as comment)
		if rule.Confidence > 0 {
			buf.WriteString(fmt.Sprintf("<!-- Confidence: %.2f%% -->\n", rule.Confidence*100))
		}
		
		// Source page metadata if available
		if rule.SourcePage != nil {
			buf.WriteString(fmt.Sprintf("<!-- Source Page: %d -->\n", *rule.SourcePage))
		}
		
		// Rule content
		content := strings.TrimSpace(rule.Content)
		if content != "" {
			// Format content as actionable guidance
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					buf.WriteString(line)
					buf.WriteString("\n\n")
				}
			}
		} else {
			buf.WriteString("_No detailed content available._\n\n")
		}
		
		// Add separator between rules
		if i < len(rules)-1 {
			buf.WriteString("---\n\n")
		}
	}
	
	// Add footer with usage instructions
	buf.WriteString("\n## Usage\n\n")
	buf.WriteString("When generating code, ensure compliance with these business rules:\n\n")
	buf.WriteString("1. **Review relevant rules** before implementing features\n")
	buf.WriteString("2. **Validate business logic** against these rules\n")
	buf.WriteString("3. **Update rules** when business requirements change\n")
	buf.WriteString("4. **Reference rule IDs** in code comments for traceability\n\n")
	
	return buf.String()
}
