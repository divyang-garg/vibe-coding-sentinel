// Architecture Analyzer - Phase 9 Implementation
// Analyzes file structure and suggests splits for oversized files

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"
	sitter "github.com/smacker/go-tree-sitter"
)

// FileContent represents a file to be analyzed
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

// ArchitectureAnalysisRequest represents a request for architecture analysis
type ArchitectureAnalysisRequest struct {
	Files []FileContent `json:"files"`
}

// ArchitectureAnalysisResponse represents the response from architecture analysis
type ArchitectureAnalysisResponse struct {
	OversizedFiles   []FileAnalysisResult `json:"oversizedFiles"`
	ModuleGraph      ModuleGraph          `json:"moduleGraph"`
	DependencyIssues []DependencyIssue    `json:"dependencyIssues"`
	Recommendations  []string             `json:"recommendations"`
}

// FileAnalysisResult represents analysis result for a single file
type FileAnalysisResult struct {
	File            string           `json:"file"`
	Lines           int              `json:"lines"`
	Status          string           `json:"status"` // ok, warning, critical, oversized
	Sections        []FileSection    `json:"sections,omitempty"`
	SplitSuggestion *SplitSuggestion `json:"splitSuggestion,omitempty"`
}

// FileSection represents a logical section within a file
type FileSection struct {
	StartLine   int    `json:"startLine"`
	EndLine     int    `json:"endLine"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lines       int    `json:"lines"`
}

// SplitSuggestion represents a suggestion for splitting a file
type SplitSuggestion struct {
	Reason                string         `json:"reason"`
	ProposedFiles         []ProposedFile `json:"proposedFiles"`
	MigrationInstructions []string       `json:"migrationInstructions"` // Text instructions only, not executable
	EstimatedEffort       string         `json:"estimatedEffort"`
}

// ProposedFile represents a proposed file in a split suggestion
type ProposedFile struct {
	Path     string   `json:"path"`
	Lines    int      `json:"lines"`
	Contents []string `json:"contents"` // Function/class names to move
}

// ModuleGraph represents the module dependency graph
type ModuleGraph struct {
	Nodes []ModuleNode `json:"nodes"`
	Edges []ModuleEdge `json:"edges"`
}

// ModuleNode represents a node in the module graph
type ModuleNode struct {
	Path  string `json:"path"`
	Lines int    `json:"lines"`
	Type  string `json:"type"` // component, service, utility, etc.
}

// ModuleEdge represents an edge in the module graph
type ModuleEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // import, extends, implements
}

// DependencyIssue represents a dependency issue found in the codebase
type DependencyIssue struct {
	Type        string   `json:"type"` // circular, tight_coupling, god_module
	Severity    string   `json:"severity"`
	Files       []string `json:"files"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion"`
}

// analyzeArchitecture performs architecture analysis on provided files
func analyzeArchitecture(files []FileContent) ArchitectureAnalysisResponse {
	var oversizedFiles []FileAnalysisResult
	var moduleGraph ModuleGraph
	var dependencyIssues []DependencyIssue
	var recommendations []string

	// Analyze each file
	for _, file := range files {
		lines := strings.Split(file.Content, "\n")
		lineCount := len(lines)

		// Check if file is oversized (using default thresholds for now)
		// In full implementation, these would come from config
		warningThreshold := 300
		criticalThreshold := 500
		maxThreshold := 1000

		if lineCount >= warningThreshold {
			status := "warning"
			if lineCount >= maxThreshold {
				status = "oversized"
			} else if lineCount >= criticalThreshold {
				status = "critical"
			}

			// Detect sections using AST (if available)
			sections := detectSections(file.Content, file.Language)

			// Generate split suggestion
			var splitSuggestion *SplitSuggestion
			if lineCount >= criticalThreshold {
				splitSuggestion = generateSplitSuggestion(file, sections)
			}

			oversizedFiles = append(oversizedFiles, FileAnalysisResult{
				File:            file.Path,
				Lines:           lineCount,
				Status:          status,
				Sections:        sections,
				SplitSuggestion: splitSuggestion,
			})
		}

		// Build module graph
		moduleGraph.Nodes = append(moduleGraph.Nodes, ModuleNode{
			Path:  file.Path,
			Lines: lineCount,
			Type:  detectModuleType(file.Path),
		})
	}

	// Detect dependency issues
	dependencyIssues = detectDependencyIssues(files, moduleGraph)

	// Generate recommendations
	recommendations = generateRecommendations(oversizedFiles, dependencyIssues)

	return ArchitectureAnalysisResponse{
		OversizedFiles:   oversizedFiles,
		ModuleGraph:      moduleGraph,
		DependencyIssues: dependencyIssues,
		Recommendations:  recommendations,
	}
}

