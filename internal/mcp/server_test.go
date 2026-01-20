// Package mcp provides unit tests for MCP server
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package mcp

import (
	"encoding/json"
	"testing"
)

func TestMCP_Initialize(t *testing.T) {
	server := NewServer()
	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage("{}"),
	}

	resp := server.handleInitialize(req)
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got %v", resp.Error)
	}
}

func TestMCP_ToolsList(t *testing.T) {
	server := NewServer()
	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  json.RawMessage("{}"),
	}

	resp := server.handleToolsList(req)
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got %v", resp.Error)
	}
}

func TestMCP_ToolsCall_Audit(t *testing.T) {
	server := NewServer()
	params := map[string]interface{}{
		"name": "sentinel_audit",
		"arguments": map[string]interface{}{
			"path": ".",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	resp := server.handleToolsCall(req)
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got %v", resp.Error)
	}
}

func TestMCP_ToolsCall_UnknownTool(t *testing.T) {
	server := NewServer()
	params := map[string]interface{}{
		"name":      "unknown_tool",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	resp := server.handleToolsCall(req)
	if resp == nil {
		t.Fatal("Response is nil")
	}

	if resp.Error == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestMCP_GetRegisteredTools(t *testing.T) {
	tools := GetRegisteredTools()
	if len(tools) < 3 {
		t.Errorf("Expected at least 3 tools, got %d", len(tools))
	}

	// Check for required tools
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	required := []string{"sentinel_get_context", "sentinel_get_patterns", "sentinel_check_file_size"}
	for _, name := range required {
		if !toolNames[name] {
			t.Errorf("Missing required tool: %s", name)
		}
	}
}
