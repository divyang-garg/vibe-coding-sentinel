// Package cli provides unit tests for CLI command handling
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package cli

import (
	"os"
	"testing"
	"time"
)

func init() {
	// Set test mode to prevent os.Exit(1) from terminating test process
	os.Setenv("SENTINEL_TEST_MODE", "1")
}

func TestExecute_UnknownCommand(t *testing.T) {
	err := Execute([]string{"unknown-command"})
	if err == nil {
		t.Error("Expected error for unknown command")
	}
}

func TestExecute_AllCommandsAccessible(t *testing.T) {
	commands := []string{
		"init", "audit", "learn", "fix", "status",
		"baseline", "history", "docs", "install-hooks",
		"validate-rules", "update-rules", "knowledge",
		"review", "doc-sync", "mcp-server", "version", "help",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			// Just verify command doesn't panic
			// Most commands will fail without proper setup, but that's OK
			_ = Execute([]string{cmd, "--help"})
		})
	}
}

func TestExecute_HelpCommand(t *testing.T) {
	err := Execute([]string{"help"})
	if err != nil {
		t.Errorf("Help command should not error: %v", err)
	}
}

func TestExecute_VersionCommand(t *testing.T) {
	err := Execute([]string{"version"})
	if err != nil {
		t.Errorf("Version command should not error: %v", err)
	}
}

func TestBaseline_AddRemove(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create .sentinel directory
	os.MkdirAll(".sentinel", 0755)

	// Test add
	err := runBaseline([]string{"add", "test.js", "42", "Test reason"})
	if err != nil {
		t.Errorf("Failed to add baseline: %v", err)
	}

	// Test list
	err = runBaseline([]string{})
	if err != nil {
		t.Errorf("Failed to list baseline: %v", err)
	}

	// Test remove
	err = runBaseline([]string{"remove", "1"})
	if err != nil {
		t.Errorf("Failed to remove baseline: %v", err)
	}
}

func TestKnowledge_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	err := runKnowledge([]string{"list"})
	if err != nil {
		t.Errorf("Knowledge list should work even when empty: %v", err)
	}
}

func TestExecuteFlags(t *testing.T) {
	t.Run("-v flag", func(t *testing.T) {
		err := Execute([]string{"-v"})
		if err != nil {
			t.Errorf("Execute(-v) error = %v", err)
		}
	})

	t.Run("--version flag", func(t *testing.T) {
		err := Execute([]string{"--version"})
		if err != nil {
			t.Errorf("Execute(--version) error = %v", err)
		}
	})

	t.Run("-h flag", func(t *testing.T) {
		err := Execute([]string{"-h"})
		if err != nil {
			t.Errorf("Execute(-h) error = %v", err)
		}
	})

	t.Run("--help flag", func(t *testing.T) {
		err := Execute([]string{"--help"})
		if err != nil {
			t.Errorf("Execute(--help) error = %v", err)
		}
	})

	t.Run("empty args shows help", func(t *testing.T) {
		err := Execute([]string{})
		if err != nil {
			t.Errorf("Execute([]) error = %v", err)
		}
	})
}

