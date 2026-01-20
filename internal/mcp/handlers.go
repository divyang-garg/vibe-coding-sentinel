// Package mcp provides MCP request handlers
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// handleInitialize handles MCP initialize request
func (s *Server) handleInitialize(req Request) *Response {
	var params InitializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.createErrorResponse(req.ID, InvalidParamsCode, "Invalid params", err.Error())
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"available": true,
				"count":     len(s.tools),
			},
		},
		ServerInfo: map[string]string{
			"name":    "sentinel",
			"version": "v24",
		},
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolsList handles MCP tools/list request
func (s *Server) handleToolsList(req Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": s.tools,
		},
	}
}

// handleToolsCall handles MCP tools/call request
func (s *Server) handleToolsCall(req Request) *Response {
	var params ToolCallParams
	if len(req.Params) == 0 {
		return s.createErrorResponse(req.ID, InvalidParamsCode, "Invalid params",
			"tools/call method requires parameters")
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.createErrorResponse(req.ID, InvalidParamsCode, "Invalid params", err.Error())
	}

	if params.Name == "" {
		return s.createErrorResponse(req.ID, InvalidParamsCode, "Invalid params",
			"tools/call requires 'name' parameter")
	}

	// Find tool
	var tool *Tool
	for i := range s.tools {
		if s.tools[i].Name == params.Name {
			tool = &s.tools[i]
			break
		}
	}

	if tool == nil {
		return s.createErrorResponse(req.ID, MethodNotFoundCode, "Tool not found",
			fmt.Sprintf("Unknown tool: %s", params.Name))
	}

	// Execute tool
	result, err := s.executeTool(params.Name, params.Arguments)
	if err != nil {
		return s.createErrorResponse(req.ID, InternalErrorCode, "Tool execution failed", err.Error())
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handlePing handles ping requests for health checks
func (s *Server) handlePing(req Request) *Response {
	uptime := time.Since(s.startTime)
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"status":         "healthy",
			"uptime_seconds": uptime.Seconds(),
			"stats": map[string]interface{}{
				"requests_total":  s.stats.RequestsTotal,
				"errors_total":    s.stats.ErrorsTotal,
				"active_requests": s.stats.ActiveRequests,
			},
		},
	}
}

// executeTool executes a tool by name
func (s *Server) executeTool(name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "sentinel_get_context":
		return s.handleGetContext(args)
	case "sentinel_get_patterns":
		return s.handleGetPatterns(args)
	case "sentinel_check_file_size":
		return s.handleCheckFileSize(args)
	case "sentinel_audit":
		return s.handleAudit(args)
	case "sentinel_vibe_check":
		return s.handleVibeCheck(args)
	case "sentinel_baseline_add":
		return s.handleBaselineAdd(args)
	case "sentinel_knowledge_search":
		return s.handleKnowledgeSearch(args)
	default:
		return nil, fmt.Errorf("tool not implemented: %s", name)
	}
}

// handleGetContext returns recent activity context
func (s *Server) handleGetContext(args map[string]interface{}) (interface{}, error) {
	codebasePath := "."
	if p, ok := args["codebasePath"].(string); ok && p != "" {
		codebasePath = p
	}

	context := map[string]interface{}{
		"version":       "v24",
		"server_uptime": time.Since(s.startTime).Seconds(),
	}

	// Check for git info
	if _, err := os.Stat(filepath.Join(codebasePath, ".git")); err == nil {
		context["git_initialized"] = true
	}

	// Check for sentinel config
	if _, err := os.Stat(filepath.Join(codebasePath, ".sentinelrc")); err == nil {
		context["sentinel_configured"] = true
	}

	// Check for patterns
	if _, err := os.Stat(filepath.Join(codebasePath, "patterns.json")); err == nil {
		context["patterns_learned"] = true
	}

	// Check for baseline
	if data, err := os.ReadFile(filepath.Join(codebasePath, ".sentinel", "baseline.json")); err == nil {
		var baseline map[string]interface{}
		if json.Unmarshal(data, &baseline) == nil {
			if entries, ok := baseline["entries"].([]interface{}); ok {
				context["baseline_entries"] = len(entries)
			}
		}
	}

	// Check for knowledge
	if data, err := os.ReadFile(filepath.Join(codebasePath, ".sentinel", "knowledge.json")); err == nil {
		var kb map[string]interface{}
		if json.Unmarshal(data, &kb) == nil {
			if entries, ok := kb["entries"].([]interface{}); ok {
				context["knowledge_entries"] = len(entries)
			}
		}
	}

	return context, nil
}

// handleGetPatterns returns learned patterns
func (s *Server) handleGetPatterns(args map[string]interface{}) (interface{}, error) {
	// Try to load patterns from patterns.json
	patternsPath := "patterns.json"
	data, err := os.ReadFile(patternsPath)
	if err != nil {
		// No patterns file, return empty
		return map[string]interface{}{
			"patterns": []string{},
			"count":    0,
			"message":  "No patterns learned yet. Run 'sentinel learn' to analyze codebase.",
		}, nil
	}

	// Parse patterns file
	var patterns map[string]interface{}
	if err := json.Unmarshal(data, &patterns); err != nil {
		return nil, fmt.Errorf("failed to parse patterns: %w", err)
	}

	// Extract pattern information
	patternList := make([]map[string]interface{}, 0)

	// Add languages
	if languages, ok := patterns["languages"].([]interface{}); ok {
		for _, lang := range languages {
			patternList = append(patternList, map[string]interface{}{
				"type":  "language",
				"value": lang,
			})
		}
	}

	// Add frameworks
	if frameworks, ok := patterns["frameworks"].([]interface{}); ok {
		for _, fw := range frameworks {
			patternList = append(patternList, map[string]interface{}{
				"type":  "framework",
				"value": fw,
			})
		}
	}

	// Add naming conventions
	if naming, ok := patterns["naming"].(map[string]interface{}); ok {
		for key, value := range naming {
			patternList = append(patternList, map[string]interface{}{
				"type":  "naming",
				"key":   key,
				"value": value,
			})
		}
	}

	return map[string]interface{}{
		"patterns": patternList,
		"count":    len(patternList),
		"raw":      patterns,
	}, nil
}
