// Package services provides documentation extraction functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"regexp"
	"strings"

	"sentinel-hub-api/ast"
)

// extractDocumentation extracts documentation from code using AST analysis
// Returns structured documentation including functions, classes, modules, and packages
func (s *CodeAnalysisServiceImpl) extractDocumentation(code, language string) map[string]interface{} {
	if code == "" {
		return map[string]interface{}{
			"functions": []interface{}{},
			"classes":   []interface{}{},
			"modules":   []string{},
			"packages": []string{},
		}
	}

	// Extract functions using AST
	functions, err := ast.ExtractFunctions(code, language, "")
	if err != nil {
		// Return empty structure on error for backward compatibility
		return map[string]interface{}{
			"functions": []interface{}{},
			"classes":   []interface{}{},
			"modules":   []string{},
			"packages": []string{},
		}
	}

	// Convert AST FunctionInfo to FunctionDoc format
	var functionDocs []map[string]interface{}
	var modules []string
	var packages []string
	moduleMap := make(map[string]bool)
	packageMap := make(map[string]bool)

	for _, fn := range functions {
		// Extract parameter names
		params := make([]string, 0, len(fn.Parameters))
		for _, param := range fn.Parameters {
			params = append(params, param.Name)
		}

		doc := map[string]interface{}{
			"name":         fn.Name,
			"line":         fn.Line,
			"column":       fn.Column,
			"parameters":   params,
			"returnType":   fn.ReturnType,
			"documentation": fn.Documentation,
			"visibility":   fn.Visibility,
		}

		functionDocs = append(functionDocs, doc)

		// Extract module/package information from metadata or code structure
		if fn.Metadata != nil {
			if module, ok := fn.Metadata["module"]; ok && module != "" {
				moduleMap[module] = true
			}
			if pkg, ok := fn.Metadata["package"]; ok && pkg != "" {
				packageMap[pkg] = true
			}
		}
	}

	// Extract modules and packages from code structure
	s.extractModulesAndPackages(code, language, &modules, &packages, moduleMap, packageMap)

	// Extract classes (language-specific)
	classes := s.extractClasses(code, language)

	return map[string]interface{}{
		"functions": functionDocs,
		"classes":   classes,
		"modules":   modules,
		"packages":  packages,
	}
}

// extractModulesAndPackages extracts module and package information from code
func (s *CodeAnalysisServiceImpl) extractModulesAndPackages(code, language string, modules *[]string, packages *[]string, moduleMap, packageMap map[string]bool) {
	lines := strings.Split(code, "\n")

	switch language {
	case "go":
		// Extract package declarations
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "package ") {
				pkg := strings.TrimSpace(strings.TrimPrefix(line, "package "))
				if pkg != "" && !packageMap[pkg] {
					*packages = append(*packages, pkg)
					packageMap[pkg] = true
				}
			}
		}
	case "javascript", "typescript":
		// Extract module declarations (import/export)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "export ") {
				// Extract module name from import/export statements
				re := regexp.MustCompile(`from\s+['"]([^'"]+)['"]`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 && !moduleMap[matches[1]] {
					*modules = append(*modules, matches[1])
					moduleMap[matches[1]] = true
				}
			}
		}
	case "python":
		// Extract module names from imports
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					module := parts[1]
					if strings.Contains(module, ".") {
						module = strings.Split(module, ".")[0]
					}
					if module != "" && !moduleMap[module] {
						*modules = append(*modules, module)
						moduleMap[module] = true
					}
				}
			}
		}
	}
}

// extractClasses extracts class definitions from code (language-specific)
func (s *CodeAnalysisServiceImpl) extractClasses(code, language string) []map[string]interface{} {
	var classes []map[string]interface{}
	lines := strings.Split(code, "\n")

	switch language {
	case "go":
		// Go structs (treated as classes)
		structRe := regexp.MustCompile(`type\s+(\w+)\s+struct`)
		for i, line := range lines {
			matches := structRe.FindStringSubmatch(line)
			if len(matches) > 1 {
				classes = append(classes, map[string]interface{}{
					"name": matches[1],
					"line": i + 1,
					"type": "struct",
				})
			}
		}
	case "javascript", "typescript":
		// JavaScript/TypeScript classes
		classRe := regexp.MustCompile(`(?:class|interface)\s+(\w+)`)
		for i, line := range lines {
			matches := classRe.FindStringSubmatch(line)
			if len(matches) > 1 {
				classType := "class"
				if strings.Contains(line, "interface") {
					classType = "interface"
				}
				classes = append(classes, map[string]interface{}{
					"name": matches[1],
					"line": i + 1,
					"type": classType,
				})
			}
		}
	case "python":
		// Python classes
		classRe := regexp.MustCompile(`class\s+(\w+)`)
		for i, line := range lines {
			matches := classRe.FindStringSubmatch(line)
			if len(matches) > 1 {
				classes = append(classes, map[string]interface{}{
					"name": matches[1],
					"line": i + 1,
					"type": "class",
				})
			}
		}
	}

	return classes
}