func TestAllCommands(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)
	os.MkdirAll(".cursor/rules", 0755)
	os.MkdirAll(".git/hooks", 0755)

	t.Run("baseline command", func(t *testing.T) {
		err := runBaseline([]string{})
		if err != nil {
			t.Errorf("runBaseline() error = %v", err)
		}
	})

	t.Run("baseline help", func(t *testing.T) {
		err := runBaseline([]string{"help"})
		if err != nil {
			t.Errorf("runBaseline(help) error = %v", err)
		}
	})

	t.Run("knowledge help", func(t *testing.T) {
		err := runKnowledge([]string{"help"})
		if err != nil {
			t.Errorf("runKnowledge(help) error = %v", err)
		}
	})

	t.Run("history command", func(t *testing.T) {
		err := runHistory([]string{})
		if err != nil {
			t.Errorf("runHistory() error = %v", err)
		}
	})

	t.Run("status command", func(t *testing.T) {
		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("validate-rules command", func(t *testing.T) {
		// Create a valid rule file with YAML frontmatter
		validRule := `---
description: Test Rule
globs: ["*.go"]
---
# Test Rule

Content here.`
		os.WriteFile(".cursor/rules/test.md", []byte(validRule), 0644)
		err := runValidateRules([]string{})
		// May return validation error, but should not panic
		_ = err
	})

	t.Run("update-rules command", func(t *testing.T) {
		err := runUpdateRules([]string{})
		if err != nil {
			t.Errorf("runUpdateRules() error = %v", err)
		}
	})

	t.Run("docs command", func(t *testing.T) {
		err := runDocs([]string{"."})
		if err != nil {
			t.Errorf("runDocs() error = %v", err)
		}
	})

	t.Run("review command", func(t *testing.T) {
		err := runReview([]string{})
		if err != nil {
			t.Errorf("runReview() error = %v", err)
		}
	})

	t.Run("review command with pending entries", func(t *testing.T) {
		// Add draft entries for review
		_ = runKnowledge([]string{"add", "Draft Entry 1", "Content 1", "requirement"})
		_ = runKnowledge([]string{"add", "Draft Entry 2", "Content 2", "requirement"})
		// Review command will try to read stdin, so it will timeout/fail
		// but we're testing the code path
		err := runReview([]string{})
		// May error on stdin read, but should not panic
		_ = err
	})

	t.Run("doc-sync command", func(t *testing.T) {
		err := runDocSync([]string{})
		if err != nil {
			t.Errorf("runDocSync() error = %v", err)
		}
	})

	t.Run("learn command", func(t *testing.T) {
		err := runLearn([]string{})
		if err != nil {
			t.Errorf("runLearn() error = %v", err)
		}
	})

	t.Run("learn command with flags", func(t *testing.T) {
		testCases := []struct {
			name string
			args []string
		}{
			{"naming flag", []string{"--naming"}},
			{"imports flag", []string{"--imports"}},
			{"structure flag", []string{"--structure"}},
			{"output json", []string{"--output", "json"}},
			{"output=json", []string{"--output=json"}},
			{"include business rules", []string{"--include-business-rules"}},
			{"project id", []string{"--project-id", "test123"}},
			{"hub url", []string{"--hub-url", "http://localhost:8080"}},
			{"hub api key", []string{"--hub-api-key", "test-key"}},
			{"custom path", []string{"."}},
			{"multiple flags", []string{"--naming", "--output", "json"}},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := runLearn(tc.args)
				// Most will fail without proper setup, but should not panic
				_ = err
			})
		}
	})

	t.Run("fix dry-run command", func(t *testing.T) {
		err := runFix([]string{"--dry-run"})
		if err != nil {
			t.Errorf("runFix(--dry-run) error = %v", err)
		}
	})

	t.Run("install-hooks command", func(t *testing.T) {
		err := runInstallHooks([]string{})
		if err != nil {
			t.Errorf("runInstallHooks() error = %v", err)
		}
	})

	t.Run("init command", func(t *testing.T) {
		err := runInit([]string{})
		if err != nil {
			t.Errorf("runInit() error = %v", err)
		}
	})

	t.Run("audit command", func(t *testing.T) {
		err := runAudit([]string{"."})
		if err != nil {
			t.Errorf("runAudit() error = %v", err)
		}
	})
}

func TestPrintHelp(t *testing.T) {
	err := printHelp()
	if err != nil {
		t.Errorf("printHelp() error = %v", err)
	}
}

func TestRunVersion(t *testing.T) {
	err := runVersion()
	if err != nil {
		t.Errorf("runVersion() error = %v", err)
	}
}

