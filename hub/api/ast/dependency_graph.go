// Package ast provides dependency graph for cross-file analysis
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

// Dependency represents a dependency relationship between files
type Dependency struct {
	FromFile string
	ToFile   string
	Type     string // "import", "require", "include"
	Line     int
	Column   int
}

// DependencyGraph manages file dependencies
type DependencyGraph struct {
	dependencies map[string][]*Dependency // file -> dependencies
	reverse      map[string][]string      // file -> files that depend on it
	mu           sync.RWMutex
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[string][]*Dependency),
		reverse:      make(map[string][]string),
	}
}

// AddDependency adds a dependency relationship
func (dg *DependencyGraph) AddDependency(dep *Dependency) error {
	if dep == nil {
		return fmt.Errorf("dependency cannot be nil")
	}
	if dep.FromFile == "" || dep.ToFile == "" {
		return fmt.Errorf("dependency must have both from and to files")
	}

	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Add to forward dependencies
	dg.dependencies[dep.FromFile] = append(dg.dependencies[dep.FromFile], dep)

	// Add to reverse dependencies
	if !containsString(dg.reverse[dep.ToFile], dep.FromFile) {
		dg.reverse[dep.ToFile] = append(dg.reverse[dep.ToFile], dep.FromFile)
	}

	return nil
}

// GetDependencies returns all dependencies for a file
func (dg *DependencyGraph) GetDependencies(filePath string) []*Dependency {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	return dg.dependencies[filePath]
}

// GetDependents returns all files that depend on the given file
func (dg *DependencyGraph) GetDependents(filePath string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	return dg.reverse[filePath]
}

// FindCircularDependencies finds circular dependency chains
func (dg *DependencyGraph) FindCircularDependencies() [][]string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cycles := [][]string{}

	for file := range dg.dependencies {
		if !visited[file] {
			cycle := dg.dfsFindCycle(file, visited, recStack, []string{})
			if len(cycle) > 0 {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

// dfsFindCycle performs DFS to find cycles
func (dg *DependencyGraph) dfsFindCycle(file string, visited, recStack map[string]bool, path []string) []string {
	visited[file] = true
	recStack[file] = true
	path = append(path, file)

	for _, dep := range dg.dependencies[file] {
		nextFile := dep.ToFile
		if !visited[nextFile] {
			cycle := dg.dfsFindCycle(nextFile, visited, recStack, path)
			if len(cycle) > 0 {
				return cycle
			}
		} else if recStack[nextFile] {
			// Found a cycle
			cycleStart := indexOf(path, nextFile)
			if cycleStart >= 0 {
				return path[cycleStart:]
			}
		}
	}

	recStack[file] = false
	return nil
}

// ExtractDependenciesFromFile extracts import/require statements from a file
func ExtractDependenciesFromFile(rootNode *sitter.Node, code, filePath, language string) ([]*Dependency, error) {
	if rootNode == nil {
		return nil, fmt.Errorf("root node cannot be nil")
	}

	dependencies := []*Dependency{}
	visitor := func(node *sitter.Node) bool {
		var dep *Dependency

		switch language {
		case "go":
			dep = extractGoImport(node, code, filePath)
		case "javascript", "typescript":
			dep = extractJSImport(node, code, filePath)
		case "python":
			dep = extractPythonImport(node, code, filePath)
		}

		if dep != nil {
			dependencies = append(dependencies, dep)
		}

		return true
	}

	TraverseAST(rootNode, visitor)
	return dependencies, nil
}

// extractGoImport extracts import statements from Go code
func extractGoImport(node *sitter.Node, code, filePath string) *Dependency {
	if node.Type() != "import_declaration" && node.Type() != "import_spec_list" {
		return nil
	}

	// Find import path
	var importPath string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "interpreted_string_literal" || child.Type() == "raw_string_literal" {
			importPath = safeSlice(code, child.StartByte(), child.EndByte())
			// Remove quotes
			if len(importPath) >= 2 {
				importPath = importPath[1 : len(importPath)-1]
			}
			break
		}
	}

	if importPath == "" {
		return nil
	}

	line, col := getLineColumn(code, int(node.StartByte()))
	return &Dependency{
		FromFile: filePath,
		ToFile:   importPath, // In Go, this is a package path
		Type:     "import",
		Line:     line,
		Column:   col,
	}
}

// extractJSImport extracts import/require statements from JavaScript/TypeScript
func extractJSImport(node *sitter.Node, code, filePath string) *Dependency {
	nodeType := node.Type()
	if nodeType != "import_statement" && nodeType != "call_expression" {
		return nil
	}

	// Handle import statement
	if nodeType == "import_statement" {
		var importPath string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if child.Type() == "string" {
				importPath = safeSlice(code, child.StartByte(), child.EndByte())
				if len(importPath) >= 2 {
					importPath = importPath[1 : len(importPath)-1]
				}
				break
			}
		}

		if importPath == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &Dependency{
			FromFile: filePath,
			ToFile:   importPath,
			Type:     "import",
			Line:     line,
			Column:   col,
		}
	}

	// Handle require() call
	if nodeType == "call_expression" {
		// Check if it's a require call
		firstChild := node.Child(0)
		if firstChild == nil || firstChild.Type() != "identifier" {
			return nil
		}
		funcName := safeSlice(code, firstChild.StartByte(), firstChild.EndByte())
		if funcName != "require" {
			return nil
		}

		// Get the argument (module path)
		var importPath string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child == nil {
				continue
			}
			if child.Type() == "arguments" {
				for j := 0; j < int(child.ChildCount()); j++ {
					arg := child.Child(j)
					if arg != nil && arg.Type() == "string" {
						importPath = safeSlice(code, arg.StartByte(), arg.EndByte())
						if len(importPath) >= 2 {
							importPath = importPath[1 : len(importPath)-1]
						}
						break
					}
				}
			}
		}

		if importPath == "" {
			return nil
		}

		line, col := getLineColumn(code, int(node.StartByte()))
		return &Dependency{
			FromFile: filePath,
			ToFile:   importPath,
			Type:     "require",
			Line:     line,
			Column:   col,
		}
	}

	return nil
}

// extractPythonImport extracts import statements from Python code
func extractPythonImport(node *sitter.Node, code, filePath string) *Dependency {
	if node.Type() != "import_statement" && node.Type() != "import_from_statement" {
		return nil
	}

	var importPath string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "dotted_name" || child.Type() == "relative_import" {
			importPath = safeSlice(code, child.StartByte(), child.EndByte())
			break
		}
	}

	if importPath == "" {
		return nil
	}

	line, col := getLineColumn(code, int(node.StartByte()))
	return &Dependency{
		FromFile: filePath,
		ToFile:   importPath,
		Type:     "import",
		Line:     line,
		Column:   col,
	}
}

// Helper functions
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
