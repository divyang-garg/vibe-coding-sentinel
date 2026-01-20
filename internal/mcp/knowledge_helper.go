// Package mcp provides knowledge helper to avoid circular dependencies
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package mcp

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

// KnowledgeEntry represents a knowledge base entry
type KnowledgeEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
}

// KnowledgeBase represents the knowledge base
type KnowledgeBase struct {
	Version string           `json:"version"`
	Entries []KnowledgeEntry `json:"entries"`
}

// searchKnowledgeBase searches knowledge entries
func searchKnowledgeBase(query string) ([]map[string]interface{}, error) {
	kbPath := ".sentinel/knowledge.json"

	// Load knowledge base
	data, err := os.ReadFile(kbPath)
	if err != nil {
		return nil, err
	}

	var kb KnowledgeBase
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, err
	}

	// Search entries
	queryLower := strings.ToLower(query)
	results := make([]map[string]interface{}, 0)

	for _, entry := range kb.Entries {
		// Check title, content, and tags
		if strings.Contains(strings.ToLower(entry.Title), queryLower) ||
			strings.Contains(strings.ToLower(entry.Content), queryLower) ||
			containsTag(entry.Tags, queryLower) {

			results = append(results, map[string]interface{}{
				"id":      entry.ID,
				"title":   entry.Title,
				"content": truncateContent(entry.Content, 200),
				"type":    entry.Type,
				"tags":    entry.Tags,
				"status":  entry.Status,
			})
		}
	}

	return results, nil
}

// containsTag checks if any tag contains the query
func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// truncateContent truncates content to maxLen
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
