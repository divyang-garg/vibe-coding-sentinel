// Package cli provides extended tests for main CLI routing
package cli

import (
	"os"
	"testing"
)

func TestExecute_AllCommands(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("empty args", func(t *testing.T) {
		err := Execute([]string{})
		if err != nil {
			t.Errorf("Execute() with empty args error = %v", err)
		}
	})

	t.Run("init command", func(t *testing.T) {
		err := Execute([]string{"init"})
		_ = err
	})

	t.Run("version command", func(t *testing.T) {
		err := Execute([]string{"version"})
		if err != nil {
			t.Errorf("Execute() with version error = %v", err)
		}
	})

	t.Run("version with --version", func(t *testing.T) {
		err := Execute([]string{"--version"})
		if err != nil {
			t.Errorf("Execute() with --version error = %v", err)
		}
	})

	t.Run("version with -v", func(t *testing.T) {
		err := Execute([]string{"-v"})
		if err != nil {
			t.Errorf("Execute() with -v error = %v", err)
		}
	})

	t.Run("help command", func(t *testing.T) {
		err := Execute([]string{"help"})
		if err != nil {
			t.Errorf("Execute() with help error = %v", err)
		}
	})

	t.Run("help with --help", func(t *testing.T) {
		err := Execute([]string{"--help"})
		if err != nil {
			t.Errorf("Execute() with --help error = %v", err)
		}
	})

	t.Run("help with -h", func(t *testing.T) {
		err := Execute([]string{"-h"})
		if err != nil {
			t.Errorf("Execute() with -h error = %v", err)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		err := Execute([]string{"unknown"})
		if err == nil {
			t.Error("Expected error for unknown command")
		}
	})

	t.Run("hook command", func(t *testing.T) {
		err := Execute([]string{"hook", "pre-commit"})
		_ = err
	})

	t.Run("install-hooks command", func(t *testing.T) {
		err := Execute([]string{"install-hooks"})
		_ = err
	})

	t.Run("mcp-server command", func(t *testing.T) {
		err := Execute([]string{"mcp-server"})
		_ = err
	})
}
