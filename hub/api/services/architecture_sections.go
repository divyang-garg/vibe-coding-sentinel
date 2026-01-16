// Package architecture_sections - Section detection functions for architecture analysis
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"strings"

	"github.com/smacker/go-tree-sitter"
)

// traverseASTForSections traverses AST tree and calls callback for each node (for section detection)
func traverseASTForSections(node *sitter.Node, callback func(*sitter.Node)) {
	callback(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			traverseASTForSections(child, callback)
		}
	}
}

// extractNodeName extracts the name from an AST node
func extractNodeName(node *sitter.Node, content string) string {
	// Look for name field in node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && (child.Type() == "identifier" || child.Type() == "type_identifier") {
			start := int(child.StartByte())
			end := int(child.EndByte())
			if start < len(content) && end <= len(content) {
				return strings.TrimSpace(content[start:end])
			}
		}
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
