// Package mcp provides tests for MCP handlers
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package mcp

import (
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