// detectSections detects logical sections within a file using AST
func detectSections(content string, language string) []FileSection {
	var sections []FileSection

	// Try to use AST parser if available
	parser, err := ast.GetParser(language)
	if err == nil {
		// Use AST to detect functions/classes
		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, []byte(content))
		if err == nil {
			defer tree.Close()
			sections = extractSectionsFromAST(tree.RootNode(), content)
		}
	}

	// Fallback to pattern-based detection if AST fails
	if len(sections) == 0 {
		sections = detectSectionsPattern(content, language)
	}

	return sections
}

// extractSectionsFromAST extracts sections from AST tree
func extractSectionsFromAST(node *sitter.Node, content string) []FileSection {
	var sections []FileSection

	// Traverse AST to find function/class declarations
	ast.TraverseAST(node, func(n *sitter.Node) bool {
		nodeType := n.Type()
		startPoint := n.StartPoint()
		endPoint := n.EndPoint()

		// Detect function/class declarations based on node type
		if isFunctionOrClassNode(nodeType) {
			sectionName := extractNodeName(n, content)
			if sectionName != "" {
				sections = append(sections, FileSection{
					StartLine:   int(startPoint.Row) + 1, // Convert from 0-based to 1-based
					EndLine:     int(endPoint.Row) + 1,
					Name:        sectionName,
					Description: getNodeDescription(nodeType),
					Lines:       int(endPoint.Row-startPoint.Row) + 1,
				})
			}
		}
		return true // Continue traversal
	})

	return sections
}

// isFunctionOrClassNode checks if a node type represents a function or class declaration
func isFunctionOrClassNode(nodeType string) bool {
	// Common function/class node types across languages
	functionTypes := []string{
		"function_declaration", "function_definition", "method_declaration",
		"class_declaration", "type_declaration", "struct_declaration",
		"func", "def", "class", "type",
	}
	for _, ft := range functionTypes {
		if nodeType == ft || strings.Contains(nodeType, ft) {
			return true
		}
	}
	return false
}

// getNodeDescription returns a human-readable description for a node type
func getNodeDescription(nodeType string) string {
	if strings.Contains(nodeType, "function") || strings.Contains(nodeType, "func") || strings.Contains(nodeType, "def") {
		return "Function definition"
	}
	if strings.Contains(nodeType, "class") || strings.Contains(nodeType, "type") || strings.Contains(nodeType, "struct") {
		return "Type/Class definition"
	}
	return "Code section"
}

// extractNodeName extracts the name from an AST node
func extractNodeName(node *sitter.Node, content string) string {
	// Try to find identifier child node
	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		childType := child.Type()
		// Identifier nodes typically contain the name
		if childType == "identifier" || childType == "type_identifier" || childType == "field_identifier" {
			startByte := int(child.StartByte())
			endByte := int(child.EndByte())
			if startByte < len(content) && endByte <= len(content) {
				return strings.TrimSpace(content[startByte:endByte])
			}
		}
		// For some languages, the name might be in the second child
		if i == 1 && (childType == "identifier" || childType == "type_identifier") {
			startByte := int(child.StartByte())
			endByte := int(child.EndByte())
			if startByte < len(content) && endByte <= len(content) {
				return strings.TrimSpace(content[startByte:endByte])
			}
		}
	}
	// Fallback: extract from first line of node
	startPoint := node.StartPoint()
	if int(startPoint.Row) < len(strings.Split(content, "\n")) {
		lines := strings.Split(content, "\n")
		line := lines[int(startPoint.Row)]
		// Try to extract name from line
		return extractFunctionName(line, "")
	}
	return "unknown"
}

// detectSectionsPattern detects sections using pattern matching (fallback)
func detectSectionsPattern(content string, language string) []FileSection {
	var sections []FileSection
	lines := strings.Split(content, "\n")

	// Pattern-based detection for common languages
	switch language {
	case "go", "golang":
		sections = detectGoSections(lines)
	case "javascript", "typescript", "js", "ts", "jsx", "tsx":
		sections = detectJSSections(lines)
	case "python", "py":
		sections = detectPythonSections(lines)
	default:
		// Generic detection
		sections = detectGenericSections(lines)
	}

	return sections
}

