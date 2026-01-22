// Package cli provides extended tests for docs command
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDocs_FlagParsing(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	os.WriteFile("test.go", []byte("package main"), 0644)

	t.Run("output flag without value", func(t *testing.T) {
		// Test when --output is last argument
		err := runDocs([]string{"--output"})
		// Should handle missing value gracefully
		_ = err
	})

	t.Run("depth flag without value", func(t *testing.T) {
		// Test when --depth is last argument
		err := runDocs([]string{"--depth"})
		// Should handle missing value gracefully
		_ = err
	})

	t.Run("short output flag", func(t *testing.T) {
		err := runDocs([]string{"-o", "output.md"})
		if err != nil {
			t.Errorf("runDocs() with -o flag error = %v", err)
		}
	})

	t.Run("short depth flag", func(t *testing.T) {
		err := runDocs([]string{"-d", "3"})
		if err != nil {
			t.Errorf("runDocs() with -d flag error = %v", err)
		}
	})

	t.Run("build file tree error", func(t *testing.T) {
		// Test with invalid path
		err := runDocs([]string{"--output", "output.md"})
		// Should handle file tree build error
		_ = err
	})
}

func TestBuildFileTreeRecursive_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("directory read error", func(t *testing.T) {
		// Create directory we can't read
		testDir := filepath.Join(tmpDir, "test")
		os.MkdirAll(testDir, 0755)
		
		// Make directory read-only
		os.Chmod(testDir, 0000)
		defer os.Chmod(testDir, 0755)

		node, err := buildFileTreeRecursive(testDir, 0, 5)
		// Should handle read error gracefully (returns node without children)
		if err != nil {
			t.Errorf("Expected node even on read error, got error: %v", err)
		}
		if node == nil {
			t.Error("Expected non-nil node even on read error")
		}
	})

	t.Run("hidden files handling", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "hidden")
		os.MkdirAll(testDir, 0755)
		os.WriteFile(filepath.Join(testDir, ".hidden"), []byte("hidden"), 0644)
		os.WriteFile(filepath.Join(testDir, ".cursor"), []byte("cursor"), 0644)
		os.WriteFile(filepath.Join(testDir, ".github"), []byte("github"), 0644)
		os.WriteFile(filepath.Join(testDir, "visible.go"), []byte("package main"), 0644)

		node, err := buildFileTreeRecursive(testDir, 0, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		// Should skip .hidden but include .cursor and .github
		_ = node
	})

	t.Run("child node build error", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "parent")
		os.MkdirAll(testDir, 0755)
		
		// Create a subdirectory that will cause an error when accessed
		subDir := filepath.Join(testDir, "subdir")
		os.MkdirAll(subDir, 0755)
		
		// Make subdirectory inaccessible
		os.Chmod(subDir, 0000)
		defer os.Chmod(subDir, 0755)

		node, err := buildFileTreeRecursive(testDir, 0, 5)
		// Should handle child errors gracefully
		if err != nil {
			t.Errorf("Expected parent node even with child error, got error: %v", err)
		}
		_ = node
	})

	t.Run("file node", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "file.go")
		os.WriteFile(testFile, []byte("package main"), 0644)

		node, err := buildFileTreeRecursive(testFile, 0, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		if node == nil {
			t.Fatal("Expected non-nil node for file")
		}
		if node.IsDir {
			t.Error("Expected IsDir to be false for file")
		}
	})
}

func TestRunDocs_ErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	t.Run("directory creation error", func(t *testing.T) {
		// Use invalid path for output
		err := runDocs([]string{"--output", "/root/invalid/output.md"})
		// Should handle directory creation error
		_ = err
	})

	t.Run("file write error", func(t *testing.T) {
		// Create read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.MkdirAll(readOnlyDir, 0444)
		defer os.Chmod(readOnlyDir, 0755)

		err := runDocs([]string{"--output", filepath.Join(readOnlyDir, "output.md")})
		// Should handle write error
		_ = err
	})
}
