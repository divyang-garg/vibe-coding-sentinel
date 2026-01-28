// Package services - Tests for architecture import resolution
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"path/filepath"
	"testing"
)

func TestResolveImport_GoSkipStdlib(t *testing.T) {
	paths := pathSet{"a.go": true, "b.go": true}
	out := resolveImport("a.go", "fmt", "go", paths)
	if len(out) != 0 {
		t.Errorf("stdlib fmt should not resolve: got %v", out)
	}
	out = resolveImport("a.go", "strings", "go", paths)
	if len(out) != 0 {
		t.Errorf("stdlib strings should not resolve: got %v", out)
	}
}

func TestResolveImport_GoSkipExternal(t *testing.T) {
	paths := pathSet{"a.go": true}
	out := resolveImport("a.go", "github.com/foo/bar", "go", paths)
	if len(out) != 0 {
		t.Errorf("external package should not resolve: got %v", out)
	}
}

func TestResolveImport_GoMatchLastSegment(t *testing.T) {
	paths := pathSet{
		filepath.Join("pkg", "services", "x.go"): true,
		filepath.Join("pkg", "services", "y.go"): true,
		"other.go":                               true,
	}
	out := resolveImport("a.go", "my/module/services", "go", paths)
	if len(out) != 2 {
		t.Errorf("expected 2 matches for services, got %d: %v", len(out), out)
	}
}

func TestResolveImport_JSRelative(t *testing.T) {
	paths := pathSet{
		"dir/utils.js": true,
		"dir/other.js": true,
	}
	out := resolveImport("dir/index.js", "./utils", "javascript", paths)
	if len(out) != 1 || (len(out) > 0 && out[0] != "dir/utils.js") {
		t.Errorf("expected [dir/utils.js], got %v", out)
	}
}

func TestNormalizeArchitectureLanguage(t *testing.T) {
	tests := []struct{ in, want string }{
		{"go", "go"},
		{"golang", "go"},
		{"js", "javascript"},
		{"ts", "typescript"},
		{"py", "python"},
	}
	for _, tt := range tests {
		got := normalizeArchitectureLanguage(tt.in)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
