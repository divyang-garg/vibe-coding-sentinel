// Package mcp provides MCP tool definitions
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package mcp

// GetRegisteredTools returns all registered MCP tools
func GetRegisteredTools() []Tool {
	return []Tool{
		{
			Name:        "sentinel_get_context",
			Description: "Get recent activity context including git status, recent commits, and error logs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"codebasePath": map[string]interface{}{
						"type":        "string",
						"description": "Path to codebase root (optional, defaults to current directory)",
					},
				},
			},
		},
		{
			Name:        "sentinel_get_patterns",
			Description: "Get learned patterns and project conventions from intent analysis",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"patternType": map[string]interface{}{
						"type":        "string",
						"description": "Filter patterns by type (optional)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of patterns to return (default: 50)",
						"default":     50,
					},
				},
			},
		},
		{
			Name:        "sentinel_check_file_size",
			Description: "Check if a file exceeds size limits according to CODING_STANDARDS.md",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Path to file to check",
					},
				},
				"required": []string{"file"},
			},
		},
		{
			Name:        "sentinel_audit",
			Description: "Run security audit on specified path",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to codebase to audit (optional, defaults to current directory)",
					},
					"vibeCheck": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable vibe coding detection",
					},
					"deep": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable Hub-based AST analysis",
					},
				},
			},
		},
		{
			Name:        "sentinel_vibe_check",
			Description: "Detect vibe coding issues (duplicate functions, orphaned code, etc.)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to codebase to check",
					},
				},
			},
		},
		{
			Name:        "sentinel_baseline_add",
			Description: "Add a finding to the baseline",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "File path",
					},
					"line": map[string]interface{}{
						"type":        "integer",
						"description": "Line number",
					},
					"reason": map[string]interface{}{
						"type":        "string",
						"description": "Reason for adding to baseline",
					},
				},
				"required": []string{"file", "line"},
			},
		},
		{
			Name:        "sentinel_knowledge_search",
			Description: "Search knowledge base entries",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}