// detectGoSections detects sections in Go files
func detectGoSections(lines []string) []FileSection {
	var sections []FileSection
	currentSection := FileSection{StartLine: 1, Name: "package", Description: "Package declaration"}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Function declaration
		if strings.HasPrefix(trimmed, "func ") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			// Extract function name
			funcName := extractFunctionName(trimmed, "func ")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        funcName,
				Description: "Function definition",
			}
		}
		// Type declaration
		if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, "struct") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			typeName := extractFunctionName(trimmed, "type ")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        typeName,
				Description: "Type definition",
			}
		}
	}

	// Add last section
	if currentSection.StartLine > 0 {
		currentSection.EndLine = len(lines)
		currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
		sections = append(sections, currentSection)
	}

	return sections
}

// detectJSSections detects sections in JavaScript/TypeScript files
func detectJSSections(lines []string) []FileSection {
	var sections []FileSection
	currentSection := FileSection{StartLine: 1, Name: "imports", Description: "Import statements"}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Function declaration
		if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, "= (") ||
			strings.HasPrefix(trimmed, "export function ") || strings.HasPrefix(trimmed, "export const ") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			funcName := extractFunctionName(trimmed, "")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        funcName,
				Description: "Function/const definition",
			}
		}
		// Class declaration
		if strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "export class ") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			className := extractFunctionName(trimmed, "class ")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        className,
				Description: "Class definition",
			}
		}
	}

	// Add last section
	if currentSection.StartLine > 0 {
		currentSection.EndLine = len(lines)
		currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
		sections = append(sections, currentSection)
	}

	return sections
}

// detectPythonSections detects sections in Python files
func detectPythonSections(lines []string) []FileSection {
	var sections []FileSection
	currentSection := FileSection{StartLine: 1, Name: "imports", Description: "Import statements"}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Function definition
		if strings.HasPrefix(trimmed, "def ") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			funcName := extractFunctionName(trimmed, "def ")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        funcName,
				Description: "Function definition",
			}
		}
		// Class definition
		if strings.HasPrefix(trimmed, "class ") {
			if currentSection.StartLine > 0 {
				currentSection.EndLine = i
				currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
				sections = append(sections, currentSection)
			}

			className := extractFunctionName(trimmed, "class ")
			currentSection = FileSection{
				StartLine:   i + 1,
				Name:        className,
				Description: "Class definition",
			}
		}
	}

	// Add last section
	if currentSection.StartLine > 0 {
		currentSection.EndLine = len(lines)
		currentSection.Lines = currentSection.EndLine - currentSection.StartLine + 1
		sections = append(sections, currentSection)
	}

	return sections
}

// detectGenericSections detects sections using generic patterns
func detectGenericSections(lines []string) []FileSection {
	var sections []FileSection
	// Simple heuristic: split into chunks of ~200 lines
	chunkSize := 200
	for i := 0; i < len(lines); i += chunkSize {
		end := i + chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		sections = append(sections, FileSection{
			StartLine:   i + 1,
			EndLine:     end,
			Name:        fmt.Sprintf("Section %d", len(sections)+1),
			Description: fmt.Sprintf("Lines %d-%d", i+1, end),
			Lines:       end - i,
		})
	}
	return sections
}

// extractFunctionName extracts function/class name from a line
func extractFunctionName(line string, prefix string) string {
	// Remove prefix
	if prefix != "" {
		line = strings.TrimPrefix(line, prefix)
	}

	// Extract name (first word after prefix)
	parts := strings.Fields(line)
	if len(parts) > 0 {
		name := parts[0]
		// Remove parentheses, brackets, etc.
		name = strings.Trim(name, "()[]{}=:")
		return name
	}
	return "unknown"
}

