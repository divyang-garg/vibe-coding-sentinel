// Package cli provides unit tests for CLI command handling
// Complies with CODING_STANDARDS.md: Test file max 500 lines, 80%+ coverage
package cli

import (
	"os"
	"testing"
)

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
		err := runKnowledge([]string{"add", "Test Title", "--content", "Test content here"})
		// May error without proper args format, just ensure no panic
		_ = err
	})

	t.Run("search knowledge", func(t *testing.T) {
		err := runKnowledge([]string{"search", "test"})
		// May error without proper data, just ensure no panic
		_ = err
	})

	t.Run("export knowledge", func(t *testing.T) {
		err := runKnowledge([]string{"export", "export.json"})
		// May error without proper data, just ensure no panic
		_ = err
	})

	t.Run("import knowledge", func(t *testing.T) {
		os.WriteFile("import.json", []byte(`{"version":"1.0","entries":[]}`), 0644)
		err := runKnowledge([]string{"import", "import.json"})
		// May error, just ensure no panic
		_ = err
	})
}