func TestKnowledgeCommands(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("add knowledge", func(t *testing.T) {
		err := runKnowledge([]string{"add", "Test Title", "Test content here", "requirement", "tag1", "tag2"})
		if err != nil {
			t.Errorf("runKnowledge(add) error = %v", err)
		}
	})

	t.Run("add knowledge minimal", func(t *testing.T) {
		err := runKnowledge([]string{"add", "Title", "Content"})
		if err != nil {
			t.Errorf("runKnowledge(add minimal) error = %v", err)
		}
	})

	t.Run("search knowledge", func(t *testing.T) {
		// First add some knowledge
		_ = runKnowledge([]string{"add", "Auth Flow", "User must authenticate", "requirement"})
		err := runKnowledge([]string{"search", "auth"})
		if err != nil {
			t.Errorf("runKnowledge(search) error = %v", err)
		}
	})

	t.Run("search knowledge with tags", func(t *testing.T) {
		_ = runKnowledge([]string{"add", "Test", "Content", "requirement", "security", "auth"})
		err := runKnowledge([]string{"search", "security"})
		if err != nil {
			t.Errorf("runKnowledge(search tags) error = %v", err)
		}
	})

	t.Run("search knowledge no matches", func(t *testing.T) {
		err := runKnowledge([]string{"search", "nonexistent"})
		if err != nil {
			t.Errorf("runKnowledge(search no matches) error = %v", err)
		}
	})

	t.Run("export knowledge", func(t *testing.T) {
		// Add some data first
		_ = runKnowledge([]string{"add", "Export Test", "Test content", "requirement"})
		err := runKnowledge([]string{"export", "export.json"})
		if err != nil {
			t.Errorf("runKnowledge(export) error = %v", err)
		}
		// Verify file exists
		if _, err := os.Stat("export.json"); os.IsNotExist(err) {
			t.Error("Export file was not created")
		}
	})

	t.Run("import knowledge", func(t *testing.T) {
		kbJSON := `{"version":"1.0","entries":[{"id":"test1","title":"Imported","content":"Content","type":"requirement","tags":[],"status":"approved","created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z"}]}`
		os.WriteFile("import.json", []byte(kbJSON), 0644)
		err := runKnowledge([]string{"import", "import.json"})
		if err != nil {
			t.Errorf("runKnowledge(import) error = %v", err)
		}
	})

	t.Run("list knowledge with entries", func(t *testing.T) {
		_ = runKnowledge([]string{"add", "List Test", "Content", "requirement"})
		err := runKnowledge([]string{"list"})
		if err != nil {
			t.Errorf("runKnowledge(list) error = %v", err)
		}
	})

	t.Run("knowledge extract help", func(t *testing.T) {
		err := runKnowledge([]string{"extract", "--help"})
		if err != nil {
			t.Errorf("runKnowledge(extract --help) error = %v", err)
		}
	})
}

func TestKnowledgeFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.MkdirAll(".sentinel", 0755)

	t.Run("loadKnowledge new file", func(t *testing.T) {
		kb, err := loadKnowledge()
		if err == nil {
			t.Error("Expected error when loading non-existent file")
		}
		if kb == nil {
			t.Error("Expected non-nil KnowledgeBase even on error")
		}
	})

	t.Run("loadKnowledge corrupted file", func(t *testing.T) {
		os.WriteFile(".sentinel/knowledge.json", []byte("invalid json"), 0644)
		kb, err := loadKnowledge()
		if err == nil {
			t.Error("Expected error when loading corrupted file")
		}
		if kb != nil {
			t.Error("Expected nil KnowledgeBase on parse error")
		}
	})

	t.Run("saveKnowledge creates directory", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{},
		}
		err := saveKnowledge(kb)
		if err != nil {
			t.Errorf("saveKnowledge() error = %v", err)
		}
	})

	t.Run("saveKnowledge with entries", func(t *testing.T) {
		kb := &KnowledgeBase{
			Version: "1.0",
			Entries: []KnowledgeEntry{
				{
					ID:        "test1",
					Title:     "Test",
					Content:   "Content",
					Type:      "requirement",
					Status:    "draft",
					CreatedAt: time.Now(),
				},
			},
		}
		err := saveKnowledge(kb)
		if err != nil {
			t.Errorf("saveKnowledge() with entries error = %v", err)
		}
	})

	t.Run("containsTag", func(t *testing.T) {
		tags := []string{"security", "auth", "user"}
		if !containsTag(tags, "security") {
			t.Error("Expected containsTag to find 'security'")
		}
		if containsTag(tags, "nonexistent") {
			t.Error("Expected containsTag to not find 'nonexistent'")
		}
		// containsTag is case-insensitive due to strings.ToLower in searchKnowledge
		// but the function itself does case-sensitive substring matching
		// Let's test what it actually does
		if !containsTag(tags, "sec") {
			t.Error("Expected containsTag to find partial match 'sec'")
		}
	})

	t.Run("truncate", func(t *testing.T) {
		if truncate("short", 10) != "short" {
			t.Error("Expected truncate to return original for short strings")
		}
		truncated := truncate("this is a very long string that should be truncated", 20)
		if len(truncated) <= 20 {
			t.Error("Expected truncate to add ellipsis")
		}
		if !endsWith(truncated, "...") {
			t.Error("Expected truncate to end with '...'")
		}
	})
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
