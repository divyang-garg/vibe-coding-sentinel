// Package services - Tests for architecture dependency detection
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"strings"
	"testing"
)

func TestDetectDependencyIssues_GodModule(t *testing.T) {
	lines := make([]string, 1001)
	for i := range lines {
		lines[i] = "package main"
	}
	files := []FileContent{
		{Path: "huge.go", Content: strings.Join(lines, "\n"), Language: "go"},
	}
	graph := ModuleGraph{
		Nodes: []ModuleNode{{Path: "huge.go", Lines: 1001}},
		Edges: nil,
	}
	issues := detectDependencyIssues(files, graph)
	var found bool
	for _, i := range issues {
		if i.Type == "god_module" && len(i.Files) == 1 && i.Files[0] == "huge.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected god_module issue for huge.go, got %v", issues)
	}
}

func TestDetectDependencyIssues_Circular(t *testing.T) {
	files := []FileContent{
		{Path: "a.go", Content: "package p\nfunc A() {}", Language: "go"},
		{Path: "b.go", Content: "package p\nfunc B() {}", Language: "go"},
	}
	graph := ModuleGraph{
		Nodes: []ModuleNode{{Path: "a.go"}, {Path: "b.go"}},
		Edges: []ModuleEdge{
			{From: "a.go", To: "b.go", Type: "import"},
			{From: "b.go", To: "a.go", Type: "import"},
		},
	}
	issues := detectDependencyIssues(files, graph)
	var found bool
	for _, i := range issues {
		if i.Type == "circular" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected circular issue, got %v", issues)
	}
}

func TestDetectDependencyIssues_TightCoupling(t *testing.T) {
	// MaxFanOut default 15; create a node with 16 edges
	edges := make([]ModuleEdge, 16)
	for i := range edges {
		edges[i] = ModuleEdge{From: "busy.go", To: "dep" + string(rune('a'+i)) + ".go", Type: "import"}
	}
	graph := ModuleGraph{
		Nodes: []ModuleNode{{Path: "busy.go"}},
		Edges: edges,
	}
	files := []FileContent{{Path: "busy.go", Content: "package p", Language: "go"}}
	issues := detectDependencyIssues(files, graph)
	var found bool
	for _, i := range issues {
		if i.Type == "tight_coupling" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected tight_coupling issue, got %v", issues)
	}
}

func TestDetectModuleType(t *testing.T) {
	tests := []struct{ path, want string }{
		{"foo/service.go", "service"},
		{"bar/util.go", "utility"},
		{"baz/helper.go", "utility"},
		{"ui/component.tsx", "component"},
		{"x_test.go", "test"},
		{"other.go", "module"},
	}
	for _, tt := range tests {
		got := detectModuleType(tt.path)
		if got != tt.want {
			t.Errorf("detectModuleType(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestFanOutByFile(t *testing.T) {
	graph := ModuleGraph{
		Edges: []ModuleEdge{
			{From: "a.go", To: "b.go", Type: "import"},
			{From: "a.go", To: "c.go", Type: "import"},
			{From: "b.go", To: "c.go", Type: "import"},
		},
	}
	out := fanOutByFile(graph)
	if out["a.go"] != 2 || out["b.go"] != 1 {
		t.Errorf("fanOutByFile: got %v", out)
	}
}
