// Package architecture_analysis - Main architecture analysis functions
// Complies with CODING_STANDARDS.md: Services max 400 lines

package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// AnalyzeArchitecture performs architecture analysis on the provided request.
// Exported for use by HTTP handlers. CodebasePath is optional and used for import resolution.
func AnalyzeArchitecture(req ArchitectureAnalysisRequest) ArchitectureAnalysisResponse {
	return analyzeArchitecture(req.Files, req.CodebasePath)
}

// analyzeArchitecture performs architecture analysis on provided files
func analyzeArchitecture(files []FileContent, codebasePath string) ArchitectureAnalysisResponse {
	var oversizedFiles []FileAnalysisResult
	var moduleGraph ModuleGraph
	var dependencyIssues []DependencyIssue
	var recommendations []string

	// Analyze each file
	for _, file := range files {
		lines := strings.Split(file.Content, "\n")
		lineCount := len(lines)
		cleanPath := filepath.Clean(file.Path)

		ac := GetArchitectureConfig()
		if lineCount >= ac.WarningLines {
			status := "warning"
			if lineCount >= ac.MaxLines {
				status = "oversized"
			} else if lineCount >= ac.CriticalLines {
				status = "critical"
			}

			// Detect sections using AST (if available)
			sections := detectSections(file.Content, file.Language)

			// Generate split suggestion
			var splitSuggestion *SplitSuggestion
			if lineCount >= ac.CriticalLines {
				splitSuggestion = generateSplitSuggestion(file, sections)
			}

			oversizedFiles = append(oversizedFiles, FileAnalysisResult{
				File:            cleanPath,
				Lines:           lineCount,
				Status:          status,
				Sections:        sections,
				SplitSuggestion: splitSuggestion,
			})
		}

		// Build module graph (use cleaned paths so Nodes match Edges)
		moduleGraph.Nodes = append(moduleGraph.Nodes, ModuleNode{
			Path:  cleanPath,
			Lines: lineCount,
			Type:  detectModuleType(file.Path),
		})
	}

	// Build module graph edges (extract deps, resolve, add edges)
	moduleGraph.Edges = buildModuleGraphEdges(files, codebasePath, moduleGraph.Nodes)

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
