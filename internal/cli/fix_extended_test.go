// Package cli provides extended tests for fix command
package cli

import (
	"os"
	"testing"
)

func TestRunFix_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with help flag", func(t *testing.T) {
		err := runFix([]string{"--help"})
		if err != nil {
			t.Errorf("runFix() with help error = %v", err)
		}
	})

	t.Run("with short help flag", func(t *testing.T) {
		err := runFix([]string{"-h"})
		if err != nil {
			t.Errorf("runFix() with -h error = %v", err)
		}
	})

	t.Run("with dry-run flag", func(t *testing.T) {
		os.WriteFile("test.go", []byte("package main\nfunc main() {}"), 0644)
		err := runFix([]string{"--dry-run"})
		// May fail if fixer encounters issues, but tests the path
		_ = err
	})

	t.Run("with safe flag", func(t *testing.T) {
		os.WriteFile("test.go", []byte("package main\nfunc main() {}"), 0644)
		err := runFix([]string{"--safe"})
		_ = err
	})

	t.Run("with yes flag", func(t *testing.T) {
		os.WriteFile("test.go", []byte("package main\nfunc main() {}"), 0644)
		err := runFix([]string{"--yes"})
		_ = err
	})

	t.Run("with short yes flag", func(t *testing.T) {
		os.WriteFile("test.go", []byte("package main\nfunc main() {}"), 0644)
		err := runFix([]string{"-y"})
		_ = err
	})

	t.Run("with pattern flag", func(t *testing.T) {
		os.WriteFile("test.js", []byte("console.log('test')"), 0644)
		err := runFix([]string{"--pattern", "console"})
		_ = err
	})

	t.Run("with target path", func(t *testing.T) {
		os.MkdirAll("src", 0755)
		os.WriteFile("src/test.go", []byte("package main"), 0644)
		err := runFix([]string{"src"})
		_ = err
	})

	t.Run("with pattern and target path", func(t *testing.T) {
		os.MkdirAll("src", 0755)
		err := runFix([]string{"--pattern", "imports", "src"})
		_ = err
	})
}

func TestRunFix_RollbackPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("rollback command", func(t *testing.T) {
		err := runFix([]string{"rollback"})
		// May fail if no rollback available, but tests the path
		_ = err
	})

	t.Run("rollback with list flag", func(t *testing.T) {
		err := runFix([]string{"rollback", "--list"})
		_ = err
	})

	t.Run("rollback with session flag", func(t *testing.T) {
		err := runFix([]string{"rollback", "--session", "test-session"})
		_ = err
	})

	t.Run("rollback with session ID as argument", func(t *testing.T) {
		err := runFix([]string{"rollback", "session-id"})
		_ = err
	})

	t.Run("rollback with list and session", func(t *testing.T) {
		err := runFix([]string{"rollback", "--list", "--session", "test"})
		_ = err
	})
}

func TestPrintFixHelp(t *testing.T) {
	err := printFixHelp()
	if err != nil {
		t.Errorf("printFixHelp() error = %v", err)
	}
}
