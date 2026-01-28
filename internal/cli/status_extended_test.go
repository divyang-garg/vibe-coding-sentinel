// Package cli provides extended tests for status command
package cli

import (
	"os"
	"os/exec"
	"testing"
)

func TestRunStatus_AllPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("with .sentinelrc", func(t *testing.T) {
		os.WriteFile(".sentinelrc", []byte("{}"), 0644)
		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with project indicators", func(t *testing.T) {
		os.WriteFile("package.json", []byte("{}"), 0644)
		os.WriteFile("go.mod", []byte("module test"), 0644)
		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with git repository clean", func(t *testing.T) {
		os.MkdirAll(".git", 0755)
		// Initialize git if available
		exec.Command("git", "init").Run()

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with git repository with changes", func(t *testing.T) {
		os.MkdirAll(".git", 0755)
		exec.Command("git", "init").Run()
		os.WriteFile("modified.go", []byte("package main"), 0644)
		exec.Command("git", "add", "modified.go").Run()

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with test files", func(t *testing.T) {
		os.MkdirAll("tests", 0755)
		os.WriteFile("tests/test.go", []byte("package main"), 0644)
		os.WriteFile("main_test.go", []byte("package main"), 0644)

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with patterns.json", func(t *testing.T) {
		os.MkdirAll(".sentinel", 0755)
		os.WriteFile(".sentinel/patterns.json", []byte("{}"), 0644)

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with cursor rules", func(t *testing.T) {
		os.MkdirAll(".cursor/rules", 0755)

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("with recommendations", func(t *testing.T) {
		// Test path where recommendations are shown
		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("test file detection with various patterns", func(t *testing.T) {
		os.WriteFile("test_file.go", []byte("package main"), 0644)
		os.WriteFile("spec_file.js", []byte("test"), 0644)
		os.MkdirAll("testdir", 0755)
		os.WriteFile("testdir/test.go", []byte("package main"), 0644)

		err := runStatus([]string{})
		if err != nil {
			t.Errorf("runStatus() error = %v", err)
		}
	})

	t.Run("git status command error", func(t *testing.T) {
		os.MkdirAll(".git", 0755)
		// Git command may fail in test environment
		err := runStatus([]string{})
		// Should handle git command errors gracefully
		if err != nil {
			t.Errorf("runStatus() should handle git errors, error = %v", err)
		}
	})
}

func TestRunStatus_WalkError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("filepath walk with error", func(t *testing.T) {
		// Create a directory structure
		os.MkdirAll("testdir", 0755)

		err := runStatus([]string{})
		// Should handle walk errors gracefully
		if err != nil {
			t.Errorf("runStatus() should handle walk errors, error = %v", err)
		}
	})
}
