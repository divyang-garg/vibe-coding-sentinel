// Package services - Tests for architecture cycle detection
// Complies with CODING_STANDARDS.md: Tests max 500 lines

package services

import (
	"testing"
)

func TestFindCyclesInModuleGraph_NoEdges(t *testing.T) {
	g := ModuleGraph{
		Nodes: []ModuleNode{{Path: "a.go", Lines: 10}, {Path: "b.go", Lines: 20}},
		Edges: nil,
	}
	cycles := findCyclesInModuleGraph(g)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %d", len(cycles))
	}
}

func TestFindCyclesInModuleGraph_NoCycle(t *testing.T) {
	g := ModuleGraph{
		Nodes: []ModuleNode{{Path: "a.go"}, {Path: "b.go"}, {Path: "c.go"}},
		Edges: []ModuleEdge{
			{From: "a.go", To: "b.go", Type: "import"},
			{From: "b.go", To: "c.go", Type: "import"},
		},
	}
	cycles := findCyclesInModuleGraph(g)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %d: %v", len(cycles), cycles)
	}
}

func TestFindCyclesInModuleGraph_SimpleCycle(t *testing.T) {
	g := ModuleGraph{
		Nodes: []ModuleNode{{Path: "a.go"}, {Path: "b.go"}},
		Edges: []ModuleEdge{
			{From: "a.go", To: "b.go", Type: "import"},
			{From: "b.go", To: "a.go", Type: "import"},
		},
	}
	cycles := findCyclesInModuleGraph(g)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	if len(cycles[0]) != 2 {
		t.Errorf("expected cycle length 2, got %d", len(cycles[0]))
	}
}

func TestFindCyclesInModuleGraph_ThreeCycle(t *testing.T) {
	g := ModuleGraph{
		Nodes: []ModuleNode{{Path: "a.go"}, {Path: "b.go"}, {Path: "c.go"}},
		Edges: []ModuleEdge{
			{From: "a.go", To: "b.go", Type: "import"},
			{From: "b.go", To: "c.go", Type: "import"},
			{From: "c.go", To: "a.go", Type: "import"},
		},
	}
	cycles := findCyclesInModuleGraph(g)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d: %v", len(cycles), cycles)
	}
	if len(cycles[0]) != 3 {
		t.Errorf("expected cycle length 3, got %d", len(cycles[0]))
	}
}

func TestCycleKey(t *testing.T) {
	k1 := cycleKey([]string{"a", "b", "c"})
	k2 := cycleKey([]string{"b", "c", "a"})
	k3 := cycleKey([]string{"c", "a", "b"})
	if k1 != k2 || k2 != k3 {
		t.Errorf("same cycle should produce same key: %q %q %q", k1, k2, k3)
	}
}
