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
	if err == nil && parser != nil {
		ctx := context.Background()
		tree, parseErr := parser.ParseCtx(ctx, nil, []byte(content))
		if parseErr == nil && tree != nil {
			defer tree.Close()
			rootNode := tree.RootNode()
			if rootNode != nil {
				// Use AST to detect functions/classes
				sections = extractSectionsFromAST(rootNode, content, language)
			}
		}
	}

	// Fallback to pattern-based detection if AST fails
	if len(sections) == 0 {
		sections = detectSectionsPattern(content, language)
	}

	return sections
}

// extractSectionsFromAST extracts sections from AST tree
func extractSectionsFromAST(node *sitter.Node, content string, language string) []FileSection {
	var sections []FileSection

	// Traverse AST to find function/class definitions
	TraverseAST(node, func(n *sitter.Node) bool {
		nodeType := n.Type()
		isSection := false

		switch language {
		case "go":
			isSection = nodeType == "function_declaration" || nodeType == "method_declaration" || nodeType == "type_declaration"
		case "javascript", "typescript":
			isSection = nodeType == "function_declaration" || nodeType == "class_declaration"
		case "python":
			isSection = nodeType == "function_definition" || nodeType == "class_definition"
		}

		if isSection {
			startPoint := n.StartPoint()
			endPoint := n.EndPoint()
			startLine := int(startPoint.Row) + 1
			endLine := int(endPoint.Row) + 1
			lines := endLine - startLine + 1

			// Extract name (simplified - extract first identifier)
			name := "unknown"
			for i := 0; i < int(n.ChildCount()); i++ {
				child := n.Child(i)
				if child != nil && (child.Type() == "identifier" || child.Type() == "field_identifier" || child.Type() == "property_identifier") {
					name = content[child.StartByte():child.EndByte()]
					break
				}
			}

			sections = append(sections, FileSection{
				StartLine:   startLine,
				EndLine:     endLine,
				Name:        name,
				Description: fmt.Sprintf("%s definition", nodeType),
				Lines:       lines,
			})
		}

		return true
	})

	return sections
}

// extractSectionsFromAST extracts sections from AST tree
// NOTE: Disabled until tree-sitter integration is complete
// func extractSectionsFromAST(node *sitter.Node, content string) []FileSection {
// 	var sections []FileSection
//
// 	// Traverse AST to find function/class definitions
// 	traverseASTForSections(node, func(n *sitter.Node) {
// 		nodeType := n.Type()
// 		if nodeType == "function_declaration" || nodeType == "method_declaration" ||
// 			nodeType == "class_declaration" || nodeType == "type_declaration" {
//
// 			startLine := int(n.StartPoint().Row) + 1
// 			endLine := int(n.EndPoint().Row) + 1
// 			lines := endLine - startLine + 1
//
// 			// Extract name
// 			name := extractNodeName(n, content)
//
// 			sections = append(sections, FileSection{
// 				StartLine:   startLine,
// 				EndLine:     endLine,
// 				Name:        name,
// 				Description: fmt.Sprintf("%s definition", nodeType),
// 				Lines:       lines,
// 			})
// 		}
// 	})
//
// 	return sections
// }

// Note: traverseASTForSections and extractNodeName are defined in architecture_sections.go
