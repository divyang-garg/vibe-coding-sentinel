// Package cli provides tests for docs command
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDocs(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Create test directory structure
	os.MkdirAll("src/components", 0755)
	os.WriteFile("src/components/test.go", []byte("package components"), 0644)
	os.WriteFile("readme.md", []byte("# Test"), 0644)

	t.Run("default output", func(t *testing.T) {
		err := runDocs([]string{})
		if err != nil {
			t.Errorf("runDocs() error = %v", err)
		}
		if _, err := os.Stat("docs/FILE_STRUCTURE.md"); os.IsNotExist(err) {
			t.Error("Expected FILE_STRUCTURE.md to be created")
		}
	})

	t.Run("custom output", func(t *testing.T) {
		err := runDocs([]string{"--output", "custom.md"})
		if err != nil {
			t.Errorf("runDocs() error = %v", err)
		}
		if _, err := os.Stat("custom.md"); os.IsNotExist(err) {
			t.Error("Expected custom.md to be created")
		}
	})

	t.Run("with depth", func(t *testing.T) {
		err := runDocs([]string{"--depth", "2"})
		if err != nil {
			t.Errorf("runDocs() error = %v", err)
		}
	})
}

func TestBuildFileTree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.MkdirAll(filepath.Join(tmpDir, "src", "components"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "src", "components", "test.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Test"), 0644)

	t.Run("valid tree", func(t *testing.T) {
		tree, err := buildFileTree(tmpDir, 5)
		if err != nil {
			t.Fatalf("buildFileTree() error = %v", err)
		}
		if tree == nil {
			t.Fatal("Expected non-nil tree")
		}
		if tree.Name != filepath.Base(tmpDir) {
			t.Errorf("Expected tree name %s, got %s", filepath.Base(tmpDir), tree.Name)
		}
	})

	t.Run("max depth", func(t *testing.T) {
		tree, err := buildFileTree(tmpDir, 0)
		if err != nil {
			t.Fatalf("buildFileTree() error = %v", err)
		}
		if tree != nil && len(tree.Children) > 0 {
			t.Error("Expected no children at depth 0")
		}
	})
}

func TestBuildFileTreeRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "test"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "test", "file.txt"), []byte("test"), 0644)

	t.Run("file node", func(t *testing.T) {
		node, err := buildFileTreeRecursive(filepath.Join(tmpDir, "test", "file.txt"), 0, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		if node == nil {
			t.Fatal("Expected non-nil node")
		}
		if node.IsDir {
			t.Error("Expected IsDir to be false for file")
		}
	})

	t.Run("directory node", func(t *testing.T) {
		node, err := buildFileTreeRecursive(tmpDir, 0, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		if node == nil {
			t.Fatal("Expected non-nil node")
		}
		if !node.IsDir {
			t.Error("Expected IsDir to be true for directory")
		}
	})

	t.Run("skip directories", func(t *testing.T) {
		os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
		node, err := buildFileTreeRecursive(filepath.Join(tmpDir, "node_modules"), 0, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		if node != nil {
			t.Error("Expected nil for skipped directory")
		}
	})

	t.Run("exceeds max depth", func(t *testing.T) {
		node, err := buildFileTreeRecursive(tmpDir, 10, 5)
		if err != nil {
			t.Fatalf("buildFileTreeRecursive() error = %v", err)
		}
		if node != nil {
			t.Error("Expected nil when depth exceeds max")
		}
	})
}

func TestGenerateMarkdown(t *testing.T) {
	node := &FileNode{
		Name:  "test",
		Path:  "/test",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:  "file.go",
				Path:  "/test/file.go",
				IsDir: false,
				Size:  100,
			},
		},
	}

	markdown := generateMarkdown(node)
	if !strings.Contains(markdown, "Project File Structure") {
		t.Error("Expected markdown to contain title")
	}
	if !strings.Contains(markdown, "test") {
		t.Error("Expected markdown to contain node name")
	}
}

func TestWriteNode(t *testing.T) {
	var sb strings.Builder

	t.Run("root node", func(t *testing.T) {
		sb.Reset()
		node := &FileNode{
			Name:  ".",
			Path:  ".",
			IsDir: true,
		}
		writeNode(&sb, node, "", true)
		output := sb.String()
		if !strings.Contains(output, ".") {
			t.Error("Expected output to contain '.'")
		}
	})

	t.Run("directory node", func(t *testing.T) {
		sb.Reset()
		node := &FileNode{
			Name:  "test",
			Path:  "test",
			IsDir: true,
			Children: []*FileNode{
				{Name: "file.go", IsDir: false},
			},
		}
		writeNode(&sb, node, "", true)
		output := sb.String()
		if !strings.Contains(output, "test") {
			t.Error("Expected output to contain directory name")
		}
	})

	t.Run("file node", func(t *testing.T) {
		sb.Reset()
		node := &FileNode{
			Name:  "file.go",
			Path:  "file.go",
			IsDir: false,
		}
		writeNode(&sb, node, "", false)
		output := sb.String()
		if !strings.Contains(output, "file.go") {
			t.Error("Expected output to contain file name")
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		sb.Reset()
		node := &FileNode{
			Name:  "src",
			Path:  "src",
			IsDir: true,
			Children: []*FileNode{
				{
					Name:  "file.go",
					IsDir: false,
				},
				{
					Name:  "subdir",
					IsDir: true,
					Children: []*FileNode{
						{Name: "file2.go", IsDir: false},
					},
				},
			},
		}
		writeNode(&sb, node, "", true)
		output := sb.String()
		if !strings.Contains(output, "src") || !strings.Contains(output, "file.go") {
			t.Error("Expected output to contain nested structure")
		}
	})
}

func TestShouldSkipDir(t *testing.T) {
	testCases := []struct {
		name string
		want bool
	}{
		{"node_modules", true},
		{"vendor", true},
		{"build", true},
		{"dist", true},
		{"target", true},
		{"bin", true},
		{"obj", true},
		{".git", true},
		{".sentinel", true},
		{"__pycache__", true},
		{".next", true},
		{".nuxt", true},
		{"src", false},
		{"components", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDir(tc.name)
			if got != tc.want {
				t.Errorf("shouldSkipDir(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
