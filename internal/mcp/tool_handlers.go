// Package mcp provides MCP tool handler implementations
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package mcp

import (
	"fmt"
	"os"
	"strings"
)

// handleCheckFileSize checks if a file exceeds size limits
func (s *Server) handleCheckFileSize(args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file parameter required")
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Count lines
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := 1
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}

	// Determine file type and limits based on CODING_STANDARDS.md
	var limit int
	var fileType string

	if strings.HasSuffix(filePath, "_test.go") {
		limit = 500
		fileType = "test"
	} else if strings.Contains(filePath, "/cli/") {
		limit = 300
		fileType = "cli_handler"
	} else if strings.Contains(filePath, "/api/handlers/") {
		limit = 200
		fileType = "http_handler"
	} else if strings.Contains(filePath, "/services/") || strings.Contains(filePath, "/scanner/") {
		limit = 400
		fileType = "business_service"
	} else if strings.Contains(filePath, "/models/") || strings.HasSuffix(filePath, "_types.go") {
		limit = 200
		fileType = "types"
	} else {
		limit = 250
		fileType = "utility"
	}

	valid := lines <= limit
	violation := ""
	if !valid {
		violation = fmt.Sprintf("File exceeds %d line limit for %s files (has %d lines)", limit, fileType, lines)
	}

	return map[string]interface{}{
		"valid":      valid,
		"lines":      lines,
		"limit":      limit,
		"file_type":  fileType,
		"size_bytes": info.Size(),
		"violation":  violation,
	}, nil
}

// handleAudit runs security audit
func (s *Server) handleAudit(args map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	vibeCheck := false
	if v, ok := args["vibeCheck"].(bool); ok {
		vibeCheck = v
	}

	deep := false
	if d, ok := args["deep"].(bool); ok {
		deep = d
	}

	// Run actual scan
	result, err := runAuditScan(path, vibeCheck, deep)
	if err != nil {
		return nil, fmt.Errorf("audit failed: %w", err)
	}

	return result, nil
}

// handleVibeCheck detects vibe coding issues
func (s *Server) handleVibeCheck(args map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	result, err := runVibeCheck(path)
	if err != nil {
		return nil, fmt.Errorf("vibe check failed: %w", err)
	}

	return result, nil
}

// handleBaselineAdd adds finding to baseline
func (s *Server) handleBaselineAdd(args map[string]interface{}) (interface{}, error) {
	file, _ := args["file"].(string)
	line, _ := args["line"].(float64)
	reason, _ := args["reason"].(string)

	if file == "" || line == 0 {
		return nil, fmt.Errorf("file and line are required")
	}

	if reason == "" {
		reason = "Added via MCP"
	}

	// Add to baseline using helper
	if err := addToBaselineFile(file, int(line), reason); err != nil {
		return nil, fmt.Errorf("failed to add to baseline: %w", err)
	}

	return map[string]interface{}{
		"status":  "added",
		"file":    file,
		"line":    int(line),
		"reason":  reason,
		"message": fmt.Sprintf("Added to baseline: %s:%d", file, int(line)),
	}, nil
}

// handleKnowledgeSearch searches knowledge base
func (s *Server) handleKnowledgeSearch(args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Search knowledge base
	results, err := searchKnowledgeBase(query)
	if err != nil {
		// Return empty results if knowledge base doesn't exist
		return map[string]interface{}{
			"query":   query,
			"results": []interface{}{},
			"count":   0,
			"message": "No knowledge base found. Use 'sentinel knowledge add' to create entries.",
		}, nil
	}

	return map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}, nil
}
