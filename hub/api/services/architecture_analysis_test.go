// Package services - Tests for architecture analysis
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"strings"
	"testing"
)

func TestAnalyzeArchitecture_EmptyFiles(t *testing.T) {
	req := ArchitectureAnalysisRequest{Files: []FileContent{}}
	resp := AnalyzeArchitecture(req)
	if len(resp.ModuleGraph.Nodes) != 0 || len(resp.OversizedFiles) != 0 {
		t.Errorf("empty files: expected empty nodes/oversized, got nodes=%d oversized=%d",
			len(resp.ModuleGraph.Nodes), len(resp.OversizedFiles))
	}
	if len(resp.Recommendations) == 0 {
		t.Error("expected at least one recommendation")
	}
}

func TestAnalyzeArchitecture_SmallFile(t *testing.T) {
	req := ArchitectureAnalysisRequest{
		Files: []FileContent{
			{Path: "a.go", Content: "package p\n\nfunc F() {}", Language: "go"},
		},
	}
	resp := AnalyzeArchitecture(req)
	if len(resp.ModuleGraph.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.ModuleGraph.Nodes))
	}
	if resp.ModuleGraph.Nodes[0].Path != "a.go" || resp.ModuleGraph.Nodes[0].Lines != 3 {
		t.Errorf("node: path=%s lines=%d", resp.ModuleGraph.Nodes[0].Path, resp.ModuleGraph.Nodes[0].Lines)
	}
	if len(resp.OversizedFiles) != 0 {
		t.Errorf("expected no oversized, got %d", len(resp.OversizedFiles))
	}
}

func TestAnalyzeArchitecture_Oversized(t *testing.T) {
	ac := GetArchitectureConfig()
	lines := make([]string, ac.MaxLines+1)
	for i := range lines {
		lines[i] = "package p"
	}
	req := ArchitectureAnalysisRequest{
		Files: []FileContent{
			{Path: "big.go", Content: strings.Join(lines, "\n"), Language: "go"},
		},
	}
	resp := AnalyzeArchitecture(req)
	if len(resp.OversizedFiles) != 1 {
		t.Fatalf("expected 1 oversized, got %d", len(resp.OversizedFiles))
	}
	if resp.OversizedFiles[0].Status != "oversized" {
		t.Errorf("expected status oversized, got %s", resp.OversizedFiles[0].Status)
	}
}

func TestAnalyzeArchitecture_WithCodebasePath(t *testing.T) {
	req := ArchitectureAnalysisRequest{
		Files: []FileContent{
			{Path: "x.go", Content: "package p\nfunc F() {}", Language: "go"},
		},
		CodebasePath: "/some/root",
	}
	resp := AnalyzeArchitecture(req)
	if len(resp.ModuleGraph.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(resp.ModuleGraph.Nodes))
	}
}
