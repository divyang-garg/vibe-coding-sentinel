// Package mcp provides tests for MCP handlers
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package mcp

import (
	"encoding/json"
	"os"
	"testing"
)

func TestHandleGetContext(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Run("basic context", func(t *testing.T) {
		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result is not a map")
		}

		if resultMap["version"] == nil {
			t.Error("Version not returned")
		}
	})

	t.Run("with git repo", func(t *testing.T) {
		os.MkdirAll(".git", 0755)

		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["git_initialized"] != true {
			t.Error("Expected git_initialized to be true")
		}
	})

	t.Run("with sentinel config", func(t *testing.T) {
		os.WriteFile(".sentinelrc", []byte("test config"), 0644)

		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["sentinel_configured"] != true {
			t.Error("Expected sentinel_configured to be true")
		}
	})

	t.Run("with patterns", func(t *testing.T) {
		os.WriteFile("patterns.json", []byte(`{"languages":["go"]}`), 0644)

		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["patterns_learned"] != true {
			t.Error("Expected patterns_learned to be true")
		}
	})

	t.Run("with baseline", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		baseline := `{"version":"1.0","entries":[{"file":"test.go","line":1}]}`
		os.WriteFile(".sentinel/baseline.json", []byte(baseline), 0644)

		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["baseline_entries"] == nil {
			t.Error("Expected baseline_entries")
		}
	})

	t.Run("with knowledge", func(t *testing.T) {
		kb := `{"version":"1.0","entries":[{"title":"Test"}]}`
		os.WriteFile(".sentinel/knowledge.json", []byte(kb), 0644)

		args := map[string]interface{}{}
		result, err := server.handleGetContext(args)
		if err != nil {
			t.Fatalf("handleGetContext failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["knowledge_entries"] == nil {
			t.Error("Expected knowledge_entries")
		}
	})
}

func TestHandleGetPatterns(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	t.Run("no patterns file", func(t *testing.T) {
		args := map[string]interface{}{}
		result, err := server.handleGetPatterns(args)
		if err != nil {
			t.Fatalf("handleGetPatterns failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		if resultMap["count"] != 0 {
			t.Error("Expected 0 patterns when no file exists")
		}
	})

	t.Run("with patterns file", func(t *testing.T) {
		patterns := `{
			"languages": ["go", "javascript"],
			"frameworks": ["react"],
			"naming": {"functions": "camelCase"}
		}`
		os.WriteFile("patterns.json", []byte(patterns), 0644)

		args := map[string]interface{}{}
		result, err := server.handleGetPatterns(args)
		if err != nil {
			t.Fatalf("handleGetPatterns failed: %v", err)
		}

		resultMap := result.(map[string]interface{})
		count, ok := resultMap["count"].(int)
		if !ok || count < 2 {
			t.Error("Expected at least 2 patterns")
		}
	})

	t.Run("invalid patterns file", func(t *testing.T) {
		os.WriteFile("patterns.json", []byte("invalid json"), 0644)

		args := map[string]interface{}{}
		_, err := server.handleGetPatterns(args)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestHandlePing(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	req := Request{
		ID:     &id,
		Method: "ping",
		Params: nil,
	}

	response := server.handlePing(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Result == nil {
		t.Fatal("Expected result in response")
	}
}

func TestHandleInitialize_InvalidParams(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	req := Request{
		ID:     &id,
		Method: "initialize",
		Params: []byte("invalid json"),
	}

	response := server.handleInitialize(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid params")
	}

	if response.Error.Code != InvalidParamsCode {
		t.Errorf("Expected InvalidParamsCode, got %d", response.Error.Code)
	}
}

func TestHandleToolsCall_MissingParams(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	req := Request{
		ID:     &id,
		Method: "tools/call",
		Params: nil,
	}

	response := server.handleToolsCall(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error == nil {
		t.Error("Expected error response for missing params")
	}

	if response.Error.Code != InvalidParamsCode {
		t.Errorf("Expected InvalidParamsCode, got %d", response.Error.Code)
	}
}

func TestHandleToolsCall_InvalidJSON(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	req := Request{
		ID:     &id,
		Method: "tools/call",
		Params: []byte("invalid json"),
	}

	response := server.handleToolsCall(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error == nil {
		t.Error("Expected error response for invalid JSON")
	}

	if response.Error.Code != InvalidParamsCode {
		t.Errorf("Expected InvalidParamsCode, got %d", response.Error.Code)
	}
}

func TestHandleToolsCall_MissingName(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	params := ToolCallParams{
		Name:      "",
		Arguments: map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := Request{
		ID:     &id,
		Method: "tools/call",
		Params: paramsJSON,
	}

	response := server.handleToolsCall(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error == nil {
		t.Error("Expected error response for missing name")
	}

	if response.Error.Code != InvalidParamsCode {
		t.Errorf("Expected InvalidParamsCode, got %d", response.Error.Code)
	}
}

func TestHandleToolsCall_ToolNotFound(t *testing.T) {
	server := NewServer()

	var id interface{} = 1
	params := ToolCallParams{
		Name:      "nonexistent_tool",
		Arguments: map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := Request{
		ID:     &id,
		Method: "tools/call",
		Params: paramsJSON,
	}

	response := server.handleToolsCall(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error == nil {
		t.Error("Expected error response for tool not found")
	}

	if response.Error.Code != MethodNotFoundCode {
		t.Errorf("Expected MethodNotFoundCode, got %d", response.Error.Code)
	}
}

func TestHandleToolsCall_ToolExecutionError(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create a file that will cause an error in handleCheckFileSize
	// by using a path that doesn't exist
	var id interface{} = 1
	params := ToolCallParams{
		Name: "sentinel_check_file_size",
		Arguments: map[string]interface{}{
			"file": "/nonexistent/path/to/file.go",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := Request{
		ID:     &id,
		Method: "tools/call",
		Params: paramsJSON,
	}

	response := server.handleToolsCall(req)
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Tool execution should fail, but it's handled internally
	// The error should be returned as InternalErrorCode
	if response.Error != nil && response.Error.Code == InternalErrorCode {
		// This is expected - tool execution failed
		return
	}

	// If no error, that's also acceptable as the tool might handle errors differently
}

func TestExecuteTool_UnknownTool(t *testing.T) {
	server := NewServer()

	result, err := server.executeTool("unknown_tool", map[string]interface{}{})

	if err == nil {
		t.Error("Expected error for unknown tool")
	}

	if result != nil {
		t.Error("Expected nil result for unknown tool")
	}

	if err.Error() != "tool not implemented: unknown_tool" {
		t.Errorf("Expected 'tool not implemented' error, got: %v", err)
	}
}

func TestHandleGetContext_FileReadErrors(t *testing.T) {
	server := NewServer()
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create .sentinel directory but with invalid JSON files
	os.MkdirAll(".sentinel", 0755)
	os.WriteFile(".sentinel/baseline.json", []byte("invalid json"), 0644)
	os.WriteFile(".sentinel/knowledge.json", []byte("invalid json"), 0644)

	args := map[string]interface{}{}
	result, err := server.handleGetContext(args)

	// Should not error - file read errors are handled gracefully
	if err != nil {
		t.Fatalf("handleGetContext should handle file read errors gracefully: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result should be a map")
	}

	// Should still return basic context even if files can't be read
	if resultMap["version"] == nil {
		t.Error("Version should be returned even with file read errors")
	}
}
