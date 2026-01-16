// Package architecture_analysis - Main architecture analysis functions
// Complies with CODING_STANDARDS.md: Services max 400 lines

package services

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

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
	parser, err := getParser(language)
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

	// Traverse AST to find function/class definitions
	traverseASTForSections(node, func(n *sitter.Node) {
		nodeType := n.Type()
		if nodeType == "function_declaration" || nodeType == "method_declaration" ||
			nodeType == "class_declaration" || nodeType == "type_declaration" {

			startLine := int(n.StartPoint().Row) + 1
			endLine := int(n.EndPoint().Row) + 1
			lines := endLine - startLine + 1

			// Extract name
			name := extractNodeName(n, content)

			sections = append(sections, FileSection{
				StartLine:   startLine,
				EndLine:     endLine,
				Name:        name,
				Description: fmt.Sprintf("%s definition", nodeType),
				Lines:       lines,
			})
		}
	})

	return sections
}