// generateSplitSuggestion generates a split suggestion for an oversized file
func generateSplitSuggestion(file FileContent, sections []FileSection) *SplitSuggestion {
	if len(sections) < 2 {
		return nil // Not enough sections to split
	}

	var proposedFiles []ProposedFile
	var migrationInstructions []string

	// Group sections into logical files
	// Simple strategy: split into ~3-4 files of roughly equal size
	targetFiles := 3
	if len(sections) > 6 {
		targetFiles = 4
	}

	sectionsPerFile := len(sections) / targetFiles
	if sectionsPerFile < 1 {
		sectionsPerFile = 1
	}

	basePath := file.Path
	ext := filepath.Ext(basePath)
	baseName := strings.TrimSuffix(filepath.Base(basePath), ext)
	dir := filepath.Dir(basePath)

	for i := 0; i < targetFiles; i++ {
		startIdx := i * sectionsPerFile
		endIdx := startIdx + sectionsPerFile
		if i == targetFiles-1 {
			endIdx = len(sections) // Last file gets remaining sections
		}

		if startIdx >= len(sections) {
			break
		}

		var sectionNames []string
		for j := startIdx; j < endIdx && j < len(sections); j++ {
			sectionNames = append(sectionNames, sections[j].Name)
		}

		fileNum := i + 1
		newPath := filepath.Join(dir, fmt.Sprintf("%s_part%d%s", baseName, fileNum, ext))

		proposedFiles = append(proposedFiles, ProposedFile{
			Path:     newPath,
			Lines:    calculateTotalLines(sections[startIdx:min(endIdx, len(sections))]),
			Contents: sectionNames,
		})
	}

	// Generate migration instructions
	migrationInstructions = []string{
		fmt.Sprintf("1. Create new files: %s", strings.Join(getProposedPaths(proposedFiles), ", ")),
		"2. Move sections to respective files as suggested",
		"3. Update imports in files that reference moved functions/classes",
		"4. Create index file to re-export public APIs if needed",
		"5. Update tests to reference new file locations",
		"6. Run tests to verify functionality",
		"7. Remove original file after verification",
	}

	estimatedEffort := "Medium"
	if len(sections) > 10 {
		estimatedEffort = "High"
	}

	return &SplitSuggestion{
		Reason:                fmt.Sprintf("File has %d lines and %d logical sections. Splitting into %d files will improve maintainability.", len(strings.Split(file.Content, "\n")), len(sections), len(proposedFiles)),
		ProposedFiles:         proposedFiles,
		MigrationInstructions: migrationInstructions,
		EstimatedEffort:       estimatedEffort,
	}
}

// Helper functions
func getProposedPaths(files []ProposedFile) []string {
	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return paths
}

func calculateTotalLines(sections []FileSection) int {
	total := 0
	for _, s := range sections {
		total += s.Lines
	}
	return total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// detectModuleType detects the type of module based on file path
func detectModuleType(path string) string {
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, "component") {
		return "component"
	}
	if strings.Contains(pathLower, "service") {
		return "service"
	}
	if strings.Contains(pathLower, "util") || strings.Contains(pathLower, "helper") {
		return "utility"
	}
	if strings.Contains(pathLower, "test") {
		return "test"
	}
	return "module"
}

// detectDependencyIssues detects dependency issues in the codebase
func detectDependencyIssues(files []FileContent, graph ModuleGraph) []DependencyIssue {
	var issues []DependencyIssue

	// Simple detection: look for circular imports (basic pattern matching)
	// Full implementation would build actual dependency graph
	for _, file := range files {
		// Check for very large files (potential god module)
		lines := strings.Split(file.Content, "\n")
		if len(lines) > 1000 {
			issues = append(issues, DependencyIssue{
				Type:        "god_module",
				Severity:    "high",
				Files:       []string{file.Path},
				Description: fmt.Sprintf("File %s is very large (%d lines), indicating it may be doing too much", file.Path, len(lines)),
				Suggestion:  "Consider splitting into smaller, focused modules",
			})
		}
	}

	return issues
}

// generateRecommendations generates recommendations based on analysis
func generateRecommendations(oversizedFiles []FileAnalysisResult, issues []DependencyIssue) []string {
	var recommendations []string

	if len(oversizedFiles) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d oversized files. Consider refactoring to improve maintainability.", len(oversizedFiles)))
	}

	if len(issues) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d dependency issues. Review and refactor to reduce coupling.", len(issues)))
	}

	if len(oversizedFiles) == 0 && len(issues) == 0 {
		recommendations = append(recommendations, "No major architecture issues detected. Keep file sizes manageable as codebase grows.")
	}

	return recommendations
}
